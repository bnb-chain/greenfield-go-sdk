package chain

import (
	"context"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/bnb-chain/greenfield/sdk/client"
	jsonrpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"sync"
)

type (
	GreenfieldClient       = client.GreenfieldClient
	TmClient               = client.TendermintClient
	GreenfieldClientOption = client.GreenfieldClientOption
)

var WithKeyManager = client.WithKeyManager
var WithGrpcDialOption = client.WithGrpcDialOption
var NewGreenfieldClient = client.NewGreenfieldClient

type TendermintClient struct {
	RpcClient     *TmClient
	JsonRpcClient *jsonrpcclient.Client // for interacting with votepool
}

func NewTendermintClient(provider string) *TendermintClient {
	rpcClient := client.NewTendermintClient(provider)
	jsonRpc, err := jsonrpcclient.New(provider)
	if err != nil {
		panic(err)
	}
	return &TendermintClient{
		&rpcClient,
		jsonRpc,
	}
}

type GnfdCompositeClient struct {
	*GreenfieldClient
	*TendermintClient
	Height int64
}

type GnfdCompositeClients struct {
	clients []*GnfdCompositeClient
}

func NewGnfdCompositClients(grpcAddrs, rpcAddrs []string, chainId string, opts ...GreenfieldClientOption) *GnfdCompositeClients {
	if len(grpcAddrs) == 0 || len(rpcAddrs) == 0 {
		panic(types.ErrorUrlNotProvided)
	}
	if len(grpcAddrs) != len(rpcAddrs) {
		panic(types.ErrorUrlsMismatch)
	}

	clients := make([]*GnfdCompositeClient, 0)

	for i := 0; i < len(grpcAddrs); i++ {
		tmClient := NewTendermintClient(rpcAddrs[i])
		clients = append(clients, &GnfdCompositeClient{
			NewGreenfieldClient(grpcAddrs[i], chainId, opts...),
			tmClient,
			0,
		})
	}
	return &GnfdCompositeClients{
		clients: clients,
	}
}

func (gc *GnfdCompositeClients) GetClient() *GnfdCompositeClient {
	wg := new(sync.WaitGroup)
	wg.Add(len(gc.clients))
	clientCh := make(chan *GnfdCompositeClient)
	waitCh := make(chan struct{})
	go func() {
		for _, c := range gc.clients {
			go getClientBlockHeight(clientCh, wg, c)
		}
		wg.Wait()
		close(waitCh)
	}()
	var maxHeight int64
	maxHeightClient := gc.clients[0]
	for {
		select {
		case c := <-clientCh:
			if c.Height > maxHeight {
				maxHeight = c.Height
				maxHeightClient = c
			}
		case <-waitCh:
			return maxHeightClient
		}
	}
}

func getClientBlockHeight(clientChan chan *GnfdCompositeClient, wg *sync.WaitGroup, client *GnfdCompositeClient) {
	defer wg.Done()
	status, err := client.RpcClient.TmClient.Status(context.Background())
	if err != nil {
		return
	}
	client.Height = status.SyncInfo.LatestBlockHeight
	clientChan <- client
}
