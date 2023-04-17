package e2e

import (
	"context"
	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
)

func Test_CrossChian_Transfer(t *testing.T) {
	mnemonic := ParseValidatorMnemonic(0)
	account, err := types.NewAccountFromMnemonic("test", mnemonic)
	assert.NoError(t, err)
	cli, err := client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: account,
	})
	assert.NoError(t, err)
	ctx := context.Background()

	resp, err := cli.TransferOut(ctx, "0xA4A2957E858529FFABBBb483D1D704378a9fca6b", math.NewInt(1000), &gnfdsdktypes.TxOption{})
	assert.NoError(t, err)
	assert.Equal(t, resp.Code, uint32(0))
}
