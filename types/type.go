package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultGasLimit = uint64(210000)
)

type SendTokenRequest struct {
	Token     string
	Amount    int64
	ToAddress string
}

// TxBroadcastResponse Generic tx response
type TxBroadcastResponse struct {
	Ok     bool   `json:"ok"`
	Log    string `json:"log"`
	TxHash string `json:"txHash"`
	Code   uint32 `json:"code"`
	Data   string `json:"data"`
}

type TxOption struct {
	Async     bool // default sync mode if not provided
	GasLimit  uint64
	Memo      string
	FeeAmount sdk.Coins
	FeePayer  sdk.AccAddress
}
