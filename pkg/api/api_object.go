package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	sdkerror "github.com/bnb-chain/greenfield-go-sdk/pkg/error"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/s3util"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/types"
)

// GetRedundancyParams query and return the data shards, parity shards and segment size of redundancy
// configuration on chain
func (c *Client) GetRedundancyParams() (uint32, uint32, uint64, error) {
	query := storageTypes.QueryParamsRequest{}
	queryResp, err := c.chainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return 0, 0, 0, err
	}

	params := queryResp.Params
	return params.GetRedundantDataChunkNum(), params.GetRedundantParityChunkNum(), params.GetMaxSegmentSize(), nil
}

// ComputeHashRoots return the hash roots list and content size
func (c *Client) ComputeHashRoots(reader io.Reader) ([][]byte, int64, storageTypes.RedundancyType, error) {
	dataBlocks, parityBlocks, segSize, err := c.GetRedundancyParams()
	if err != nil {
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}

	// get hash and objectSize from reader
	return hashlib.ComputeIntegrityHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks))
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *Client) CreateObject(ctx context.Context, bucketName, objectName string,
	reader io.Reader, opts client.CreateObjectOptions) (string, error) {
	if reader == nil {
		return "", errors.New("fail to compute hash of payload, reader is nil")
	}

	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	// compute hash root of payload
	expectCheckSums, size, redundancyType, err := c.ComputeHashRoots(reader)
	if err != nil {
		return "", err
	}

	var contentType string
	if opts.ContentType != "" {
		contentType = opts.ContentType
	} else {
		contentType = types.ContentDefault
	}

	var visibility storageTypes.VisibilityType
	if opts.Visibility == storageTypes.VISIBILITY_TYPE_UNSPECIFIED {
		visibility = storageTypes.VISIBILITY_TYPE_INHERIT // set default visibility type
	} else {
		visibility = opts.Visibility
	}

	createObjectMsg := storageTypes.NewMsgCreateObject(km.GetAddr(), bucketName, objectName,
		uint64(size), visibility, expectCheckSums, contentType, redundancyType, math.MaxUint, nil, opts.SecondarySPAccs)
	err = createObjectMsg.ValidateBasic()
	if err != nil {
		return "", err
	}

	signedCreateObjectMsg, err := c.GetCreateObjectApproval(ctx, createObjectMsg, NewAuthInfo(false, ""))
	if err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{signedCreateObjectMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, err
}

// DeleteObject send DeleteBucket txn to greenfield chain and return txn hash
func (c *Client) DeleteObject(bucketName, objectName string, opt client.DeleteObjectOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	delObjectMsg := storageTypes.NewMsgDeleteObject(km.GetAddr(), bucketName, objectName)

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{delObjectMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// CancelCreateObject send CancelCreateObject txn to greenfield chain
func (c *Client) CancelCreateObject(bucketName, objectName string, opt client.CancelCreateOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	cancelCreateMsg := storageTypes.NewMsgCancelCreateObject(km.GetAddr(), bucketName, objectName)

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{cancelCreateMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// PutObject supports the second stage of uploading the object to bucket.
// txnHash should be the str which hex.encoding from txn hash bytes
func (c *Client) PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
	reader io.Reader, authInfo AuthInfo, opt client.PutObjectOption,
) (err error) {
	if txnHash == "" {
		return errors.New("txn hash empty")
	}

	if objectSize <= 0 {
		return errors.New("object size not set")
	}

	var contentType string
	if opt.ContentType != "" {
		contentType = opt.ContentType
	} else {
		contentType = types.ContentDefault
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: types.EmptyStringSHA256,
		contentLength: objectSize,
		contentType:   contentType,
	}

	sendOpt := sendOptions{
		method:  http.MethodPut,
		body:    reader,
		txnHash: txnHash,
	}

	endpoint, err := c.getSPUrlFromBucket(bucketName)
	if err != nil {
		return err
	}

	_, err = c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return err
	}

	return nil
}

// FPutObject supports uploading object from local file
func (c *Client) FPutObject(ctx context.Context, bucketName, objectName,
	filePath, txnHash, contentType string, authInfo AuthInfo,
) (err error) {
	fReader, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		return err
	}
	defer fReader.Close()

	// Save the file stat.
	stat, err := fReader.Stat()
	if err != nil {
		return err
	}

	return c.PutObject(ctx, bucketName, objectName, txnHash, stat.Size(), fReader, authInfo, client.PutObjectOption{ContentType: contentType})
}

// GetObject download s3 object payload and return the related object info
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string,
	opts client.GetObjectOption, authInfo AuthInfo) (io.ReadCloser, ObjectInfo, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return nil, ObjectInfo{}, err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return nil, ObjectInfo{}, err
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: types.EmptyStringSHA256,
	}

	if opts.Range != "" {
		reqMeta.Range = opts.Range
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlFromBucket(bucketName)
	if err != nil {
		return nil, ObjectInfo{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return nil, ObjectInfo{}, err
	}

	ObjInfo, err := getObjInfo(objectName, resp.Header)
	if err != nil {
		utils.CloseResponse(resp)
		return nil, ObjectInfo{}, err
	}

	return resp.Body, ObjInfo, nil
}

// FGetObject download s3 object payload adn write the object content into local file specified by filePath
func (c *Client) FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts client.GetObjectOption, authinfo AuthInfo) error {
	// Verify if destination already exists.
	st, err := os.Stat(filePath)
	if err == nil {
		// If the destination exists and is a directory.
		if st.IsDir() {
			return errors.New("fileName is a directory.")
		}
	}

	// If file exist, open it in append mode
	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o660)
	if err != nil {
		return err
	}

	body, _, err := c.GetObject(ctx, bucketName, objectName, opts, authinfo)
	if err != nil {
		return err
	}
	defer body.Close()

	_, err = io.Copy(fd, body)
	fd.Close()
	if err != nil {
		return err
	}

	return nil
}

