package kv

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
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
		rawBytesDvsRequest, err := proto.Marshal(result.DvsRequest)
		if err != nil {
			return err
		}
		hash := types.DvsRequest(rawBytesDvsRequest).Hash()

		if result.ResponseProcessDvsRequest != nil {
			// index DVS Request by events
			err = dvsReqIdx.indexEvents(result.ResponseProcessDvsRequest.Events, hash, storeBatch)
			if err != nil {
				return err
			}
		}

		if result.ResponseProcessDvsResponse != nil {
			// index DVS Response by events
			err = dvsReqIdx.indexEvents(result.ResponseProcessDvsResponse.Events, hash, storeBatch)
			if err != nil {
				return err
			}
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

	rawBytesDvsRequest, err := proto.Marshal(result.DvsRequest)
	if err != nil {
		return err
	}

	hash := types.DvsRequest(rawBytesDvsRequest).Hash()

	if result.ResponseProcessDvsRequest != nil {
		// index DVS Request by events
		err = dvsReqIdx.indexEvents(result.ResponseProcessDvsRequest.Events, hash, batch)
		if err != nil {
			return err
		}
	}

	if result.ResponseProcessDvsResponse != nil {
		// index DVS Response by events
		err = dvsReqIdx.indexEvents(result.ResponseProcessDvsResponse.Events, hash, batch)
		if err != nil {
			return err
		}
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

func (dvsReqIdx *DvsRequestIndex) indexEvents(events []avsi.Event, hash []byte, store dbm.Batch) error {

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
				err := store.Set(keyForEvent(compositeTag, attr.Value, dvsReqIdx.eventSeq), hash)
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
	}

	return nil
}

// func keyForEvent(key string, value string, result *avsi.ResponseProcessDVSRequest, eventSeq int64) []byte {
func keyForEvent(key string, value string, eventSeq *big.Int) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s",
		key,
		value,
		eventSeqSeparator+eventSeq.String(), //strconv.FormatInt(eventSeq, 10),
	))
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

	// var hashesInitialized bool
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

	// for all other conditions
	for _, c := range conditions {
		filteredHashes = dvsReqIdx.match(ctx, c, startKeyForCondition(c, 0), filteredHashes, true)
	}

	results := make([]*avsi.DVSRequestResult, 0, len(filteredHashes))
	resultMap := make(map[string]struct{})
RESULTS_LOOP:
	for _, h := range filteredHashes {

		res, err := dvsReqIdx.Get(h)
		if err != nil {
			return nil, fmt.Errorf("failed to get DVSRequest{%X}: %w", h, err)
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

func (*DvsRequestIndex) setTmpHashes(tmpHeights map[string][]byte, key, value []byte) {
	eventSeq := extractEventSeqFromKey(key)

	// Copy the value because the iterator will be reused.
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	tmpHeights[string(valueCopy)+eventSeq] = valueCopy
}

// match returns all matching DVS Requests by hash that meet a given condition and start
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
	value := key[indices[0]+1 : indices[len(indices)-2]]

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

func startKeyForCondition(c syntax.Condition, height int64) []byte {
	if height > 0 {
		return startKey(c.Tag, c.Arg.Value(), height)
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
