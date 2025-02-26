package rpc

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	operatorinfoprovider "github.com/0xPellNetwork/pelldvs-interactor/interactor/operator_info_provider"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/store"
	"github.com/0xPellNetwork/pelldvs-interactor/types"
)

var _ AggregatorDVSInteractorer = &AggregatorDVSInteractor{}

type AggregatorDVSInteractorer interface {
	GetOperatorsDVSStateAtBlock(groupNumbers types.GroupNumbers, blockNumber uint32) (map[types.OperatorID]types.OperatorDVSState, error)
	GetGroupsDVSStateAtBlock(groupNumbers types.GroupNumbers, blockNumber uint32) (map[types.GroupNumber]types.GroupDVSState, error)
	GetOperatorInfoByID(operatorID types.OperatorID) (types.OperatorInfo, error)
	getOperatorState(groupNumbers types.GroupNumbers, blockNumber uint32) (*OperatorStateInfo, error)
	GetCheckSignaturesIndices(blockNumber uint32, groupNumbers types.GroupNumbers, nonSignerOperatorIds []types.OperatorID) (types.CheckSignaturesIndices, error)
}

type AggregatorDVSInteractor struct {
	operatorStateRetrieverInteractor operatorinfoprovider.OperatorInfoProvider
	pellStore                        store.PellStore
}

func NewAggregatorDVSInteractor(
	operatorStateRetrieverInteractor operatorinfoprovider.OperatorInfoProvider,
	pellStore store.PellStore,
) *AggregatorDVSInteractor {
	return &AggregatorDVSInteractor{
		operatorStateRetrieverInteractor: operatorStateRetrieverInteractor,
		pellStore:                        pellStore,
	}
}

func (a *AggregatorDVSInteractor) GetOperatorInfoByID(operatorID types.OperatorID) (types.OperatorInfo, error) {
	return a.pellStore.GetOperatorInfoByID(operatorID)
}

func (a *AggregatorDVSInteractor) GetOperatorInfoByAddress(operatorAddress common.Address) (types.OperatorInfo, error) {
	return a.pellStore.GetOperatorInfoByAddress(operatorAddress)
}

func (a *AggregatorDVSInteractor) GetOperatorsDVSStateAtBlock(groupNumbers types.GroupNumbers, blockNumber uint32) (map[types.OperatorID]types.OperatorDVSState, error) {
	return a.operatorStateRetrieverInteractor.GetOperatorsDVSStateAtBlock(groupNumbers, blockNumber)
}

func (a *AggregatorDVSInteractor) GetGroupsDVSStateAtBlock(groupNumbers types.GroupNumbers, blockNumber uint32) (map[types.GroupNumber]types.GroupDVSState, error) {
	return a.operatorStateRetrieverInteractor.GetGroupsDVSStateAtBlock(groupNumbers, blockNumber)
}

func (a *AggregatorDVSInteractor) GetCheckSignaturesIndices(blockNumber uint32, groupNumbers types.GroupNumbers, nonSignerOperatorIds []types.OperatorID) (types.CheckSignaturesIndices, error) {
	return a.operatorStateRetrieverInteractor.GetCheckSignaturesIndices(blockNumber, groupNumbers, nonSignerOperatorIds)
}

func (a *AggregatorDVSInteractor) getOperatorState(groupNumbers types.GroupNumbers, blockNumber uint32) (*OperatorStateInfo, error) {
	operatorState, err := a.operatorStateRetrieverInteractor.GetOperatorsStateAtBlock(groupNumbers, uint64(blockNumber))
	if err != nil {
		return nil, err
	}

	result := &OperatorStateInfo{
		Operators:        make(map[types.OperatorID]common.Address),
		GroupStakes:      make(map[types.GroupNumber]*big.Int),
		GroupOperatorMap: make(map[types.GroupNumber][]OperatorStakeInfo),
	}

	for operatorID, dvsState := range operatorState {
		result.Operators[operatorID] = dvsState.OperatorAddress
		for groupNumber, stake := range dvsState.StakePerGroup {
			if result.GroupOperatorMap[groupNumber] == nil {
				result.GroupOperatorMap[groupNumber] = make([]OperatorStakeInfo, 0)
			}
			result.GroupOperatorMap[groupNumber] = append(result.GroupOperatorMap[groupNumber], OperatorStakeInfo{
				Operator:   dvsState.OperatorAddress,
				OperatorID: operatorID,
				Stake:      stake,
			})

			if result.GroupStakes[groupNumber] == nil {
				result.GroupStakes[groupNumber] = new(big.Int)
			}
			result.GroupStakes[groupNumber].Add(result.GroupStakes[groupNumber], stake)
		}

	}

	return result, nil
}
