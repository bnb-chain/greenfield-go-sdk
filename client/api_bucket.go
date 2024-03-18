package client

import (
	"context"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/rs/zerolog/log"

	gnfdsdk "github.com/bnb-chain/greenfield/sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/s3util"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

// IBucketClient interface defines functions related to bucket.
// The concept of "bucket" is the same as the concept of a bucket in AWS S3 storage.
type IBucketClient interface {
	GetCreateBucketApproval(ctx context.Context, createBucketMsg *storageTypes.MsgCreateBucket) (*storageTypes.MsgCreateBucket, error)
	CreateBucket(ctx context.Context, bucketName string, primaryAddr string, opts types.CreateBucketOptions) (string, error)
	DeleteBucket(ctx context.Context, bucketName string, opt types.DeleteBucketOption) (string, error)
	UpdateBucketVisibility(ctx context.Context, bucketName string, visibility storageTypes.VisibilityType, opt types.UpdateVisibilityOption) (string, error)
	UpdateBucketInfo(ctx context.Context, bucketName string, opts types.UpdateBucketOptions) (string, error)
	UpdateBucketPaymentAddr(ctx context.Context, bucketName string, paymentAddr sdk.AccAddress, opt types.UpdatePaymentOption) (string, error)
	ToggleSPAsDelegatedAgent(ctx context.Context, bucketName string, opt types.UpdateBucketOptions) (string, error)
	HeadBucket(ctx context.Context, bucketName string) (*storageTypes.BucketInfo, error)
	HeadBucketByID(ctx context.Context, bucketID string) (*storageTypes.BucketInfo, error)
	PutBucketPolicy(ctx context.Context, bucketName string, principal types.Principal, statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	DeleteBucketPolicy(ctx context.Context, bucketName string, principal types.Principal, opt types.DeletePolicyOption) (string, error)
	GetBucketPolicy(ctx context.Context, bucketName string, principalAddr string) (*permTypes.Policy, error)
	IsBucketPermissionAllowed(ctx context.Context, userAddr string, bucketName string, action permTypes.ActionType) (permTypes.Effect, error)
	ListBuckets(ctx context.Context, opts types.ListBucketsOptions) (types.ListBucketsResult, error)
	ListBucketReadRecord(ctx context.Context, bucketName string, opts types.ListReadRecordOptions) (types.QuotaRecordInfo, error)
	GetQuotaUpdateTime(ctx context.Context, bucketName string) (int64, error)
	BuyQuotaForBucket(ctx context.Context, bucketName string, targetQuota uint64, opt types.BuyQuotaOption) (string, error)
	GetBucketReadQuota(ctx context.Context, bucketName string) (types.QuotaInfo, error)
	ListBucketsByBucketID(ctx context.Context, bucketIds []uint64, opts types.EndPointOptions) (types.ListBucketsByBucketIDResponse, error)
	GetMigrateBucketApproval(ctx context.Context, migrateBucketMsg *storageTypes.MsgMigrateBucket) (*storageTypes.MsgMigrateBucket, error)
	MigrateBucket(ctx context.Context, bucketName string, dstPrimarySPID uint32, opts types.MigrateBucketOptions) (string, error)
	CancelMigrateBucket(ctx context.Context, bucketName string, opts types.CancelMigrateBucketOptions) (string, error)
	GetBucketMigrationProgress(ctx context.Context, bucketName string, destSP uint32) (types.MigrationProgress, error)
	ListBucketsByPaymentAccount(ctx context.Context, paymentAccount string, opts types.ListBucketsByPaymentAccountOptions) (types.ListBucketsByPaymentAccountResult, error)
}

// GetCreateBucketApproval - Send create bucket approval request to SP and returns the signature info for the approval of preCreating resources.
//
// - ctx: Context variables for the current API call.
//
// - createBucketMsg: The msg of create bucket which defined by the greenfield chain.
//
// - ret1: The msg of create bucket which contain the approval signature from the storage provider
//
// - ret2: Return error when get approval failed, otherwise return nil.
func (c *Client) GetCreateBucketApproval(ctx context.Context, createBucketMsg *storageTypes.MsgCreateBucket) (*storageTypes.MsgCreateBucket, error) {
	unsignedBytes := createBucketMsg.GetSignBytes()

	// set the action type
	urlVal := make(url.Values)
	urlVal["action"] = []string{types.CreateBucketAction}

	reqMeta := requestMeta{
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: types.EmptyStringSHA256,
		txnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method: http.MethodGet,
		adminInfo: AdminAPIInfo{
			isAdminAPI:   true,
			adminVersion: types.AdminV1Version,
		},
	}

	primarySPAddr := createBucketMsg.GetPrimarySpAddress()
	endpoint, err := c.getSPUrlByAddr(primarySPAddr)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("route endpoint by addr: %s failed, err: %s", primarySPAddr, err.Error()))
		return nil, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return nil, err
	}

	// fetch primary signed msg from sp response
	signedRawMsg := resp.Header.Get(types.HTTPHeaderSignedMsg)
	if signedRawMsg == "" {
		return nil, errors.New("fail to fetch pre createBucket signature")
	}

	signedMsgBytes, err := hex.DecodeString(signedRawMsg)
	if err != nil {
		return nil, err
	}

	var signedMsg storageTypes.MsgCreateBucket
	storageTypes.ModuleCdc.MustUnmarshalJSON(signedMsgBytes, &signedMsg)

	return &signedMsg, nil
}

