package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	cmtjson "github.com/0xPellNetwork/pelldvs/libs/json"
	"github.com/0xPellNetwork/pelldvs/privval"
)

// GenValidatorCmd allows the generation of a keypair for a
// validator.
var GenValidatorCmd = &cobra.Command{
	Use:   "gen-validator",
	Short: "Generate new validator keypair",
	Run:   genValidator,
}

func genValidator(*cobra.Command, []string) {
	pv := privval.GenFilePV("")
	jsbz, err := cmtjson.Marshal(pv)
	if err != nil {
		panic(err)
	}
	fmt.Printf(`%v
`, string(jsbz))
}
