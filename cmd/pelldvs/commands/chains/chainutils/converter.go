package chainutils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

func ConvStrToUint8(s string) (uint8, error) {
	i, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}
	// check max value
	if i > 255 {
		return 0, fmt.Errorf("value out of range: %v", i)
	}
	res, err := uint8(i), nil //nolint:gosec
	return res, err
}

func ConvStrToUint64(s string) (uint64, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func ConvStrToUint32(s string) (uint32, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil //nolint:gosec
}

func ConvStrToBigInt(s string) (*big.Int, error) {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse s to bigInt: %v", s)
	}
	return i, nil
}

func ConvUint64ToBigInt(i uint64) *big.Int {
	return big.NewInt(int64(i)) //nolint:gosec
}

func ConvStrToEthAddress(s string) (gethcommon.Address, error) {
	if !gethcommon.IsHexAddress(s) {
		return gethcommon.Address{}, fmt.Errorf("invalid address: %v", s)
	}
	return gethcommon.HexToAddress(s), nil
}

func ConvStrsToUint8List(s string) []uint8 {
	slist := strings.Split(s, ",")
	var res []uint8
	for _, s2 := range slist {
		for _, c := range s2 {
			if c >= '0' && c <= '9' {
				res = append(res, uint8(c-'0'))
			}
		}
	}

	return res
}
