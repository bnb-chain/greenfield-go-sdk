package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	gnfdTypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/s3util"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/types"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
)

// CreateBucket get approval of creating bucket and send createBucket txn to greenfield chain
func (c *Client) CreateBucket(ctx context.Context, bucketName string, primaryAddr sdk.AccAddress, opts client.CreateBucketOptions) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	var visibility storageTypes.VisibilityType
	if opts.Visibility == storageTypes.VISIBILITY_TYPE_UNSPECIFIED {
		visibility = storageTypes.VISIBILITY_TYPE_PRIVATE // set default visibility type
	} else {
		visibility = opts.Visibility
	}

	createBucketMsg := storageTypes.NewMsgCreateBucket(km.GetAddr(), bucketName,
		visibility, primaryAddr, opts.PaymentAddress, 0, nil, opts.ChargedQuota)

	err = createBucketMsg.ValidateBasic()
	if err != nil {
		return "", err
	}
	signedMsg, err := c.GetCreateBucketApproval(ctx, createBucketMsg, NewAuthInfo(false, ""))
	if err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{signedMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// DeleteBucket send DeleteBucket txn to greenfield chain and return txn hash
func (c *Client) DeleteBucket(bucketName string, opt client.DeleteBucketOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	delBucketMsg := storageTypes.NewMsgDeleteBucket(km.GetAddr(), bucketName)

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{delBucketMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

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

// UpdateBucketVisibility update the visibilityType of bucket
func (c *Client) UpdateBucketVisibility(ctx context.Context, bucketName string,
	visibility storageTypes.VisibilityType, opt client.UpdateVisibilityOption) (string, error) {
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

	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, &bucketInfo.ChargedReadQuota, paymentAddr, visibility)

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
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

	endpoint, err := c.getSPUrlFromBucket(bucketName)
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
func (c *Client) ListBucketReadRecord(ctx context.Context, bucketName string,
	maxRecords int, opt client.ListReadRecordOption, authInfo AuthInfo) (client.QuotaRecordInfo, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return client.QuotaRecordInfo{}, err
	}
	timeNow := time.Now()
	timeToday := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())
	if opt.StartTimeStamp < 0 {
		return client.QuotaRecordInfo{}, errors.New("start timestamp  less than 0")
	}
	var startTimeStamp int64
	if opt.StartTimeStamp == 0 {
		// the timestamp of the first day of this month
		startTimeStamp = timeToday.AddDate(0, 0, -timeToday.Day()+1).UnixMicro()
	} else {
		startTimeStamp = opt.StartTimeStamp
	}
	// the timestamp of the last day of this month
	timeMonthEnd := timeToday.AddDate(0, 1, -timeToday.Day()+1).UnixMicro()

	if timeMonthEnd < startTimeStamp {
		return client.QuotaRecordInfo{}, errors.New("start timestamp larger than the end timestamp of this month")
	}

	params := url.Values{}
	params.Set("list-read-record", "")
	params.Set("max-records", strconv.Itoa(maxRecords))
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

	endpoint, err := c.getSPUrlFromBucket(bucketName)
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

// HeadBucket query the bucketInfo on chain, return the bucket info if exists
// return err info if bucket not exist
func (c *Client) HeadBucket(ctx context.Context, bucketName string) (*storageTypes.BucketInfo, error) {
	queryHeadBucketRequest := storageTypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := c.chainClient.HeadBucket(ctx, &queryHeadBucketRequest)
	if err != nil {
		return nil, err
	}

	return queryHeadBucketResponse.BucketInfo, nil
}

// HeadBucketByID query the bucketInfo on chain by bucketId, return the bucket info if exists
// return err info if bucket not exist
func (c *Client) HeadBucketByID(ctx context.Context, bucketID string) (*storageTypes.BucketInfo, error) {
	headBucketRequest := &storageTypes.QueryHeadBucketByIdRequest{
		BucketId: bucketID,
	}

	headBucketResponse, err := c.chainClient.HeadBucketById(ctx, headBucketRequest)
	if err != nil {
		return nil, err
	}

	return headBucketResponse.BucketInfo, nil
}

// PutBucketPolicy apply bucket policy to the principal, return the txn hash
func (c *Client) PutBucketPolicy(bucketName string, principalStr client.Principal,
	statements []*permTypes.Statement, opt client.PutPolicyOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	resource := gnfdTypes.NewBucketGRN(bucketName)
	principal := &permTypes.Principal{}
	if err = principal.Unmarshal([]byte(principalStr)); err != nil {
		return "", err
	}

	putPolicyMsg := storageTypes.NewMsgPutPolicy(km.GetAddr(), resource.String(),
		principal, statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(putPolicyMsg, opt.TxOpts)
}

// DeleteBucketPolicy delete the bucket policy of the principal
func (c *Client) DeleteBucketPolicy(bucketName string, principalAddr sdk.AccAddress, opt client.DeletePolicyOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	resource := gnfdTypes.NewBucketGRN(bucketName).String()
	principal := permTypes.NewPrincipalWithAccount(principalAddr)

	return c.sendDelPolicyTxn(km.GetAddr(), resource, principal, opt.TxOpts)
}

// IsBucketPermissionAllowed check if the permission of bucket is allowed to the user
func (c *Client) IsBucketPermissionAllowed(ctx context.Context, user sdk.AccAddress,
	bucketName string, action permTypes.ActionType) (permTypes.Effect, error) {
	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   user.String(),
		BucketName: bucketName,
		ActionType: action,
	}

	verifyResp, err := c.chainClient.VerifyPermission(ctx, &verifyReq)
	if err != nil {
		return permTypes.EFFECT_DENY, err
	}

	return verifyResp.Effect, nil
}

// GetBucketPolicy get the bucket policy info of the user specified by principalAddr
func (c *Client) GetBucketPolicy(ctx context.Context, bucketName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	queryPolicy := storageTypes.QueryPolicyForAccountRequest{
		Resource:         resource,
		PrincipalAddress: principalAddr.String(),
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// ListBuckets list buckets for the owner
func (c *Client) ListBuckets(ctx context.Context, userInfo client.UserInfo, authInfo AuthInfo) (client.ListBucketsResponse, error) {
	if userInfo.Address == "" {
		return client.ListBucketsResponse{}, errors.New("fail to get user address")
	}

	reqMeta := requestMeta{
		contentSHA256: types.EmptyStringSHA256,
		userInfo: client.UserInfo{
			Address: userInfo.Address,
		},
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlFromAddr(userInfo.Address)
	if err != nil {
		return client.ListBucketsResponse{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return client.ListBucketsResponse{}, err
	}
	defer utils.CloseResponse(resp)

	listBucketsResult := client.ListBucketsResponse{}
	//unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return client.ListBucketsResponse{}, err
	}

	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &listBucketsResult)

	//TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil && listBucketsResult.Buckets == nil {
		return client.ListBucketsResponse{}, err
	}

	return listBucketsResult, nil
}
