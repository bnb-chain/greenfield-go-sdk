package api

import (
	"net/http"
	"net/url"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	account2 "github.com/bnb-chain/greenfield-go-sdk/pkg/account"
	sdkclient "github.com/bnb-chain/greenfield/sdk/client"
)

type Client struct {
	// chainClients
	chainClient *sdkclient.GreenfieldClient
	// tendermintClient
	tendermintClient *sdkclient.TendermintClient
	// httpClient
	httpClient *http.Client
	// account
	account *account2.Account
	// spEndpoints
	spEndpoints map[string]*url.URL
}

// New - instantiate greenfield chain with options
func New(chainID string, grpcAddress, rpcAddress string, opts *client.Options) (*Client, error) {
	tc := sdkclient.NewTendermintClient(rpcAddress)
	cc := sdkclient.NewGreenfieldClient(grpcAddress, chainID,
		sdkclient.WithAccount(account),
		sdkclient.WithGrpcDialOption(opts.grpcDialOption))

	return &Client{
		chainClient:      cc,
		tendermintClient: &tc,
		httpClient:       &http.Client{},
	}, nil
}

func (c Client) bucketRouter(bucketName string) (*url.URL, error) {
	// 1. headBucket()
	// 2. get from spEndpoints
	// 3. get from chain
}
