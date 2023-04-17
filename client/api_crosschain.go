package client

import (
	"context"
	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CrossChain interface {
	// TransferOut makes a transfer from Greenfield to BSC
	TransferOut(ctx context.Context, toAddress string, amount math.Int, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)

	Claims(ctx context.Context, fromAddr string, srcShainId, destChainId uint32, sequence uint64, timestamp uint64, payload []byte, voteAddrSet []uint64, aggSignature []byte, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	GetChannelSendSequence(ctx context.Context, channelId uint32) (uint64, error)
	GetChannelReceiveSequence(ctx context.Context, channelId uint32) (uint64, error)

	MirrorGroup(ctx context.Context, operatorAddr string, id sdkmath.Uint, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	MirrorBucket(ctx context.Context, operatorAddr string, id sdkmath.Uint, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	MirrorObject(ctx context.Context, operatorAddr string, id sdkmath.Uint, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
}

func (c *client) TransferOut(ctx context.Context, toAddress string, amount math.Int, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgTransferOut := bridgetypes.NewMsgTransferOut(c.MustGetDefaultAccount().GetAddress().String(),
		toAddress,
		&sdk.Coin{Denom: gnfdSdkTypes.Denom, Amount: amount},
	)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgTransferOut}, txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

func (c *client) Claims(ctx context.Context, fromAddr string, srcShainId, destChainId uint32, sequence uint64,
	timestamp uint64, payload []byte, voteAddrSet []uint64, aggSignature []byte, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {

	msg := oracletypes.NewMsgClaim(
		fromAddr,
		srcShainId,
		destChainId,
		sequence,
		timestamp,
		payload,
		voteAddrSet,
		aggSignature)

	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

func (c *client) GetChannelSendSequence(ctx context.Context, channelId uint32) (uint64, error) {
	resp, err := c.chainClient.CrosschainQueryClient.SendSequence(
		ctx,
		&crosschaintypes.QuerySendSequenceRequest{ChannelId: channelId},
	)
	if err != nil {
		return 0, err
	}
	return resp.Sequence, nil
}

func (c *client) GetChannelReceiveSequence(ctx context.Context, channelId uint32) (uint64, error) {
	resp, err := c.chainClient.CrosschainQueryClient.ReceiveSequence(
		ctx,
		&crosschaintypes.QueryReceiveSequenceRequest{ChannelId: channelId},
	)
	if err != nil {
		return 0, err
	}
	return resp.Sequence, nil
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
