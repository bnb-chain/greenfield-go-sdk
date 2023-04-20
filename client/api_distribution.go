package client

import (
	"context"
	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type Distribution interface {
	SetWithdrawAddress(ctx context.Context, withdrawAddr string, txOption gnfdsdktypes.TxOption) (string, error)
	WithdrawValidatorCommission(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error)
	WithdrawDelegatorReward(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error)
	FundCommunityPool(ctx context.Context, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
}

func (c *client) SetWithdrawAddress(ctx context.Context, withdrawAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	withdraw, err := sdk.AccAddressFromHexUnsafe(withdrawAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgSetWithdrawAddress(c.MustGetDefaultAccount().GetAddress(), withdraw)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) WithdrawValidatorCommission(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error) {
	msg := distrtypes.NewMsgWithdrawValidatorCommission(c.MustGetDefaultAccount().GetAddress())
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) WithdrawDelegatorReward(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := distrtypes.NewMsgWithdrawDelegatorReward(c.MustGetDefaultAccount().GetAddress(), validator)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) FundCommunityPool(ctx context.Context, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
	msg := distrtypes.NewMsgFundCommunityPool(sdk.Coins{sdk.Coin{Denom: gnfdsdktypes.Denom, Amount: amount}}, c.MustGetDefaultAccount().GetAddress())
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}
