package client

import (
	"context"
	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"time"
)

type FeeGrant interface {
	GrantBasicAllowance(ctx context.Context, granteeAddr string, feeAllowanceAmount math.Int, expiration *time.Time, txOption gnfdsdktypes.TxOption) (string, error)
	QueryBasicAllowance(ctx context.Context, granterAddr, granteeAddr string) (*feegrant.BasicAllowance, error)

	GrantAllowance(ctx context.Context, granteeAddr string, allowance feegrant.FeeAllowanceI, txOption gnfdsdktypes.TxOption) (string, error)
	QueryAllowance(ctx context.Context, granterAddr, granteeAddr string) (*feegrant.Grant, error)

	RevokeAllowance(ctx context.Context, granteeAddr string, txOption gnfdsdktypes.TxOption) (string, error)
}

// GrantBasicAllowance grants the grantee the BasicAllowance with specified amount and expiration. If the
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
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// GrantAllowance provides a generic way to grant 3 types of allowance(BasicAllowance, PeriodicAllowance, AllowedMsgAllowance), the user needs to construct the desired type of allowance
func (c *client) GrantAllowance(ctx context.Context, granteeAddr string, allowance feegrant.FeeAllowanceI, txOption gnfdsdktypes.TxOption) (string, error) {
	grantee, err := sdk.AccAddressFromHexUnsafe(granteeAddr)
	if err != nil {
		return "", err
	}
	msg, err := feegrant.NewMsgGrantAllowance(allowance, c.defaultAccount.GetAddress(), grantee)
	if err != nil {
		return "", err
	}
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) RevokeAllowance(ctx context.Context, granteeAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	grantee, err := sdk.AccAddressFromHexUnsafe(granteeAddr)
	if err != nil {
		return "", err
	}
	msg := feegrant.NewMsgRevokeAllowance(c.defaultAccount.GetAddress(), grantee)
	if err != nil {
		return "", err
	}
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{&msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) QueryBasicAllowance(ctx context.Context, granterAddr, granteeAddr string) (*feegrant.BasicAllowance, error) {
	allowance, err := c.QueryAllowance(ctx, granterAddr, granteeAddr)
	if err != nil {
		return nil, err
	}
	basicAllowance := &feegrant.BasicAllowance{}
	if err = c.chainClient.GetCodec().Unmarshal(allowance.Allowance.GetValue(), basicAllowance); err != nil {
		// Return an error if there was an issue unmarshalling the account data.
		return nil, err
	}
	// Unmarshal the raw account data from the response into a BaseAccount object.
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
	// Unmarshal the raw account data from the response into a BaseAccount object.
	return response.Allowance, nil
}
