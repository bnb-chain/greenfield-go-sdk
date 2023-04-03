package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	sdkerror "github.com/bnb-chain/greenfield-go-sdk/pkg/error"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/s3util"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
)

type Object interface {
	GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject,
		authInfo types.AuthInfo) (*storageTypes.MsgCreateObject, error)
	CreateObject(ctx context.Context, bucketName, objectName string,
		reader io.Reader, opts *CreateObjectOptions) (string, error)
	PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
		reader io.Reader, opt types.PutObjectOption) (err error)
	CancelCreateObject(ctx context.Context, bucketName, objectName string, opt types.CancelCreateOption) (string, error)
	DeleteObject(ctx context.Context, bucketName, objectName string, opt DeleteObjectOption) (string, error)
	GetObject(ctx context.Context, bucketName, objectName string, opt types.GetObjectOption) (io.ReadCloser, GetObjectResult, error)
	// HeadObject query the objectInfo on chain to check th object id, return the object info if exists
	// return err info if object not exist
	HeadObject(ctx context.Context, bucketName, objectName string) (*storageTypes.ObjectInfo, error)
	// HeadObjectByID query the objectInfo on chain by object id, return the object info if exists
	// return err info if object not exist
	HeadObjectByID(ctx context.Context, objID string) (*storageTypes.ObjectInfo, error)
	// PutObjectPolicy apply object policy to the principal, return the txn hash
	PutObjectPolicy(ctx context.Context, bucketName, objectName string, principalStr types.Principal,
		statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	DeleteObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress, opt types.DeletePolicyOption) (string, error)
	// GetObjectPolicy get the object policy info of the user specified by principalAddr
	GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error)
	ListObjects(ctx context.Context, bucketName string, authInfo types.AuthInfo) (ListObjectsResult, error)
}

// CreateObjectOptions indicates the metadata to construct `createObject` message of storage module
type CreateObjectOptions struct {
	Visibility      storageTypes.VisibilityType
	TxOpts          *gnfdSdkTypes.TxOption
	SecondarySPAccs []sdk.AccAddress
	ContentType     string
	IsReplicaType   bool // indicates whether the object use REDUNDANCY_REPLICA_TYPE
}

type DeleteObjectOption struct {
	TxOpts *gnfdSdkTypes.TxOption
}

// GetRedundancyParams query and return the data shards, parity shards and segment size of redundancy
// configuration on chain
func (c *client) GetRedundancyParams() (uint32, uint32, uint64, error) {
	query := storageTypes.QueryParamsRequest{}
	queryResp, err := c.chainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return 0, 0, 0, err
	}

	params := queryResp.Params
	return params.GetRedundantDataChunkNum(), params.GetRedundantParityChunkNum(), params.GetMaxSegmentSize(), nil
}

