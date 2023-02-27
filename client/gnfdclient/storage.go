package gnfdclient

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"

	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/bnb-chain/greenfield/sdk/types"
	storageType "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

// CreateBucketOptions indicates the meta to construct createBucket msg of storage module
type CreateBucketOptions struct {
	IsPublic       bool
	TxOpts         *types.TxOption
	PaymentAddress sdk.AccAddress
}

// CreateObjectOptions indicates the meta to construct createObject msg of storage module
type CreateObjectOptions struct {
	IsPublic        bool
	TxOpts          *types.TxOption
	SecondarySPAccs []sdk.AccAddress
	ContentType     string
}

// ComputeHashOptions  indicates the meta of redundancy strategy
type ComputeHashOptions struct {
	SegmentSize  uint64
	DataShards   uint32
	ParityShards uint32
}

type GnfdResponse struct {
	TxnHash string
	Err     error
	TxnType string
}

// CreateBucket get approval of creating bucket and send createBucket txn to greenfield chain
func (c *GnfdClient) CreateBucket(ctx context.Context, bucketName string, primarySPAddress sdk.AccAddress, opts CreateBucketOptions) GnfdResponse {
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreateBucket"}
	}

	createBucketMsg := storageType.NewMsgCreateBucket(km.GetAddr(), bucketName, opts.IsPublic, primarySPAddress, opts.PaymentAddress, 0, nil)

	err = createBucketMsg.ValidateBasic()
	if err != nil {
		return GnfdResponse{"", err, "CreateBucket"}
	}
	signedMsg, err := c.SPClient.GetCreateBucketApproval(ctx, createBucketMsg, sp.NewAuthInfo(false, ""))
	if err != nil {
		return GnfdResponse{"", err, "CreateBucket"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{signedMsg}, opts.TxOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateBucket"}
	}

	log.Info().Msg("get createBucket txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "CreateBucket"}
}

// DelBucket send DeleteBucket txn to greenfield chain and return txn hash
func (c *GnfdClient) DelBucket(bucketName string, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteBucket"}
	}
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "DeleteBucket"}
	}
	delBucketMsg := storageType.NewMsgDeleteBucket(km.GetAddr(), bucketName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "DeleteBucket"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "DeleteBucket"}
}

func (c *GnfdClient) ComputeHash(reader io.Reader, opts ComputeHashOptions) ([]string, int64, error) {
	var dataBlocks, parityBlocks uint32
	var segSize uint64
	if opts.DataShards == 0 || opts.ParityShards == 0 || opts.SegmentSize == 0 {
		query := storageType.QueryParamsRequest{}
		queryResp, err := c.ChainClient.StorageQueryClient.Params(context.Background(), &query)
		if err != nil {
			return nil, 0, err
		}
		dataBlocks = queryResp.Params.GetRedundantDataChunkNum()
		parityBlocks = queryResp.Params.GetRedundantParityChunkNum()
		segSize = queryResp.Params.GetMaxSegmentSize()
	} else {
		dataBlocks = opts.DataShards
		parityBlocks = opts.ParityShards
		segSize = opts.SegmentSize
	}

	log.Info().Msg(fmt.Sprintf("get segSize %d, DataShards: %d , ParityShards: %d", segSize, dataBlocks, parityBlocks))
	// get hash and objectSize from reader
	return hashlib.ComputerHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks))
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *GnfdClient) CreateObject(ctx context.Context, bucketName, objectName string,
	reader io.Reader, opts CreateObjectOptions) GnfdResponse {
	if reader == nil {
		return GnfdResponse{"", errors.New("fail to compute hash of payload, reader is nil"), "CreateObject"}
	}

	if err := utils.VerifyBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	if err := utils.VerifyObjectName(objectName); err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreateBucket"}
	}
	// compute hash root of payload
	pieceHashRoots, size, err := c.ComputeHash(reader, ComputeHashOptions{})
	if err != nil {
		log.Error().Msg("get hash roots fail" + err.Error())
		return GnfdResponse{"", err, "CreateObject"}
	}

	expectCheckSums := make([][]byte, len(pieceHashRoots))
	for index, hash := range pieceHashRoots {
		hashByte, err := hex.DecodeString(hash)
		if err != nil {
			return GnfdResponse{"", err, "CreateObject"}
		}
		expectCheckSums[index] = hashByte
	}

	var contentType string
	if opts.ContentType != "" {
		contentType = opts.ContentType
	} else {
		contentType = sp.ContentDefault
	}

	approveOpts := sp.ApproveObjectOptions{}
	if opts.SecondarySPAccs != nil {
		approveOpts.SecondarySPAccs = opts.SecondarySPAccs
	}

	approveOpts.IsPublic = opts.IsPublic

	createObjectMsg := storageType.NewMsgCreateObject(km.GetAddr(), bucketName, objectName, uint64(size), opts.IsPublic, expectCheckSums, contentType, math.MaxUint, nil, nil)
	err = createObjectMsg.ValidateBasic()
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	signedCreateObjectMsg, err := c.SPClient.GetCreateObjectApproval(ctx, createObjectMsg, sp.NewAuthInfo(false, ""))
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{signedCreateObjectMsg}, opts.TxOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	log.Info().Msg("get createObject txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "CreateObject"}
}

