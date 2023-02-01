package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/types"
	"github.com/bnb-chain/gnfd-go-sdk/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/grpc"
)

func (c *GreenfieldClient) SendToken(req types.SendTokenRequest, sync bool, opts ...grpc.CallOption) (*types.TxBroadcastResponse, error) {
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
	km, err := c.GetKeyManager()
	if err != nil {
		return nil, err
	}
	transferMsg := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(req.Token, req.Amount)))
	res, err := c.BroadcastTx(sync, []sdk.Msg{transferMsg}, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}
