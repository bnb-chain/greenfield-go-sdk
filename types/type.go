package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultGasLimit = uint64(210000)
)

type TxOption struct {
	Async     bool // default sync mode if not provided
	GasLimit  uint64
	Memo      string
	FeeAmount sdk.Coins
	FeePayer  sdk.AccAddress
}