// getObjInfo generates objectInfo base on the response http header content
func getObjInfo(objectName string, h http.Header) (ObjectInfo, error) {
	// Parse content length is exists
	var size int64 = -1
	var err error
	contentLength := h.Get(types.HTTPHeaderContentLength)
	if contentLength != "" {
		size, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return ObjectInfo{}, sdkerror.ErrResponse{
				Code:    "InternalError",
				Message: fmt.Sprintf("Content-Length parse error %v", err),
			}
		}
	}

	// fetch content type
	contentType := strings.TrimSpace(h.Get("Content-Type"))
	if contentType == "" {
		contentType = types.ContentDefault
	}

	return ObjectInfo{
		ObjectName:  objectName,
		ContentType: contentType,
		Size:        size,
	}, nil
}

// HeadObject query the objectInfo on chain to check th object id, return the object info if exists
// return err info if object not exist
func (c *Client) HeadObject(ctx context.Context, bucketName, objectName string) (*storageTypes.ObjectInfo, error) {
	queryHeadObjectRequest := storageTypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := c.chainClient.HeadObject(ctx, &queryHeadObjectRequest)
	if err != nil {
		return nil, err
	}

	return queryHeadObjectResponse.ObjectInfo, nil
}

// HeadObjectByID query the objectInfo on chain by object id, return the object info if exists
// return err info if object not exist
func (c *Client) HeadObjectByID(ctx context.Context, objID string) (*storageTypes.ObjectInfo, error) {
	headObjectRequest := storageTypes.QueryHeadObjectByIdRequest{
		ObjectId: objID,
	}
	queryHeadObjectResponse, err := c.chainClient.HeadObjectById(ctx, &headObjectRequest)
	if err != nil {
		return nil, err
	}

	return queryHeadObjectResponse.ObjectInfo, nil
}

// PutObjectPolicy apply object policy to the principal, return the txn hash
func (c *Client) PutObjectPolicy(bucketName, objectName string, principalStr client.Principal,
	statements []*permTypes.Statement, opt client.PutPolicyOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)

	principal := &permTypes.Principal{}
	if err = principal.Unmarshal([]byte(principalStr)); err != nil {
		return "", err
	}

	putPolicyMsg := storageTypes.NewMsgPutPolicy(km.GetAddr(), resource.String(),
		principal, statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(putPolicyMsg, opt.TxOpts)
}

func (c *Client) DeleteObjectPolicy(bucketName, objectName string, principalAddr sdk.AccAddress, opt client.DeletePolicyOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	principal := permTypes.NewPrincipalWithAccount(principalAddr)
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	return c.sendDelPolicyTxn(km.GetAddr(), resource.String(), principal, opt.TxOpts)
}

// IsObjectPermissionAllowed check if the permission of the object is allowed to the user
func (c *Client) IsObjectPermissionAllowed(ctx context.Context, user sdk.AccAddress,
	bucketName, objectName string, action permTypes.ActionType) (permTypes.Effect, error) {
	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   user.String(),
		BucketName: bucketName,
		ObjectName: objectName,
		ActionType: action,
	}

	verifyResp, err := c.chainClient.VerifyPermission(ctx, &verifyReq)
	if err != nil {
		return permTypes.EFFECT_DENY, err
	}

	return verifyResp.Effect, nil
}

// GetObjectPolicy get the object policy info of the user specified by principalAddr
func (c *Client) GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	queryPolicy := storageTypes.QueryPolicyForAccountRequest{
		Resource:         resource.String(),
		PrincipalAddress: principalAddr.String(),
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// ListObjects return object list of the specific bucket
func (c *Client) ListObjects(ctx context.Context, bucketName string, authInfo AuthInfo) (client.ListObjectsResponse, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return client.ListObjectsResponse{}, err
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlFromBucket(bucketName)
	if err != nil {
		return client.ListObjectsResponse{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return client.ListObjectsResponse{}, err
	}
	defer utils.CloseResponse(resp)

	ListObjectsResult := client.ListObjectsResponse{}
	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return client.ListObjectsResponse{}, err
	}

	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &ListObjectsResult)
	//TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil && ListObjectsResult.Objects == nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return client.ListObjectsResponse{}, err
	}

	return ListObjectsResult, nil
}
