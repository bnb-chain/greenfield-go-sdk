package client

import (
	"github.com/tendermint/tendermint/rpc/client"
	chttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

type RPCClient struct {
	RpcClient client.Client
}

func httpClient(addr string) *chttp.HTTP {
	httpCli, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		panic(err)
	}
	cli, err := chttp.NewWithClient(addr, "/websocket", httpCli)
	if err != nil {
		panic(err)
	}
	return cli
}

func NewRPCClient(addr string) RPCClient {
	return RPCClient{
		RpcClient: httpClient(addr),
	}
}