// CreateBucket - Create a new bucket in greenfield.
//
// This API sends a request to the storage provider to get approval for creating  bucket and sends the createBucket transaction to the Greenfield.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to be created.
//
// - primaryAddr: The primary SP address to which the bucket will be created on.
//
// - opts: The Options indicates the meta to construct createBucket msg and the way to send transaction
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error if create bucket failed, otherwise return nil.
func (c *Client) CreateBucket(ctx context.Context, bucketName string, primaryAddr string, opts types.CreateBucketOptions) (string, error) {
	address, err := sdk.AccAddressFromHexUnsafe(primaryAddr)
	if err != nil {
		return "", err
	}

	var visibility storageTypes.VisibilityType
	if opts.Visibility == storageTypes.VISIBILITY_TYPE_UNSPECIFIED {
		visibility = storageTypes.VISIBILITY_TYPE_PRIVATE // set default visibility type
	} else {
		visibility = opts.Visibility
	}

	var paymentAddr sdk.AccAddress
	if opts.PaymentAddress != "" {
		paymentAddr, err = sdk.AccAddressFromHexUnsafe(opts.PaymentAddress)
		if err != nil {
			return "", err
		}
	}

	createBucketMsg := storageTypes.NewMsgCreateBucket(c.MustGetDefaultAccount().GetAddress(), bucketName, visibility, address, paymentAddr, 0, nil, opts.ChargedQuota)

	err = createBucketMsg.ValidateBasic()
	if err != nil {
		return "", err
	}
	signedMsg, err := c.GetCreateBucketApproval(ctx, createBucketMsg)
	if err != nil {
		return "", err
	}

	// set the default txn broadcast mode as block mode
	if opts.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opts.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}
	msgs := []sdk.Msg{signedMsg}

	if opts.Tags != nil {
		// Set tag
		grn := gnfdTypes.NewBucketGRN(bucketName)
		msgSetTag := storageTypes.NewMsgSetTag(c.MustGetDefaultAccount().GetAddress(), grn.String(), opts.Tags)
		msgs = append(msgs, msgSetTag)
	}
	resp, err := c.BroadcastTx(ctx, msgs, opts.TxOpts)
	if err != nil {
		return "", err
	}
	txnHash := resp.TxResponse.TxHash
	if !opts.IsAsyncMode {
		ctxTimeout, cancel := context.WithTimeout(ctx, types.ContextTimeout)
		defer cancel()
		txnResponse, err := c.WaitForTx(ctxTimeout, txnHash)
		if err != nil {
			return txnHash, fmt.Errorf("the transaction has been submitted, please check it later:%v", err)
		}
		if txnResponse.TxResult.Code != 0 {
			return txnHash, fmt.Errorf("the createBucket txn has failed with response code: %d, codespace:%s", txnResponse.TxResult.Code, txnResponse.TxResult.Codespace)
		}
	}
	return txnHash, nil
}

