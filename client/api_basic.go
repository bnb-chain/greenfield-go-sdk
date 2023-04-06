package client

import (
	"context"
	"strings"
	"time"

	"cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/tendermint/tendermint/proto/tendermint/p2p"
	"google.golang.org/grpc"
)

type Basic interface {
	GetNodeInfo(ctx context.Context) (*p2p.DefaultNodeInfo, *tmservice.VersionInfo, error)
	BroadcastRawTx(ctx context.Context, txBytes []byte, sync bool) (*sdk.TxResponse, error)
	SimulateRawTx(ctx context.Context, txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
	WaitForBlockHeight(ctx context.Context, height int64) error
	WaitForTx(ctx context.Context, hash string) (*sdk.TxResponse, error)
	LatestBlockHeight(ctx context.Context) (int64, error)
	LatestBlock(ctx context.Context) (*tmservice.Block, error)
	SimulateTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
	BroadcastTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
	GetSyncing(ctx context.Context) (bool, error)
	GetBlockByHeight(ctx context.Context, height int64) (*tmservice.Block, error)
	GetValidatorSet(ctx context.Context, request *query.PageRequest) (int64, []*tmservice.Validator, *query.PageResponse, error)
}

// GetNodeInfo returns the current node info of the greenfield that the client is connected to.
// It takes a context as input and returns a ResultStatus object and an error (if any).
func (c *client) GetNodeInfo(ctx context.Context) (*p2p.DefaultNodeInfo, *tmservice.VersionInfo, error) {
	nodeInfoResponse, err := c.chainClient.TmClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	if err != nil {
		return nil, nil, err
	}
	return nodeInfoResponse.DefaultNodeInfo, nodeInfoResponse.ApplicationVersion, nil
}

// BroadcastRawTx broadcasts raw transaction bytes to a Tendermint node.
// It takes a context, transaction bytes, and a sync boolean.
// If sync is true, the transaction is broadcast synchronously.
// If sync is false, the transaction is broadcast asynchronously.
func (c *client) BroadcastRawTx(ctx context.Context, txBytes []byte, sync bool) (*sdk.TxResponse, error) {
	mode := tx.BroadcastMode_BROADCAST_MODE_UNSPECIFIED
	if sync {
		mode = tx.BroadcastMode_BROADCAST_MODE_SYNC
	} else {
		mode = tx.BroadcastMode_BROADCAST_MODE_ASYNC
	}
	broadcastTxResponse, err := c.chainClient.TxClient.BroadcastTx(ctx, &tx.BroadcastTxRequest{TxBytes: txBytes, Mode: mode})
	if err != nil {
		return nil, err
	}
	return broadcastTxResponse.TxResponse, nil
}

// SimulateRawTx simulates the execution of a raw transaction on the blockchain without broadcasting it to the network.
// It takes a context, transaction bytes, and any additional gRPC call options.
// It returns a SimulateResponse object and an error (if any).
func (c *client) SimulateRawTx(ctx context.Context, txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	simulateResponse, err := c.chainClient.TxClient.Simulate(
		ctx,
		&tx.SimulateRequest{
			TxBytes: txBytes,
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return simulateResponse, nil
}

func (c *client) LatestBlock(ctx context.Context) (*tmservice.Block, error) {
	resp, err := c.chainClient.TmClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return nil, err
	}
	return resp.SdkBlock, nil
}

func (c *client) LatestBlockHeight(ctx context.Context) (int64, error) {
	resp, err := c.LatestBlock(ctx)
	if err != nil {
		return 0, nil
	}
	return resp.Header.Height, nil
}

// WaitForBlockHeight waits until block height h is committed, or returns an
// error if ctx is canceled.
func (c *client) WaitForBlockHeight(ctx context.Context, h int64) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		latestBlockHeight, err := c.LatestBlockHeight(ctx)
		if err != nil {
			return err
		}
		if latestBlockHeight >= h {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "timeout exceeded waiting for block")
		case <-ticker.C:
		}
	}

}

// WaitForNextBlock waits until next block is committed.
// It reads the current block height and then waits for another block to be
// committed, or returns an error if ctx is canceled.
func (c *client) WaitForNextBlock(ctx context.Context) error {
	return c.WaitForNBlocks(ctx, 1)
}

// WaitForNBlocks reads the current block height and then waits for another n
// blocks to be committed, or returns an error if ctx is canceled.
func (c *client) WaitForNBlocks(ctx context.Context, n int64) error {
	start, err := c.LatestBlock(ctx)
	if err != nil {
		return err
	}
	return c.WaitForBlockHeight(ctx, start.Header.Height+n)
}

// WaitForTx requests the tx from hash, if not found, waits for next block and
// tries again. Returns an error if ctx is canceled.
func (c *client) WaitForTx(ctx context.Context, hash string) (*sdk.TxResponse, error) {
	for {
		txResponse, err := c.chainClient.TxClient.GetTx(ctx, &tx.GetTxRequest{Hash: hash})
		if err != nil {
			return nil, err
		}
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// Tx not found, wait for next block and try again
				err := c.WaitForNextBlock(ctx)
				if err != nil {
					return nil, errors.Wrap(err, "waiting for next block")
				}
				continue
			}
			return nil, errors.Wrapf(err, "fetching tx '%s'", hash)
		}
		// Tx found
		return txResponse.TxResponse, nil
	}
}

func (c *client) BroadcastTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error) {
	return c.chainClient.BroadcastTx(ctx, msgs, &txOpt, opts...)
}

func (c *client) SimulateTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	return c.chainClient.SimulateTx(ctx, msgs, &txOpt, opts...)
}

func (c *client) GetSyncing(ctx context.Context) (bool, error) {
	syncing, err := c.chainClient.GetSyncing(ctx, &tmservice.GetSyncingRequest{})
	if err != nil {
		return false, err
	}
	return syncing.Syncing, nil
}

func (c *client) GetBlockByHeight(ctx context.Context, height int64) (*tmservice.Block, error) {
	blockByHeight, err := c.chainClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: height})
	if err != nil {
		return nil, err
	}
	return blockByHeight.SdkBlock, nil

}

func (c *client) GetValidatorSet(ctx context.Context, request *query.PageRequest) (int64, []*tmservice.Validator, *query.PageResponse, error) {
	validatorSetResponse, err := c.chainClient.TmClient.GetLatestValidatorSet(ctx, &tmservice.GetLatestValidatorSetRequest{Pagination: request})
	if err != nil {
		return 0, nil, nil, err
	}
	return validatorSetResponse.BlockHeight, validatorSetResponse.Validators, validatorSetResponse.Pagination, nil
}
