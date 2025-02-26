package requestindex

import (
	"context"
	"errors"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/libs/query"
)

// XXX/TODO: These types should be moved to the indexer package.

//go:generate ../../scripts/mockery_generate.sh DvsRequestIndexer

// TxIndexer interface defines methods to index and search transactions.
type DvsRequestIndexer interface {
	// AddBatch analyzes, indexes and stores a batch of transactions.
	AddBatch(b *Batch) error

	// Index analyzes, indexes and stores a single transaction.
	Index(result *avsi.DVSRequestResult) error

	// Get returns the transaction specified by hash or nil if the transaction is not indexed
	// or stored.
	Get(hash []byte) (*avsi.DVSRequestResult, error)

	//Set Logger
	SetLogger(l log.Logger)

	// Search allows you to query for transactions.
	Search(ctx context.Context, q *query.Query) ([]*avsi.DVSRequestResult, error)
}

// Batch groups together multiple Index operations to be performed at the same time.
// NOTE: Batch is NOT thread-safe and must not be modified after starting its execution.
type Batch struct {
	Ops []*avsi.DVSRequestResult
}

// NewBatch creates a new Batch.
func NewBatch(n int64) *Batch {
	return &Batch{
		Ops: make([]*avsi.DVSRequestResult, 0, n),
	}
}

// Add or update an entry for the given result.Index.
func (b *Batch) Add(result *avsi.DVSRequestResult) error {
	b.Ops = append(b.Ops, result)
	return nil
}

// Size returns the total number of operations inside the batch.
func (b *Batch) Size() int {
	return len(b.Ops)
}

// ErrorEmptyHash indicates empty hash
var ErrorEmptyHash = errors.New("transaction hash cannot be empty")
