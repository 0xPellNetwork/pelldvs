package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs/version"
)

var verbose bool

// VersionCmd ...
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		cmtVersion := version.TMCoreSemVer
		if version.TMGitCommitHash != "" {
			cmtVersion += "+" + version.TMGitCommitHash
		}

		if verbose {
			values, _ := json.MarshalIndent(struct {
				PellDVS       string `json:"pelldvs"`
				AVSI          string `json:"avsi"`
				BlockProtocol uint64 `json:"block_protocol"`
				P2PProtocol   uint64 `json:"p2p_protocol"`
			}{
				PellDVS:       cmtVersion,
				AVSI:          version.AVSIVersion,
				BlockProtocol: version.BlockProtocol,
				P2PProtocol:   version.P2PProtocol,
			}, "", "  ")
			fmt.Println(string(values))
		} else {
			fmt.Println(cmtVersion)
		}
	},
}

func init() {
	VersionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show protocol and library versions")
}
