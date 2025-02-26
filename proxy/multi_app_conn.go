package proxy

import (
	"fmt"

	cmtlog "github.com/0xPellNetwork/pelldvs-libs/log"
	avsicli "github.com/0xPellNetwork/pelldvs/avsi/client"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
	"github.com/0xPellNetwork/pelldvs/libs/service"
)

const (
	connDvs   = "dvs"
	connQuery = "query"
)

// AppConns is the PellDVS's interface to the application that consists of
// multiple connections.
type AppConns interface {
	service.Service
	// Dvs connection
	Dvs() AppConnDvs
	// Query connection
	Query() AppConnQuery
}

// NewAppConns calls NewMultiAppConn.
func NewAppConns(clientCreator ClientCreator, metrics *Metrics) AppConns {
	return NewMultiAppConn(clientCreator, metrics)
}

// multiAppConn implements AppConns.
//
// A multiAppConn is made of a few appConns and manages their underlying avsi
// clients.
// TODO: on app restart, clients must reboot together
type multiAppConn struct {
	service.BaseService

	metrics *Metrics

	dvsConn   AppConnDvs
	queryConn AppConnQuery

	dvsConnClient   avsicli.Client
	queryConnClient avsicli.Client

	clientCreator ClientCreator
}

// NewMultiAppConn makes all necessary avsi connections to the application.
func NewMultiAppConn(clientCreator ClientCreator, metrics *Metrics) AppConns {
	multiAppConn := &multiAppConn{
		metrics:       metrics,
		clientCreator: clientCreator,
	}
	multiAppConn.BaseService = *service.NewBaseService(nil, "multiAppConn", multiAppConn)
	return multiAppConn
}

func (app *multiAppConn) Dvs() AppConnDvs {
	return app.dvsConn
}

func (app *multiAppConn) Query() AppConnQuery {
	return app.queryConn
}

func (app *multiAppConn) OnStart() error {

	c, err := app.avsiClientFor(connDvs)
	if err != nil {
		app.stopAllClients()
		return err
	}

	app.dvsConnClient = c
	app.dvsConn = NewAppConnDvs(c, app.metrics)

	c, err = app.avsiClientFor(connQuery)
	if err != nil {
		return err
	}
	app.queryConnClient = c
	app.queryConn = NewAppConnQuery(c, app.metrics)

	// Kill PellDVS if the avsi application crashes.
	go app.killTMOnClientError()

	return nil
}

func (app *multiAppConn) OnStop() {
	app.stopAllClients()
}

func (app *multiAppConn) killTMOnClientError() {
	killFn := func(conn string, err error, logger cmtlog.Logger) {
		logger.Error(
			fmt.Sprintf("%s connection terminated. Did the application crash? Please restart PellDVS", conn),
			"err", err)
		killErr := cmtos.Kill()
		if killErr != nil {
			logger.Error("Failed to kill this process - please do so manually", "err", killErr)
		}
	}

	select {
	case <-app.dvsConnClient.Quit():
		if err := app.dvsConnClient.Error(); err != nil {
			killFn(connDvs, err, app.Logger)
		}

	case <-app.queryConnClient.Quit():
		if err := app.queryConnClient.Error(); err != nil {
			killFn(connQuery, err, app.Logger)
		}
	}
}

func (app *multiAppConn) stopAllClients() {
	if app.dvsConnClient != nil {
		if err := app.dvsConnClient.Stop(); err != nil {
			app.Logger.Error("error while stopping consensus client", "error", err)
		}
	}

	if app.queryConnClient != nil {
		if err := app.queryConnClient.Stop(); err != nil {
			app.Logger.Error("error while stopping query client", "error", err)
		}
	}
}

func (app *multiAppConn) avsiClientFor(conn string) (avsicli.Client, error) {
	c, err := app.clientCreator.NewAVSIClient()
	if err != nil {
		return nil, fmt.Errorf("error creating AVSI client (%s connection): %w", conn, err)
	}
	c.SetLogger(app.Logger.With("module", "avsi-client", "connection", conn))
	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("error starting AVSI client (%s connection): %w", conn, err)
	}
	return c, nil
}