// DeleteBucket - Send DeleteBucket msg to greenfield chain and return txn hash.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to be deleted
//
// - opt: The Options for customizing the transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error if delete bucket failed, otherwise return nil.
func (c *Client) DeleteBucket(ctx context.Context, bucketName string, opt types.DeleteBucketOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}
	delBucketMsg := storageTypes.NewMsgDeleteBucket(c.MustGetDefaultAccount().GetAddress(), bucketName)
	return c.sendTxn(ctx, delBucketMsg, opt.TxOpts)
}

// UpdateBucketVisibility - Update the visibilityType of bucket.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to be updated.
//
// - visibility: The VisibilityType defines on greenfield which can be PUBLIC_READ, PRIVATE or INHERIT
//
// - opt: The Options for customizing the transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error if update visibility failed, otherwise return nil.
func (c *Client) UpdateBucketVisibility(ctx context.Context, bucketName string,
	visibility storageTypes.VisibilityType, opt types.UpdateVisibilityOption,
) (string, error) {
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}

	paymentAddr, err := sdk.AccAddressFromHexUnsafe(bucketInfo.PaymentAddress)
	if err != nil {
		return "", err
	}

	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(c.MustGetDefaultAccount().GetAddress(), bucketName, &bucketInfo.ChargedReadQuota, paymentAddr, visibility)
	return c.sendTxn(ctx, updateBucketMsg, opt.TxOpts)
}

// UpdateBucketPaymentAddr - Update the payment address of bucket. It will send the MsgUpdateBucketInfo msg to greenfield to update the meta.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to be updated.
//
// - paymentAddr: The payment address from which deduct the cost of bucket storage or quota.
//
// - opt: The Options for customizing the transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error if update payment address failed, otherwise return nil.
func (c *Client) UpdateBucketPaymentAddr(ctx context.Context, bucketName string,
	paymentAddr sdk.AccAddress, opt types.UpdatePaymentOption,
) (string, error) {
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}

	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(c.MustGetDefaultAccount().GetAddress(), bucketName, &bucketInfo.ChargedReadQuota, paymentAddr, bucketInfo.Visibility)
	return c.sendTxn(ctx, updateBucketMsg, opt.TxOpts)
}

// UpdateBucketInfo - Update the bucket meta on chain, including read quota, payment address or visibility. It will send the MsgUpdateBucketInfo msg to greenfield to update the meta.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to be updated.
//
// - paymentAddr: The payment address from which deduct the cost of bucket storage or quota.
//
// - opts: The Options used to specify which metas need to be updated and the option to send transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error if update bucket meta failed, otherwise return nil.
func (c *Client) UpdateBucketInfo(ctx context.Context, bucketName string, opts types.UpdateBucketOptions) (string, error) {
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}

	if opts.Visibility == bucketInfo.Visibility && opts.PaymentAddress == "" && opts.ChargedQuota == nil {
		return "", errors.New("no meta need to update")
	}

	var visibility storageTypes.VisibilityType
	var chargedReadQuota uint64
	var paymentAddr sdk.AccAddress

	if opts.Visibility != bucketInfo.Visibility {
		visibility = opts.Visibility
	} else {
		visibility = bucketInfo.Visibility
	}

	if opts.PaymentAddress != "" {
		paymentAddr, err = sdk.AccAddressFromHexUnsafe(opts.PaymentAddress)
		if err != nil {
			return "", err
		}
	} else {
		paymentAddr, err = sdk.AccAddressFromHexUnsafe(bucketInfo.PaymentAddress)
		if err != nil {
			return "", err
		}
	}

	if opts.ChargedQuota != nil {
		chargedReadQuota = *opts.ChargedQuota
	} else {
		chargedReadQuota = bucketInfo.ChargedReadQuota
	}

	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(c.MustGetDefaultAccount().GetAddress(), bucketName,
		&chargedReadQuota, paymentAddr, visibility)

	// set the default txn broadcast mode as block mode
	if opts.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opts.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}

	return c.sendTxn(ctx, updateBucketMsg, opts.TxOpts)
}

