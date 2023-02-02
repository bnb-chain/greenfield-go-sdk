package staking

import (
	"context"
	gnfdclient "github.com/bnb-chain/gnfd-go-sdk/client/grpc"
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStakingValidator(t *testing.T) {
	km, err := keys.NewPrivateKeyManager("e3ac46e277677f0f103774019d03bd89c7b4b5ecc554b2650bd5d5127992c20c")
	assert.NoError(t, err)
	client := gnfdclient.NewGreenlandClientWithKeyManager("localhost:9090", "greenfield_9000-121", km)

	query := stakingtypes.QueryValidatorRequest{
		ValidatorAddr: common.HexToAddress("0xAE713C1ebf18eA66F2Bf882F495429A5cf3bD9dF").String(),
	}
	res, err := client.StakingQueryClient.Validator(context.Background(), &query)
	assert.NoError(t, err)

	println(res.String())
}
