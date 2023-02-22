package gnfdclient

import (
	chain "github.com/bnb-chain/greenfield/sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/keys"
)

// IntegratedClient integrate the chainClient and SPClient
type IntegratedClient struct {
	ChainClient *chain.GreenfieldClient
	SPClient    *sp.SPClient
	sender      sdk.AccAddress
}

type ChainClientInfo struct {
	RpcAddr string
	ChainId string
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
	if chainInfo.RpcAddr != "" && chainInfo.ChainId != "" {
		chainClient = chain.NewGreenfieldClient(chainInfo.RpcAddr, chainInfo.ChainId)
	}

	return &IntegratedClient{
		ChainClient: chainClient,
		SPClient:    spClient,
		sender:      nil,
	}, nil
}

// NewIntegratedWithKeyManager return client with keyManager and set the sender
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

	GreenfieldClient.sender = keyManager.GetAddr()
	return GreenfieldClient, nil
}
