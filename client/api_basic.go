package client

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cometbft/cometbft/votepool"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"

	"cosmossdk.io/errors"
	gosdktypes "github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cometbft/cometbft/proto/tendermint/p2p"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	bfttypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
)

// IBasicClient interface defines basic functions of greenfield Client.
type IBasicClient interface {
	EnableTrace(outputStream io.Writer, onlyTraceErr bool)

	GetNodeInfo(ctx context.Context) (*p2p.DefaultNodeInfo, *tmservice.VersionInfo, error)

	GetStatus(ctx context.Context) (*ctypes.ResultStatus, error)
	GetCommit(ctx context.Context, height int64) (*ctypes.ResultCommit, error)
	GetLatestBlockHeight(ctx context.Context) (int64, error)
	GetLatestBlock(ctx context.Context) (*bfttypes.Block, error)
	GetSyncing(ctx context.Context) (bool, error)
	GetBlockByHeight(ctx context.Context, height int64) (*bfttypes.Block, error)
	GetBlockResultByHeight(ctx context.Context, height int64) (*ctypes.ResultBlockResults, error)

	GetValidatorSet(ctx context.Context) (int64, []*bfttypes.Validator, error)
	GetValidatorsByHeight(ctx context.Context, height int64) ([]*bfttypes.Validator, error)

	WaitForBlockHeight(ctx context.Context, height int64) error
	WaitForTx(ctx context.Context, hash string) (*ctypes.ResultTx, error)
	WaitForNBlocks(ctx context.Context, n int64) error
	WaitForNextBlock(ctx context.Context) error

	SimulateTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
	SimulateRawTx(ctx context.Context, txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
	BroadcastTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
	BroadcastRawTx(ctx context.Context, txBytes []byte, sync bool) (*sdk.TxResponse, error)

	BroadcastVote(ctx context.Context, vote votepool.Vote) error
	QueryVote(ctx context.Context, eventType int, eventHash []byte) (*ctypes.ResultQueryVote, error)
}

// EnableTrace support trace error info the request and the response
func (c *Client) EnableTrace(output io.Writer, onlyTraceErr bool) {
	if output == nil {
		output = os.Stdout
	}

	c.onlyTraceError = onlyTraceErr

	c.traceOutput = output
	c.isTraceEnabled = true
}

// GetNodeInfo - Get the current node info of the greenfield that the Client is connected to.
//
// - ctx: Context variables for the current API call.
//
// - ret1: The Node info.
//
// - ret2: The Version info.
//
// - ret3: Return error when the request failed, otherwise return nil.
func (c *Client) GetNodeInfo(ctx context.Context) (*p2p.DefaultNodeInfo, *tmservice.VersionInfo, error) {
	nodeInfoResponse, err := c.chainClient.TmClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	if err != nil {
		return nil, nil, err
	}
	return nodeInfoResponse.DefaultNodeInfo, nodeInfoResponse.ApplicationVersion, nil
}

// GetStatus - Get the status of connected Node.
//
// - ctx: Context variables for the current API call.
//
// - ret1: The detail of Node status.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetStatus(ctx context.Context) (*ctypes.ResultStatus, error) {
	return c.chainClient.GetStatus(ctx)
}

// GetCommit - Get the block commit detail.
//
// - ctx: Context variables for the current API call.
//
// - height: The block height.
//
// - ret1: The commit result.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetCommit(ctx context.Context, height int64) (*ctypes.ResultCommit, error) {
	return c.chainClient.GetCommit(ctx, height)
}

// BroadcastRawTx - Broadcast raw transaction bytes to a Tendermint node.
//
// - ctx: Context variables for the current API call.
//
// - txBytes: The transaction bytes.
//
// - sync: A flag to specify the transaction mode. If it is true, the transaction is broadcast synchronously. If it is false, the transaction is broadcast asynchronously.
//
// - ret1: Transaction response, it can indicate both success and failed transaction.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) BroadcastRawTx(ctx context.Context, txBytes []byte, sync bool) (*sdk.TxResponse, error) {
	var mode tx.BroadcastMode
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

