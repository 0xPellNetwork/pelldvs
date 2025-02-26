package null

import (
	"context"
	"errors"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/libs/query"
	requestindex "github.com/0xPellNetwork/pelldvs/state/requestindex"
)

var _ requestindex.DvsRequestIndexer = (*DvsRequestIndex)(nil)

// DvsRequestIndex acts as a /dev/null.
type DvsRequestIndex struct{}

// Get on a DvsRequestIndex is disabled and panics when invoked.
func (txi *DvsRequestIndex) Get(_ []byte) (*avsi.DVSRequestResult, error) {
	return nil, errors.New(`indexing is disabled (set 'tx_index = "kv"' in config)`)
}

// AddBatch is a noop and always returns nil.
func (txi *DvsRequestIndex) AddBatch(_ *requestindex.Batch) error {
	return nil
}

// Index is a noop and always returns nil.
func (txi *DvsRequestIndex) Index(_ *avsi.DVSRequestResult) error {
	return nil
}

func (txi *DvsRequestIndex) SetLogger(log.Logger) {

}

func (txi *DvsRequestIndex) Search(_ context.Context, _ *query.Query) ([]*avsi.DVSRequestResult, error) {
	return []*avsi.DVSRequestResult{}, nil
}
