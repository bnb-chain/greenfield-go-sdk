package staking

import (
	"context"
	gnfdclient "github.com/bnb-chain/gnfd-go-sdk/client/rpc"
	"github.com/bnb-chain/gnfd-go-sdk/client/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method [recovered]
// Panics if t.Log(res.String())
// Removing .String() lets the test case pass, but the error is printed instead
func TestStakingValidator(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryValidatorRequest{
		ValidatorAddr: testutil.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.Validator(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res)
}

// TODO: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method [recovered]
func TestStakingValidators(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryValidatorsRequest{
		Status: "",
	}
	res, err := client.StakingQueryClient.Validators(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res)
}

// TODO: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method [recovered]
func TestStakingDelagatorValidator(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: testutil.TEST_ADDR,
		ValidatorAddr: testutil.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorValidator(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

// TODO: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method [recovered]
func TestStakingDelagatorValidators(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorValidatorsRequest{
		DelegatorAddr: testutil.TEST_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorValidators(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

// TODO: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method [recovered]
func TestStakingUnbondingDelagation(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: testutil.TEST_ADDR,
		ValidatorAddr: testutil.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.UnbondingDelegation(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingDelagatorDelegations(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: testutil.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingValidatorDelegations(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: testutil.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.ValidatorDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingDelegatorUnbondingDelagation(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: testutil.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorUnbondingDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStaking(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr: testutil.TEST_VAL_ADDR,
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

// TODO: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method [recovered]
func TestStakingHistoricalInfo(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := stakingtypes.QueryHistoricalInfoRequest{
		Height: 1,
	}
	res, err := client.StakingQueryClient.HistoricalInfo(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}