func (c *Client) ToggleSPAsDelegatedAgent(ctx context.Context, bucketName string, opt types.UpdateBucketOptions,
) (string, error) {
	_, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}
	msg := storageTypes.NewMsgToggleSPAsDelegatedAgent(c.MustGetDefaultAccount().GetAddress(), bucketName)
	return c.sendTxn(ctx, msg, opt.TxOpts)
}

// HeadBucket - query the bucketInfo on chain by bucket name, return the bucket info if exists.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to query.
//
// - ret1: The bucket specific metadata information, including Visibility, payment address, charged quota and so on.
//
// - ret2: Return error if bucket not exist, otherwise return nil.
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

// HeadBucketByID - query the bucketInfo on chain by the bucket id, return the bucket info if exists.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to query.
//
// - ret1: The bucket specific metadata information, including Visibility, payment address, charged quota and so on.
//
// - ret2: Return error if bucket not exist, otherwise return nil.
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

// PutBucketPolicy - Apply bucket policy to the principal, return the txn hash.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - principalStr: Indicates the marshaled principal content of greenfield permission types, users can generate it by NewPrincipalWithAccount or NewPrincipalWithGroupId method.
//
// - statements: Policies outline the specific details of permissions, including the Effect, ActionList, and Resources.
//
// - opt: The options for customizing the policy expiration time and transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) PutBucketPolicy(ctx context.Context, bucketName string, principalStr types.Principal,
	statements []*permTypes.Statement, opt types.PutPolicyOption,
) (string, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName)
	principal := &permTypes.Principal{}
	if err := principal.Unmarshal([]byte(principalStr)); err != nil {
		return "", err
	}

	putPolicyMsg := storageTypes.NewMsgPutPolicy(c.MustGetDefaultAccount().GetAddress(), resource.String(),
		principal, statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(ctx, putPolicyMsg, opt.TxOpts)
}

// DeleteBucketPolicy - Delete the bucket policy of the principal.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - principalStr: Indicates the marshaled principal content of greenfield permission types, users can generate it by NewPrincipalWithAccount or NewPrincipalWithGroupId method.
//
// - opt: The option for send delete policy transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) DeleteBucketPolicy(ctx context.Context, bucketName string, principalStr types.Principal, opt types.DeletePolicyOption) (string, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()
	principal := &permTypes.Principal{}
	if err := principal.Unmarshal([]byte(principalStr)); err != nil {
		return "", err
	}

	return c.sendDelPolicyTxn(ctx, c.MustGetDefaultAccount().GetAddress(), resource, principal, opt.TxOpts)
}

// IsBucketPermissionAllowed - Check if the permission of bucket is allowed to the user.
//
// - ctx: Context variables for the current API call.
//
// - userAddr: The HEX-encoded string of the user address
//
// - bucketName: The bucket name identifies the bucket.
//
// - action: Indicates the permission corresponding to which type of action needs to be verified
//
// - ret1: Return EFFECT_ALLOW if the permission is allowed and EFFECT_DENY if the permission is denied
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) IsBucketPermissionAllowed(ctx context.Context, userAddr string,
	bucketName string, action permTypes.ActionType,
) (permTypes.Effect, error) {
	_, err := sdk.AccAddressFromHexUnsafe(userAddr)
	if err != nil {
		return permTypes.EFFECT_DENY, err
	}
	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   userAddr,
		BucketName: bucketName,
		ActionType: action,
	}

	verifyResp, err := c.chainClient.VerifyPermission(ctx, &verifyReq)
	if err != nil {
		return permTypes.EFFECT_DENY, err
	}

	return verifyResp.Effect, nil
}

