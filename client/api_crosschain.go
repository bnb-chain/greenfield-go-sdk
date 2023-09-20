package client

import (
	"context"

	"cosmossdk.io/math"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ICrossChainClient interface {
	TransferOut(ctx context.Context, toAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)

	Claims(ctx context.Context, srcShainId, destChainId uint32, sequence uint64, timestamp uint64, payload []byte, voteAddrSet []uint64, aggSignature []byte, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	GetChannelSendSequence(ctx context.Context, destChainId sdk.ChainID, channelId uint32) (uint64, error)
	GetChannelReceiveSequence(ctx context.Context, destChainId sdk.ChainID, channelId uint32) (uint64, error)
	GetInturnRelayer(ctx context.Context, req *oracletypes.QueryInturnRelayerRequest) (*oracletypes.QueryInturnRelayerResponse, error)
	GetCrossChainPackage(ctx context.Context, destChainId sdk.ChainID, channelId uint32, sequence uint64) ([]byte, error)

	MirrorGroup(ctx context.Context, destChainId sdk.ChainID, groupId math.Uint, groupName string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	MirrorBucket(ctx context.Context, destChainId sdk.ChainID, bucketId math.Uint, bucketName string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	MirrorObject(ctx context.Context, destChainId sdk.ChainID, objectId math.Uint, bucketName, objectName string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
}

// TransferOut - Make a transfer from Greenfield to BSC
//
// - ctx: Context variables for the current API call.
//
// - toAddress: The destination address in BSC.
//
// - amount: The amount of BNB to transfer.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction response from Greenfield.
//
// - ret2: Return error if transaction failed, otherwise return nil.
func (c *Client) TransferOut(ctx context.Context, toAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgTransferOut := bridgetypes.NewMsgTransferOut(c.MustGetDefaultAccount().GetAddress().String(),
		toAddress,
		&sdk.Coin{Denom: gnfdSdkTypes.Denom, Amount: amount},
	)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgTransferOut}, &txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

// Claims - Claim cross-chain packages from BSC to Greenfield, used by relayers which run by validators
//
// - ctx: Context variables for the current API call.
//
// - srcChainId: The source chain id.
//
// - destChainId: The destination chain id.
//
// - sequence: The sequence of the claim.
//
// - timestamp: The timestamp of the cross-chain packages.
//
// - payload: The payload of the claim.
//
// - voteAddrSet: The bitset of the voted validators.
//
// - aggSignature: The aggregated bls signature of the claim.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction response from Greenfield.
//
// - ret2: Return error if transaction failed, otherwise return nil.
func (c *Client) Claims(ctx context.Context, srcChainId, destChainId uint32, sequence uint64,
	timestamp uint64, payload []byte, voteAddrSet []uint64, aggSignature []byte, txOption gnfdSdkTypes.TxOption,
) (*sdk.TxResponse, error) {
	msg := oracletypes.NewMsgClaim(
		c.MustGetDefaultAccount().GetAddress().String(),
		srcChainId,
		destChainId,
		sequence,
		timestamp,
		payload,
		voteAddrSet,
		aggSignature)

	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

// GetChannelSendSequence - Get the next send sequence for a channel
//
// - ctx: Context variables for the current API call.
//
// - destChainId: The destination chain id.
//
// - channelId: The channel id to query.
//
// - ret1: Send sequence of the channel.
//
// - ret2: Return error if the query failed, otherwise return nil.
func (c *Client) GetChannelSendSequence(ctx context.Context, destChainId sdk.ChainID, channelId uint32) (uint64, error) {
	resp, err := c.chainClient.CrosschainQueryClient.SendSequence(
		ctx,
		&crosschaintypes.QuerySendSequenceRequest{
			DestChainId: uint32(destChainId),
			ChannelId:   channelId,
		},
	)
	if err != nil {
		return 0, err
	}
	return resp.Sequence, nil
}

// GetChannelReceiveSequence - Get the next receive sequence for a channel
//
// - ctx: Context variables for the current API call.
//
// - destChainId: The destination chain id.
//
// - channelId: The channel id to query.
//
// - ret1: Send sequence of the channel.
//
// - ret2: Return error if the query failed, otherwise return nil.
func (c *Client) GetChannelReceiveSequence(ctx context.Context, destChainId sdk.ChainID, channelId uint32) (uint64, error) {
	resp, err := c.chainClient.CrosschainQueryClient.ReceiveSequence(
		ctx,
		&crosschaintypes.QueryReceiveSequenceRequest{
			DestChainId: uint32(destChainId),
			ChannelId:   channelId,
		},
	)
	if err != nil {
		return 0, err
	}
	return resp.Sequence, nil
}

// GetInturnRelayer - Get the in-turn relayer bls public key and its relay interval
//
// - ctx: Context variables for the current API call.
//
// - req: The request to query in-turn relayer.
// - ret1: The response of the `QueryInturnRelayerRequest` query.
//
// - ret2: Return error if the query failed, otherwise return nil.
func (c *Client) GetInturnRelayer(ctx context.Context, req *oracletypes.QueryInturnRelayerRequest) (*oracletypes.QueryInturnRelayerResponse, error) {
	return c.chainClient.InturnRelayer(ctx, req)
}

// GetCrossChainPackage - Get the cross-chain package by sequence.
//
// - ctx: Context variables for the current API call.
//
// - destChainId: The destination chain id.
//
// - channelId: The channel id to query.
//
// - sequence: The sequence of the cross-chain package.
//
// - ret1: The bytes of the cross-chain package.
//
// - ret2: Return error if the query failed, otherwise return nil.
func (c *Client) GetCrossChainPackage(ctx context.Context, destChainId sdk.ChainID, channelId uint32, sequence uint64) ([]byte, error) {
	resp, err := c.chainClient.CrossChainPackage(
		ctx,
		&crosschaintypes.QueryCrossChainPackageRequest{
			DestChainId: uint32(destChainId),
			ChannelId:   channelId,
			Sequence:    sequence,
		},
	)
	if err != nil {
		return nil, err
	}
	return resp.Package, nil
}

// MirrorGroup - Mirror the group to BSC as an NFT
//
// - ctx: Context variables for the current API call.
//
// - destChainId: The destination chain id.
//
// - groupId: The group id to mirror.
//
// - groupName: The group name.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction response from Greenfield.
//
// - ret2: Return error if the transaction failed, otherwise return nil.
func (c *Client) MirrorGroup(ctx context.Context, destChainId sdk.ChainID, groupId math.Uint, groupName string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgMirrorGroup := storagetypes.NewMsgMirrorGroup(c.MustGetDefaultAccount().GetAddress(), destChainId, groupId, groupName)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgMirrorGroup}, &txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

// MirrorBucket - Mirror the bucket to BSC as an NFT
//
// - ctx: Context variables for the current API call.
//
// - destChainId: The destination chain id.
//
// - bucketId: The bucket id to mirror.
//
// - bucketName: The bucket name.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction response from Greenfield.
//
// - ret2: Return error if the transaction failed, otherwise return nil.
func (c *Client) MirrorBucket(ctx context.Context, destChainId sdk.ChainID, bucketId math.Uint, bucketName string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgMirrorBucket := storagetypes.NewMsgMirrorBucket(c.MustGetDefaultAccount().GetAddress(), destChainId, bucketId, bucketName)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgMirrorBucket}, &txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

// MirrorObject - Mirror the object to BSC as an NFT
//
// - ctx: Context variables for the current API call.
//
// - destChainId: The destination chain id.
//
// - objectId: The object id to mirror.
//
// - bucketName: The bucket name.
//
// - objectName: The object name.
//
// - txOption: The txOption for sending transactions.
//
// - ret1: Transaction response from Greenfield.
//
// - ret2: Return error if the transaction failed, otherwise return nil.
func (c *Client) MirrorObject(ctx context.Context, destChainId sdk.ChainID, objectId math.Uint, bucketName, objectName string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgMirrorObject := storagetypes.NewMsgMirrorObject(c.MustGetDefaultAccount().GetAddress(), destChainId, objectId, bucketName, objectName)
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgMirrorObject}, &txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}
