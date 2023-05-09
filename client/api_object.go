package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
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
	gnfdsdk "github.com/bnb-chain/greenfield/sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/s3util"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

type Object interface {
	GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject) (*storageTypes.MsgCreateObject, error)
	CreateObject(ctx context.Context, bucketName, objectName string, reader io.Reader, opts types.CreateObjectOptions) (string, error)
	PutObject(ctx context.Context, bucketName, objectName string, objectSize int64, reader io.Reader, opts types.PutObjectOptions) error
	FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts types.PutObjectOptions) (err error)
	CancelCreateObject(ctx context.Context, bucketName, objectName string, opt types.CancelCreateOption) (string, error)
	DeleteObject(ctx context.Context, bucketName, objectName string, opt types.DeleteObjectOption) (string, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts types.GetObjectOption) (io.ReadCloser, types.ObjectStat, error)
	FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts types.GetObjectOption) error

	// HeadObject query the objectInfo on chain to check th object id, return the object info if exists
	// return err info if object not exist
	HeadObject(ctx context.Context, bucketName, objectName string) (*storageTypes.ObjectInfo, error)
	// HeadObjectByID query the objectInfo on chain by object id, return the object info if exists
	// return err info if object not exist
	HeadObjectByID(ctx context.Context, objID string) (*storageTypes.ObjectInfo, error)

	// PutObjectPolicy apply object policy to the principal, return the txn hash
	PutObjectPolicy(ctx context.Context, bucketName, objectName string, principalStr types.Principal,
		statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	// DeleteObjectPolicy delete the object policy of the principal
	// principalAddr indicates the HEX-encoded string of the principal address
	DeleteObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr string, opt types.DeletePolicyOption) (string, error)
	// GetObjectPolicy get the object policy info of the user specified by principalAddr.
	// principalAddr indicates the HEX-encoded string of the principal address
	GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr string) (*permTypes.Policy, error)
	// IsObjectPermissionAllowed check if the permission of the object is allowed to the user
	// userAddr indicates the HEX-encoded string of the user address
	IsObjectPermissionAllowed(ctx context.Context, userAddr string, bucketName, objectName string, action permTypes.ActionType) (permTypes.Effect, error)

	ListObjects(ctx context.Context, bucketName, maxKeys, startAfter, continuationToken, delimiter, prefix string, opts types.ListObjectsOptions) (types.ListObjectsResult, error)
	// ComputeHashRoots compute the integrity hash, content size and the redundancy type of the file
	ComputeHashRoots(reader io.Reader) ([][]byte, int64, storageTypes.RedundancyType, error)

	// CreateFolder creates an empty object used as folder.
	// objectName must ending with a forward slash (/) character
	CreateFolder(ctx context.Context, bucketName, objectName string, opts types.CreateObjectOptions) (string, error)

	// GetObjectUploadProgress return the status of the uploading object
	GetObjectUploadProgress(ctx context.Context, bucketName, objectName string) (string, error)
}

// GetRedundancyParams query and return the data shards, parity shards and segment size of redundancy
// configuration on chain
func (c *client) GetRedundancyParams() (uint32, uint32, uint64, error) {
	query := storageTypes.QueryParamsRequest{}
	queryResp, err := c.chainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return 0, 0, 0, err
	}

	versionedParams := queryResp.Params.VersionedParams
	return versionedParams.GetRedundantDataChunkNum(), versionedParams.GetRedundantParityChunkNum(), versionedParams.GetMaxSegmentSize(), nil
}

// ComputeHashRoots return the integrity hash, content size and the redundancy type of the file
func (c *client) ComputeHashRoots(reader io.Reader) ([][]byte, int64, storageTypes.RedundancyType, error) {
	dataBlocks, parityBlocks, segSize, err := c.GetRedundancyParams()
	if reader == nil {
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, errors.New("fail to compute hash, reader is nil")
	}
	if err != nil {
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}

	return hashlib.ComputeIntegrityHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks))
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *client) CreateObject(ctx context.Context, bucketName, objectName string,
	reader io.Reader, opts types.CreateObjectOptions,
) (string, error) {
	if reader == nil {
		return "", errors.New("fail to compute hash of payload, reader is nil")
	}

	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
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

	createObjectMsg := storageTypes.NewMsgCreateObject(c.MustGetDefaultAccount().GetAddress(), bucketName, objectName,
		uint64(size), visibility, expectCheckSums, contentType, redundancyType, math.MaxUint, nil, opts.SecondarySPAccs)
	err = createObjectMsg.ValidateBasic()
	if err != nil {
		return "", err
	}

	signedCreateObjectMsg, err := c.GetCreateObjectApproval(ctx, createObjectMsg)
	if err != nil {
		return "", err
	}

	// set the default txn broadcast mode as block mode
	if opts.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opts.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}

	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{signedCreateObjectMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, err
}