// GetBucketPolicy - Get the bucket policy info of the user specified by principalAddr.
//
// - bucketName: The bucket name identifies the bucket.
//
// - principalAddr: The HEX-encoded string of the principal address.
//
// - ret1: The bucket policy info defined on greenfield.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetBucketPolicy(ctx context.Context, bucketName string, principalAddr string) (*permTypes.Policy, error) {
	_, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return nil, err
	}

	resource := gnfdTypes.NewBucketGRN(bucketName).String()
	queryPolicy := storageTypes.QueryPolicyForAccountRequest{
		Resource:         resource,
		PrincipalAddress: principalAddr,
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

type listBucketsByIDsResponse map[uint64]*types.BucketMeta

type bucketEntry struct {
	Id    uint64
	Value *types.BucketMeta
}

func (m *listBucketsByIDsResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = listBucketsByIDsResponse{}
	for {
		var e bucketEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		} else {
			(*m)[e.Id] = e.Value
		}
	}
	return nil
}

// ListBuckets - Lists the bucket info of the user.
//
// If the opts.Account is not set, the user is default set as the sender.
//
// - ctx: Context variables for the current API call.
//
// - opts: The options to set the meta to list the bucket
//
// - ret1: The result of list bucket under specific user address
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListBuckets(ctx context.Context, opts types.ListBucketsOptions) (types.ListBucketsResult, error) {
	params := url.Values{}
	params.Set("include-removed", strconv.FormatBool(opts.ShowRemovedBucket))

	account := opts.Account
	if account == "" {
		acc, err := c.GetDefaultAccount()
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to get default account:  %s", err.Error()))
			return types.ListBucketsResult{}, err
		}
		account = acc.GetAddress().String()
	} else {
		_, err := sdk.AccAddressFromHexUnsafe(account)
		if err != nil {
			return types.ListBucketsResult{}, err
		}
	}

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
		userAddress:   account,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListBucketsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return types.ListBucketsResult{}, err
	}
	defer utils.CloseResponse(resp)

	listBucketsResult := types.ListBucketsResult{}
	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return types.ListBucketsResult{}, err
	}

	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), &listBucketsResult)

	// TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil {
		return types.ListBucketsResult{}, err
	}

	return listBucketsResult, nil
}

// ListBucketReadRecord - List the download record info of the specific bucket of the current month.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - opts: Indicates the start timestamp of return read records and the max number of return items
//
// - ret1: The read record info of the bucket returned by SP.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListBucketReadRecord(ctx context.Context, bucketName string, opts types.ListReadRecordOptions) (types.QuotaRecordInfo, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return types.QuotaRecordInfo{}, err
	}
	timeNow := time.Now()
	timeToday := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())
	if opts.StartTimeStamp < 0 {
		return types.QuotaRecordInfo{}, errors.New("start timestamp  less than 0")
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
		return types.QuotaRecordInfo{}, errors.New("start timestamp larger than the end timestamp of this month")
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
		log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %s", bucketName, err.Error()))
		return types.QuotaRecordInfo{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.QuotaRecordInfo{}, err
	}
	defer utils.CloseResponse(resp)

	QuotaRecords := types.QuotaRecordInfo{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&QuotaRecords)
	if err != nil {
		return types.QuotaRecordInfo{}, err
	}

	return QuotaRecords, nil
}

