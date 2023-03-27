package gnfdclient

import (
	client "github.com/bnb-chain/greenfield-go-sdk/client/chain"
	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	chain "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
)

type (
	GreenfieldClient       = client.GreenfieldClient
	GreenfieldClientOption = client.GreenfieldClientOption
)

// GnfdClient integrates the chainClient and SPClient
type GnfdClient struct {
	ChainClient *chain.GreenfieldClient
	SPClient    *sp.SPClient
}

// NewGnfdClient returns GnfdClient from chain info and sp info
// km pass a keyManager for SP client to sign http request
func NewGnfdClient(grpcAddrs string, chainId string, spEndpoint string, km keys.KeyManager, secure bool, gnfdopts ...GreenfieldClientOption) (*GnfdClient, error) {
	chainClient := chain.NewGreenfieldClient(grpcAddrs, chainId, gnfdopts...)

	spClient, err := sp.NewSpClient(spEndpoint, sp.WithKeyManager(km), sp.WithSecure(secure))
	if err != nil {
		return nil, err
	}

	return &GnfdClient{
		ChainClient: chainClient,
		SPClient:    spClient,
	}, nil
}
