package types

import (
	"github.com/0xPellNetwork/pelldvs/crypto/tmhash"
	"github.com/cosmos/gogoproto/proto"
)

type DVSRequestHash []byte

func (d *DVSRequest) Hash() DVSRequestHash {
	raw, err := proto.Marshal(d)
	if err != nil {
		return nil
	}
	return tmhash.Sum(raw)
}
