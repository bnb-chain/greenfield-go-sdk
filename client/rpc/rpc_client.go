package client

import (
	"context"
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/rpc/client"
	rlient "github.com/tendermint/tendermint/rpc/client"
	chttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

type RpcClient struct {
	rpcClient  client.Client
	keyManager keys.KeyManager
}

func HttpClient(addr string) *chttp.HTTP {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		panic(err)
	}
	cli, err := chttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		panic(err)
	}
	return cli
}

func NewRPCClient(addr string) RpcClient {
	return RpcClient{
		rpcClient: HttpClient(addr),
	}
}

func NewRPCClientWithKeyManager(addr string, keyManager keys.KeyManager) RpcClient {
	return RpcClient{
		rpcClient:  HttpClient(addr),
		keyManager: keyManager,
	}
}

func (c *RpcClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return c.rpcClient.ABCIInfo(context.Background())
}

func (c *RpcClient) Status() (*ctypes.ResultStatus, error) {
	return c.rpcClient.Status(context.Background())
}

func (c *RpcClient) ABCIQueryWithOptions(path string, data bytes.HexBytes,
	opts rlient.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	return c.rpcClient.ABCIQueryWithOptions(context.Background(), path, data, opts)
}
