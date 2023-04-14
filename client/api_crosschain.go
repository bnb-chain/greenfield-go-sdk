package client

import (
	"context"
	sdkmath "cosmossdk.io/math"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CrossChain interface {
	TransferOut(ctx context.Context, toAddress string, amount int64, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	MirrorGroup(ctx context.Context, id sdkmath.Uint)
}

func (c *client) TransferOut(ctx context.Context, toAddress string, amount int64, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgTransferOut := bridgetypes.NewMsgTransferOut(c.MustGetDefaultAccount().GetAddress().String(),
		toAddress,
		&sdk.Coin{Denom: gnfdSdkTypes.Denom, Amount: sdk.NewInt(amount)},
	)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgTransferOut}, txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

func (c *client) MirrorGroup(ctx context.Context, operatorAddr string, id sdkmath.Uint, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	operator, err := sdk.AccAddressFromHexUnsafe(operatorAddr)
	if err != nil {
		return nil, err
	}
	msgMirrorGroup := storagetypes.NewMsgMirrorGroup(operator, id)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgMirrorGroup}, txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

func (c *client) MirrorBucket(ctx context.Context, operatorAddr string, id sdkmath.Uint, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	operator, err := sdk.AccAddressFromHexUnsafe(operatorAddr)
	if err != nil {
		return nil, err
	}
	msgMirrorBucket := storagetypes.NewMsgMirrorBucket(operator, id)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgMirrorBucket}, txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

func (c *client) MirrorObject(ctx context.Context, operatorAddr string, id sdkmath.Uint, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	operator, err := sdk.AccAddressFromHexUnsafe(operatorAddr)
	if err != nil {
		return nil, err
	}
	msgMirrorBucket := storagetypes.NewMsgMirrorBucket(operator, id)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgMirrorBucket}, txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}