// ComputeHashRoots return the hash roots list and content size
func (c *client) ComputeHashRoots(reader io.Reader) ([][]byte, int64, storageTypes.RedundancyType, error) {
	dataBlocks, parityBlocks, segSize, err := c.GetRedundancyParams()
	if err != nil {
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}

	// get hash and objectSize from reader
	return hashlib.ComputeIntegrityHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks))
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *client) CreateObject(ctx context.Context, bucketName, objectName string,
	reader io.Reader, opts CreateObjectOptions) (string, error) {
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

	signedCreateObjectMsg, err := c.GetCreateObjectApproval(ctx, createObjectMsg, types.NewAuthInfo(false, ""))
	if err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{signedCreateObjectMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, err
}

// DeleteObject send DeleteBucket txn to greenfield chain and return txn hash
func (c *client) DeleteObject(ctx context.Context, bucketName, objectName string, opt DeleteObjectOption) (string, error) {
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

	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{delObjectMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// CancelCreateObject send CancelCreateObject txn to greenfield chain
func (c *client) CancelCreateObject(ctx context.Context, bucketName, objectName string, opt types.CancelCreateOption) (string, error) {
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

	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{cancelCreateMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// PutObject supports the second stage of uploading the object to bucket.
// txnHash should be the str which hex.encoding from txn hash bytes
func (c *client) PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
	reader io.Reader, authInfo types.AuthInfo, opt types.PutObjectOption,
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

	endpoint, err := c.getSPUrlByBucket(bucketName)
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
func (c *client) FPutObject(ctx context.Context, bucketName, objectName,
	filePath, txnHash, contentType string, authInfo types.AuthInfo,
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

	return c.PutObject(ctx, bucketName, objectName, txnHash, stat.Size(), fReader, authInfo, types.PutObjectOption{ContentType: contentType})
}

// GetObject download s3 object payload and return the related object info
func (c *client) GetObject(ctx context.Context, bucketName, objectName string,
	opts types.GetObjectOption, authInfo types.AuthInfo) (io.ReadCloser, types.ObjectStat, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return nil, types.ObjectStat{}, err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return nil, types.ObjectStat{}, err
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

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		return nil, types.ObjectStat{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return nil, types.ObjectStat{}, err
	}

	objStat, err := getObjInfo(objectName, resp.Header)
	if err != nil {
		utils.CloseResponse(resp)
		return nil, types.ObjectStat{}, err
	}

	return resp.Body, objStat, nil
}

// FGetObject download s3 object payload adn write the object content into local file specified by filePath
func (c *client) FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts types.GetObjectOption, authInfo types.AuthInfo) error {
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

	body, _, err := c.GetObject(ctx, bucketName, objectName, opts, authInfo)
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
func getObjInfo(objectName string, h http.Header) (types.ObjectStat, error) {
	// Parse content length is exists
	var size int64 = -1
	var err error
	contentLength := h.Get(types.HTTPHeaderContentLength)
	if contentLength != "" {
		size, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return types.ObjectStat{}, sdkerror.ErrResponse{
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

	return types.ObjectStat{
		ObjectName:  objectName,
		ContentType: contentType,
		Size:        size,
	}, nil
}

// HeadObject query the objectInfo on chain to check th object id, return the object info if exists
// return err info if object not exist
func (c *client) HeadObject(ctx context.Context, bucketName, objectName string) (*storageTypes.ObjectInfo, error) {
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
func (c *client) HeadObjectByID(ctx context.Context, objID string) (*storageTypes.ObjectInfo, error) {
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
func (c *client) PutObjectPolicy(ctx context.Context, bucketName, objectName string, principalStr types.Principal,
	statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error) {
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

	return c.sendPutPolicyTxn(ctx, putPolicyMsg, opt.TxOpts)
}

func (c *client) DeleteObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress, opt types.DeletePolicyOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	principal := permTypes.NewPrincipalWithAccount(principalAddr)
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	return c.sendDelPolicyTxn(ctx, km.GetAddr(), resource.String(), principal, opt.TxOpts)
}

// IsObjectPermissionAllowed check if the permission of the object is allowed to the user
func (c *client) IsObjectPermissionAllowed(ctx context.Context, user sdk.AccAddress,
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
func (c *client) GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error) {
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
func (c *client) ListObjects(ctx context.Context, bucketName string, authInfo types.AuthInfo) (ListObjectsResult, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return ListObjectsResult{}, err
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		return ListObjectsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return ListObjectsResult{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsResult{}, err
	}

	listObjectsResult := ListObjectsResult{}
	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &listObjectsResult)
	//TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil && listObjectsResult.Objects == nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsResult{}, err
	}

	return listObjectsResult, nil
}

// GetCreateObjectApproval returns the signature info for the approval of preCreating resources
func (c *client) GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject, authInfo types.AuthInfo) (*storageTypes.MsgCreateObject, error) {
	unsignedBytes := createObjectMsg.GetSignBytes()

	// set the action type
	urlValues := url.Values{
		"action": {types.CreateObjectAction},
	}

	reqMeta := requestMeta{
		urlValues:     urlValues,
		urlRelPath:    "get-approval",
		contentSHA256: types.EmptyStringSHA256,
		TxnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	endpoint, err := c.getSPUrlByBucket(createObjectMsg.BucketName)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return nil, err
	}

	// fetch primary signed msg from sp response
	signedRawMsg := resp.Header.Get(types.HTTPHeaderSignedMsg)
	if signedRawMsg == "" {
		return nil, errors.New("fail to fetch pre createObject signature")
	}

	signedMsgBytes, err := hex.DecodeString(signedRawMsg)
	if err != nil {
		return nil, err
	}

	var signedMsg storageTypes.MsgCreateObject
	storageTypes.ModuleCdc.MustUnmarshalJSON(signedMsgBytes, &signedMsg)

	return &signedMsg, nil
}