// SimulateRawTx - Simulate the execution of a raw transaction on the blockchain without broadcasting it to the network.
//
// - ctx: Context variables for the current API call.
//
// - txBytes: The transaction bytes.
//
// - opts: The grpc option(s) if Client is using grpc connection.
//
// - ret1: The simulation result.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) SimulateRawTx(ctx context.Context, txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
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

// GetLatestBlock - Get the latest block from the chain.
//
// - ctx: Context variables for the current API call.
//
// - ret1: The block result.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetLatestBlock(ctx context.Context) (*bfttypes.Block, error) {
	res, err := c.chainClient.GetBlock(ctx, nil)
	if err != nil {
		return nil, err
	}
	return res.Block, nil
}

// GetLatestBlockHeight - Get the height of the latest block from the chain.
//
// - ctx: Context variables for the current API call.
//
// - ret1: The block height.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	resp, err := c.GetLatestBlock(ctx)
	if err != nil {
		return 0, nil
	}
	return resp.Header.Height, nil
}

// WaitForBlockHeight - Wait until a specified block height is committed.
//
// - ctx: Context variables for the current API call.
//
// - ret: Return error when the request failed, otherwise return nil.
func (c *Client) WaitForBlockHeight(ctx context.Context, h int64) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		latestBlockHeight, err := c.GetLatestBlockHeight(ctx)
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

// WaitForNextBlock - Wait until the next block is committed since current block.
//
// - ctx: Context variables for the current API call.
//
// - ret: Return error when the request failed, otherwise return nil.
func (c *Client) WaitForNextBlock(ctx context.Context) error {
	return c.WaitForNBlocks(ctx, 1)
}

// WaitForNBlocks - Wait for another n blocks to be committed since current block.
//
// - ctx: Context variables for the current API call.
//
// - n: number of blocks to be waited.
//
// - ret: Return error when the request failed, otherwise return nil.
func (c *Client) WaitForNBlocks(ctx context.Context, n int64) error {
	start, err := c.GetLatestBlock(ctx)
	if err != nil {
		return err
	}
	return c.WaitForBlockHeight(ctx, start.Header.Height+n)
}

// WaitForTx - Wait for a transaction to be confirmed onchian, if transaction not found in current block, wait for the next block. API ends when a transaction is found or context is canceled.
//
// - ctx: Context variables for the current API call.
//
// - hash: The hex representation of transaction hash.
//
// - ret1: The transaction result details.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) WaitForTx(ctx context.Context, hash string) (*ctypes.ResultTx, error) {
	for {
		var (
			txResponse *ctypes.ResultTx
			err        error
			waitTxCtx  context.Context
			cancelFunc context.CancelFunc
		)

		// when websocket conn is used, use a short timeout context to achieve the retry mechanism
		if c.useWebsocketConn {
			waitTxCtx, cancelFunc = context.WithTimeout(context.Background(), gosdktypes.WaitTxContextTimeOut)
			txResponse, err = c.chainClient.Tx(waitTxCtx, hash)
			cancelFunc()
		} else {
			txResponse, err = c.chainClient.Tx(ctx, hash)
		}
		if err != nil {
			// Tx not found, wait for next block and try again
			// If websocket conn is enabled, we also want to re-try the GetTx calls by having a timeout context
			if strings.Contains(err.Error(), "not found") || (c.useWebsocketConn && (waitTxCtx.Err() == context.DeadlineExceeded)) {

				err := c.WaitForNextBlock(ctx)
				if err != nil {
					return nil, errors.Wrap(err, "waiting for next block")
				}
				continue
			}
			return nil, errors.Wrapf(err, "fetching tx '%s'", hash)
		}
		// `nil` could mean the transaction is in the mempool, invalidated, or was not sent in the first place.
		if txResponse == nil {
			err := c.WaitForNextBlock(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "waiting for next block")
			}
			continue
		}
		// Tx found
		return txResponse, nil
	}
}

// BroadcastTx - Broadcast a transaction containing the provided message(s) to the chain.
//
// - ctx: Context variables for the current API call.
//
// - msgs: Message(s) to be broadcast to blockchain.
//
// - txOpt: txOpt contains options for customizing the transaction.
//
// - opts: The grpc option(s) if Client is using grpc connection.
//
// - ret1: transaction response, it can indicate both success and failed transaction.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) BroadcastTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error) {
	return c.chainClient.BroadcastTx(ctx, msgs, &txOpt, opts...)
}

