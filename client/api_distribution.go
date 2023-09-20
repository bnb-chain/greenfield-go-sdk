package client

import (
	"context"

	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type IDistributionClient interface {
	SetWithdrawAddress(ctx context.Context, withdrawAddr string, txOption gnfdsdktypes.TxOption) (string, error)
	WithdrawValidatorCommission(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error)
	WithdrawDelegatorReward(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error)
	FundCommunityPool(ctx context.Context, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
}

// SetWithdrawAddress - Set the withdrawal address for a delegator (or validator self-delegation).
//
// - ctx: Context variables for the current API call.
//
// - withdrawAddr: The withdrawal address to set.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction hash of the transaction.
//
// - ret2: Return error if the transaction failed, otherwise return nil.
func (c *Client) SetWithdrawAddress(ctx context.Context, withdrawAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	withdraw, err := sdk.AccAddressFromHexUnsafe(withdrawAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgSetWithdrawAddress(c.MustGetDefaultAccount().GetAddress(), withdraw)
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// WithdrawValidatorCommission - Withdraw accumulated commission by validator.
//
// - ctx: Context variables for the current API call.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction hash of the transaction.
//
// - ret2: Return error if the transaction failed, otherwise return nil.
func (c *Client) WithdrawValidatorCommission(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error) {
	msg := distrtypes.NewMsgWithdrawValidatorCommission(c.MustGetDefaultAccount().GetAddress())
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// WithdrawDelegatorReward - Withdraw rewards by a delegator.
//
// - ctx: Context variables for the current API call.
//
// - validatorAddr: The validator address to withdraw from.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction hash of the transaction.
//
// - ret2: Return error if the transaction failed, otherwise return nil.
func (c *Client) WithdrawDelegatorReward(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgWithdrawDelegatorReward(c.MustGetDefaultAccount().GetAddress(), validator)
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// FundCommunityPool - Sends coins directly from the sender to the community pool.
//
// - ctx: Context variables for the current API call.
//
// - amount: The amount of BNB to send.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction hash of the transaction.
//
// - ret2: Return error if the transaction failed, otherwise return nil.
func (c *Client) FundCommunityPool(ctx context.Context, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
	msg := distrtypes.NewMsgFundCommunityPool(sdk.Coins{sdk.Coin{Denom: gnfdsdktypes.Denom, Amount: amount}}, c.MustGetDefaultAccount().GetAddress())
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}
