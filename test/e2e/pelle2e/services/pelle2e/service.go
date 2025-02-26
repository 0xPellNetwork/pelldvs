package pelle2e

import (
	"context"
	"fmt"

	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/mocks/mockdvsservicemanager.sol"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/pelldvs-interactor/chainlibs/eth"
	"github.com/0xPellNetwork/pelldvs/libs/log"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
)

type PellDVSE2ERunner struct {
	DVSNodeRPCURL string
	logger        log.Logger

	Client         eth.Client
	serviceManager *mockdvsservicemanager.MockDvsServiceManager
}

func NewPellE2ERunner(
	ctx context.Context,
	ethRPCURL string,
	DVSNodeRPCURL string,
	serviceManagerAddress string,
	logger log.Logger,
) (*PellDVSE2ERunner, error) {
	per := &PellDVSE2ERunner{
		logger: logger.With("module", "PellDVSE2ERunner"),
	}

	per.logger.Info("NewPellE2ERunner", "DVSNodeRPCURL", DVSNodeRPCURL)

	per.DVSNodeRPCURL = DVSNodeRPCURL

	client, err := eth.NewClient(ethRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create eth client for RPCURL `%s`: %v", ethRPCURL, err)
	}
	per.Client = client

	manager, err := mockdvsservicemanager.NewMockDvsServiceManager(
		common.HexToAddress(serviceManagerAddress), client,
	)
	if err != nil {
		return nil, err
	}
	per.serviceManager = manager

	return per, nil
}

func (per *PellDVSE2ERunner) VerifyBLSSigsOnChain(eeContext *E2EContext, requestResult *ctypes.ResultDvsRequest) error {
	if requestResult == nil {
		return fmt.Errorf("VerifyBLSSigsOnChain: nil requestResult")
	}
	if requestResult.DvsResponse == nil {
		return fmt.Errorf("VerifyBLSSigsOnChain: nil requestResult.DvsResponse")
	}

	if len(requestResult.DvsResponse.SignersAggSigG1) == 0 {
		return fmt.Errorf("no responses to verify")
	}

	err := per.callVerify(eeContext, requestResult)
	return err
}