// SimulateTx - Simulate a transaction containing the provided message(s) on the chain.
//
// - ctx: Context variables for the current API call.
//
// - msgs: Message(s) to be broadcast to blockchain.
//
// - txOpt: TxOpt contains options for customizing the transaction.
//
// - opts: The grpc option(s) if Client is using grpc connection.
//
// - ret1: The simulation result.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) SimulateTx(ctx context.Context, msgs []sdk.Msg, txOpt types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	return c.chainClient.SimulateTx(ctx, msgs, &txOpt, opts...)
}

// GetSyncing - Retrieve the syncing status of the node.
//
// - ctx: Context variables for the current API call.
//
// - ret1: The boolean value which indicates whether the node has caught up the latest block.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetSyncing(ctx context.Context) (bool, error) {
	syncing, err := c.chainClient.GetSyncing(ctx, &tmservice.GetSyncingRequest{})
	if err != nil {
		return false, err
	}
	return syncing.Syncing, nil
}

// GetBlockByHeight - Retrieve the block at the given height from the chain.
//
// - ctx: Context variables for the current API call.
//
// - height: The block height.
//
// - ret1: The boolean value which indicates whether the node has caught up the latest block.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetBlockByHeight(ctx context.Context, height int64) (*bfttypes.Block, error) {
	blockByHeight, err := c.chainClient.GetBlock(ctx, &height)
	if err != nil {
		return nil, err
	}
	return blockByHeight.Block, nil
}

// GetBlockResultByHeight - Retrieve the block result at the given height from the chain.
//
// - ctx: Context variables for the current API call.
//
// - height: The block height.
//
// - ret1: The boolean value which indicates whether the node has caught up the latest block.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetBlockResultByHeight(ctx context.Context, height int64) (*ctypes.ResultBlockResults, error) {
	return c.chainClient.GetBlockResults(ctx, &height)
}

// GetValidatorSet - Retrieve the latest validator set from the chain.
//
// - ctx: Context variables for the current API call.
//
// - ret1: The latest height of block that validators set info retrieved from.
//
// - ret2: The list of validators.
//
// - ret3: Return error when the request failed, otherwise return nil.
func (c *Client) GetValidatorSet(ctx context.Context) (int64, []*bfttypes.Validator, error) {
	validatorSetResponse, err := c.chainClient.GetValidators(ctx, nil)
	if err != nil {
		return 0, nil, err
	}
	return validatorSetResponse.BlockHeight, validatorSetResponse.Validators, nil
}

// GetValidatorsByHeight - Retrieve the validator set at a given block height from the chain.
//
// - ctx: Context variables for the current API call.
//
// - height: The block height.
//
// - ret1: The list of validators.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetValidatorsByHeight(ctx context.Context, height int64) ([]*bfttypes.Validator, error) {
	validatorSetResponse, err := c.chainClient.GetValidators(ctx, &height)
	if err != nil {
		return nil, err
	}
	return validatorSetResponse.Validators, nil
}

// BroadcastVote - Broadcast a vote to the Node's VotePool, it is used by Greenfield relayer and challengers by now.
//
// - ctx: Context variables for the current API call.
//
// - vote: Contains vote details.
//
// - ret: Return error when the request failed, otherwise return nil.
func (c *Client) BroadcastVote(ctx context.Context, vote votepool.Vote) error {
	return c.chainClient.BroadcastVote(ctx, vote)
}

// QueryVote - Query a vote from the Node's VotePool, it is used by Greenfield relayer and challengers by now.
//
// - ctx: Context variables for the current API call.
//
// - eventType: The type of vote to be queried.
//
// - eventHash: The hash bytes of vote
//
// - ret1: The vote result
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) QueryVote(ctx context.Context, eventType int, eventHash []byte) (*ctypes.ResultQueryVote, error) {
	return c.chainClient.QueryVote(ctx, eventType, eventHash)
}