// GetBucketReadQuota - Query the quota info of the specific bucket of current month.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - ret1: The info of quota which contains the consumed quota, the charged quota and free quota info of the bucket
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetBucketReadQuota(ctx context.Context, bucketName string) (types.QuotaInfo, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return types.QuotaInfo{}, err
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
		log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %s", bucketName, err.Error()))
		return types.QuotaInfo{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.QuotaInfo{}, err
	}
	defer utils.CloseResponse(resp)

	QuotaResult := types.QuotaInfo{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&QuotaResult)
	if err != nil {
		return types.QuotaInfo{}, err
	}

	return QuotaResult, nil
}

// GetQuotaUpdateTime - Query the update time stamp of the bucket quota info.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - ret1: The update time stamp.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetQuotaUpdateTime(ctx context.Context, bucketName string) (int64, error) {
	resp, err := c.chainClient.QueryQuotaUpdateTime(ctx, &storageTypes.QueryQuoteUpdateTimeRequest{
		BucketName: bucketName,
	})
	if err != nil {
		return 0, err
	}
	return resp.UpdateAt, nil
}

// BuyQuotaForBucket - Buy the target quota for the specific bucket.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - targetQuota: Indicates the target quota value after setting, and can only be set to a higher value than the current value.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) BuyQuotaForBucket(ctx context.Context, bucketName string, targetQuota uint64, opt types.BuyQuotaOption) (string, error) {
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}

	paymentAddr, err := sdk.AccAddressFromHexUnsafe(bucketInfo.PaymentAddress)
	if err != nil {
		return "", err
	}
	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(c.MustGetDefaultAccount().GetAddress(), bucketName, &targetQuota, paymentAddr, bucketInfo.Visibility)

	resp, err := c.BroadcastTx(ctx, []sdk.Msg{updateBucketMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// ListBucketsByBucketID - List buckets by bucket ids.
//
// By inputting a collection of bucket IDs, we can retrieve the corresponding bucket data. If the bucket is nonexistent or has been deleted, a null value will be returned
//
// - ctx: Context variables for the current API call.
//
// - bucketIds: The list of bucket ids.
//
// - opts: The options to set the meta to list buckets by bucket id.
//
// - ret1: The result of bucket info map by given bucket ids.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListBucketsByBucketID(ctx context.Context, bucketIds []uint64, opts types.EndPointOptions) (types.ListBucketsByBucketIDResponse, error) {
	const MaximumListBucketsSize = 1000
	if len(bucketIds) == 0 || len(bucketIds) > MaximumListBucketsSize {
		return types.ListBucketsByBucketIDResponse{}, nil
	}

	bucketIDMap := make(map[uint64]bool)
	for _, id := range bucketIds {
		if _, ok := bucketIDMap[id]; ok {
			// repeat id keys in request
			return types.ListBucketsByBucketIDResponse{}, nil
		}
		bucketIDMap[id] = true
	}

	idStr := make([]string, len(bucketIds))
	for i, id := range bucketIds {
		idStr[i] = strconv.FormatUint(id, 10)
	}
	IDs := strings.Join(idStr, ",")

	params := url.Values{}
	params.Set("buckets-query", "")
	params.Set("ids", IDs)

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&opts)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListBucketsByBucketIDResponse{}, err

	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.ListBucketsByBucketIDResponse{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msgf("the list of buckets in bucket ids:%v failed: %s", bucketIds, err.Error())
		return types.ListBucketsByBucketIDResponse{}, err
	}

	buckets := types.ListBucketsByBucketIDResponse{}
	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), (*listBucketsByIDsResponse)(&buckets.Buckets))
	if err != nil && buckets.Buckets == nil {
		log.Error().Msgf("the list of buckets in bucket ids:%v failed: %s", bucketIds, err.Error())
		return types.ListBucketsByBucketIDResponse{}, err
	}

	return buckets, nil
}

