package client

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type SP interface {
	// ListSP return the storage provider info on chain
	// isInService indicates if only display the sp with STATUS_IN_SERVICE status
	ListSP(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error)
	// GetSPInfo return the sp info the sp chain address
	GetSPInfo(ctx context.Context, spAddr sdk.AccAddress) (*spTypes.StorageProvider, error)
	// GetSpAddrFromEndpoint return the chain addr according to the SP endpoint
	GetSpAddrFromEndpoint(ctx context.Context, spEndpoint string) (sdk.AccAddress, error)
	// GetStoragePrice returns the storage price for a particular storage provider, including update time, read price, store price and .etc.
	GetStoragePrice(ctx context.Context, SPAddr string) (*spTypes.SpStoragePrice, error)
	// GrantDepositForStorageProvider submit a grant transaction to allow gov module account to deduct the specified number of tokens
	GrantDepositForStorageProvider(ctx context.Context, spAddr string, depositAmount math.Int, opts GrantDepositForStorageProviderOptions) (string, error)
	// CreateStorageProvider submits a proposal to create a storage provider to the greenfield blockchain, and it returns a proposal ID
	CreateStorageProvider(ctx context.Context, fundingAddr, sealAddr, approvalAddr, gcAddr string, endpoint string, depositAmount math.Int, description spTypes.Description, opts CreateStorageProviderOptions) (uint64, string, error)
}

func (c *client) GetStoragePrice(ctx context.Context, spAddr string) (*spTypes.SpStoragePrice, error) {
	spAcc, err := sdk.AccAddressFromHexUnsafe(spAddr)
	if err != nil {
		return nil, err
	}
	resp, err := c.chainClient.QueryGetSpStoragePriceByTime(ctx, &spTypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAcc.String(),
		Timestamp: 0,
	})
	if err != nil {
		return nil, err
	}
	return &resp.SpStoragePrice, nil
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

// GetSPInfo return the sp info the sp chain address
func (c *client) GetSPInfo(ctx context.Context, SPAddr sdk.AccAddress) (*spTypes.StorageProvider, error) {
	request := &spTypes.QueryStorageProviderRequest{
		SpAddress: SPAddr.String(),
	}

	gnfdRep, err := c.chainClient.StorageProvider(ctx, request)
	if err != nil {
		return nil, err
	}

	return gnfdRep.StorageProvider, nil
}

func (c *client) getSPUrlInfo() (map[string]*url.URL, error) {
	ctx := context.Background()
	spInfo := make(map[string]*url.URL, 0)
	request := &spTypes.QueryStorageProvidersRequest{}
	gnfdRep, err := c.chainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, err
	}
	spList := gnfdRep.GetSps()
	if len(spList) == 0 {
		return nil, errors.New("no SP found on chain")
	}
	
	for _, info := range spList {
		endpoint := info.Endpoint
		urlInfo, urlErr := utils.GetEndpointURL(endpoint, c.secure)
		if urlErr != nil {
			return nil, urlErr
		}
		spInfo[info.GetOperator().String()] = urlInfo
	}

	return spInfo, nil
}

type CreateStorageProviderOptions struct {
	ReadPrice             sdk.Dec
	FreeReadQuota         uint64
	StorePrice            sdk.Dec
	ProposalDepositAmount math.Int // wei BNB
	ProposalMetaData      string
	TxOption              gnfdSdkTypes.TxOption
}

// CreateStorageProvider will submit a CreateStorageProvider proposal and return proposalID, TxHash and err if it has.
func (c *client) CreateStorageProvider(ctx context.Context, fundingAddr, sealAddr, approvalAddr, gcAddr string, endpoint string, depositAmount math.Int, description spTypes.Description, opts CreateStorageProviderOptions) (uint64, string, error) {
	defaultAccount := c.MustGetDefaultAccount()
	govModuleAddress, err := c.GetModuleAccountByName(ctx, govTypes.ModuleName)
	if err != nil {
		return 0, "", err
	}
	if opts.ProposalDepositAmount.IsNil() {
		opts.ProposalDepositAmount = math.NewIntWithDecimal(1, gnfdSdkTypes.DecimalBNB)
	}
	if opts.ReadPrice.IsNil() {
		opts.ReadPrice = sdk.NewDec(1)
	}
	if opts.StorePrice.IsNil() {
		opts.StorePrice = sdk.NewDec(1)
	}

	fundingAcc, err := sdk.AccAddressFromHexUnsafe(fundingAddr)
	if err != nil {
		return 0, "", err
	}
	sealAcc, err := sdk.AccAddressFromHexUnsafe(sealAddr)
	if err != nil {
		return 0, "", err
	}
	approvalAcc, err := sdk.AccAddressFromHexUnsafe(approvalAddr)
	if err != nil {
		return 0, "", err
	}
	gcAcc, err := sdk.AccAddressFromHexUnsafe(gcAddr)
	if err != nil {
		return 0, "", err
	}
	msgCreateStorageProvider, err := spTypes.NewMsgCreateStorageProvider(
		govModuleAddress.GetAddress(),
		defaultAccount.GetAddress(),
		fundingAcc, sealAcc, approvalAcc, gcAcc, description,
		endpoint,
		sdk.NewCoin(gnfdSdkTypes.Denom, depositAmount),
		opts.ReadPrice,
		opts.FreeReadQuota,
		opts.StorePrice,
	)
	if err != nil {
		return 0, "", err
	}
	err = msgCreateStorageProvider.ValidateBasic()
	if err != nil {
		return 0, "", err
	}

	return c.SubmitProposal(ctx, []sdk.Msg{msgCreateStorageProvider}, opts.ProposalDepositAmount, SubmitProposalOptions{Metadata: opts.ProposalMetaData, TxOption: opts.TxOption})
}

type GrantDepositForStorageProviderOptions struct {
	expiration *time.Time
	TxOption   gnfdSdkTypes.TxOption
}

func (c *client) GrantDepositForStorageProvider(ctx context.Context, spAddr string, depositAmount math.Int, opts GrantDepositForStorageProviderOptions) (string, error) {
	granter := c.MustGetDefaultAccount()
	govModuleAddress, err := c.GetModuleAccountByName(ctx, govTypes.ModuleName)
	if err != nil {
		return "", err
	}
	spAcc, err := sdk.AccAddressFromHexUnsafe(spAddr)
	if err != nil {
		return "", err
	}
	coin := sdk.NewCoin(gnfdSdkTypes.Denom, depositAmount)
	authorization := spTypes.NewDepositAuthorization(spAcc, &coin)

	if opts.expiration == nil {
		expiration := time.Now().Add(24 * time.Hour)
		opts.expiration = &expiration
	}
	msgGrant, err := authz.NewMsgGrant(granter.GetAddress(), govModuleAddress.GetAddress(), authorization, opts.expiration)
	if err != nil {
		return "", err
	}
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgGrant}, &opts.TxOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}