// DeleteObject send DeleteBucket txn to greenfield chain and return txn hash
func (c *client) DeleteObject(ctx context.Context, bucketName, objectName string, opt types.DeleteObjectOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	delObjectMsg := storageTypes.NewMsgDeleteObject(c.MustGetDefaultAccount().GetAddress(), bucketName, objectName)
	return c.sendTxn(ctx, delObjectMsg, opt.TxOpts)
}

// CancelCreateObject send CancelCreateObject txn to greenfield chain
func (c *client) CancelCreateObject(ctx context.Context, bucketName, objectName string, opt types.CancelCreateOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	cancelCreateMsg := storageTypes.NewMsgCancelCreateObject(c.MustGetDefaultAccount().GetAddress(), bucketName, objectName)
	return c.sendTxn(ctx, cancelCreateMsg, opt.TxOpts)
}

// PutObject supports the second stage of uploading the object to bucket.
// txnHash should be the str which hex.encoding from txn hash bytes
func (c *client) PutObject(ctx context.Context, bucketName, objectName string, objectSize int64,
	reader io.Reader, opts types.PutObjectOptions,
) (err error) {
	if objectSize <= 0 {
		return errors.New("object size should be more than 0")
	}

	var contentType string
	if opts.ContentType != "" {
		contentType = opts.ContentType
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

	var sendOpt sendOptions
	if opts.TxnHash != "" {
		sendOpt = sendOptions{
			method:  http.MethodPut,
			body:    reader,
			txnHash: opts.TxnHash,
		}
	} else {
		sendOpt = sendOptions{
			method: http.MethodPut,
			body:   reader,
		}
	}

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %s", bucketName, err.Error()))
		return err
	}

	_, err = c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return err
	}

	return nil
}

// FPutObject supports uploading object from local file
func (c *client) FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts types.PutObjectOptions) (err error) {
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

	return c.PutObject(ctx, bucketName, objectName, stat.Size(), fReader, opts)
}

// GetObject download s3 object payload and return the related object info
func (c *client) GetObject(ctx context.Context, bucketName, objectName string,
	opts types.GetObjectOption,
) (io.ReadCloser, types.ObjectStat, error) {
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
		reqMeta.rangeInfo = opts.Range
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed,  err: %s", bucketName, err.Error()))
		return nil, types.ObjectStat{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
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
func (c *client) FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts types.GetObjectOption) error {
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

	body, _, err := c.GetObject(ctx, bucketName, objectName, opts)
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
			return types.ObjectStat{}, types.ErrResponse{
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
	statements []*permTypes.Statement, opt types.PutPolicyOption,
) (string, error) {
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)

	principal := &permTypes.Principal{}
	if err := principal.Unmarshal([]byte(principalStr)); err != nil {
		return "", err
	}

	putPolicyMsg := storageTypes.NewMsgPutPolicy(c.MustGetDefaultAccount().GetAddress(), resource.String(),
		principal, statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(ctx, putPolicyMsg, opt.TxOpts)
}

// DeleteObjectPolicy delete the object policy of the principal
// principalAddr indicates the HEX-encoded string of the principal address
func (c *client) DeleteObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr string, opt types.DeletePolicyOption) (string, error) {
	addr, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return "", err
	}

	principal := permTypes.NewPrincipalWithAccount(addr)
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	return c.sendDelPolicyTxn(ctx, c.MustGetDefaultAccount().GetAddress(), resource.String(), principal, opt.TxOpts)
}

