package client

import (
	"context"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/prysmaticlabs/prysm/crypto/bls"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type SP interface {
	// ListStorageProviders return the storage provider info on chain
	// isInService indicates if only display the sp with STATUS_IN_SERVICE status
	ListStorageProviders(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error)
	// GetStorageProviderInfo return the sp info with the sp chain address
	GetStorageProviderInfo(ctx context.Context, SPAddr sdk.AccAddress) (*spTypes.StorageProvider, error)
	// GetStoragePrice returns the storage price for a particular storage provider, including update time, read price, store price and .etc.
	GetStoragePrice(ctx context.Context, SPAddr string) (*spTypes.SpStoragePrice, error)
	// GetSecondarySpStorePrice returns the secondary storage price, including update time and store price
	GetSecondarySpStorePrice(ctx context.Context) (*spTypes.SecondarySpStorePrice, error)
	// GrantDepositForStorageProvider submit a grant transaction to allow gov module account to deduct the specified number of tokens
	GrantDepositForStorageProvider(ctx context.Context, spAddr string, depositAmount math.Int, opts types.GrantDepositForStorageProviderOptions) (string, error)
	// CreateStorageProvider submits a proposal to create a storage provider to the greenfield blockchain, and it returns a proposal ID
	CreateStorageProvider(ctx context.Context, fundingAddr, sealAddr, approvalAddr, gcAddr, blsPubKey, blsProof, endpoint string, depositAmount math.Int, description spTypes.Description, opts types.CreateStorageProviderOptions) (uint64, string, error)
	// UpdateSpStoragePrice updates the read price, storage price and free read quota for a particular storage provider
	UpdateSpStoragePrice(ctx context.Context, spAddr string, readPrice, storePrice sdk.Dec, freeReadQuota uint64, TxOption gnfdSdkTypes.TxOption) (string, error)
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

func (c *client) GetSecondarySpStorePrice(ctx context.Context) (*spTypes.SecondarySpStorePrice, error) {
	resp, err := c.chainClient.QueryGetSecondarySpStorePriceByTime(ctx, &spTypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	if err != nil {
		return nil, err
	}
	return &resp.SecondarySpStorePrice, nil
}

// ListStorageProviders return the storage provider info on chain
// isInService indicates if only display the sp with STATUS_IN_SERVICE status
func (c *client) ListStorageProviders(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error) {
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

// GetStorageProviderInfo return the sp info with the sp chain address
func (c *client) GetStorageProviderInfo(ctx context.Context, SPAddr sdk.AccAddress) (*spTypes.StorageProvider, error) {
	request := &spTypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: SPAddr.String(),
	}

	gnfdRep, err := c.chainClient.StorageProviderByOperatorAddress(ctx, request)
	if err != nil {
		return nil, err
	}

	return gnfdRep.StorageProvider, nil
}

func (c *client) getSPUrlList() (map[uint32]*url.URL, map[string]*url.URL, error) {
	ctx := context.Background()
	spIDInfo := make(map[uint32]*url.URL, 0)
	spAddressInfo := make(map[string]*url.URL, 0)
	request := &spTypes.QueryStorageProvidersRequest{}
	gnfdRep, err := c.chainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	spList := gnfdRep.GetSps()
	if len(spList) == 0 {
		return nil, nil, errors.New("no SP found on chain")
	}

	for _, info := range spList {
		var useHttps bool
		if strings.Contains(info.Endpoint, "https") {
			useHttps = true
		} else {
			useHttps = c.secure
		}

		urlInfo, urlErr := utils.GetEndpointURL(info.Endpoint, useHttps)
		if urlErr != nil {
			return nil, nil, urlErr
		}
		spIDInfo[info.GetId()] = urlInfo
		spAddressInfo[info.GetOperatorAddress()] = urlInfo
	}

	return spIDInfo, spAddressInfo, nil
}

// CreateStorageProvider will submit a CreateStorageProvider proposal and return proposalID, TxHash and err if it has.
func (c *client) CreateStorageProvider(ctx context.Context, fundingAddr, sealAddr, approvalAddr, gcAddr, blsPubKey, blsProof, endpoint string, depositAmount math.Int, description spTypes.Description, opts types.CreateStorageProviderOptions) (uint64, string, error) {
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
	blsPubKeyBz, err := hex.DecodeString(blsPubKey)
	if err != nil {
		return 0, "", err
	}
	_, err = bls.PublicKeyFromBytes(blsPubKeyBz)
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
		blsPubKey,
		blsProof,
	)
	if err != nil {
		return 0, "", err
	}
	err = msgCreateStorageProvider.ValidateBasic()
	if err != nil {
		return 0, "", err
	}

	return c.SubmitProposal(ctx, []sdk.Msg{msgCreateStorageProvider}, opts.ProposalDepositAmount, opts.ProposalTitle, opts.ProposalSummary, types.SubmitProposalOptions{Metadata: opts.ProposalMetaData, TxOption: opts.TxOption})
}

func (c *client) GrantDepositForStorageProvider(ctx context.Context, spAddr string, depositAmount math.Int, opts types.GrantDepositForStorageProviderOptions) (string, error) {
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

	if opts.Expiration == nil {
		expiration := time.Now().Add(24 * time.Hour)
		opts.Expiration = &expiration
	}
	msgGrant, err := authz.NewMsgGrant(granter.GetAddress(), govModuleAddress.GetAddress(), authorization, opts.Expiration)
	if err != nil {
		return "", err
	}
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgGrant}, &opts.TxOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) UpdateSpStoragePrice(ctx context.Context, spAddr string, readPrice, storePrice sdk.Dec, freeReadQuota uint64, TxOption gnfdSdkTypes.TxOption) (string, error) {
	spAcc, err := sdk.AccAddressFromHexUnsafe(spAddr)
	if err != nil {
		return "", err
	}
	msgUpdateStoragePrice := &spTypes.MsgUpdateSpStoragePrice{
		SpAddress:     spAcc.String(),
		ReadPrice:     readPrice,
		StorePrice:    storePrice,
		FreeReadQuota: freeReadQuota,
	}
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgUpdateStoragePrice}, &TxOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}
