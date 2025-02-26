package privval

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/crypto/ed25519"
	cmtnet "github.com/0xPellNetwork/pelldvs/libs/net"
)

var (
	testTimeoutAccept = defaultTimeoutAcceptSeconds * time.Second

	testTimeoutReadWrite    = 100 * time.Millisecond
	testTimeoutReadWrite2o3 = 60 * time.Millisecond // 2/3 of the other one
)

type dialerTestCase struct {
	addr   string
	dialer SocketDialer
}

func newSignerListenerEndpoint(logger log.Logger, addr string, timeoutReadWrite time.Duration) *SignerListenerEndpoint {
	proto, address := cmtnet.ProtocolAndAddress(addr)

	ln, err := net.Listen(proto, address)
	logger.Info("SignerListener: Listening", "proto", proto, "address", address)
	if err != nil {
		panic(err)
	}

	var listener net.Listener

	if proto == "unix" {
		unixLn := NewUnixListener(ln)
		UnixListenerTimeoutAccept(testTimeoutAccept)(unixLn)
		UnixListenerTimeoutReadWrite(timeoutReadWrite)(unixLn)
		listener = unixLn
	} else {
		tcpLn := NewTCPListener(ln, ed25519.GenPrivKey())
		TCPListenerTimeoutAccept(testTimeoutAccept)(tcpLn)
		TCPListenerTimeoutReadWrite(timeoutReadWrite)(tcpLn)
		listener = tcpLn
	}

	return NewSignerListenerEndpoint(
		logger,
		listener,
		SignerListenerEndpointTimeoutReadWrite(testTimeoutReadWrite),
	)
}

func startListenerEndpointAsync(t *testing.T, sle *SignerListenerEndpoint, endpointIsOpenCh chan struct{}) {
	go func(sle *SignerListenerEndpoint) {
		require.NoError(t, sle.Start())
		assert.True(t, sle.IsRunning())
		close(endpointIsOpenCh)
	}(sle)
}

func getMockEndpoints(
	t *testing.T,
	addr string,
	socketDialer SocketDialer,
) (*SignerListenerEndpoint, *SignerDialerEndpoint) {

	var (
		logger           = log.TestingLogger()
		endpointIsOpenCh = make(chan struct{})

		dialerEndpoint = NewSignerDialerEndpoint(
			logger,
			socketDialer,
		)

		listenerEndpoint = newSignerListenerEndpoint(logger, addr, testTimeoutReadWrite)
	)

	SignerDialerEndpointTimeoutReadWrite(testTimeoutReadWrite)(dialerEndpoint)
	SignerDialerEndpointConnRetries(1e6)(dialerEndpoint)

	startListenerEndpointAsync(t, listenerEndpoint, endpointIsOpenCh)

	require.NoError(t, dialerEndpoint.Start())
	assert.True(t, dialerEndpoint.IsRunning())

	<-endpointIsOpenCh

	return listenerEndpoint, dialerEndpoint
}

func TestSignerListenerEndpointServiceLoop(t *testing.T) {
	listenerEndpoint := NewSignerListenerEndpoint(
		log.TestingLogger(),
		&testListener{initialErrs: 5},
	)

	require.NoError(t, listenerEndpoint.Start())
	require.NoError(t, listenerEndpoint.WaitForConnection(time.Second))
}

type testListener struct {
	net.Listener
	initialErrs int
}

func (l *testListener) Accept() (net.Conn, error) {
	if l.initialErrs > 0 {
		l.initialErrs--

		return nil, errors.New("accept error")
	}

	return nil, nil // Note this doesn't actually return a valid connection, it just doesn't error.
}
