package api

import (
	"context"
	"encoding/xml"
	"errors"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/types"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield/types/s3util"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BuyQuotaForBucket buy the target quota of the specific bucket
// targetQuota indicates the target quota to set for the bucket
func (c *Client) BuyQuotaForBucket(ctx context.Context, bucketName string, targetQuota uint64, opt client.BuyQuotaOption) (string, error) {
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
func (c *Client) GetQuotaPrice(ctx context.Context, SPAddress sdk.AccAddress) (uint64, error) {
	resp, err := c.chainClient.QueryGetSpStoragePriceByTime(ctx, &spTypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    SPAddress.String(),
		Timestamp: 0,
	})
	if err != nil {
		return 0, err
	}
	return resp.SpStoragePrice.ReadPrice.BigInt().Uint64(), nil
}

// GetBucketReadQuota return quota info of bucket of current month, include chain quota, free quota and consumed quota
// GetBucketReadQuota returns the bucket quota info of this month
func (c *Client) GetBucketReadQuota(ctx context.Context, bucketName string, authInfo AuthInfo) (client.QuotaInfo, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return client.QuotaInfo{}, err
	}

	year, month, _ := time.Now().Date()
	var date string
	if int(month) < 10 {
		date = strconv.Itoa(year) + "-" + "0" + strconv.Itoa(int(month))
	} else {
		date = strconv.Itoa(year) + "-" + strconv.Itoa(int(month))
	}

	params := url.Values{}
	params.Add("read-quota", "")
	params.Add("year-month", date)

	reqMeta := requestMeta{
		urlValues:     params,
		bucketName:    bucketName,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		return client.QuotaInfo{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return client.QuotaInfo{}, err
	}
	defer utils.CloseResponse(resp)

	QuotaResult := client.QuotaInfo{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&QuotaResult)
	if err != nil {
		return client.QuotaInfo{}, err
	}

	return QuotaResult, nil
}

// ListBucketReadRecord returns the read record of this month, the return items should be no more than maxRecords
// ListReadRecordOption indicates the start timestamp of return read records
func (c *Client) ListBucketReadRecord(ctx context.Context, bucketName string, opts client.ListReadRecordOptions, authInfo AuthInfo) (client.QuotaRecordInfo, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return client.QuotaRecordInfo{}, err
	}
	timeNow := time.Now()
	timeToday := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())
	if opts.StartTimeStamp < 0 {
		return client.QuotaRecordInfo{}, errors.New("start timestamp  less than 0")
	}
	var startTimeStamp int64
	if opts.StartTimeStamp == 0 {
		// the timestamp of the first day of this month
		startTimeStamp = timeToday.AddDate(0, 0, -timeToday.Day()+1).UnixMicro()
	} else {
		startTimeStamp = opts.StartTimeStamp
	}
	// the timestamp of the last day of this month
	timeMonthEnd := timeToday.AddDate(0, 1, -timeToday.Day()+1).UnixMicro()

	if timeMonthEnd < startTimeStamp {
		return client.QuotaRecordInfo{}, errors.New("start timestamp larger than the end timestamp of this month")
	}

	params := url.Values{}
	params.Set("list-read-record", "")
	if opts.MaxRecords > 0 {
		params.Set("max-records", strconv.Itoa(opts.MaxRecords))
	} else {
		params.Set("max-records", strconv.Itoa(math.MaxUint32))
	}

	params.Set("start-timestamp", strconv.FormatInt(startTimeStamp, 10))
	params.Set("end-timestamp", strconv.FormatInt(timeMonthEnd, 10))

	reqMeta := requestMeta{
		urlValues:     params,
		bucketName:    bucketName,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		return client.QuotaRecordInfo{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return client.QuotaRecordInfo{}, err
	}
	defer utils.CloseResponse(resp)

	QuotaRecords := client.QuotaRecordInfo{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&QuotaRecords)
	if err != nil {
		return client.QuotaRecordInfo{}, err
	}

	return QuotaRecords, nil
}
