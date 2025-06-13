package kv

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/libs/query"
	"github.com/0xPellNetwork/pelldvs/libs/query/syntax"
	requestindex "github.com/0xPellNetwork/pelldvs/state/requestindex"
	"github.com/0xPellNetwork/pelldvs/types"
)

const (
	tagKeySeparator     = "/"
	tagKeySeparatorRune = '/'
	eventSeqSeparator   = "$es$"
	eventSeqKey         = "EventSeqKey"
)

var _ requestindex.DvsRequestIndexer = (*DvsRequestIndex)(nil)

// DvsRequestIndex is the simplest possible requestindex, backed by key-value storage (levelDB).
type DvsRequestIndex struct {
	store dbm.DB
	// Number the events in the event list
	eventSeq *big.Int
	log      log.Logger
}

// NewDvsRequestIndex creates new KV requestindex.
func NewDvsRequestIndex(store dbm.DB) *DvsRequestIndex {
	return &DvsRequestIndex{
		store: store,
	}
}

func (dvsReqIdx *DvsRequestIndex) SetLogger(l log.Logger) {
	dvsReqIdx.log = l
}

func (dvsReqIdx *DvsRequestIndex) GetEventSeq() (*big.Int, error) {
	rawBytes, err := dvsReqIdx.store.Get([]byte(eventSeqKey))
	if err != nil {
		return nil, err
	}
	if rawBytes == nil {
		return big.NewInt(0), nil
	}

	return new(big.Int).SetBytes(rawBytes), nil
}

// Get gets transaction from the DvsRequestIndex storage and returns it or nil if the
// transaction is not found.
func (dvsReqIdx *DvsRequestIndex) Get(hash []byte) (*avsi.DVSRequestResult, error) {
	if len(hash) == 0 {
		return nil, requestindex.ErrorEmptyHash
	}

	rawBytes, err := dvsReqIdx.store.Get(hash)
	if err != nil {
		panic(err)
	}
	if rawBytes == nil {
		return nil, nil
	}

	requestResult := new(avsi.DVSRequestResult)
	err = proto.Unmarshal(rawBytes, requestResult)
	if err != nil {
		return nil, fmt.Errorf("error reading DVSRequestResult: %v", err)
	}

	return requestResult, nil
}

func (dvsReqIdx *DvsRequestIndex) AddBatch(batch *requestindex.Batch) error {
	storeBatch := dvsReqIdx.store.NewBatch()
	defer storeBatch.Close()

	for _, result := range batch.Ops {
		var err error
		hash := result.DvsRequest.Hash()

		if result.ResponseProcessDvsRequest != nil {
			// index DVS Request by events
			err = dvsReqIdx.indexEvents(result.DvsRequest.ChainId, result.DvsRequest.Height, result.ResponseProcessDvsRequest.Events, hash, storeBatch)
			if err != nil {
				return err
			}
		}

		if result.ResponseProcessDvsResponse != nil {
			// index DVS Response by events
			err = dvsReqIdx.indexEvents(result.DvsRequest.ChainId, result.DvsRequest.Height, result.ResponseProcessDvsResponse.Events, hash, storeBatch)
			if err != nil {
				return err
			}
		}

		// index by height (always)
		err = storeBatch.Set(keyForHeight(result, dvsReqIdx.eventSeq), hash)
		if err != nil {
			return err
		}

		rawBytes, err := proto.Marshal(result)
		if err != nil {
			return err
		}
		// index by hash (always)
		err = storeBatch.Set(hash, rawBytes)
		if err != nil {
			return err
		}
	}

	return storeBatch.WriteSync()
}

func (dvsReqIdx *DvsRequestIndex) Index(result *avsi.DVSRequestResult) error {
	batch := dvsReqIdx.store.NewBatch()
	defer batch.Close()

	var err error
	hash := result.DvsRequest.Hash()

	if result.ResponseProcessDvsRequest != nil {
		// index DVS Request by events
		err = dvsReqIdx.indexEvents(result.DvsRequest.ChainId, result.DvsRequest.Height, result.ResponseProcessDvsRequest.Events, hash, batch)
		if err != nil {
			return err
		}
	}

	if result.ResponseProcessDvsResponse != nil {
		// index DVS Response by events
		err = dvsReqIdx.indexEvents(result.DvsRequest.ChainId, result.DvsRequest.Height, result.ResponseProcessDvsResponse.Events, hash, batch)
		if err != nil {
			return err
		}
	}

	// index by height (always)
	err = batch.Set(keyForHeight(result, dvsReqIdx.eventSeq), hash)
	if err != nil {
		return err
	}

	rawBytes, err := proto.Marshal(result)
	if err != nil {
		return err
	}
	// index by hash (always)
	err = batch.Set(hash, rawBytes)
	if err != nil {
		return err
	}

	return batch.WriteSync()
}

