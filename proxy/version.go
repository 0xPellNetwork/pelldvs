package proxy

import (
	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/version"
)

// RequestInfo contains all the information for sending
// the avsi.RequestInfo message during handshake with the app.
// It contains only compile-time version information.
var RequestInfo = &avsi.RequestInfo{
	Version:      version.TMCoreSemVer,
	BlockVersion: version.BlockProtocol,
	P2PVersion:   version.P2PProtocol,
	AbciVersion:  version.AVSIVersion,
}
