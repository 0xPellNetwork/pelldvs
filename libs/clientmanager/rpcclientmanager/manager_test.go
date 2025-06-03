package rpcclientmanager

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/0xPellNetwork/pelldvs-libs/log"
)

type TestRPCServer struct {
	rpcAddress string
	isHealthy  bool
	server     *rpc.Server
	listener   net.Listener
}

func (trs *TestRPCServer) setIsHealthy() {
	trs.isHealthy = true
}

func (trs *TestRPCServer) setUnHealthy() {
	trs.isHealthy = false
}

func (trs *TestRPCServer) HealthCheck(_ struct{}, reply *bool) error {
	*reply = trs.isHealthy
	return nil
}

func (trs *TestRPCServer) Now(_ struct{}, reply *int) error {
	*reply = time.Now().Nanosecond()
	return nil
}

func newTestRPCServer() (*TestRPCServer, error) {
	// Register the TestRPCServer
	testServer := &TestRPCServer{
		rpcAddress: "localhost:12366",
		isHealthy:  true,
		server:     rpc.NewServer(),
	}
	if err := testServer.server.Register(testServer); err != nil {
		panic(err)
	}
	listener, err := net.Listen("tcp", testServer.rpcAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on address %s: %v", testServer.rpcAddress, err)
	}

	testServer.listener = listener
	// Start the RPC server in a goroutine
	go testServer.server.Accept(testServer.listener)

	return testServer, err
}

func TestGetClient(t *testing.T) {
	var logger = log.NewLogger(os.Stdout)
	server, err := newTestRPCServer()
	if err != nil {
		t.Fatalf("failed to create test RPC server: %v", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			t.Fatalf("failed to close listener: %v", err)
		}
	}(server.listener)

	clientManager, err := NewRPCClientManager(server.rpcAddress, "TestRPCServer.HealthCheck", logger)
	if err != nil {
		t.Fatalf("failed to create RPC client manager: %v", err)
	}

	client, err := clientManager.GetClient()
	if err != nil {
		t.Fatalf("failed to get RPC client: %v", err)
	}
	var now int
	err = client.Call("TestRPCServer.Now", struct{}{}, &now)
	t.Log("now1", now)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, now > 0)

	// Test health check
	var isHealthy bool
	err = client.Call("TestRPCServer.HealthCheck", struct{}{}, &isHealthy)
	assert.Equal(t, err, nil)
	assert.Equal(t, isHealthy, true)

	server.setUnHealthy()
	err = client.Call("TestRPCServer.HealthCheck", struct{}{}, &isHealthy)
	assert.Equal(t, err, nil)
	assert.Equal(t, isHealthy, false)

	// test reconnection
	// recreate testserver

	client.Close()
	server.setIsHealthy()

	// set now to zero
	now = 0
	// this should fail because the server is restart
	err = client.Call("TestRPCServer.Now", struct{}{}, &now)
	t.Log("err===", err)
	assert.Equal(t, err, rpc.ErrShutdown)

	time.Sleep(2 * time.Second) // wait for manager to reconnect

	// test reconnection
	client, err = clientManager.GetClient()
	if err != nil {
		t.Fatalf("failed to get RPC client after server restart: %v", err)
	}
	err = client.Call("TestRPCServer.Now", struct{}{}, &now)
	if err != nil {
		if errors.Is(err, rpc.ErrShutdown) {
			client, err = clientManager.GetClient()
			if err != nil {
				t.Fatalf("failed to get RPC client after server restart: %v", err)
			}
		}
	}
	err = client.Call("TestRPCServer.Now", struct{}{}, &now)
	t.Log("err======", err)
	t.Log("now2", now)
	assert.Equal(t, err, nil)
	assert.Equal(t, true, now > 0)
}
