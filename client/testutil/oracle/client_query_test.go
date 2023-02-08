package bank

import (
	"context"
	gnfdclient "github.com/bnb-chain/gnfd-go-sdk/client/rpc"
	"github.com/bnb-chain/gnfd-go-sdk/client/testutil"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOracleParams(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := oracletypes.QueryParamsRequest{}
	res, err := client.OracleQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.GetParams())
}
