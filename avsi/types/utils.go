package types

import (
	"github.com/cosmos/gogoproto/proto"

	"github.com/0xPellNetwork/pelldvs/crypto/tmhash"
)

type DVSRequestHash []byte

func (d *DVSRequest) Hash() DVSRequestHash {
	raw, err := proto.Marshal(d)
	if err != nil {
		return nil
	}
	return tmhash.Sum(raw)
}
