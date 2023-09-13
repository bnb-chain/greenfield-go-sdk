package client

import (
	"context"
	"time"

	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

type FeeGrant interface {
	GrantBasicAllowance(ctx context.Context, granteeAddr string, feeAllowanceAmount math.Int, expiration *time.Time, txOption gnfdsdktypes.TxOption) (string, error)
	QueryBasicAllowance(ctx context.Context, granterAddr, granteeAddr string) (*feegrant.BasicAllowance, error)

	// for generic allowance(BasicAllowance, PeriodicAllowance, AllowedMsgAllowance)
	GrantAllowance(ctx context.Context, granteeAddr string, allowance feegrant.FeeAllowanceI, txOption gnfdsdktypes.TxOption) (string, error)
	QueryAllowance(ctx context.Context, granterAddr, granteeAddr string) (*feegrant.Grant, error)
	QueryAllowances(ctx context.Context, granteeAddr string) ([]*feegrant.Grant, error)

	RevokeAllowance(ctx context.Context, granteeAddr string, txOption gnfdsdktypes.TxOption) (string, error)
}

// GrantBasicAllowance grants the grantee the BasicAllowance with specified amount and expiration.
func (c *client) GrantBasicAllowance(ctx context.Context, granteeAddr string, feeAllowanceAmount math.Int, expiration *time.Time, txOption gnfdsdktypes.TxOption) (string, error) {
	grantee, err := sdk.AccAddressFromHexUnsafe(granteeAddr)
	if err != nil {
		return "", err
	}
	bnb := sdk.NewCoins(sdk.NewCoin(gnfdsdktypes.Denom, feeAllowanceAmount))
	allowance := feegrant.BasicAllowance{
		SpendLimit: bnb,
		Expiration: expiration,
	}
	msg, err := feegrant.NewMsgGrantAllowance(&allowance, c.defaultAccount.GetAddress(), grantee)
	if err != nil {
		return "", err
	}
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// GrantAllowance provides a generic way to grant different types of allowance(BasicAllowance, PeriodicAllowance, AllowedMsgAllowance), the user needs to construct the desired type of allowance
func (c *client) GrantAllowance(ctx context.Context, granteeAddr string, allowance feegrant.FeeAllowanceI, txOption gnfdsdktypes.TxOption) (string, error) {
	grantee, err := sdk.AccAddressFromHexUnsafe(granteeAddr)
	if err != nil {
		return "", err
	}
	msg, err := feegrant.NewMsgGrantAllowance(allowance, c.defaultAccount.GetAddress(), grantee)
	if err != nil {
		return "", err
	}
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// RevokeAllowance revokes allowance on a grantee by the granter
func (c *client) RevokeAllowance(ctx context.Context, granteeAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	grantee, err := sdk.AccAddressFromHexUnsafe(granteeAddr)
	if err != nil {
		return "", err
	}
	msg := feegrant.NewMsgRevokeAllowance(c.defaultAccount.GetAddress(), grantee)
	if err != nil {
		return "", err
	}
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{&msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// QueryBasicAllowance queries the BasicAllowance
func (c *client) QueryBasicAllowance(ctx context.Context, granterAddr, granteeAddr string) (*feegrant.BasicAllowance, error) {
	allowance, err := c.QueryAllowance(ctx, granterAddr, granteeAddr)
	if err != nil {
		return nil, err
	}
	basicAllowance := &feegrant.BasicAllowance{}
	if err = c.chainClient.GetCodec().Unmarshal(allowance.Allowance.GetValue(), basicAllowance); err != nil {
		return nil, err
	}
	return basicAllowance, nil
}

func (c *client) QueryAllowance(ctx context.Context, granterAddr, granteeAddr string) (*feegrant.Grant, error) {
	_, err := sdk.AccAddressFromHexUnsafe(granterAddr)
	if err != nil {
		return nil, err
	}
	_, err = sdk.AccAddressFromHexUnsafe(granteeAddr)
	if err != nil {
		return nil, err
	}
	req := &feegrant.QueryAllowanceRequest{
		Granter: granterAddr,
		Grantee: granteeAddr,
	}
	response, err := c.chainClient.FeegrantQueryClient.Allowance(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.Allowance, nil
}

func (c *client) QueryAllowances(ctx context.Context, granteeAddr string) ([]*feegrant.Grant, error) {
	_, err := sdk.AccAddressFromHexUnsafe(granteeAddr)
	if err != nil {
		return nil, err
	}
	req := &feegrant.QueryAllowancesRequest{
		Grantee: granteeAddr,
	}
	response, err := c.chainClient.FeegrantQueryClient.Allowances(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.Allowances, nil
}

func (c *client) QueryGranterAllowances(ctx context.Context, granterAddr string) ([]*feegrant.Grant, error) {
	_, err := sdk.AccAddressFromHexUnsafe(granterAddr)
	if err != nil {
		return nil, err
	}
	req := &feegrant.QueryAllowancesByGranterRequest{
		Granter: granterAddr,
	}
	response, err := c.chainClient.FeegrantQueryClient.AllowancesByGranter(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.Allowances, nil
}
