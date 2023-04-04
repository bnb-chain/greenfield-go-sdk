package types

import (
	"github.com/bnb-chain/greenfield/sdk/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Account struct {
	name string
	km   keys.KeyManager
}

func NewAccountFromPrivateKey(name, privKey string) (*Account, error) {
	km, err := keys.NewPrivateKeyManager(privKey)
	if err != nil {
		return nil, err
	}
	return &Account{
		name: name,
		km:   km,
	}, nil
}

func NewAccountFromMnemonic(name, mnemonic string) (*Account, error) {
	km, err := keys.NewMnemonicKeyManager(mnemonic)
	if err != nil {
		return nil, err
	}
	return &Account{
		name: name,
		km:   km,
	}, nil
}

func (a *Account) GetKeyManager() keys.KeyManager {
	return a.km
}

func (a *Account) GetAddress() sdk.AccAddress {
	return a.km.GetAddr()
}

func (a *Account) Sign(unsignBytes []byte) ([]byte, error) {
	return a.km.Sign(unsignBytes)
}
