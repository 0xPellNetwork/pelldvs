package pelle2e

import (
	"fmt"
	"os/exec"
)

func (per *PellDVSE2ERunner) TriggerAnvilNewBlocks(times int, privateKey, receiverAddress, rpcURL string) error {
	for i := 0; i < times; i++ {
		err := per.genNewBlock(privateKey, receiverAddress, rpcURL)
		if err != nil {
			per.logger.Error("Failed to trigger new block", "err", err)
		}
	}
	return nil
}

func (per *PellDVSE2ERunner) genNewBlock(privateKey, receiverAddress, rpcURL string) error {
	value := "1000000000000000000" // 1 ETH in Wei

	args := []string{
		"send",
		receiverAddress,
		"--value", value,
		"--rpc-url", rpcURL,
		"--private-key", privateKey,
	}
	per.logger.Info("Triggering new block, cast args", "args", args)

	cmd := exec.Command("cast", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		per.logger.Error(fmt.Sprintf("Failed to execute cast send: %v\nOutput: %s", err, string(output)))
		return err
	}

	//fmt.Printf("Command output: %s", string(output))
	return nil
}