// GetMigrateBucketApproval - Send migrate get approval request to the storage provider and return the signed MsgMigrateBucket by SP.
//
// - ctx: Context variables for the current API call.
//
// - migrateBucketMsg: Indicates msg of migrating bucket which defined by greenfield
//
// - ret1: The msg of migrating bucket which contain the approval signature from the storage provider.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetMigrateBucketApproval(ctx context.Context, migrateBucketMsg *storageTypes.MsgMigrateBucket) (*storageTypes.MsgMigrateBucket, error) {
	unsignedBytes := migrateBucketMsg.GetSignBytes()

	// set the action type
	urlVal := make(url.Values)
	urlVal["action"] = []string{types.MigrateBucketAction}

	reqMeta := requestMeta{
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: types.EmptyStringSHA256,
		txnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method: http.MethodGet,
		adminInfo: AdminAPIInfo{
			isAdminAPI:   true,
			adminVersion: types.AdminV1Version,
		},
	}

	primarySPID := migrateBucketMsg.DstPrimarySpId
	endpoint, err := c.getSPUrlByID(primarySPID)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("route endpoint by addr: %d failed, err: %s", primarySPID, err.Error()))
		return nil, err
	}
	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return nil, err
	}

	// fetch primary signed msg from sp response
	signedRawMsg := resp.Header.Get(types.HTTPHeaderSignedMsg)
	if signedRawMsg == "" {
		return nil, errors.New("fail to fetch pre createBucket signature")
	}

	signedMsgBytes, err := hex.DecodeString(signedRawMsg)
	if err != nil {
		return nil, err
	}

	var signedMsg storageTypes.MsgMigrateBucket
	storageTypes.ModuleCdc.MustUnmarshalJSON(signedMsgBytes, &signedMsg)

	return &signedMsg, nil
}

// MigrateBucket - Get approval of migrating from SP, send the signed migrate bucket msg to greenfield chain and return the txn hash.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket to be migrated.
//
// - dstPrimarySPID: The sp id of the migration target SP.
//
// - opt: The options of send transaction of migration bucket
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request of getting approval or sending transaction failed, otherwise return nil.
func (c *Client) MigrateBucket(ctx context.Context, bucketName string, dstPrimarySPID uint32, opts types.MigrateBucketOptions) (string, error) {
	migrateBucketMsg := storageTypes.NewMsgMigrateBucket(c.MustGetDefaultAccount().GetAddress(), bucketName, dstPrimarySPID)

	err := migrateBucketMsg.ValidateBasic()
	if err != nil {
		return "", err
	}
	signedMsg, err := c.GetMigrateBucketApproval(ctx, migrateBucketMsg)
	if err != nil {
		return "", err
	}

	// set the default txn broadcast mode as block mode
	if opts.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opts.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}

	resp, err := c.BroadcastTx(ctx, []sdk.Msg{signedMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}
	txnHash := resp.TxResponse.TxHash
	if !opts.IsAsyncMode {
		ctxTimeout, cancel := context.WithTimeout(ctx, types.ContextTimeout)
		defer cancel()
		txnResponse, err := c.WaitForTx(ctxTimeout, txnHash)
		if err != nil {
			return txnHash, fmt.Errorf("the transaction has been submitted, please check it later:%v", err)
		}
		if txnResponse.TxResult.Code != 0 {
			return txnHash, fmt.Errorf("the migrateBucket txn has failed with response code: %d, codespace:%s", txnResponse.TxResult.Code, txnResponse.TxResult.Codespace)
		}
	}
	return txnHash, nil
}

