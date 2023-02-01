package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TransactionClient interface {
	BroadcastTx(sync bool, msgs ...sdk.Msg) (*types.TxBroadcastResponse, error)
	SendToken(req types.SendTokenRequest, sync bool) (*types.TxBroadcastResponse, error)
}
