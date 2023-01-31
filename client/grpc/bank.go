package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/types"
	"github.com/bnb-chain/gnfd-go-sdk/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type SendTokenResult struct {
	types.TxBroadcastResponse
}

func (c *GreenfieldClient) SendToken(token, toAddr string, amount int64, sync bool) (*SendTokenResult, error) {
	if err := util.ValidateToken(token); err != nil {
		return nil, err
	}
	if err := util.ValidateAmount(amount); err != nil {
		return nil, err
	}

	to, err := sdk.AccAddressFromHexUnsafe(toAddr)
	if err != nil {
		return nil, err
	}
	transferMsg := banktypes.NewMsgSend(c.keyManager.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(token, amount)))
	res, err := c.BroadcastTx(sync, transferMsg)
	if err != nil {
		return nil, err
	}
	return &SendTokenResult{
		*res,
	}, nil

}
