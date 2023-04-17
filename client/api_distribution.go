package client

import (
	"context"
	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type Distribution interface {
	SetWithdrawAddress(ctx context.Context, delAddr, withdrawAddr string, txOption gnfdsdktypes.TxOption) (string, error)
	WithdrawValidatorCommission(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error)
	WithdrawDelegatorReward(ctx context.Context, delAddr, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error)
	FundCommunityPool(ctx context.Context, amount math.Int, depositorAddr string, txOption gnfdsdktypes.TxOption) (string, error)
}

func (c *client) SetWithdrawAddress(ctx context.Context, delAddr, withdrawAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	del, err := sdk.AccAddressFromHexUnsafe(delAddr)
	if err != nil {
		return "", err
	}
	withdraw, err := sdk.AccAddressFromHexUnsafe(withdrawAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgSetWithdrawAddress(del, withdraw)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) WithdrawValidatorCommission(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgWithdrawValidatorCommission(validator)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) WithdrawDelegatorReward(ctx context.Context, delAddr, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	del, err := sdk.AccAddressFromHexUnsafe(delAddr)
	if err != nil {
		return "", err
	}
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgWithdrawDelegatorReward(del, validator)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) FundCommunityPool(ctx context.Context, amount math.Int, depositorAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	depositor, err := sdk.AccAddressFromHexUnsafe(depositorAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgFundCommunityPool(sdk.Coins{sdk.Coin{Denom: gnfdsdktypes.Denom, Amount: amount}}, depositor)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}
