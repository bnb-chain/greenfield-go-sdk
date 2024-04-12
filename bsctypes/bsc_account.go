package bsctypes

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// BscAccount indicates the user's identity information used for interaction with BSC.
type BscAccount struct {
	name string
	km   KeyManager
}

type KeyManager interface {
	GetPrivateKey() *ecdsa.PrivateKey
	GetAddr() *common.Address
}

type BscKeyManager struct {
	privateKey *ecdsa.PrivateKey
	// TODO BARRY replace it to  sdk.AccAddress
	address *common.Address
}

func (k *BscKeyManager) GetPrivateKey() *ecdsa.PrivateKey {
	return k.privateKey
}

func (k *BscKeyManager) GetAddr() *common.Address {
	return k.address
}

func NewBscKeyManager(privateKeyHex string) (KeyManager, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &BscKeyManager{
		privateKey: privateKey,
		address:    &address,
	}, nil
}

// GetKeyManager - Get the key manager of the account.
func (a *BscAccount) GetKeyManager() KeyManager {
	return a.km
}

// GetAddress - Get the address of the account.
func (a *BscAccount) GetAddress() *common.Address {
	return a.km.GetAddr()
}

func NewBscAccountFromPrivateKey(name, privKey string) (*BscAccount, error) {
	km, err := NewBscKeyManager(privKey)
	if err != nil {
		return nil, err
	}
	return &BscAccount{
		name: name,
		km:   km,
	}, nil
}
