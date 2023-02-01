package bank

import (
	"context"
	gnfdclient "github.com/bnb-chain/gnfd-go-sdk/client/grpc"
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBankParams(t *testing.T) {
	km, err := keys.NewPrivateKeyManager("e3ac46e277677f0f103774019d03bd89c7b4b5ecc554b2650bd5d5127992c20c")
	assert.NoError(t, err)
	client := gnfdclient.NewGreenlandClientWithKeyManager("localhost:9090", "greenfield_9000-121", km)

	query := banktypes.QueryParamsRequest{}
	res, err := client.BankQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	println(res.Params)
	println(res.GetParams())
	println(res.String())
}
