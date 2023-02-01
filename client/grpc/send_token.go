package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/types"
	"github.com/bnb-chain/gnfd-go-sdk/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (c *GreenfieldClient) SendToken(req types.SendTokenRequest, sync bool) (*types.TxBroadcastResponse, error) {
	if err := util.ValidateToken(req.Token); err != nil {
		return nil, err
	}
	if err := util.ValidateAmount(req.Amount); err != nil {
		return nil, err
	}

	to, err := sdk.AccAddressFromHexUnsafe(req.ToAddress)
	if err != nil {
		return nil, err
	}
	transferMsg := banktypes.NewMsgSend(c.keyManager.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(req.Token, req.Amount)))
	res, err := c.BroadcastTx(sync, transferMsg)
	if err != nil {
		return nil, err
	}
	return res, nil
}
