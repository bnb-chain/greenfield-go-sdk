package client

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Basic interface {
	Status()
	BroadcastRawTx(txBytes []byte)
	SimulateRaeTx()
	WaitForBlockHeight(ctx context.Context, height int64) error
	WaitForTx(ctx context.Context, hash string) error
}

func (c *client) GetSPAddrInfo() (map[string]*url.URL, error) {
	ctx := context.Background()
	spInfo := make(map[string]*url.URL, 0)
	request := &spTypes.QueryStorageProvidersRequest{}
	gnfdRep, err := c.chainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, err
	}
	spList := gnfdRep.GetSps()
	for _, info := range spList {
		endpoint := info.Endpoint
		urlInfo, urlErr := utils.GetEndpointURL(endpoint, c.Secure)
		if urlErr != nil {
			return nil, urlErr
		}
		spInfo[info.GetOperator().String()] = urlInfo
	}

	return spInfo, nil
}

// ListSP return the storage provider info on chain
// isInService indicates if only display the sp with STATUS_IN_SERVICE status
func (c *client) ListSP(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error) {
	request := &spTypes.QueryStorageProvidersRequest{}
	gnfdRep, err := c.chainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, err
	}

	spList := gnfdRep.GetSps()
	spInfoList := make([]spTypes.StorageProvider, 0)
	for _, info := range spList {
		if isInService && info.Status != spTypes.STATUS_IN_SERVICE {
			continue
		}
		spInfoList = append(spInfoList, *info)
	}

	return spInfoList, nil
}

// GetSpAddrFromEndpoint return the chain addr according to the SP endpoint
func (c *client) GetSpAddrFromEndpoint(ctx context.Context, spEndpoint string) (sdk.AccAddress, error) {
	spList, err := c.ListSP(ctx, false)
	if err != nil {
		return nil, err
	}

	if strings.Contains(spEndpoint, "http") {
		s := strings.Split(spEndpoint, "//")
		spEndpoint = s[1]
	}

	for _, spInfo := range spList {
		endpoint := spInfo.GetEndpoint()
		if strings.Contains(endpoint, "http") {
			s := strings.Split(endpoint, "//")
			endpoint = s[1]
		}
		if endpoint == spEndpoint {
			addr := spInfo.GetOperatorAddress()
			if addr == "" {
				return nil, errors.New("fail to get addr")
			}
			return sdk.MustAccAddressFromHex(spInfo.GetOperatorAddress()), nil
		}
	}
	return nil, errors.New("fail to get addr")
}