// CancelMigrateBucket - Cancel migrate migration by sending the MsgCancelMigrateBucket msg.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the migrating bucket to be canceled.
//
// - opt: The options of the proposal meta and transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request of cancel migration failed, otherwise return nil.
func (c *Client) CancelMigrateBucket(ctx context.Context, bucketName string, opts types.CancelMigrateBucketOptions) (string, error) {

	cancelMigrateBucketMsg := storageTypes.NewMsgCancelMigrateBucket(c.MustGetDefaultAccount().GetAddress(), bucketName)

	err := cancelMigrateBucketMsg.ValidateBasic()
	if err != nil {
		return "", err
	}

	// set the default txn broadcast mode as sync mode
	if opts.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opts.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}

	resp, err := c.BroadcastTx(ctx, []sdk.Msg{cancelMigrateBucketMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}

	var txnResponse *ctypes.ResultTx
	txnHash := resp.TxResponse.TxHash
	if !opts.IsAsyncMode {
		ctxTimeout, cancel := context.WithTimeout(ctx, types.ContextTimeout)
		defer cancel()

		txnResponse, err = c.WaitForTx(ctxTimeout, txnHash)
		if err != nil {
			return txnHash, fmt.Errorf("the transaction has been submitted, please check it later:%v", err)
		}

		if txnResponse.TxResult.Code != 0 {
			return txnHash, fmt.Errorf("the createBucket txn has failed with response code: %d", txnResponse.TxResult.Code)
		}
	}

	return txnHash, nil
}

// ListBucketsByPaymentAccount - List bucket info by payment account.
//
// By inputting a collection of bucket IDs, we can retrieve the corresponding bucket data. If the bucket is nonexistent or has been deleted, a null value will be returned
//
// - ctx: Context variables for the current API call.
//
// - paymentAccount: Payment account defines the address of payment account.
//
// - opts: The options to set the meta to list buckets by payment account.
//
// - ret1: The result of bucket info under specific payment account.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListBucketsByPaymentAccount(ctx context.Context, paymentAccount string, opts types.ListBucketsByPaymentAccountOptions) (types.ListBucketsByPaymentAccountResult, error) {

	_, err := sdk.AccAddressFromHexUnsafe(paymentAccount)
	if err != nil {
		return types.ListBucketsByPaymentAccountResult{}, err
	}

	params := url.Values{}
	params.Set("payment-buckets", "")
	params.Set("payment-account", paymentAccount)

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListBucketsByPaymentAccountResult{}, err

	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.ListBucketsByPaymentAccountResult{}, err
	}
	defer utils.CloseResponse(resp)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return types.ListBucketsByPaymentAccountResult{}, errors.New("copy the response error" + err.Error())
	}

	buckets := types.ListBucketsByPaymentAccountResult{}
	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), &buckets)
	if err != nil {
		return types.ListBucketsByPaymentAccountResult{}, errors.New("unmarshal response error" + err.Error())
	}

	return buckets, nil
}

// GetBucketMigrationProgress - Query the migration progress info of the specific bucket.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - ret1: The info of migration progress which contains the progress info of the bucket
//
// - ret2: Return error when the request failed, otherwise return nil.

// GetBucketMigrationProgress return the status of object including the uploading progress
func (c *Client) GetBucketMigrationProgress(ctx context.Context, bucketName string, destSP uint32) (types.MigrationProgress, error) {
	_, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return types.MigrationProgress{}, err
	}

	// get object status from sp
	migrationProgress, err := c.getMigrationStateFromSP(ctx, bucketName, destSP)
	if err != nil {
		return types.MigrationProgress{}, errors.New("fail to fetch bucket migration progress from sp" + err.Error())
	}
	return migrationProgress, nil
}

func (c *Client) getMigrationStateFromSP(ctx context.Context, bucketName string, destSP uint32) (types.MigrationProgress, error) {
	params := url.Values{}
	params.Set("bucket-migration-progress", "")

	reqMeta := requestMeta{
		urlValues:     params,
		bucketName:    bucketName,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlByID(destSP)
	if err != nil {
		return types.MigrationProgress{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.MigrationProgress{}, err
	}

	defer utils.CloseResponse(resp)

	migrationProgress := types.MigrationProgress{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&migrationProgress)
	if err != nil {
		return types.MigrationProgress{}, err
	}

	return migrationProgress, nil
}
