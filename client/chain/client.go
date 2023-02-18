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
	JsonRpcClient *jsonrpcclient.Client // need it for interacting with votepool
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

type GnfdClient struct {
	*GreenfieldClient
	*TendermintClient
	Provider string
	Height   int64
}

type GreenfieldClients struct {
	mutex   sync.RWMutex
	clients []*GnfdClient
}

func NewGreenfieldClients(grpcAddrs, rpcAddrs []string, chainId string, opts ...GreenfieldClientOption) *GreenfieldClients {
	if len(grpcAddrs) == 0 || len(rpcAddrs) == 0 {
		panic(types.ErrorUrlNotProvided)
	}
	if len(grpcAddrs) != len(rpcAddrs) {
		panic(types.ErrorUrlsMismatch)
	}

	clients := make([]*GnfdClient, 0)

	for i := 0; i < len(grpcAddrs); i++ {
		tmClient := NewTendermintClient(rpcAddrs[i])
		clients = append(clients, &GnfdClient{
			NewGreenfieldClient(grpcAddrs[i], chainId, opts...),
			tmClient,
			rpcAddrs[i],
			0,
		})

	}
	return &GreenfieldClients{
		clients: clients,
	}
}

func (gc *GreenfieldClients) GetClient() (*GnfdClient, error) {
	wg := new(sync.WaitGroup)
	wg.Add(len(gc.clients))
	clientCh := make(chan *GnfdClient)
	errCh := make(chan error)
	waitCh := make(chan struct{})
	go func() {
		for _, c := range gc.clients {
			go calClientHeight(clientCh, errCh, wg, c)
		}
		wg.Wait()
		close(waitCh)
	}()
	var maxHeight int64
	maxHeightClient := gc.clients[0]
	for {
		select {
		case err := <-errCh:
			return nil, err
		case c := <-clientCh:
			if c.Height > maxHeight {
				gc.mutex.Lock()
				maxHeight = c.Height
				maxHeightClient = c
				gc.mutex.Unlock()
			}
		case <-waitCh:
			return maxHeightClient, nil
		}
	}
}

func calClientHeight(clientChan chan *GnfdClient, errChan chan error, wg *sync.WaitGroup, client *GnfdClient) {
	defer wg.Done()
	status, err := client.RpcClient.TmClient.Status(context.Background())
	if err != nil {
		errChan <- err
		return
	}
	latestHeight := status.SyncInfo.LatestBlockHeight
	client.Height = latestHeight
	clientChan <- client
}