// DelObject send DeleteBucket txn to greenfield chain and return txn hash
func (c *GnfdClient) DelObject(bucketName, objectName string,
	txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteObject"}
	}
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "DeleteObject"}
	}

	if err := utils.VerifyObjectName(objectName); err != nil {
		return GnfdResponse{"", err, "DeleteObject"}
	}
	delObjectMsg := storageType.NewMsgDeleteObject(km.GetAddr(), bucketName, objectName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delObjectMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "DeleteObject"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "DeleteObject"}
}

// CancelCreateObject send CancelCreateObject txn to greenfield chain
func (c *GnfdClient) CancelCreateObject(bucketName, objectName string, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CancelCreateObject"}
	}
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "CancelCreateObject"}
	}

	if err := utils.VerifyObjectName(objectName); err != nil {
		return GnfdResponse{"", err, "CancelCreateObject"}
	}

	cancelCreateMsg := storageType.NewMsgCancelCreateObject(km.GetAddr(), bucketName, objectName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{cancelCreateMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "CancelCreateObject"}
	}

	log.Info().Msg("get txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "CancelCreateObject"}
}

// UploadObject upload payload of object to storage provider
func (c *GnfdClient) UploadObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
	reader io.Reader, opt sp.UploadOption) (res sp.UploadResult, err error) {
	return c.SPClient.PutObject(ctx, bucketName, objectName, txnHash,
		objectSize, reader, sp.NewAuthInfo(false, ""), opt)
}

// DownloadObject download the object from primary storage provider
func (c *GnfdClient) DownloadObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, sp.ObjectInfo, error) {
	return c.SPClient.GetObject(ctx, bucketName, objectName, sp.DownloadOption{}, sp.NewAuthInfo(false, ""))
}

// BuyQuotaForBucket increase the quota to reach storage service of Sender
func (c *GnfdClient) BuyQuotaForBucket(bucketName string,
	quota storageType.ReadQuota, paymentAcc sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "UpdateBucketInfo"}
	}
	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storageType.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := c.ChainClient.HeadBucket(ctx, &queryHeadBucketRequest)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	newQuota := queryHeadBucketResponse.BucketInfo.GetReadQuota() + quota
	updateBucketMsg := storageType.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, newQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "UpdateBucketInfo"}
}

// UpdateBucket update the bucket read quota on chain
func (c *GnfdClient) UpdateBucket(bucketName string,
	readQuota storageType.ReadQuota, paymentAcc sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "UpdateBucketInfo"}
	}

	updateBucketMsg := storageType.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, readQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	log.Info().Msg("get updateBucketInfo txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "UpdateBucketInfo"}
}

// BucketInfo represent the bucket basic meta on greenfield chain
type BucketInfo struct {
	BucketId uint64
	Owner    string
}

// HeadBucket query the bucketInfo on chain to check the bucketId
// if bucket exist, return true and the bucketId
func (c *GnfdClient) HeadBucket(bucketName string) (BucketInfo, error) {
	ctx := context.Background()
	queryHeadBucketRequest := storageType.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := c.ChainClient.HeadBucket(ctx, &queryHeadBucketRequest)
	if err != nil {
		return BucketInfo{}, err
	}

	info := queryHeadBucketResponse.BucketInfo
	return BucketInfo{
		BucketId: info.Id.Uint64(),
		Owner:    info.Owner,
	}, nil
}

// ObjectInfo represent the object basic meta on greenfield chain
type ObjectInfo struct {
	ObjectId uint64
	Status   string
	Size     uint64
}

// HeadObject query the objectInfo on chain to check th ObjectId
// if object exist, return true and the objectId
func (c *GnfdClient) HeadObject(bucketName, objectName string) (ObjectInfo, error) {
	ctx := context.Background()
	queryHeadObjectRequest := storageType.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := c.ChainClient.HeadObject(ctx, &queryHeadObjectRequest)
	if err != nil {
		return ObjectInfo{}, err
	}

	info := queryHeadObjectResponse.ObjectInfo
	return ObjectInfo{
		ObjectId: info.Id.Uint64(),
		Status:   info.GetObjectStatus().String(),
		Size:     info.GetPayloadSize(),
	}, nil
}
