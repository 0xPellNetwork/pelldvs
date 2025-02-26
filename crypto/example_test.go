package crypto_test

import (
	"fmt"

	"github.com/0xPellNetwork/pelldvs/crypto"
)

func ExampleSha256() {
	sum := crypto.Sha256([]byte("This is PellDVS"))
	fmt.Printf("%x\n", sum)
	// Output:
	// 8df447f279c3714896cd6c1ae70cbbbb46b3530bde51db86e92c0850c6daa463
}
