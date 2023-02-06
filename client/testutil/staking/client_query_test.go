package staking

import (
	"context"
	gnfdclient "github.com/bnb-chain/gnfd-go-sdk/client/rpc"
	"github.com/bnb-chain/gnfd-go-sdk/client/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStakingValidator(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryValidatorRequest{
		ValidatorAddr: common.HexToAddress(testutil.TEST_ADDR).String(),
	}
	res, err := client.StakingQueryClient.Validator(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingValidators(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryValidatorsRequest{
		Status: "active",
	}
	res, err := client.StakingQueryClient.Validators(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingDelagatorValidator(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: "gov",
		ValidatorAddr: testutil.TEST_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorValidator(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingDelagatorValidators(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorValidatorsRequest{
		DelegatorAddr: "gov",
	}
	res, err := client.StakingQueryClient.DelegatorValidators(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingUnbondingDelagation(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: "gov",
		ValidatorAddr: testutil.TEST_ADDR,
	}
	res, err := client.StakingQueryClient.UnbondingDelegation(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingDelagatorDelegations(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: "gov",
	}
	res, err := client.StakingQueryClient.DelegatorDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingValidatorDelegations(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: testutil.TEST_ADDR,
	}
	res, err := client.StakingQueryClient.ValidatorDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingDelegatorUnbondingDelagation(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: "gov",
	}
	res, err := client.StakingQueryClient.DelegatorUnbondingDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

// TODO: Change delegator addr
func TestStaking(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr:    "gov",
		SrcValidatorAddr: testutil.TEST_ADDR,
		DstValidatorAddr: testutil.TEST_ADDR,
	}
	res, err := client.StakingQueryClient.Redelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingParams(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryParamsRequest{}
	res, err := client.StakingQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingPool(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryPoolRequest{}
	res, err := client.StakingQueryClient.Pool(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingHistoricalInfo(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryHistoricalInfoRequest{
		Height: 10,
	}
	res, err := client.StakingQueryClient.HistoricalInfo(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}
