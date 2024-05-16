package bsc

import (
	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

type IAccountClient interface {
	SetDefaultAccount(account *bsctypes.BscAccount)
	GetDefaultAccount() (*bsctypes.BscAccount, error)
}

// SetDefaultAccount - Set the default account of the Client.
//
// If you call other APIs without specifying the account, it will be assumed that you are operating on the default
// account. This includes sending transactions and other actions.
//
// - account: The account to be set as the default account, should be created using a private key or a mnemonic phrase.
func (c *Client) SetDefaultAccount(account *bsctypes.BscAccount) {
	c.defaultAccount = account
}

// GetDefaultAccount - Get the default account of the Client.
//
// - ret1: The default account of the Client.
//
// - ret2: Return error when default account doesn't exist, otherwise return nil.
func (c *Client) GetDefaultAccount() (*bsctypes.BscAccount, error) {
	if c.defaultAccount == nil {
		return nil, types.ErrorDefaultAccountNotExist
	}
	return c.defaultAccount, nil
}
