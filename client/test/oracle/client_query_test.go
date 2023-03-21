package bank

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	client "github.com/bnb-chain/greenfield-go-sdk/client/chain"
	"github.com/bnb-chain/greenfield-go-sdk/client/test"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	"github.com/stretchr/testify/assert"
)

func TestOracleParams(t *testing.T) {
	client := client.NewGreenfieldClient(test.TEST_GRPC_ADDR,
		test.TEST_CHAIN_ID,
		client.WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))

	query := oracletypes.QueryParamsRequest{}
	res, err := client.OracleQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.GetParams())
}
