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
	RunE:  genValidator,
}

func genValidator(*cobra.Command, []string) error {
	pv, err := privval.GenFilePV("")
	if err != nil {
		return fmt.Errorf("cannot generate file pv: %w", err)
	}
	jsbz, err := cmtjson.Marshal(pv)
	if err != nil {
		panic(err)
	}
	fmt.Printf(`%v
`, string(jsbz))
	return nil
}
