package client

import (
	"context"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	paymentTypes "github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// IAccountClient - Client APIs for operating Greenfield accounts.
type IAccountClient interface {
	SetDefaultAccount(account *types.Account)
	GetDefaultAccount() (*types.Account, error)
	MustGetDefaultAccount() *types.Account

	GetAccount(ctx context.Context, address string) (authTypes.AccountI, error)
	GetAccountBalance(ctx context.Context, address string) (*sdk.Coin, error)
	GetPaymentAccount(ctx context.Context, address string) (*paymentTypes.PaymentAccount, error)
	GetModuleAccounts(ctx context.Context) ([]authTypes.ModuleAccountI, error)
	GetModuleAccountByName(ctx context.Context, name string) (authTypes.ModuleAccountI, error)
	GetPaymentAccountsByOwner(ctx context.Context, owner string) ([]*paymentTypes.PaymentAccount, error)

	CreatePaymentAccount(ctx context.Context, address string, txOption gnfdSdkTypes.TxOption) (string, error)
	Transfer(ctx context.Context, toAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error)
	MultiTransfer(ctx context.Context, details []types.TransferDetail, txOption gnfdSdkTypes.TxOption) (string, error)
}

// SetDefaultAccount - Set the default account of the Client.
//
// If you call other APIs without specifying the account, it will be assumed that you are operating on the default
// account. This includes sending transactions and other actions.
//
// - account: The account to be set as the default account, should be created using a private key or a mnemonic phrase.
func (c *Client) SetDefaultAccount(account *types.Account) {
	c.defaultAccount = account
	c.chainClient.SetKeyManager(account.GetKeyManager())
}

// GetDefaultAccount - Get the default account of the Client.
//
// - ret1: The default account of the Client.
//
// - ret2: Return error when default account doesn't exist, otherwise return nil.
func (c *Client) GetDefaultAccount() (*types.Account, error) {
	if c.defaultAccount == nil {
		return nil, types.ErrorDefaultAccountNotExist
	}
	return c.defaultAccount, nil
}

// MustGetDefaultAccount - Get the default account of the Client, panic when account not found.
//
// - ret1: The default account of the Client.
func (c *Client) MustGetDefaultAccount() *types.Account {
	if c.defaultAccount == nil {
		panic("Default account not exist, Use SetDefaultAccount to set ")
	}
	return c.defaultAccount
}

// GetAccount - Retrieve on-chain account information for a given address.
//
// - ctx: Context variables for the current API call.
//
// - address: The given address for retrieving.
//
// - ret1: The account interface for the given address.
//
// - ret2: Return error when getting account failed, otherwise return nil.
func (c *Client) GetAccount(ctx context.Context, address string) (authTypes.AccountI, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(address)
	if err != nil {
		return nil, err
	}
	// Call the DefaultAccount method of the chain Client with a QueryAccountRequest containing the address.
	response, err := c.chainClient.Account(ctx, &authTypes.QueryAccountRequest{Address: accAddress.String()})
	if err != nil {
		// Return an error if there was an issue retrieving the account.
		return nil, err
	}

	// Unmarshal the raw account data from the response into a BaseAccount object.
	baseAccount := authTypes.BaseAccount{}
	err = c.chainClient.GetCodec().Unmarshal(response.Account.GetValue(), &baseAccount)
	if err != nil {
		// Return an error if there was an issue unmarshalling the account data.
		return nil, err
	}

	// Return the BaseAccount object as an AccountI interface.
	return &baseAccount, err
}

// CreatePaymentAccount - Create a new payment account for the given address.
//
// The payment account is used to pay for the storage and read quota fee of objects. When you need to pay for different
// buckets separately, you can create different payment accounts to do so. Note that the payment account does not have a
// private key, and only the owner of the payment account can withdraw funds from it. Once the owner revokes permission
// for withdrawal, the funds in the payment account can only be utilized to cover storage and read quota fees.
//
// - ctx: Context variables for the current API call.
//
// - address: The owner address of the new payment account.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Return the transaction hash if created successfully, otherwise return empty string.
//
// - ret2: Return error when created failed, otherwise return nil.
func (c *Client) CreatePaymentAccount(ctx context.Context, address string, txOption gnfdSdkTypes.TxOption) (string, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(address)
	if err != nil {
		return "", err
	}
	msgCreatePaymentAccount := paymentTypes.NewMsgCreatePaymentAccount(accAddress.String())
	tx, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgCreatePaymentAccount}, &txOption)
	if err != nil {
		return "", err
	}
	return tx.TxResponse.TxHash, nil
}

