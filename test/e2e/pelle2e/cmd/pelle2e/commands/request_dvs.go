package commands

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/random"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/chainlibs/eth"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/rpc/client/http"
)

var RequestDVSCmdFlagETHRPCURL = &chainflags.StringFlag{
	Name:    "eth-rpc-url",
	Usage:   "eth rpc url",
	Default: "http://eth:8545",
}

var RequestDVSCmdFlagDVSNodeURL = &chainflags.StringFlag{
	Name:    "node-url",
	Usage:   "dvs node url",
	Default: "http://127.0.0.1:26657",
}

var RequestDVSCmdFlagGroupNumber = &chainflags.IntFlag{
	Name:    "group",
	Usage:   "group number",
	Default: 0,
}

var RequestDVSCmdFlagThreshold = &chainflags.IntFlag{
	Name:    "threshold",
	Usage:   "Threshold",
	Default: 2,
}

var RequestDVSCmdFlagBlockNumber = &chainflags.IntFlag{
	Name:  "block-number",
	Usage: "block number",
}

var RequestDVSCmdFlagKey = &chainflags.StringFlag{
	Name: "key",
}

var RequestDVSCmdFlagValue = &chainflags.StringFlag{
	Name: "value",
}

func init() {
	RequestDVSCmdFlagETHRPCURL.AddToCmdFlag(RequestDVSCmd)
	RequestDVSCmdFlagDVSNodeURL.AddToCmdFlag(RequestDVSCmd)
	RequestDVSCmdFlagGroupNumber.AddToCmdFlag(RequestDVSCmd)
	RequestDVSCmdFlagThreshold.AddToCmdFlag(RequestDVSCmd)
	RequestDVSCmdFlagBlockNumber.AddToCmdFlag(RequestDVSCmd)

	RequestDVSCmdFlagKey.AddToCmdFlag(RequestDVSCmd)
	RequestDVSCmdFlagValue.AddToCmdFlag(RequestDVSCmd)
}

var RequestDVSCmd = &cobra.Command{
	Use:  "request-dvs",
	RunE: requestDVSCmdProcess,
}

func requestDVSCmdProcess(cmd *cobra.Command, args []string) error {
	var ethClient eth.Client
	var err error
	ctx := cmd.Context()
	var lastBlockNumber uint64

	ethClient, err = eth.NewClient(RequestDVSCmdFlagETHRPCURL.Value)
	if err != nil {
		return err
	}
	if RequestDVSCmdFlagBlockNumber.Value == 0 {
		lastBlockNumber, err = ethClient.BlockNumber(ctx)
		if err != nil {
			return err
		}
		RequestDVSCmdFlagBlockNumber.Value = int(lastBlockNumber)
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return err
	}

	if RequestDVSCmdFlagKey.Value == "" {
		RequestDVSCmdFlagKey.Value = random.String(10)
	}
	if RequestDVSCmdFlagValue.Value == "" {
		RequestDVSCmdFlagValue.Value = random.String(10)
	}

	blockNumber := int64(RequestDVSCmdFlagBlockNumber.Value)

	data := []byte(fmt.Sprintf("%s=%s", RequestDVSCmdFlagKey.Value, RequestDVSCmdFlagValue.Value))

	groupNumbers := []uint32{uint32(RequestDVSCmdFlagGroupNumber.Value)}
	groupThresholdPercentages := []uint32{uint32(RequestDVSCmdFlagThreshold.Value)}

	logger.Info("RequestDVSAsync",
		"dvsNodeRPCURL", RequestDVSCmdFlagDVSNodeURL.Value,
		"data", string(data),
		"blockNumber", blockNumber,
		"chainID", chainID.Int64(),
		"groupNumbers", groupNumbers,
		"groupThresholdPercentages",
		groupThresholdPercentages,
	)

	err = processRequestDVSCmd(
		ctx,
		RequestDVSCmdFlagDVSNodeURL.Value,
		data,
		blockNumber,
		chainID.Int64(),
		groupNumbers,
		groupThresholdPercentages,
	)
	return err
}

func processRequestDVSCmd(
	ctx context.Context,
	dvsNodeRPCURL string,
	data []byte,
	blockNumber int64,
	chainID int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) error {
	httpClient, err := http.New(dvsNodeRPCURL, "")
	if err != nil {
		return err
	}
	result, err := httpClient.RequestDVS(
		ctx,
		data,
		blockNumber,
		chainID,
		groupNumbers,
		groupThresholdPercentages,
	)
	logger.Info("RequestDVSAsync", "result", result, "err", err)
	return nil
}
