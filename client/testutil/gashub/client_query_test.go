package gashub

import (
	"context"
	gnfdclient "github.com/bnb-chain/gnfd-go-sdk/client/rpc"
	"github.com/bnb-chain/gnfd-go-sdk/client/testutil"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method [recovered]
func TestGashubParams(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := gashubtypes.QueryParamsRequest{}
	res, err := client.GashubQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}