// IsObjectPermissionAllowed check if the permission of the object is allowed to the user
func (c *client) IsObjectPermissionAllowed(ctx context.Context, userAddr string,
	bucketName, objectName string, action permTypes.ActionType,
) (permTypes.Effect, error) {
	_, err := sdk.AccAddressFromHexUnsafe(userAddr)
	if err != nil {
		return permTypes.EFFECT_DENY, err
	}
	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   userAddr,
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
func (c *client) GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr string) (*permTypes.Policy, error) {
	_, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return nil, err
	}

	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	queryPolicy := storageTypes.QueryPolicyForAccountRequest{
		Resource:         resource.String(),
		PrincipalAddress: principalAddr,
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// ListObjects return object list of the specific bucket
func (c *client) ListObjects(ctx context.Context, bucketName, maxKeys, startAfter, continuationToken, delimiter, prefix string, opts types.ListObjectsOptions) (types.ListObjectsResult, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return types.ListObjectsResult{}, err
	}

	if maxKeys != "" {
		if maxKeysVal, err := utils.StringToUint64(maxKeys); err != nil || maxKeysVal == 0 {
			return types.ListObjectsResult{}, err
		}
	}

	if startAfter != "" {
		if err := s3util.CheckValidObjectName(startAfter); err != nil {
			return types.ListObjectsResult{}, err
		}
	}

	if continuationToken != "" {
		decodedContinuationToken, err := base64.StdEncoding.DecodeString(continuationToken)
		if err != nil {
			return types.ListObjectsResult{}, err
		}
		continuationToken = string(decodedContinuationToken)

		if err = s3util.CheckValidObjectName(continuationToken); err != nil {
			return types.ListObjectsResult{}, err
		}

		if !strings.HasPrefix(continuationToken, prefix) {
			return types.ListObjectsResult{}, fmt.Errorf("continuationToken does not match the input prefix")
		}
	}

	if ok := utils.IsValidObjectPrefix(prefix); !ok {
		return types.ListObjectsResult{}, fmt.Errorf("invalid object prefix")
	}

	params := url.Values{}
	params.Set("max-keys", maxKeys)
	params.Set("start-after", startAfter)
	params.Set("continuation-token", continuationToken)
	params.Set("delimiter", delimiter)
	params.Set("prefix", prefix)
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
		return types.ListObjectsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.ListObjectsResult{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return types.ListObjectsResult{}, err
	}

	listObjectsResult := types.ListObjectsResult{}
	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &listObjectsResult)
	// TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil && listObjectsResult.Objects == nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return types.ListObjectsResult{}, err
	}

	if opts.ShowRemovedObject {
		return listObjectsResult, nil
	}

	// default only return the object that has not been removed
	objectMetaList := make([]*types.ObjectMeta, 0)
	for _, objectInfo := range listObjectsResult.Objects {
		if objectInfo.Removed {
			continue
		}

		objectMetaList = append(objectMetaList, objectInfo)
	}

	listObjectsResult.Objects = objectMetaList
	listObjectsResult.KeyCount = uint64(len(objectMetaList))
	return listObjectsResult, nil
}

// GetCreateObjectApproval returns the signature info for the approval of preCreating resources
func (c *client) GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject) (*storageTypes.MsgCreateObject, error) {
	unsignedBytes := createObjectMsg.GetSignBytes()

	// set the action type
	urlValues := url.Values{
		"action": {types.CreateObjectAction},
	}

	reqMeta := requestMeta{
		urlValues:     urlValues,
		urlRelPath:    "get-approval",
		contentSHA256: types.EmptyStringSHA256,
		txnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	bucketName := createObjectMsg.BucketName
	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %s", bucketName, err.Error()))
		return nil, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
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

// CreateFolder send create empty object txn to greenfield chain
func (c *client) CreateFolder(ctx context.Context, bucketName, objectName string, opts types.CreateObjectOptions) (string, error) {
	if !strings.HasSuffix(objectName, "/") {
		return "", errors.New("failed to create folder. Folder names must end with a forward slash (/) character")
	}

	reader := bytes.NewReader([]byte(``))
	txHash, err := c.CreateObject(ctx, bucketName, objectName, reader, opts)
	return txHash, err
}

// GetObjectUploadProgress return the status of object including the uploading progress
func (c *client) GetObjectUploadProgress(ctx context.Context, bucketName, objectName string) (string, error) {
	status, err := c.HeadObject(ctx, bucketName, objectName)
	if err != nil {
		return "", err
	}

	// get object status from sp
	if status.ObjectStatus == storageTypes.OBJECT_STATUS_CREATED &&
		status.ObjectStatus != storageTypes.OBJECT_STATUS_SEALED {
		uploadProgressInfo, err := c.getObjectStatusFromSP(ctx, bucketName, objectName)
		if err != nil {
			return "", errors.New("fail to fetch object uploading progress from sp" + err.Error())
		}
		return uploadProgressInfo.ProgressDescription, nil
	}

	return status.ObjectStatus.String(), nil
}

func (c *client) getObjectStatusFromSP(ctx context.Context, bucketName, objectName string) (types.UploadProgress, error) {
	params := url.Values{}
	params.Set("upload-progress", "")

	reqMeta := requestMeta{
		urlValues:     params,
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		return types.UploadProgress{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.UploadProgress{}, err
	}

	defer utils.CloseResponse(resp)

	objectStatus := types.UploadProgress{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&objectStatus)
	if err != nil {
		return types.UploadProgress{}, err
	}

	return objectStatus, nil
}
