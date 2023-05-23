package types

import (
	"encoding/hex"

	"cosmossdk.io/math"

	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Will add an AccountRegister struct to manage multi account.

type Account struct {
	name string
	km   keys.KeyManager
}

type TransferDetail struct {
	ToAddress string
	Amount    math.Int
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

// TODO: return mnemonic to user

func NewAccount(name string) (*Account, string, error) {
	privKey := secp256k1.GenPrivKey()
	km, err := keys.NewPrivateKeyManager(hex.EncodeToString(privKey.Bytes()))
	if err != nil {
		return nil, "", err
	}
	return &Account{
		name: name,
		km:   km,
	}, hex.EncodeToString(privKey.Bytes()), nil
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
