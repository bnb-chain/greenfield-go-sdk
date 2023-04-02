package client

import (
	"context"
	"errors"

	types2 "github.com/bnb-chain/greenfield-go-sdk/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Account interface {
	BuyQuotaForBucket(ctx context.Context, bucketName string, targetQuota uint64, opt types2.BuyQuotaOption) (string, error)
	GetQuotaPrice(ctx context.Context, SPAddress sdk.AccAddress) (uint64, error)
}

// BuyQuotaForBucket buy the target quota of the specific bucket
// targetQuota indicates the target quota to set for the bucket
func (c *client) BuyQuotaForBucket(ctx context.Context, bucketName string, targetQuota uint64, opt types2.BuyQuotaOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}

	paymentAddr, err := sdk.AccAddressFromHexUnsafe(bucketInfo.PaymentAddress)
	if err != nil {
		return "", err
	}
	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, &targetQuota, paymentAddr, bucketInfo.Visibility)

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// GetQuotaPrice return the quota price of the SP
func (c *client) GetQuotaPrice(ctx context.Context, SPAddress sdk.AccAddress) (uint64, error) {
	resp, err := c.chainClient.QueryGetSpStoragePriceByTime(ctx, &spTypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    SPAddress.String(),
		Timestamp: 0,
	})
	if err != nil {
		return 0, err
	}
	return resp.SpStoragePrice.ReadPrice.BigInt().Uint64(), nil
}
