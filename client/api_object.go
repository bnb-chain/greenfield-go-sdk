package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
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
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/rs/zerolog/log"

	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdsdk "github.com/bnb-chain/greenfield/sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/s3util"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// IObjectClient interface defines functions related to object operations.
// The concept of "object" is the same as the concept of the object in AWS S3 storage.
type IObjectClient interface {
	GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject) (*storageTypes.MsgCreateObject, error)
	CreateObject(ctx context.Context, bucketName, objectName string, reader io.Reader, opts types.CreateObjectOptions) (string, error)
	UpdateObjectContent(ctx context.Context, bucketName, objectName string, reader io.Reader, opts types.UpdateObjectOptions) (string, error)
	CancelUpdateObjectContent(ctx context.Context, bucketName, objectName string, opts types.CancelUpdateObjectOption) (string, error)
	PutObject(ctx context.Context, bucketName, objectName string, objectSize int64, reader io.Reader, opts types.PutObjectOptions) error
	DelegatePutObject(ctx context.Context, bucketName, objectName string, objectSize int64, reader io.Reader, opts types.PutObjectOptions) error
	DelegateUpdateObjectContent(ctx context.Context, bucketName, objectName string, objectSize int64, reader io.Reader, opts types.PutObjectOptions) error
	FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts types.PutObjectOptions) (err error)
	CancelCreateObject(ctx context.Context, bucketName, objectName string, opt types.CancelCreateOption) (string, error)
	DeleteObject(ctx context.Context, bucketName, objectName string, opt types.DeleteObjectOption) (string, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts types.GetObjectOptions) (io.ReadCloser, types.ObjectStat, error)
	FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts types.GetObjectOptions) error
	FGetObjectResumable(ctx context.Context, bucketName, objectName, filePath string, opts types.GetObjectOptions) error
	HeadObject(ctx context.Context, bucketName, objectName string) (*types.ObjectDetail, error)
	HeadObjectByID(ctx context.Context, objID string) (*types.ObjectDetail, error)
	UpdateObjectVisibility(ctx context.Context, bucketName, objectName string, visibility storageTypes.VisibilityType, opt types.UpdateObjectOption) (string, error)
	PutObjectPolicy(ctx context.Context, bucketName, objectName string, principal types.Principal,
		statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	DeleteObjectPolicy(ctx context.Context, bucketName, objectName string, principal types.Principal, opt types.DeletePolicyOption) (string, error)
	GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr string) (*permTypes.Policy, error)
	IsObjectPermissionAllowed(ctx context.Context, userAddr string, bucketName, objectName string, action permTypes.ActionType) (permTypes.Effect, error)
	ListObjects(ctx context.Context, bucketName string, opts types.ListObjectsOptions) (types.ListObjectsResult, error)
	ComputeHashRoots(reader io.Reader, isSerial bool) ([][]byte, int64, storageTypes.RedundancyType, error)
	CreateFolder(ctx context.Context, bucketName, objectName string, opts types.CreateObjectOptions) (string, error)
	GetObjectUploadProgress(ctx context.Context, bucketName, objectName string) (string, error)
	ListObjectsByObjectID(ctx context.Context, objectIds []uint64, opts types.EndPointOptions) (types.ListObjectsByObjectIDResponse, error)
	ListObjectPolicies(ctx context.Context, objectName, bucketName string, actionType uint32, opts types.ListObjectPoliciesOptions) (types.ListObjectPoliciesResponse, error)
}

// GetRedundancyParams query and return the data shards, parity shards and segment size of redundancy
// configuration on chain
func (c *Client) GetRedundancyParams() (uint32, uint32, uint64, error) {
	query := storageTypes.QueryParamsRequest{}
	queryResp, err := c.chainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return 0, 0, 0, err
	}

	versionedParams := queryResp.Params.VersionedParams
	return versionedParams.GetRedundantDataChunkNum(), versionedParams.GetRedundantParityChunkNum(), versionedParams.GetMaxSegmentSize(), nil
}

// GetParams query and return the data shards, parity shards and segment size of redundancy
// configuration on chain
func (c *Client) GetParams() (storageTypes.Params, error) {
	query := storageTypes.QueryParamsRequest{}
	queryResp, err := c.chainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return storageTypes.Params{}, err
	}

	return queryResp.Params, nil
}

