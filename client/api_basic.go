package client

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
)

type Basic interface {
	Status(ctx context.Context) (*ctypes.ResultStatus, error)
	BroadcastRawTx(ctx context.Context, txBytes []byte, sync bool) (*ctypes.ResultBroadcastTx, error)
	SimulateRawTx(ctx context.Context, txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
	WaitForBlockHeight(ctx context.Context, height int64) error
	WaitForTx(ctx context.Context, hash string) (*ctypes.ResultTx, error)
	LatestBlockHeight(ctx context.Context) (int64, error)
	SimulateTx(msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
	BroadcastTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
}

// Status returns the current status of the Tendermint node that the client is connected to.
// It takes a context as input and returns a ResultStatus object and an error (if any).
func (c *client) Status(ctx context.Context) (*ctypes.ResultStatus, error) {
	return c.tendermintClient.TmClient.Status(ctx)
}

// BroadcastRawTx broadcasts raw transaction bytes to a Tendermint node.
// It takes a context, transaction bytes, and a sync boolean.
// If sync is true, the transaction is broadcast synchronously.
// If sync is false, the transaction is broadcast asynchronously.
func (c *client) BroadcastRawTx(ctx context.Context, txBytes []byte, sync bool) (*ctypes.ResultBroadcastTx, error) {
	if sync {
		return c.tendermintClient.TmClient.BroadcastTxSync(ctx, txBytes)
	} else {
		return c.tendermintClient.TmClient.BroadcastTxAsync(ctx, txBytes)
	}
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

// LatestBlockHeight returns the latest block height of the blockchain.
// It takes a context as input and returns an int64 representing the latest block height and an error (if any).
func (c *client) LatestBlockHeight(ctx context.Context) (int64, error) {
	resp, err := c.Status(ctx)
	if err != nil {
		return 0, err
	}
	return resp.SyncInfo.LatestBlockHeight, nil
}

// WaitForBlockHeight waits until block height h is committed, or returns an
// error if ctx is canceled.
func (c *client) WaitForBlockHeight(ctx context.Context, h int64) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		latestHeight, err := c.LatestBlockHeight(ctx)
		if err != nil {
			return err
		}
		if latestHeight >= h {
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
	start, err := c.LatestBlockHeight(ctx)
	if err != nil {
		return err
	}
	return c.WaitForBlockHeight(ctx, start+n)
}

// WaitForTx requests the tx from hash, if not found, waits for next block and
// tries again. Returns an error if ctx is canceled.
func (c *client) WaitForTx(ctx context.Context, hash string) (*ctypes.ResultTx, error) {
	bz, err := hex.DecodeString(hash)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decode tx hash '%s'", hash)
	}
	for {
		resp, err := c.tendermintClient.TmClient.Tx(ctx, bz, false)
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
		return resp, nil
	}
}

func (c *client) BroadcastTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error) {
	return c.chainClient.BroadcastTx(ctx, msgs, &txOpt, opts...)
}

func (c *client) SimulateTx(msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	return c.chainClient.SimulateTx(msgs, &txOpt, opts...)
}
