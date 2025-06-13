package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/test/e2e/pelle2e/services/pelle2e"
)

var CheckBLSAggrSigCmdFlagDVSNodeURL = &chainflags.StringFlag{
	Name:    "node-url",
	Usage:   "dvs node url",
	Default: "http://127.0.0.1:26657",
}

var CheckBLSAggrSigCmdFlagGroupNumber = &chainflags.IntFlag{
	Name:    "group",
	Usage:   "group number",
	Default: 0,
}

var CheckBLSAggrSigCmdFlagThreshold = &chainflags.IntFlag{
	Name:    "threshold",
	Usage:   "Threshold",
	Default: 60,
}

var CheckBLSAggrSigCmdFlagDVSServiceManagerAddress = &chainflags.StringFlag{
	Name: "service-manager",
	Aliases: []string{
		"service-manager-address",
	},
}

var CheckBLSAggrSigCmdFlagETHRPCURL = &chainflags.StringFlag{
	Name:    "rpc-url",
	Default: "http://eth:8545",
	Aliases: []string{"eth-rpc-url"},
}

var CheckBLSAggrSigCmdFlagSenderPrivateKey = &chainflags.StringFlag{
	Name:    "sender-private-key",
	Default: "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
}

var CheckBLSAggrSigCmdFlagReceiverAddress = &chainflags.StringFlag{
	Name:    "receiver-address",
	Default: "4860f78301d7ef2dd42a1a4a0a230cc8c38d1996",
}

var CheckBLSAggrSigCmdFlagTimesForTriggerNewBlock = &chainflags.IntFlag{
	Name:    "trigger-times",
	Default: 2,
}

var CheckBLSAggrSigCmd = &cobra.Command{
	Use:  "check-aggr-sigs",
	RunE: checkBLSAggrSig,
}

func init() {
	CheckBLSAggrSigCmdFlagDVSNodeURL.AddToCmdFlag(CheckBLSAggrSigCmd)
	CheckBLSAggrSigCmdFlagDVSServiceManagerAddress.AddToCmdFlag(CheckBLSAggrSigCmd)
	CheckBLSAggrSigCmdFlagGroupNumber.AddToCmdFlag(CheckBLSAggrSigCmd)
	CheckBLSAggrSigCmdFlagThreshold.AddToCmdFlag(CheckBLSAggrSigCmd)

	// flags for trigger new block
	CheckBLSAggrSigCmdFlagETHRPCURL.AddToCmdFlag(CheckBLSAggrSigCmd)
	CheckBLSAggrSigCmdFlagSenderPrivateKey.AddToCmdFlag(CheckBLSAggrSigCmd)
	CheckBLSAggrSigCmdFlagReceiverAddress.AddToCmdFlag(CheckBLSAggrSigCmd)
	CheckBLSAggrSigCmdFlagTimesForTriggerNewBlock.AddToCmdFlag(CheckBLSAggrSigCmd)

	// flags required
	_ = CheckBLSAggrSigCmdFlagDVSNodeURL.MarkRequired(CheckBLSAggrSigCmd)
	_ = CheckBLSAggrSigCmdFlagDVSServiceManagerAddress.MarkRequired(CheckBLSAggrSigCmd)
	_ = CheckBLSAggrSigCmdFlagETHRPCURL.MarkRequired(CheckBLSAggrSigCmd)
}

func checkBLSAggrSig(cmd *cobra.Command, args []string) error {
	// check flags
	if CheckBLSAggrSigCmdFlagTimesForTriggerNewBlock.Value == 0 {
		CheckBLSAggrSigCmdFlagTimesForTriggerNewBlock.Value = CheckBLSAggrSigCmdFlagTimesForTriggerNewBlock.Default
	}

	return _checkBLSAggrSig(cmd)
}

