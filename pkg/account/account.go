package account

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Account struct {
}

func (a Account) Sign(unsignedBytes []byte) ([]byte, error) {
	return nil, nil
}

func (a Account) GetAddr() sdk.AccAddress {
	return nil
}

func NewAccount() *Account {
	return nil
}

func NewAccountWithMnemonic() *Account {
	return nil
}
