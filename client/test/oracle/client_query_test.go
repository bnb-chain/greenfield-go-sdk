package bank

import (
	"context"
<<<<<<< HEAD
	client "github.com/bnb-chain/greenfield-go-sdk/client/chain"
=======
	"testing"

	gnfdclient "github.com/bnb-chain/greenfield-go-sdk/client/chain"
>>>>>>> 70dd951 (support get approval new version)
	"github.com/bnb-chain/greenfield-go-sdk/client/test"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	"github.com/stretchr/testify/assert"
)

func TestOracleParams(t *testing.T) {
	client := client.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := oracletypes.QueryParamsRequest{}
	res, err := client.OracleQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.GetParams())
}
