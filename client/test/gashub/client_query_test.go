package gashub

import (
	"context"
<<<<<<< HEAD
	client "github.com/bnb-chain/greenfield-go-sdk/client/chain"
=======
	"testing"

	gnfdclient "github.com/bnb-chain/greenfield-go-sdk/client/chain"
>>>>>>> 70dd951 (support get approval new version)
	"github.com/bnb-chain/greenfield-go-sdk/client/test"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	"github.com/stretchr/testify/assert"
)

func TestGashubParams(t *testing.T) {
	client := client.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := gashubtypes.QueryParamsRequest{}
	res, err := client.GashubQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}