// ComputeHashRoots return the integrity hash, content size and the redundancy type of the file
func (c *Client) ComputeHashRoots(reader io.Reader, isSerial bool) ([][]byte, int64, storageTypes.RedundancyType, error) {
	dataBlocks, parityBlocks, segSize, err := c.GetRedundancyParams()
	if reader == nil {
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, errors.New("fail to compute hash, reader is nil")
	}
	if err != nil {
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}

	return hashlib.ComputeIntegrityHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks), isSerial)
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain,
// it returns the transaction hash value and error
func (c *Client) CreateObject(ctx context.Context, bucketName, objectName string,
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
	expectCheckSums, size, redundancyType, err := c.ComputeHashRoots(reader, opts.IsSerialComputeMode)
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
		uint64(size), visibility, expectCheckSums, contentType, redundancyType, math.MaxUint, nil)

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
	msgs := []sdk.Msg{signedCreateObjectMsg}

	if opts.Tags != nil {
		// Set tag
		grn := gnfdTypes.NewObjectGRN(bucketName, objectName)
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
			return txnHash, fmt.Errorf("the createObject txn has failed with response code: %d, codespace:%s", txnResponse.TxResult.Code, txnResponse.TxResult.Codespace)
		}
	}
	return txnHash, nil
}

// UpdateObjectContent sends updateObjectContent tx to greenfield chain,
// it returns the transaction hash value and error
func (c *Client) UpdateObjectContent(ctx context.Context, bucketName, objectName string,
	reader io.Reader, opts types.UpdateObjectOptions,
) (string, error) {
	if reader == nil {
		return "", errors.New("fail to compute hash of payload, reader is nil")
	}
	object, err := c.HeadObject(ctx, bucketName, objectName)
	if err != nil {
		return "", err
	}
	if object.ObjectInfo.ObjectStatus != storageTypes.OBJECT_STATUS_SEALED {
		return "", errors.New("object not sealed can not be updated")
	}
	// compute hash root of payload
	expectCheckSums, size, _, err := c.ComputeHashRoots(reader, opts.IsSerialComputeMode)
	if err != nil {
		return "", err
	}
	updateObjectContentMsg := storageTypes.NewMsgUpdateObjectContent(c.MustGetDefaultAccount().GetAddress(), bucketName, objectName,
		uint64(size), expectCheckSums)
	if opts.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opts.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{updateObjectContentMsg}, opts.TxOpts)
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
			return txnHash, fmt.Errorf("the updateObjectContent txn has failed with response code: %d, codespace:%s", txnResponse.TxResult.Code, txnResponse.TxResult.Codespace)
		}
	}
	return txnHash, nil
}

// CancelUpdateObjectContent sends CancelUpdateObjectContent tx to greenfield chain,
// it returns the transaction hash value and error
func (c *Client) CancelUpdateObjectContent(ctx context.Context, bucketName, objectName string, opts types.CancelUpdateObjectOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	msg := storageTypes.NewMsgCancelUpdateObjectContent(c.MustGetDefaultAccount().GetAddress(), bucketName, objectName)
	return c.sendTxn(ctx, msg, opts.TxOpts)
}

