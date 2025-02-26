package kvstore

import (
	"bytes"
	"context"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/avsi/types"
)

var (
	kvPairPrefixKey = []byte("kvPairKey:")
)

const (
	ValidatorPrefix        = "val="
	AppVersion      uint64 = 1
)

var _ types.Application = (*Application)(nil)

// Application is the kvstore state machine. It complies with the abci.Application interface.
// It takes transactions in the form of key=value and saves them in a database. This is
// a somewhat trivial example as there is no real state execution
type Application struct {
	types.BaseApplication

	state  State
	logger log.Logger
}

// NewApplication creates an instance of the kvstore from the provided database
func NewApplication(db dbm.DB) *Application {
	return &Application{
		logger: log.NewNopLogger(),
		state: State{
			db: db,
		},
	}
}

// NewPersistentApplication creates a new application using the goleveldb database engine
func NewPersistentApplication(dbDir string) *Application {
	name := "kvstore"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		panic(fmt.Errorf("failed to create persistent app at %s: %w", dbDir, err))
	}
	return NewApplication(db)
}

// NewInMemoryApplication creates a new application from an in memory database.
// Nothing will be persisted.
func NewInMemoryApplication() *Application {
	return NewApplication(dbm.NewMemDB())
}

// Info returns information about the state of the application. This is generally used everytime a Tendermint instance
// begins and let's the application know what Tendermint versions it's interacting with. Based from this information,
// Tendermint will ensure it is in sync with the application by potentially replaying the blocks it has. If the
// Application returns a 0 appBlockHeight, Tendermint will call InitChain to initialize the application with consensus related data
func (app *Application) Info(context.Context, *types.RequestInfo) (*types.ResponseInfo, error) {
	return &types.ResponseInfo{}, nil
}

func (app *Application) ProcessDVSRequest(_ context.Context, req *types.RequestProcessDVSRequest) (*types.ResponseProcessDVSRequest, error) {

	parts := bytes.Split(req.Request.Data, []byte("="))
	if len(parts) != 2 {
		panic(fmt.Sprintf("unexpected tx format. Expected 2 got %d: %s", len(parts), parts))
	}
	key, value := string(parts[0]), string(parts[1])
	err := app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	if err != nil {
		panic(err)
	}

	events := make([]types.Event, 0, 10)
	attrs := make([]types.EventAttribute, 0)
	attrs = append(attrs, types.EventAttribute{
		Key:   "First Event Key",
		Value: "First Event Value",
		Index: true,
	})

	events = append(events, types.Event{
		Type:       "FirstEventType",
		Attributes: attrs,
	})

	attrs = append(attrs, types.EventAttribute{
		Key:   "SecondEventKey",
		Value: "SecondEventValue",
		Index: true,
	})

	events = append(events, types.Event{
		Type:       "SecondEventType",
		Attributes: attrs,
	})

	attrs = append(attrs, types.EventAttribute{
		Key:   "ThirdEventKey",
		Value: "Third Event Value",
		Index: true,
	})

	events = append(events, types.Event{
		Type:       "ThirdEventType",
		Attributes: attrs,
	})

	return &types.ResponseProcessDVSRequest{
		Response:       []byte(key),
		ResponseDigest: []byte(value),
		Events:         events,
	}, nil
}

func (app *Application) ProcessDVSResponse(_ context.Context, req *types.RequestProcessDVSResponse) (*types.ResponseProcessDVSResponse, error) {

	events := make([]types.Event, 0, 10)
	attrs := make([]types.EventAttribute, 0)
	attrs = append(attrs, types.EventAttribute{
		Key:   "FourthEventKey",
		Value: "Fourth Event Value",
		Index: true,
	})

	events = append(events, types.Event{
		Type:       "FourthEventType",
		Attributes: attrs,
	})

	return &types.ResponseProcessDVSResponse{
		Events: events,
	}, nil
}

// Returns an associated value or nil if missing.
func (app *Application) Query(_ context.Context, reqQuery *types.RequestQuery) (*types.ResponseQuery, error) {

	v, err := app.state.db.Get(prefixKey(reqQuery.Data))
	if err != nil {
		return nil, err
	}

	resQuery := &types.ResponseQuery{
		Key:   reqQuery.Data,
		Value: v,
	}

	return resQuery, nil
}

func (app *Application) Close() error {
	return app.state.db.Close()
}

type State struct {
	db dbm.DB
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}
