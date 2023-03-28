package client

import (
	"github.com/bnb-chain/greenfield-go-sdk/pkg/account"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BasicOption struct {
	TxOption types.TxOption
	Account  account.Account
}

type CreateBucketOptions struct {
	BasicOption
	PaymentAddress   sdk.AccAddress
	PrimarySPAddress sdk.AccAddress
	ChargedQuota     uint64
}

type CreateObjectOptions struct {
	BasicOption
	SecondarySPAddress []sdk.AccAddress
	ContentType        string
	IsReplicaType      bool
}