// DeleteObject - Send DeleteObject msg to greenfield chain and return txn hash.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The name of the bucket which contain the object.
//
// - objectName: The name of the object to be deleted.
//
// - opt: The Options for customizing the DeleteObject transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error if delete bucket failed, otherwise return nil.
func (c *Client) DeleteObject(ctx context.Context, bucketName, objectName string, opt types.DeleteObjectOption) (string, error) {
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
func (c *Client) CancelCreateObject(ctx context.Context, bucketName, objectName string, opt types.CancelCreateOption) (string, error) {
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
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, objectSize int64,
	reader io.Reader, opts types.PutObjectOptions,
) (err error) {
	if objectSize <= 0 {
		return errors.New("object size should be more than 0")
	}
	params, err := c.GetParams()
	if err != nil {
		return err
	}
	// minPartSize: 16MB
	if opts.PartSize == 0 {
		opts.PartSize = types.MinPartSize
	}
	if opts.PartSize%params.GetMaxSegmentSize() != 0 {
		return errors.New("part size should be an integer multiple of the segment size")
	}

	// upload an entire object to the storage provider in a single request
	if objectSize <= int64(opts.PartSize) || opts.DisableResumable {
		return c.putObject(ctx, bucketName, objectName, objectSize, reader, opts)
	}

	// resumableupload
	return c.putObjectResumable(ctx, bucketName, objectName, objectSize, reader, opts)
}

func (c *Client) putObject(ctx context.Context, bucketName, objectName string, objectSize int64,
	reader io.Reader, opts types.PutObjectOptions,
) (err error) {
	if err := c.headSPObjectInfo(ctx, bucketName, objectName); err != nil {
		log.Error().Msg(fmt.Sprintf("fail to head object %s , err %v ", objectName, err))
		return err
	}

	var contentType string
	if opts.ContentType != "" {
		contentType = opts.ContentType
	} else {
		contentType = types.ContentDefault
	}
	urlValues := make(url.Values)
	if opts.Delegated {
		urlValues.Set("delegate", "")
		urlValues.Set("is_update", strconv.FormatBool(opts.IsUpdate))
		urlValues.Set("payload_size", strconv.FormatInt(objectSize, 10))
		urlValues.Set("visibility", strconv.FormatInt(int64(opts.Visibility), 10))
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: types.EmptyStringSHA256,
		contentLength: objectSize,
		contentType:   contentType,
		urlValues:     urlValues,
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

// UploadSegmentHook is for testing usage
type uploadSegmentHook func(id int) error

var UploadSegmentHooker uploadSegmentHook = DefaultUploadSegment

func DefaultUploadSegment(id int) error {
	return nil
}

func (c *Client) putObjectResumable(ctx context.Context, bucketName, objectName string, objectSize int64,
	reader io.Reader, opts types.PutObjectOptions,
) (err error) {
	if err := c.headSPObjectInfo(ctx, bucketName, objectName); err != nil {
		return err
	}

	offset, err := c.getObjectResumableUploadOffset(ctx, bucketName, objectName)
	if err != nil {
		return err
	}

	// Total data read and written to server. should be equal to
	// 'size' at the end of the call.
	var totalUploadedSize int64

	// Calculate the optimal parts info for a given size.
	totalPartsCount, partSize, _, err := c.SplitPartInfo(objectSize, opts.PartSize)
	if err != nil {
		return err
	}

	// Part number always starts with '1'.
	partNumber := 1
	startPartNumber := int(offset/opts.PartSize + 1)

	// Create a buffer.
	buf := make([]byte, partSize)
	complete := false

	//  TODO(chris): Skip successful segments or add a verification file check.
	for partNumber < startPartNumber {
		length, rErr := utils.ReadFull(reader, buf)
		if rErr == io.EOF && partNumber > 1 {
			break
		}
		// Increment part number.
		log.Debug().Msg(fmt.Sprintf("skip partNumber:%d, length:%d", partNumber, length))
		// Save successfully uploaded size.
		totalUploadedSize += int64(length)
		partNumber++
	}

	for partNumber <= totalPartsCount {
		if partNumber == totalPartsCount {
			complete = true
		}
		if err = UploadSegmentHooker(partNumber); err != nil {
			return err
		}
		length, rErr := utils.ReadFull(reader, buf)
		if rErr == io.EOF && partNumber > 1 {
			break
		}

		if rErr != nil && rErr != io.ErrUnexpectedEOF && rErr != io.EOF {
			return err
		}

		log.Debug().Msg(fmt.Sprintf("partNumber:%d, length:%d", partNumber, length))

		// Update progress reader appropriately to the latest offset
		// as we read from the source.
		rd := bytes.NewReader(buf[:length])

		var contentType string
		if opts.ContentType != "" {
			contentType = opts.ContentType
		} else {
			contentType = types.ContentDefault
		}

		// Initialize url queries.
		urlValues := make(url.Values)
		urlValues.Set("offset", strconv.FormatInt(totalUploadedSize, 10))
		urlValues.Set("complete", strconv.FormatBool(complete))
		if opts.Delegated {
			urlValues.Set("delegate", "")
			urlValues.Set("is_update", strconv.FormatBool(opts.IsUpdate))
		}
		reqMeta := requestMeta{
			bucketName:    bucketName,
			objectName:    objectName,
			contentLength: int64(length),
			contentType:   contentType,
			urlValues:     urlValues,
		}

		var sendOpt sendOptions
		if opts.TxnHash != "" {
			sendOpt = sendOptions{
				method:  http.MethodPost,
				body:    rd,
				txnHash: opts.TxnHash,
			}
		} else {
			sendOpt = sendOptions{
				method: http.MethodPost,
				body:   rd,
			}
		}

		endpoint, err := c.getSPUrlByBucket(bucketName)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %s", bucketName, err.Error()))
			return err
		}

		// Proceed to upload the part.
		_, err = c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
		if err != nil {
			return err
		}

		// Save successfully uploaded size.
		totalUploadedSize += int64(length)

		// Increment part number.
		partNumber++

		// For unknown size, Read EOF we break away.
		// We do not have to upload till totalPartsCount.
		if rErr == io.EOF {
			break
		}
	}

	return nil
}

func (c *Client) headSPObjectInfo(ctx context.Context, bucketName, objectName string) error {
	backoffDelay := types.HeadBackOffDelay
	for retry := 0; retry < types.MaxHeadTryTime; retry++ {
		_, err := c.getObjectStatusFromSP(ctx, bucketName, objectName)
		if err == nil {
			return nil
		}
		// if the error is not "no such object", ignore it
		if !strings.Contains(strings.ToLower(err.Error()), types.NoSuchObjectErr) {
			return nil
		}

		if retry == types.MaxHeadTryTime-1 {
			return fmt.Errorf(" sp failed to head info of the object: %s, please try putObject later", objectName)
		}

		time.Sleep(backoffDelay)
		backoffDelay *= 2
	}

	return nil
}

// FPutObject supports uploading object from local file
func (c *Client) FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts types.PutObjectOptions) (err error) {
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
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string,
	opts types.GetObjectOptions,
) (io.ReadCloser, types.ObjectStat, error) {
	var err error
	if err = s3util.CheckValidBucketName(bucketName); err != nil {
		return nil, types.ObjectStat{}, err
	}

	if err = s3util.CheckValidObjectName(objectName); err != nil {
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

	var endpoint *url.URL

	if c.forceToUseSpecifiedSpEndpointForDownloadOnly != nil {
		endpoint = c.forceToUseSpecifiedSpEndpointForDownloadOnly
	} else {
		endpoint, err = c.getSPUrlByBucket(bucketName)

		if err != nil {
			log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed,  err: %s", bucketName, err.Error()))
			return nil, types.ObjectStat{}, err
		}
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
func (c *Client) FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts types.GetObjectOptions) error {
	// Verify if destination already exists.
	st, err := os.Stat(filePath)
	if err == nil {
		// If the destination exists and is a directory.
		if st.IsDir() {
			return errors.New("download file path is a directory")
		}
		return errors.New("download file already exist")
	}

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

// getSegmentEnd calculates the end position
func getSegmentEnd(begin int64, total int64, per int64) int64 {
	if begin+per > total {
		return total - 1
	}
	return begin + per - 1
}

// FGetObjectResumable download s3 object payload with resumable download
func (c *Client) FGetObjectResumable(ctx context.Context, bucketName, objectName, filePath string, opts types.GetObjectOptions) error {
	// Get the object detailed meta for object whole size
	meta, err := c.HeadObject(ctx, bucketName, objectName)
	if err != nil {
		return err
	}

	tempFilePath := filePath + "_" + c.defaultAccount.GetAddress().String() + opts.Range + types.TempFileSuffix

	var (
		startOffset    int64
		endOffset      int64
		partEndOffset  int64
		maxSegmentSize int64
		objectOption   types.GetObjectOptions
		segNum         int64
		partSize       int64
	)

	// 1) check paramter
	params, err := c.GetParams()
	if err != nil {
		return err
	}
	maxSegmentSize = int64(params.GetMaxSegmentSize())

	// minPartSize: 32MB
	if opts.PartSize == 0 {
		partSize = types.MinPartSize
	} else {
		partSize = int64(opts.PartSize)
	}

	if partSize%maxSegmentSize != 0 {
		return errors.New("part size should be an integer multiple of the segment size")
	}

	isRange, rangeStart, rangeEnd := utils.ParseRange(opts.Range)
	if isRange && (rangeEnd < 0 || rangeEnd >= int64(meta.ObjectInfo.GetPayloadSize())) {
		rangeEnd = int64(meta.ObjectInfo.GetPayloadSize()) - 1
	}

	if isRange {
		startOffset = rangeStart
		endOffset = rangeEnd
	} else {
		startOffset = 0
		endOffset = int64(meta.ObjectInfo.GetPayloadSize()) - 1
	}

	// 2)prepare and check temp file
	fileInfo, err := os.Stat(tempFilePath)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			return err
		}
	} else {
		/*
				| -----seg1 ------ | -----seg2 ------ | -----seg3 ------ | -----seg4 ------ |
				        ^                                                         ^
				        |													      |
				 startOffset													  endOffset
			            |----------|
		*/
		fileSize := fileInfo.Size()
		var (
			firstSegSize   int64
			file           *os.File
			truncateOffset int64
		)

		if isRange {
			firstSegSize = partSize - rangeStart%partSize
		} else {
			firstSegSize = partSize
		}
		fileSizeWithoutFirstSeg := fileSize - firstSegSize
		if fileSize > firstSegSize {
			truncateOffset = ((fileSize-firstSegSize)/partSize)*partSize + firstSegSize
			startOffset = ((fileSize-firstSegSize)/partSize)*partSize + partSize
		} else {
			truncateOffset = 0
			startOffset = 0
		}
		log.Debug().Msgf("The file:%s size:%d, startOffset:%d, range:%s\n", tempFilePath, fileSize, startOffset, opts.Range)

		// truncated file to part size integer multiples
		if fileSizeWithoutFirstSeg%partSize != 0 {
			file, err = os.OpenFile(tempFilePath, os.O_RDWR, 0o644)
			if err != nil {
				return err
			}
			defer file.Close()

			err = file.Truncate(truncateOffset)
			if err != nil {
				return err
			}
			log.Debug().Msgf("The file was truncated to the specified size.%d\n", truncateOffset)
			// TODO(chris): verify file's segment
		}
	}

	// Create the file if not exists. Otherwise the segments download will overwrite it.
	fd, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, types.FilePermMode)
	if err != nil {
		return err
	}
	_, err = fd.Seek(startOffset, io.SeekStart)
	if err != nil {
		fd.Close()
		return err
	}

	log.Debug().Msg(fmt.Sprintf("get object resumeable begin segment Range: %s, startOffset: %d, endOffset:%d", opts.Range, startOffset, endOffset))

	// 3) Downloading Parts Sequentially based on partSize
	segNum = startOffset / partSize
	for partStartOffset := startOffset; partStartOffset < endOffset; partStartOffset += partSize {
		// hook for test
		if err = DownloadSegmentHooker(segNum); err != nil {
			return err
		}

		partEndOffset = getSegmentEnd(partStartOffset, endOffset+1, partSize)
		err = objectOption.SetRange(partStartOffset, partEndOffset)
		if err != nil {
			return err
		}

		startT := time.Now().UnixNano() / 1000 / 1000 / 1000

		rd, _, err := c.GetObject(ctx, bucketName, objectName, objectOption)
		if err != nil {
			return err
		}
		defer rd.Close()

		_, err = io.Copy(fd, rd)
		log.Debug().Msg(fmt.Sprintf("get object for segment Range: %s, current partStartOffset: %d, segNum: %d", objectOption.Range, partStartOffset, segNum))
		endT := time.Now().UnixNano() / 1000 / 1000 / 1000
		if err != nil {
			log.Error().Msg(fmt.Sprintf("get seg error,cost:%d second,seg number:%d,error:%s.\n", endT-startT, segNum, err.Error()))
			fd.Close()
		}

		segNum++
	}

	fd.Close()

	// 4) rename temp file
	err = os.Rename(tempFilePath, filePath)
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
func (c *Client) HeadObject(ctx context.Context, bucketName, objectName string) (*types.ObjectDetail, error) {
	queryHeadObjectRequest := storageTypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := c.chainClient.HeadObject(ctx, &queryHeadObjectRequest)
	if err != nil {
		return nil, err
	}

	return &types.ObjectDetail{
		ObjectInfo:         queryHeadObjectResponse.ObjectInfo,
		GlobalVirtualGroup: queryHeadObjectResponse.GlobalVirtualGroup,
	}, nil
}

// HeadObjectByID query the objectInfo on chain by object id, return the object info if exists
// return err info if object not exist
func (c *Client) HeadObjectByID(ctx context.Context, objID string) (*types.ObjectDetail, error) {
	headObjectRequest := storageTypes.QueryHeadObjectByIdRequest{
		ObjectId: objID,
	}
	queryHeadObjectResponse, err := c.chainClient.HeadObjectById(ctx, &headObjectRequest)
	if err != nil {
		return nil, err
	}

	return &types.ObjectDetail{
		ObjectInfo:         queryHeadObjectResponse.ObjectInfo,
		GlobalVirtualGroup: queryHeadObjectResponse.GlobalVirtualGroup,
	}, nil
}

// PutObjectPolicy apply object policy to the principal, return the txn hash
func (c *Client) PutObjectPolicy(ctx context.Context, bucketName, objectName string, principalStr types.Principal,
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
func (c *Client) DeleteObjectPolicy(ctx context.Context, bucketName, objectName string, principalStr types.Principal, opt types.DeletePolicyOption) (string, error) {
	principal := &permTypes.Principal{}
	if err := principal.Unmarshal([]byte(principalStr)); err != nil {
		return "", err
	}

	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	return c.sendDelPolicyTxn(ctx, c.MustGetDefaultAccount().GetAddress(), resource.String(), principal, opt.TxOpts)
}

// IsObjectPermissionAllowed check if the permission of the object is allowed to the user
func (c *Client) IsObjectPermissionAllowed(ctx context.Context, userAddr string,
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
func (c *Client) GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr string) (*permTypes.Policy, error) {
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

// ListObjects - Lists the object info of the bucket. If opts.ShowRemovedObject set to false, these objects will be skipped.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - opts: The options to set the meta to list the objects
//
// - ret1: The result of list objects under specific bucket
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListObjects(ctx context.Context, bucketName string, opts types.ListObjectsOptions) (types.ListObjectsResult, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return types.ListObjectsResult{}, err
	}

	const listObjectsDefaultMaxKeys = 1000
	if opts.MaxKeys == 0 {
		opts.MaxKeys = listObjectsDefaultMaxKeys
	}

	if opts.StartAfter != "" {
		if err := s3util.CheckValidObjectName(opts.StartAfter); err != nil {
			return types.ListObjectsResult{}, err
		}
	}

	if opts.ContinuationToken != "" {
		decodedContinuationToken, err := base64.StdEncoding.DecodeString(opts.ContinuationToken)
		if err != nil {
			return types.ListObjectsResult{}, err
		}
		objectName := string(decodedContinuationToken)
		if err = s3util.CheckValidObjectName(objectName); err != nil {
			return types.ListObjectsResult{}, err
		}
		if !strings.HasPrefix(objectName, opts.Prefix) {
			return types.ListObjectsResult{}, fmt.Errorf("continuation-token does not match the input prefix")
		}
	}

	if ok := utils.IsValidObjectPrefix(opts.Prefix); !ok {
		return types.ListObjectsResult{}, fmt.Errorf("invalid object prefix")
	}

	params := url.Values{}
	params.Set("max-keys", strconv.FormatUint(opts.MaxKeys, 10))
	params.Set("start-after", opts.StartAfter)
	params.Set("continuation-token", opts.ContinuationToken)
	params.Set("delimiter", opts.Delimiter)
	params.Set("prefix", opts.Prefix)
	params.Set("include-removed", strconv.FormatBool(opts.ShowRemovedObject))
	reqMeta := requestMeta{
		urlValues:     params,
		bucketName:    bucketName,
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
	err = xml.Unmarshal([]byte(bufStr), &listObjectsResult)
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
	listObjectsResult.KeyCount = strconv.Itoa(len(objectMetaList))
	return listObjectsResult, nil
}

// GetCreateObjectApproval returns the signature info for the approval of preCreating resources
func (c *Client) GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject) (*storageTypes.MsgCreateObject, error) {
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
		method: http.MethodGet,
		adminInfo: AdminAPIInfo{
			isAdminAPI:   true,
			adminVersion: types.AdminV1Version,
		},
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
func (c *Client) CreateFolder(ctx context.Context, bucketName, objectName string, opts types.CreateObjectOptions) (string, error) {
	if !strings.HasSuffix(objectName, "/") {
		return "", errors.New("failed to create folder. Folder names must end with a forward slash (/) character")
	}

	reader := bytes.NewReader([]byte(``))
	txHash, err := c.CreateObject(ctx, bucketName, objectName, reader, opts)
	return txHash, err
}

// GetObjectUploadProgress return the status of object including the uploading progress
func (c *Client) GetObjectUploadProgress(ctx context.Context, bucketName, objectName string) (string, error) {
	status, err := c.HeadObject(ctx, bucketName, objectName)
	if err != nil {
		return "", err
	}

	// get object status from sp
	if status.ObjectInfo.ObjectStatus == storageTypes.OBJECT_STATUS_CREATED {
		uploadProgressInfo, err := c.getObjectStatusFromSP(ctx, bucketName, objectName)
		if err != nil {
			return "", errors.New("fail to fetch object uploading progress from sp" + err.Error())
		}
		return uploadProgressInfo.ProgressDescription, nil
	}

	return status.ObjectInfo.ObjectStatus.String(), nil
}

// getObjectResumableUploadOffset return the status of object including the uploading progress
func (c *Client) getObjectResumableUploadOffset(ctx context.Context, bucketName, objectName string) (uint64, error) {
	status, err := c.HeadObject(ctx, bucketName, objectName)
	if err != nil {
		return 0, err
	}

	// get object status from sp
	if status.ObjectInfo.ObjectStatus == storageTypes.OBJECT_STATUS_CREATED {
		uploadOffsetInfo, err := c.getObjectOffsetFromSP(ctx, bucketName, objectName)
		if err != nil {
			return 0, errors.New("fail to fetch object uploading offset from sp" + err.Error())
		}
		log.Debug().Msgf("get object resumable upload offset %d from sp", uploadOffsetInfo.Offset)
		return uploadOffsetInfo.Offset, nil
	}

	// TODO(chris): may error
	return 0, nil
}

func (c *Client) getObjectOffsetFromSP(ctx context.Context, bucketName, objectName string) (types.UploadOffset, error) {
	params := url.Values{}
	params.Set("upload-context", "")

	reqMeta := requestMeta{
		urlValues:  params,
		bucketName: bucketName,
		objectName: objectName,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		return types.UploadOffset{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		// not exist
		if find := strings.Contains(err.Error(), "no uploading record"); find {
			return types.UploadOffset{Offset: 0}, nil
		} else {
			return types.UploadOffset{}, err
		}
	}

	defer utils.CloseResponse(resp)

	objectOffset := types.UploadOffset{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&objectOffset)
	if err != nil {
		return types.UploadOffset{}, err
	}

	return objectOffset, nil
}

func (c *Client) getObjectStatusFromSP(ctx context.Context, bucketName, objectName string) (types.UploadProgress, error) {
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

func (c *Client) UpdateObjectVisibility(ctx context.Context, bucketName, objectName string,
	visibility storageTypes.VisibilityType, opt types.UpdateObjectOption,
) (string, error) {
	object, err := c.HeadObject(ctx, bucketName, objectName)
	if err != nil {
		return "", fmt.Errorf("object:%s not exists: %s\n", objectName, err.Error())
	}

	if object.ObjectInfo.GetVisibility() == visibility {
		return "", fmt.Errorf("the visibility of object:%s is already %s \n", objectName, visibility.String())
	}

	updateObjectMsg := storageTypes.NewMsgUpdateObjectInfo(c.MustGetDefaultAccount().GetAddress(), bucketName, objectName, visibility)

	// set the default txn broadcast mode as sync mode
	if opt.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opt.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}

	return c.sendTxn(ctx, updateObjectMsg, opt.TxOpts)
}

type listObjectsByIDsResponse map[uint64]*types.ObjectMeta

type objectEntry struct {
	Id    uint64
	Value *types.ObjectMeta
}

func (m *listObjectsByIDsResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = listObjectsByIDsResponse{}
	for {
		var e objectEntry

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

// ListObjectsByObjectID - List objects by object ids. If opts.ShowRemovedObject set to false, these objects will be skipped.
//
// By inputting a collection of object IDs, we can retrieve the corresponding object data. If the object is nonexistent or has been deleted, a null value will be returned
//
// - ctx: Context variables for the current API call.
//
// - objectIds: The list of object ids.
//
// - opts: The options to set the meta to list objects by object id.
//
// - ret1: The result of object info map by given object ids.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListObjectsByObjectID(ctx context.Context, objectIds []uint64, opts types.EndPointOptions) (types.ListObjectsByObjectIDResponse, error) {
	const MaximumListObjectsSize = 100
	if len(objectIds) == 0 || len(objectIds) > MaximumListObjectsSize {
		return types.ListObjectsByObjectIDResponse{}, nil
	}

	objectIDMap := make(map[uint64]bool)
	for _, id := range objectIds {
		if _, ok := objectIDMap[id]; ok {
			// repeat id keys in request
			return types.ListObjectsByObjectIDResponse{}, nil
		}
		objectIDMap[id] = true
	}

	idStr := make([]string, len(objectIds))
	for i, id := range objectIds {
		idStr[i] = strconv.FormatUint(id, 10)
	}
	IDs := strings.Join(idStr, ",")

	params := url.Values{}
	params.Set("objects-query", "")
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
		return types.ListObjectsByObjectIDResponse{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.ListObjectsByObjectIDResponse{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msgf("the list of objects in object ids:%v failed: %s", objectIds, err.Error())
		return types.ListObjectsByObjectIDResponse{}, err
	}

	objects := types.ListObjectsByObjectIDResponse{}
	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), (*listObjectsByIDsResponse)(&objects.Objects))
	if err != nil && objects.Objects == nil {
		log.Error().Msgf("the list of objects in object ids:%v failed: %s", objectIds, err.Error())
		return types.ListObjectsByObjectIDResponse{}, err
	}

	return objects, nil
}

// ListObjectPolicies - List object policies by object info and action type.
//
// If the limit is set to 0, it will default to 50. If the limit exceeds 1000, only 1000 records will be returned.
//
// - ctx: Context variables for the current API call.
//
// - objectName: The object name identifies the object.
//
// - bucketName: The bucket name identifies the bucket.
//
//   - actionType: The action type defines the requested action type of permission.
//     | Value | Description                |
//     | ----- | -------------------------- |
//     | 0     | ACTION_UNSPECIFIED         |89
//     | 1     | ACTION_UPDATE_BUCKET_INFO  |
//     | 2     | ACTION_DELETE_BUCKET       |
//     | 3     | ACTION_CREATE_OBJECT       |
//     | 4     | ACTION_DELETE_OBJECT       |
//     | 5     | ACTION_COPY_OBJECT         |
//     | 6     | ACTION_GET_OBJECT          |
//     | 7     | ACTION_EXECUTE_OBJECT      |
//     | 8     | ACTION_LIST_OBJECT         |
//     | 9     | ACTION_UPDATE_GROUP_MEMBER |
//     | 10    | ACTION_DELETE_GROUP        |
//     | 11    | ACTION_UPDATE_OBJECT_INFO  |
//     | 12    | ACTION_UPDATE_GROUP_EXTRA  |
//     | 99    | ACTION_TYPE_ALL            |
//
// - opts: The options to set the meta to list object policies
//
// - ret1: The result of object policy meta
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListObjectPolicies(ctx context.Context, objectName, bucketName string, actionType uint32, opts types.ListObjectPoliciesOptions) (types.ListObjectPoliciesResponse, error) {
	params := url.Values{}
	params.Set("object-policies", "")
	// StartAfter is used to input the policy id for pagination purposes
	params.Set("start-after", opts.StartAfter)
	// If the limit is set to 0, it will default to 50.
	// If the limit exceeds 1000, only 1000 records will be returned.
	params.Set("limit", strconv.FormatInt(opts.Limit, 10))
	params.Set("action-type", strconv.FormatUint(uint64(actionType), 10))

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

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListObjectPoliciesResponse{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.ListObjectPoliciesResponse{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msgf("the list object policies in bucket name:%s, object name:%s failed: %s", bucketName, objectName, err.Error())
		return types.ListObjectPoliciesResponse{}, err
	}

	policies := types.ListObjectPoliciesResponse{}
	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), &policies)
	if err != nil {
		log.Error().Msgf("the list object policies in bucket name:%s, object name:%s failed: %s", bucketName, objectName, err.Error())
		return types.ListObjectPoliciesResponse{}, err
	}

	return policies, nil
}

func (c *Client) DelegatePutObject(ctx context.Context, bucketName, objectName string, objectSize int64,
	reader io.Reader, opts types.PutObjectOptions,
) (err error) {
	if objectSize <= 0 {
		return errors.New("object size should be more than 0")
	}
	params, err := c.GetParams()
	if err != nil {
		return err
	}
	opts.Delegated = true
	// minPartSize: 16MB
	if opts.PartSize == 0 {
		opts.PartSize = types.MinPartSize
	}
	if opts.PartSize%params.GetMaxSegmentSize() != 0 {
		return errors.New("part size should be an integer multiple of the segment size")
	}

	// upload an entire object to the storage provider in a single request
	if objectSize <= int64(opts.PartSize) || opts.DisableResumable {
		return c.putObject(ctx, bucketName, objectName, objectSize, reader, opts)
	}

	// resumableupload
	return c.putObjectResumable(ctx, bucketName, objectName, objectSize, reader, opts)
}

func (c *Client) DelegateUpdateObjectContent(ctx context.Context, bucketName, objectName string, objectSize int64,
	reader io.Reader, opts types.PutObjectOptions,
) (err error) {
	opts.IsUpdate = true
	return c.DelegatePutObject(ctx, bucketName, objectName, objectSize, reader, opts)
}
