package tools

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type OperatorID [32]byte

func GenOperatorIDByAddress(operatorAddress common.Address) OperatorID {
	operatorID := crypto.Keccak256Hash(operatorAddress.Bytes())
	return OperatorID(operatorID[:])
}
