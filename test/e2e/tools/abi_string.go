package tools

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetResponseDigestFromString(value string) ([32]byte, error) {
	arguments := abi.Arguments{
		{
			Type: abi.Type{
				T: abi.StringTy,
			},
		},
	}

	encoded, err := arguments.Pack(value)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to encode string: %v", err)
	}

	return crypto.Keccak256Hash(encoded), nil
}