// GetModuleAccountByName - Get module account by module name.
//
// - ctx: Context variables for the current API call.
//
// - name: Module name.
//
// - ret1: The account interface for the given module name.
//
// - ret2: Return error when getting failed, otherwise return nil.
func (c *Client) GetModuleAccountByName(ctx context.Context, name string) (authTypes.ModuleAccountI, error) {
	response, err := c.chainClient.ModuleAccountByName(ctx, &authTypes.QueryModuleAccountByNameRequest{Name: name})
	if err != nil {
		return nil, err
	}
	// Unmarshal the raw account data from the response into a BaseAccount object.
	moduleAccount := authTypes.ModuleAccount{}
	err = c.chainClient.GetCodec().Unmarshal(response.Account.GetValue(), &moduleAccount)
	if err != nil {
		// Return an error if there was an issue unmarshalling the account data.
		return nil, err
	}

	// Return the BaseAccount object as an AccountI interface.
	return &moduleAccount, err
}

// GetModuleAccounts - Get all module accounts.
//
// - ctx: Context variables for the current API call.
//
// - ret1: The account interface lists for all the module accounts.
//
// - ret2: Return error when getting failed, otherwise return nil.
func (c *Client) GetModuleAccounts(ctx context.Context) ([]authTypes.ModuleAccountI, error) {
	response, err := c.chainClient.ModuleAccounts(ctx, &authTypes.QueryModuleAccountsRequest{})
	if err != nil {
		return nil, err
	}
	var accounts []authTypes.ModuleAccountI
	for _, accValue := range response.Accounts {
		moduleAccount := authTypes.ModuleAccount{}
		err = c.chainClient.GetCodec().Unmarshal(accValue.Value, &moduleAccount)
		if err != nil {
			// Return an error if there was an issue unmarshalling the account data.
			return nil, err
		}
		accounts = append(accounts, &moduleAccount)
	}
	return accounts, err
}

// GetAccountBalance - Get the bank balance for the given address.
//
// - ctx: Context variables for the current API call.
//
// - address: The given address for retrieving.
//
// - ret1: The balance info for the given address, in sdk.Coin format.
//
// - ret2: Return error when getting failed, otherwise return nil.
func (c *Client) GetAccountBalance(ctx context.Context, address string) (*sdk.Coin, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(address)
	if err != nil {
		return nil, err
	}
	response, err := c.chainClient.BankQueryClient.Balance(ctx, &bankTypes.QueryBalanceRequest{Address: accAddress.String(), Denom: gnfdSdkTypes.Denom})
	if err != nil {
		return nil, err
	}

	return response.Balance, nil
}

// GetPaymentAccount - Get payment account by the payment account's address.
//
// - ctx: Context variables for the current API call.
//
// - address: The given payment account address for retrieving.
//
// - ret1: The payment account info for the given address.
//
// - ret2: Return error when getting failed, otherwise return nil.
func (c *Client) GetPaymentAccount(ctx context.Context, address string) (*paymentTypes.PaymentAccount, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(address)
	if err != nil {
		return nil, err
	}
	pa, err := c.chainClient.PaymentAccount(ctx, &paymentTypes.QueryPaymentAccountRequest{Addr: accAddress.String()})
	if err != nil {
		return nil, err
	}
	return &pa.PaymentAccount, nil
}