func (dvsReqIdx *DvsRequestIndex) indexEvents(chainid, height int64, events []avsi.Event, hash []byte, store dbm.Batch) error {

	if dvsReqIdx.eventSeq == nil {
		seq, err := dvsReqIdx.GetEventSeq()
		if err != nil {
			return err
		}
		dvsReqIdx.eventSeq = seq
	}

	preEventSeq := *dvsReqIdx.eventSeq
	for _, event := range events {
		dvsReqIdx.eventSeq = dvsReqIdx.eventSeq.Add(dvsReqIdx.eventSeq, big.NewInt(1))
		// only index events with a non-empty type
		if len(event.Type) == 0 {
			continue
		}

		for _, attr := range event.Attributes {
			if len(attr.Key) == 0 {
				continue
			}

			// index if `index: true` is set
			compositeTag := fmt.Sprintf("%s.%s", event.Type, attr.Key)
			// ensure event does not conflict with a reserved prefix key
			if compositeTag == types.DVSHashKey {
				return fmt.Errorf("event type and attribute key \"%s\" is reserved; please use a different key", compositeTag)
			}
			if attr.GetIndex() {
				// TODO: The result.Height and result.Index are not used here, as they may lead to duplicates.
				// store.Set(keyForEvent(compositeTag, attr.Value, result, dvsReqIdx.eventSeq), hash)
				// result.Height,
				// result.Index,

				dvsReqIdx.log.Info("event key", "keyForEvent(compositeTag, attr.Value, dvsReqIdx.eventSeq)",
					string(keyForEvent(compositeTag, attr.Value, chainid, height, dvsReqIdx.eventSeq)))
				err := store.Set(keyForEvent(compositeTag, attr.Value, chainid, height, dvsReqIdx.eventSeq), hash)
				if err != nil {
					return err
				}
			}
		}
	}

	if preEventSeq.Cmp(dvsReqIdx.eventSeq) != 0 {
		err := store.Set([]byte(eventSeqKey), dvsReqIdx.eventSeq.Bytes())
		if err != nil {
			return err
		}
	} else {
		dvsReqIdx.eventSeq = dvsReqIdx.eventSeq.Add(dvsReqIdx.eventSeq, big.NewInt(1))
		err := store.Set([]byte(eventSeqKey), dvsReqIdx.eventSeq.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

// Search performs a search using the given query.
// Search will exit early and return any result fetched so far,
// when a message is received on the context chan.
func (dvsReqIdx *DvsRequestIndex) Search(ctx context.Context, q *query.Query) ([]*avsi.DVSRequestResult, error) {
	select {
	case <-ctx.Done():
		return make([]*avsi.DVSRequestResult, 0), nil

	default:
	}

	var hashesInitialized bool
	filteredHashes := make(map[string][]byte)

	conditions := q.Syntax()

	// if there is a hash condition, return the result immediately
	hash, ok, err := lookForHash(conditions)
	if err != nil {
		return nil, fmt.Errorf("error during searching for a hash in the query: %w", err)
	} else if ok {
		res, err := dvsReqIdx.Get(hash)
		switch {
		case err != nil:
			return []*avsi.DVSRequestResult{}, fmt.Errorf("error while retrieving the result: %w", err)
		case res == nil:
			return []*avsi.DVSRequestResult{}, nil
		default:
			return []*avsi.DVSRequestResult{res}, nil
		}
	}

	// conditions to skip because they're handled before "everything else"
	skipIndexes := make([]int, 0)
	var heightInfo HeightInfo

	// If we are not matching events and tx.height = 3 occurs more than once, the later value will
	// overwrite the first one.
	conditions, heightInfo = dedupHeight(conditions)

	if !heightInfo.onlyHeightEq {
		skipIndexes = append(skipIndexes, heightInfo.heightEqIdx)
	}
	// extract ranges
	// if both upper and lower bounds exist, it's better to get them in order not
	// no iterate over kvs that are not within range.
	ranges, rangeIndexes, heightRange := LookForRangesWithHeight(conditions)
	heightInfo.heightRange = heightRange

	dvsReqIdx.log.Info("heightInfo",
		"heightInfo.height", heightInfo.height,
		"heightInfo.heightEqIdx", heightInfo.heightEqIdx,
		"heightInfo.onlyHeightEq", heightInfo.onlyHeightEq,
		"heightInfo.onlyHeightRange", heightInfo.onlyHeightRange,
		"len(ranges)", len(ranges))

	if len(ranges) > 0 {
		skipIndexes = append(skipIndexes, rangeIndexes...)

		for _, qr := range ranges {

			// If we have a query range over height and want to still look for
			// specific event values we do not want to simply return all
			// transactios in this height range. We remember the height range info
			// and pass it on to match() to take into account when processing events.
			//if qr.Key == types.TxHeightKey && !heightInfo.onlyHeightRange {
			if qr.Key == types.DVSHeightKey && !heightInfo.onlyHeightRange {
				continue
			}
			if !hashesInitialized {

				filteredHashes = dvsReqIdx.matchRange(ctx, qr, startKey(qr.Key), filteredHashes, true, heightInfo)
				hashesInitialized = true

				// Ignore any remaining conditions if the first condition resulted
				// in no matches (assuming implicit AND operand).
				if len(filteredHashes) == 0 {
					break
				}
			} else {
				filteredHashes = dvsReqIdx.matchRange(ctx, qr, startKey(qr.Key), filteredHashes, false, heightInfo)
			}
		}
	}

	// if there is a height condition ("tx.height=3"), extract it

	conditions, chainID := lookForChainIDAndRemove(conditions)

	// for all other conditions
	for i, c := range conditions {
		if intInSlice(i, skipIndexes) {
			continue
		}

		dvsReqIdx.log.Info("match",
			"heightInfo.height", heightInfo.height,
			"startKeyForCondition(c, heightInfo.height)", string(startKeyForCondition(c, chainID, heightInfo.height)),
			"c.Op", c.Op)

		if !hashesInitialized {
			filteredHashes = dvsReqIdx.match(ctx, c, startKeyForCondition(c, chainID, heightInfo.height), filteredHashes, true, heightInfo)
			hashesInitialized = true
			// Ignore any remaining conditions if the first condition resulted
			// in no matches (assuming implicit AND operand).
			if len(filteredHashes) == 0 {
				break
			}
		} else {
			filteredHashes = dvsReqIdx.match(ctx, c, startKeyForCondition(c, chainID, heightInfo.height), filteredHashes, false, heightInfo)
		}
	}

	results := make([]*avsi.DVSRequestResult, 0, len(filteredHashes))
	resultMap := make(map[string]struct{})
RESULTS_LOOP:
	for _, h := range filteredHashes {

		res, err := dvsReqIdx.Get(h)
		if err != nil {
			return nil, fmt.Errorf("failed to get Tx{%X}: %w", h, err)
		}
		hashString := string(h)
		if _, ok := resultMap[hashString]; !ok {
			resultMap[hashString] = struct{}{}
			results = append(results, res)
		}
		// Potentially exit early.
		select {
		case <-ctx.Done():
			break RESULTS_LOOP
		default:
		}
	}

	return results, nil
}

func lookForHash(conditions []syntax.Condition) (hash []byte, ok bool, err error) {
	for _, c := range conditions {
		if c.Tag == types.DVSHashKey {
			decoded, err := hex.DecodeString(c.Arg.Value())
			return decoded, true, err
		}
	}
	return
}

func lookForChainIDAndRemove(conditions []syntax.Condition) ([]syntax.Condition, int64) {

	chainID := int64(-1)
	for _, c := range conditions {
		if c.Tag == types.DVSChainID {
			res, _ := c.Arg.Number().Int64()
			chainID = res
		}
	}

	return removeChainIDElements(conditions), chainID
}

func removeChainIDElements(conditions []syntax.Condition) []syntax.Condition {
	result := make([]syntax.Condition, 0, len(conditions))
	for _, c := range conditions {
		if c.Tag != types.DVSChainID {
			result = append(result, c)
		}
	}
	return result
}

func (*DvsRequestIndex) setTmpHashes(tmpHeights map[string][]byte, key, value []byte) {
	eventSeq := extractEventSeqFromKey(key)

	// Copy the value because the iterator will be reused.
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	_ = eventSeq
	//tmpHeights[string(valueCopy)+eventSeq] = valueCopy
	tmpHeights[string(valueCopy)] = valueCopy
}

// matchRange returns all matching txs by hash that meet a given queryRange and
// start key. An already filtered result (filteredHashes) is provided such that
// any non-intersecting matches are removed.
//
// NOTE: filteredHashes may be empty if no previous condition has matched.
func (dvsReqIdx *DvsRequestIndex) matchRange(
	ctx context.Context,
	qr QueryRange,
	startKey []byte,
	filteredHashes map[string][]byte,
	firstRun bool,
	heightInfo HeightInfo,
) map[string][]byte {
	// A previous match was attempted but resulted in no matches, so we return
	// no matches (assuming AND operand).
	if !firstRun && len(filteredHashes) == 0 {
		return filteredHashes
	}

	tmpHashes := make(map[string][]byte)

	it, err := dbm.IteratePrefix(dvsReqIdx.store, startKey)
	if err != nil {
		panic(err)
	}
	defer it.Close()
	bigIntValue := new(big.Int)

LOOP:
	for ; it.Valid(); it.Next() {
		// TODO: We need to make a function for getting it.Key() as a byte slice with no copies.
		// It currently copies the source data (which can change on a subsequent .Next() call) but that
		// is not an issue for us.
		key := it.Key()

		if !isTagKey(key) {
			continue
		}

		if _, ok := qr.AnyBound().(*big.Float); ok {

			value := extractValueFromKey(key)
			v, ok := bigIntValue.SetString(value, 10)

			var vF *big.Float
			if !ok {
				vF, _, err = big.ParseFloat(value, 10, 125, big.ToNearestEven)
				if err != nil {
					continue LOOP
				}

			}
			if qr.Key != types.DVSHeightKey {
				keyHeight, err := extractHeightFromKey(key)

				if err != nil {
					dvsReqIdx.log.Error("failure to parse height from key:", err)
					continue
				}

				withinBounds, err := checkHeightConditions(heightInfo, keyHeight)
				if err != nil {
					dvsReqIdx.log.Error("failure checking for height bounds:", err)
					continue
				}
				if !withinBounds {
					continue
				}
			}
			var withinBounds bool
			var err error
			if !ok {
				withinBounds, err = checkBounds(qr, vF)
			} else {
				withinBounds, err = checkBounds(qr, v)
			}
			if err != nil {
				dvsReqIdx.log.Error("failed to parse bounds:", err)
			} else if withinBounds {
				dvsReqIdx.setTmpHashes(tmpHashes, key, it.Value())
			}

			// XXX: passing time in a ABCI Events is not yet implemented
			// case time.Time:
			// 	v := strconv.ParseInt(extractValueFromKey(it.Key()), 10, 64)
			// 	if v == r.upperBound {
			// 		break
			// 	}
		}

		// Potentially exit early.
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}
	}
	if err := it.Error(); err != nil {
		panic(err)
	}

	if len(tmpHashes) == 0 || firstRun {
		// Either:
		//
		// 1. Regardless if a previous match was attempted, which may have had
		// results, but no match was found for the current condition, then we
		// return no matches (assuming AND operand).
		//
		// 2. A previous match was not attempted, so we return all results.
		return tmpHashes
	}

	// Remove/reduce matches in filteredHashes that were not found in this
	// match (tmpHashes).
REMOVE_LOOP:
	for k, v := range filteredHashes {
		tmpHash := tmpHashes[k]
		if tmpHash == nil || !bytes.Equal(tmpHashes[k], v) {
			delete(filteredHashes, k)

			// Potentially exit early.
			select {
			case <-ctx.Done():
				break REMOVE_LOOP
			default:
			}
		}
	}

	return filteredHashes
}

// match returns all matching txs by hash that meet a given condition and start
// key. An already filtered result (filteredHashes) is provided such that any
// non-intersecting matches are removed.
//
// NOTE: filteredHashes may be empty if no previous condition has matched.
func (dvsReqIdx *DvsRequestIndex) match(
	ctx context.Context,
	c syntax.Condition,
	startKeyBz []byte,
	filteredHashes map[string][]byte,
	firstRun bool,
	heightInfo HeightInfo,
) map[string][]byte {

	// A previous match was attempted but resulted in no matches, so we return
	// no matches (assuming AND operand).
	if !firstRun && len(filteredHashes) == 0 {
		return filteredHashes
	}

	tmpHashes := make(map[string][]byte)

	switch {
	case c.Op == syntax.TEq:

		it, err := dbm.IteratePrefix(dvsReqIdx.store, startKeyBz)
		if err != nil {
			panic(err)
		}
		defer it.Close()

	EQ_LOOP:
		for ; it.Valid(); it.Next() {
			// If we have a height range in a query, we need only transactions
			// for this height
			key := it.Key()

			keyHeight, err := extractHeightFromKey(key)
			if err != nil {
				dvsReqIdx.log.Error("failure to parse height from key:", err)
				continue
			}

			withinBounds, err := checkHeightConditions(heightInfo, keyHeight)
			if err != nil {
				dvsReqIdx.log.Error("failure checking for height bounds:", err)
				continue
			}

			if !withinBounds {
				continue
			}

			dvsReqIdx.setTmpHashes(tmpHashes, key, it.Value())

			// Potentially exit early.
			select {
			case <-ctx.Done():
				break EQ_LOOP
			default:
			}
		}
		if err := it.Error(); err != nil {
			panic(err)
		}

	case c.Op == syntax.TExists:
		// XXX: can't use startKeyBz here because c.Operand is nil
		// (e.g. "account.owner/<nil>/" won't match w/ a single row)
		it, err := dbm.IteratePrefix(dvsReqIdx.store, startKey(c.Tag))
		if err != nil {
			panic(err)
		}
		defer it.Close()

	EXISTS_LOOP:
		for ; it.Valid(); it.Next() {
			key := it.Key()
			keyHeight, err := extractHeightFromKey(key)
			if err != nil {
				dvsReqIdx.log.Error("failure to parse height from key:", err)
				continue
			}
			withinBounds, err := checkHeightConditions(heightInfo, keyHeight)
			if err != nil {
				dvsReqIdx.log.Error("failure checking for height bounds:", err)
				continue
			}
			if !withinBounds {
				continue
			}
			dvsReqIdx.setTmpHashes(tmpHashes, key, it.Value())

			// Potentially exit early.
			select {
			case <-ctx.Done():
				break EXISTS_LOOP
			default:
			}
		}
		if err := it.Error(); err != nil {
			panic(err)
		}

	case c.Op == syntax.TContains:

		// XXX: startKey does not apply here.
		// For example, if startKey = "account.owner/an/" and search query = "account.owner CONTAINS an"
		// we can't iterate with prefix "account.owner/an/" because we might miss keys like "account.owner/Ulan/"
		it, err := dbm.IteratePrefix(dvsReqIdx.store, startKey(c.Tag))
		if err != nil {
			panic(err)
		}
		defer it.Close()

	CONTAINS_LOOP:
		for ; it.Valid(); it.Next() {
			if !isTagKey(it.Key()) {
				continue
			}

			if strings.Contains(extractValueFromKey(it.Key()), c.Arg.Value()) {
				key := it.Key()
				keyHeight, err := extractHeightFromKey(key)
				if err != nil {
					dvsReqIdx.log.Error("failure to parse height from key:", err)
					continue
				}
				withinBounds, err := checkHeightConditions(heightInfo, keyHeight)
				if err != nil {
					dvsReqIdx.log.Error("failure checking for height bounds:", err)
					continue
				}
				if !withinBounds {
					continue
				}
				dvsReqIdx.setTmpHashes(tmpHashes, key, it.Value())
			}

			// Potentially exit early.
			select {
			case <-ctx.Done():
				break CONTAINS_LOOP
			default:
			}
		}
		if err := it.Error(); err != nil {
			panic(err)
		}
	default:
		panic("other operators should be handled already")
	}

	if len(tmpHashes) == 0 || firstRun {
		// Either:
		//
		// 1. Regardless if a previous match was attempted, which may have had
		// results, but no match was found for the current condition, then we
		// return no matches (assuming AND operand).
		//
		// 2. A previous match was not attempted, so we return all results.
		return tmpHashes
	}

	// Remove/reduce matches in filteredHashes that were not found in this
	// match (tmpHashes).
REMOVE_LOOP:
	for k, v := range filteredHashes {
		tmpHash := tmpHashes[k]
		if tmpHash == nil || !bytes.Equal(tmpHash, v) {
			delete(filteredHashes, k)

			// Potentially exit early.
			select {
			case <-ctx.Done():
				break REMOVE_LOOP
			default:
			}
		}
	}

	return filteredHashes
}

// Keys
func isTagKey(key []byte) bool {
	// Normally, if the event was indexed with an event sequence, the number of
	// tags should 4. Alternatively it should be 3 if the event was not indexed
	// with the corresponding event sequence. However, some attribute values in
	// production can contain the tag separator. Therefore, the condition is >= 3.
	numTags := 0
	for i := 0; i < len(key); i++ {
		if key[i] == tagKeySeparatorRune {
			numTags++
			if numTags >= 3 {
				return true
			}
		}
	}
	return false
}

func extractHeightFromKey(key []byte) (int64, error) {
	// the height is the second last element in the key.
	// Find the position of the last occurrence of tagKeySeparator
	endPos := bytes.LastIndexByte(key, tagKeySeparatorRune)
	if endPos == -1 {
		return 0, errors.New("separator not found")
	}

	// Find the position of the second last occurrence of tagKeySeparator
	startPos := bytes.LastIndexByte(key[:endPos-1], tagKeySeparatorRune)
	if startPos == -1 {
		return 0, errors.New("second last separator not found")
	}

	// Extract the height part of the key
	height, err := strconv.ParseInt(string(key[startPos+1:endPos]), 10, 64)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func extractValueFromKey(key []byte) string {
	// Find the positions of tagKeySeparator in the byte slice
	var indices []int
	for i, b := range key {
		if b == tagKeySeparatorRune {
			indices = append(indices, i)
		}
	}

	// If there are less than 2 occurrences of tagKeySeparator, return an empty string
	if len(indices) < 2 {
		return ""
	}

	// Extract the value between the first and second last occurrence of tagKeySeparator
	value := key[indices[0]+1 : indices[len(indices)-3]]

	// Trim any leading or trailing whitespace
	value = bytes.TrimSpace(value)

	// TODO: Do an unsafe cast to avoid an extra allocation here
	return string(value)
}

func extractEventSeqFromKey(key []byte) string {
	parts := strings.SplitN(string(key), tagKeySeparator, -1)

	lastEl := parts[len(parts)-1]

	if strings.Contains(lastEl, eventSeqSeparator) {
		return strings.SplitN(lastEl, eventSeqSeparator, 2)[1]
	}
	return "0"
}

func startKeyForCondition(c syntax.Condition, chainID, height int64) []byte {

	if chainID > 0 {
		if height > 0 {
			return startKey(c.Tag, c.Arg.Value(), chainID, height)
		} else {
			return startKey(c.Tag, c.Arg.Value(), chainID)
		}
	}

	return startKey(c.Tag, c.Arg.Value())
}

func startKey(fields ...interface{}) []byte {
	var b bytes.Buffer
	for _, f := range fields {
		b.Write([]byte(fmt.Sprintf("%v", f) + tagKeySeparator))
	}
	return b.Bytes()
}

func keyForEvent(key string, value string, chainid, height int64, eventSeq *big.Int) []byte {
	return []byte(fmt.Sprintf("%s/%s/%d/%d/%s",
		key,
		value,
		chainid,
		height,
		eventSeqSeparator+eventSeq.String(), //strconv.FormatInt(eventSeq, 10),
	))
}

func keyForHeight(result *avsi.DVSRequestResult, idx *big.Int) []byte {
	return []byte(fmt.Sprintf("%s/%d/%d/%d/%s%s",
		types.DVSHeightKey,
		result.DvsRequest.Height,
		result.DvsRequest.ChainId,
		result.DvsRequest.Height,
		idx.String(),
		// Added to facilitate having the eventSeq in event keys
		// Otherwise queries break expecting 5 entries
		eventSeqSeparator+"0",
	))
}
