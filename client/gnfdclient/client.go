package gnfdclient

import (
	chain "github.com/bnb-chain/greenfield/sdk/client"

	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/keys"
)

// IntegratedClient integrate the chainClient and SPClient
type IntegratedClient struct {
	ChainClient *chain.GreenfieldClient
	SPClient    *sp.SPClient
}

type ChainClientInfo struct {
	RpcAddr  string
	GrpcAddr string
}

type SPClientInfo struct {
	Endpoint string
	opt      *sp.Option
}

func NewIntegratedClient(chainInfo ChainClientInfo, spInfo SPClientInfo) (*IntegratedClient, error) {
	var err error
	spClient := &sp.SPClient{}
	if spInfo.Endpoint != "" {
		if spInfo.opt == nil {
			spClient, err = sp.NewSpClient(spInfo.Endpoint, &sp.Option{})
			if err != nil {
				return nil, err
			}
		} else {
			spClient, err = sp.NewSpClient(spInfo.Endpoint, spInfo.opt)
			if err != nil {
				return nil, err
			}
		}
	}

	chainClient := &chain.GreenfieldClient{}
	if chainInfo.RpcAddr != "" && chainInfo.GrpcAddr != "" {
		chainClient = chain.NewGreenfieldClient(chainInfo.RpcAddr, chainInfo.GrpcAddr)
	}

	return &IntegratedClient{
		ChainClient: chainClient,
		SPClient:    spClient,
	}, nil
}

func NewIntegratedWithKeyManager(chainInfo ChainClientInfo, spInfo SPClientInfo,
	keyManager keys.KeyManager) (*IntegratedClient, error) {
	GreenfieldClient, err := NewIntegratedClient(chainInfo, spInfo)
	if err != nil {
		return nil, err
	}

	GreenfieldClient.ChainClient.SetKeyManager(keyManager)
	err = GreenfieldClient.SPClient.SetKeyManager(keyManager)
	if err != nil {
		return nil, err
	}
	return GreenfieldClient, nil
}