// GetPaymentAccountsByOwner - Get all payment accounts owned by the given owner address.
//
// - ctx: Context variables for the current API call.
//
// - address: The given owner account address for retrieving.
//
// - ret1: The payment accounts list for the given owner address.
//
// - ret2: Return error when getting failed, otherwise return nil.
func (c *Client) GetPaymentAccountsByOwner(ctx context.Context, owner string) ([]*paymentTypes.PaymentAccount, error) {
	ownerAcc, err := sdk.AccAddressFromHexUnsafe(owner)
	if err != nil {
		return nil, err
	}
	// Call the GetPaymentAccountsByOwner method of the chain Client with a QueryGetPaymentAccountsByOwnerRequest containing the owner address.
	accountsByOwnerResponse, err := c.chainClient.PaymentAccountsByOwner(ctx, &paymentTypes.QueryPaymentAccountsByOwnerRequest{Owner: ownerAcc.String()})
	if err != nil {
		return nil, err
	}

	// Initialize a slice of PaymentAccount pointers.
	paymentAccounts := make([]*paymentTypes.PaymentAccount, 0, len(accountsByOwnerResponse.PaymentAccounts))

	// Iterate over each account address returned in the response.
	for _, accAddress := range accountsByOwnerResponse.PaymentAccounts {
		// Call the GetPaymentAccount method of the Client to retrieve the PaymentAccount object for the given address.
		pa, err := c.GetPaymentAccount(ctx, accAddress)
		if err != nil {
			return nil, err
		}
		// Append the PaymentAccount object to the slice.
		paymentAccounts = append(paymentAccounts, pa)
	}

	// Return the slice of PaymentAccount pointers.
	return paymentAccounts, nil
}

// Transfer - Transfer BNB from sender to receiver.
//
// - ctx: Context variables for the current API call.
//
// - toAddress: The address who will receive the BNB.
//
// - amount: The BNB amount to transfer, 1e18 denotes 1BNB.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Return the transaction hash if transferred successfully, otherwise return empty string.
//
// - ret2: Return error if transferred failed, otherwise return nil.
func (c *Client) Transfer(ctx context.Context, toAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error) {
	toAddr, err := sdk.AccAddressFromHexUnsafe(toAddress)
	if err != nil {
		return "", err
	}
	msgSend := bankTypes.NewMsgSend(c.MustGetDefaultAccount().GetAddress(), toAddr, sdk.Coins{sdk.Coin{Denom: gnfdSdkTypes.Denom, Amount: amount}})
	tx, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgSend}, &txOption)
	if err != nil {
		return "", err
	}
	return tx.TxResponse.TxHash, nil
}

// MultiTransfer - Transfer BNB from sender to multiple receivers.
//
// - ctx: Context variables for the current API call.
//
// - details: The receiver address and transfer amount list.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Return the transaction hash if transferred successfully, otherwise return empty string.
//
// - ret2: Return error if transferred failed, otherwise return nil.
func (c *Client) MultiTransfer(ctx context.Context, details []types.TransferDetail, txOption gnfdSdkTypes.TxOption) (string, error) {
	outputs := make([]bankTypes.Output, 0)
	denom := gnfdSdkTypes.Denom
	sum := math.NewInt(0)
	for i := 0; i < len(details); i++ {
		outputs = append(outputs, bankTypes.Output{
			Address: details[i].ToAddress,
			Coins:   []sdk.Coin{{Denom: denom, Amount: details[i].Amount}},
		})
		sum = sum.Add(details[i].Amount)
	}
	in := bankTypes.Input{
		Address: c.MustGetDefaultAccount().GetAddress().String(),
		Coins:   []sdk.Coin{{Denom: denom, Amount: sum}},
	}
	msg := &bankTypes.MsgMultiSend{
		Inputs:  []bankTypes.Input{in},
		Outputs: outputs,
	}
	tx, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return tx.TxResponse.TxHash, nil
}
