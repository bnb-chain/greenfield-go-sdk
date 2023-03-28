package account

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Account struct {
}

func (a Account) Sign() {

}

func (a Account) GetAddr() sdk.AccAddress {

}

func NewAccount() *Account {
	return nil
}

func NewAccountWithMnemonic() *Account {
	return nil
}
