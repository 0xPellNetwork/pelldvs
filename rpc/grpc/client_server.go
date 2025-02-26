package coregrpc

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cmtnet "github.com/0xPellNetwork/pelldvs/libs/net"
	"github.com/0xPellNetwork/pelldvs/rpc/core"
)

// Config is an gRPC server configuration.
type Config struct {
	MaxOpenConnections int
}

func StartGRPCServer(env *core.Environment, ln net.Listener) error {
	grpcServer := grpc.NewServer()
	RegisterDVSRequestAPIServer(grpcServer, &DVSRequestAPIServerAPI{env: env})
	return grpcServer.Serve(ln)
}

func StartGRPCClient(protoAddr string) DVSRequestAPIClient {

	conn, err := grpc.NewClient(protoAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerFunc))
	if err != nil {
		panic(err)
	}
	return NewDVSRequestAPIClient(conn)
}

func dialerFunc(_ context.Context, addr string) (net.Conn, error) {
	return cmtnet.Connect(addr)
}
