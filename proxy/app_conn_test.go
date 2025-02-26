package proxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/avsi/example/kvstore"
	"github.com/0xPellNetwork/pelldvs/avsi/server"
	"github.com/0xPellNetwork/pelldvs/avsi/types"
	cmtrand "github.com/0xPellNetwork/pelldvs/libs/rand"
)

var SOCKET = "socket"

func TestKV(t *testing.T) {
	sockPath := fmt.Sprintf("unix:///tmp/echo_%v.sock", cmtrand.Str(6))
	clientCreator := NewRemoteClientCreator(sockPath, SOCKET, true)

	// Start server
	s := server.NewSocketServer(sockPath, kvstore.NewInMemoryApplication())
	s.SetLogger(log.TestingLogger().With("module", "avsi-server"))
	if err := s.Start(); err != nil {
		t.Fatalf("Error starting socket server: %v", err.Error())
	}
	t.Cleanup(func() {
		if err := s.Stop(); err != nil {
			t.Error(err)
		}
	})

	// Start client
	cli, err := clientCreator.NewAVSIClient()
	if err != nil {
		t.Fatalf("Error creating avsi client: %v", err.Error())
	}
	cli.SetLogger(log.TestingLogger().With("module", "avsi-client"))
	if err := cli.Start(); err != nil {
		t.Fatalf("Error starting avsi client: %v", err.Error())
	}

	taskProxy := NewAppConnDvs(cli, NopMetrics())

	t.Log("Connected")

	key := "key123"
	value := "value456"

	taskRes, err := taskProxy.ProcessDVSRequest(context.Background(), &types.RequestProcessDVSRequest{Request: &types.DVSRequest{Data: []byte(key + "=" + value)}})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, key, string(taskRes.Response))
	require.Equal(t, value, string(taskRes.ResponseDigest))

	queryProxy := NewAppConnQuery(cli, NopMetrics())
	queryRes, err := queryProxy.Query(context.Background(), &types.RequestQuery{
		Data: []byte(key),
	})

	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(queryRes.Key), string(queryRes.Value))
	require.Equal(t, key, string(queryRes.Key))
	require.Equal(t, value, string(queryRes.Value))
}