func _checkBLSAggrSig(cmd *cobra.Command) error {

	ctx := cmd.Context()
	groupNumbers := []uint32{uint32(CheckBLSAggrSigCmdFlagGroupNumber.Value)}
	groupThresholdPercentages := []uint32{uint32(CheckBLSAggrSigCmdFlagThreshold.Value)}

	per, err := pelle2e.NewPellE2ERunner(ctx,
		CheckBLSAggrSigCmdFlagETHRPCURL.Value,
		CheckBLSAggrSigCmdFlagDVSNodeURL.Value,
		CheckBLSAggrSigCmdFlagDVSServiceManagerAddress.Value,
		logger,
	)
	if err != nil {
		return err
	}

	eectx, req, err := per.PrepareRequest(ctx, groupNumbers, groupThresholdPercentages)
	if err != nil {
		return err
	}

	logger.Info("Requesting DVS request data",
		"eectx", fmt.Sprintf("%v", eectx),
		"req", req,
	)

	reqResp, err := per.RequestDVSAsync(ctx, req)
	if err != nil {
		logger.Error("Failed to request DVS", "error", err)
		return err
	}
	if reqResp == nil {
		logger.Error("reqResp is nil")
		return fmt.Errorf("reqResp is nil")
	}
	if reqResp.Hash == nil {
		logger.Error("reqResp.Hash is nil")
		return fmt.Errorf("reqResp.Hash is nil")
	}

	logger.Info("RequestDVSAsync result", "resp", reqResp)

	var secondsForRequestToBeProcessed = 10 * time.Second
	logger.Info("⌛ waiting for the request to be processed", "seconds", secondsForRequestToBeProcessed)
	time.Sleep(secondsForRequestToBeProcessed)

	logger.Info("Querying request by using hash", "hash", reqResp.Hash.String())
	taskResult, err := per.QueryRequest(ctx, reqResp.Hash.String())
	if err != nil {
		logger.Error("failed to query request", "error", err)
		return err
	}

	if taskResult == nil {
		logger.Error("taskResult is nil")
		return fmt.Errorf("taskResult is nil")
	}

	logger.Info("taskResult",
		"hashHex", reqResp.Hash.String(),
		"hash", reqResp,
		"taskResult", taskResult,
	)

	// test for search request result
	request, err := per.SearchRequest(ctx, "SecondEventType.SecondEventKey='SecondEventValue'",
		nil, nil)
	if err != nil {
		logger.Error("SearchRequest failed",
			"query", "SecondEventType.SecondEventKey='SecondEventValue'",
			"error", err)
		return err
	}

	if request == nil {
		logger.Error("SearchRequest returned no results",
			"query", "SecondEventType.SecondEventKey='SecondEventValue'")
		return fmt.Errorf("search request returned no results")
	} else {
		logger.Info("SearchRequest successful",
			"query", "SecondEventType.SecondEventKey='SecondEventValue'",
			"results", request)
	}
	fmt.Println()
	fmt.Println()

	// Trigger new blocks
	logger.Info("Triggering new blocks")
	err = per.TriggerAnvilNewBlocks(
		CheckBLSAggrSigCmdFlagTimesForTriggerNewBlock.Value,
		CheckBLSAggrSigCmdFlagSenderPrivateKey.Value,
		CheckBLSAggrSigCmdFlagReceiverAddress.Value,
		CheckBLSAggrSigCmdFlagETHRPCURL.Value,
	)
	if err != nil {
		logger.Error("failed to trigger new blocks", "error", err)
		return err
	}

	logger.Info("")
	logger.Info("")

	var secondsForNewBlocksToBeGenerated = 5 * time.Second
	logger.Info("⌛ wainting for new blocks to be generated",
		"seconds", secondsForNewBlocksToBeGenerated,
	)
	time.Sleep(secondsForNewBlocksToBeGenerated)

	logger.Info("")
	logger.Info("")

	// verify BLS signatures on chain
	logger.Info("Checking BLS signatures on chain after new blocks are generated")
	err = per.VerifyBLSSigsOnChain(eectx, taskResult)
	if err != nil {
		logger.Error("Failed to verify BLS signatures on chain", "error", err)
		return err
	}

	fmt.Println("✅ BLS signatures verified successfully")

	return nil
}
