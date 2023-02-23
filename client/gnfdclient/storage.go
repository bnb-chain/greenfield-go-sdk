package gnfdclient

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	lib "github.com/bnb-chain/greenfield-common/go"
	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/bnb-chain/greenfield/sdk/types"
	storage_type "github.com/bnb-chain/greenfield/x/storage/types"
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
	TxOpts         types.TxOption
	PaymentAddress sdk.AccAddress
}

// CreateObjectOptions indicates the meta to construct createObject msg of storage module
type CreateObjectOptions struct {
	IsPublic        bool
	TxOpts          types.TxOption
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
func (c *GnfdClient) CreateBucket(ctx context.Context, bucketName string, primarySPAddr sdk.AccAddress,
	opts CreateBucketOptions) GnfdResponse {
	approveOpts := sp.ApproveBucketOptions{}
	if opts.PaymentAddress != nil {
		approveOpts.PaymentAddress = opts.PaymentAddress
	}

	approveOpts.IsPublic = opts.IsPublic
	// get approval of creating bucket from sp
	signedCreateBucketMsg, err := c.SPClient.GetCreateBucketApproval(ctx, bucketName, primarySPAddr,
		sp.NewAuthInfo(false, ""), approveOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateBucket"}
	}

	decodedMsg, err := hex.DecodeString(signedCreateBucketMsg)
	if err != nil {
		return GnfdResponse{"", err, "CreateBucket"}
	}

	var signedMsg storage_type.MsgCreateBucket

	ModuleCdc.MustUnmarshalJSON(decodedMsg, &signedMsg)

	txOpts := types.TxOption{}
	if opts.TxOpts.Mode != nil {
		txOpts = opts.TxOpts
	}

	_, err = c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreatBucket"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{&signedMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateBucket"}
	}

	log.Info().Msg("get createBucket txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "CreateBucket"}
}

// DelBucket send DeleteBucket txn to chain
func (c *GnfdClient) DelBucket(operator sdk.AccAddress, bucketName string, txOpts types.TxOption) GnfdResponse {
	_, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteBucket"}
	}
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "DeleteBucket"}
	}
	delBucketMsg := storage_type.NewMsgDeleteBucket(operator, bucketName)

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
		query := storage_type.QueryParamsRequest{}
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
	return lib.ComputerHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks))
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *GnfdClient) CreateObject(ctx context.Context, bucketName, objectName string,
	reader io.Reader, opts CreateObjectOptions) GnfdResponse {
	if reader == nil {
		return GnfdResponse{"", errors.New("fail to compute hash of payload, reader is nil"), "CreateObject"}
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

	// get approval of creating bucket from sp
	signedCreateObjectMsg, err := c.SPClient.GetCreateObjectApproval(ctx, bucketName, objectName,
		contentType, uint64(size), expectCheckSums, sp.NewAuthInfo(false, ""), approveOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	decodedMsg, err := hex.DecodeString(signedCreateObjectMsg)
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}
	var signedMsg storage_type.MsgCreateObject
	ModuleCdc.MustUnmarshalJSON(decodedMsg, &signedMsg)

	txOpts := types.TxOption{}
	if opts.TxOpts.Mode != nil {
		txOpts = opts.TxOpts
	}

	_, err = c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreateObject"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{&signedMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	log.Info().Msg("get createObject txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "CreateObject"}
}

// DelObject send DeleteBucket txn to chain
func (c *GnfdClient) DelObject(operator sdk.AccAddress, bucketName, objectName string,
	txOpts types.TxOption) GnfdResponse {
	_, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteObject"}
	}
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "DeleteObject"}
	}

	if err := utils.IsValidObjectName(objectName); err != nil {
		return GnfdResponse{"", err, "DeleteObject"}
	}
	delObjectMsg := storage_type.NewMsgDeleteObject(operator, bucketName, objectName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delObjectMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "DeleteObject"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "DeleteObject"}
}

// CancelCreateObject send CancelCreateObject txn to chain
func (c *GnfdClient) CancelCreateObject(operator sdk.AccAddress, bucketName,
	objectName string, txOpts types.TxOption) GnfdResponse {
	_, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CancelCreateObject"}
	}
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "CancelCreateObject"}
	}

	if err := utils.IsValidObjectName(objectName); err != nil {
		return GnfdResponse{"", err, "CancelCreateObject"}
	}

	cancelCreateMsg := storage_type.NewMsgCancelCreateObject(operator, bucketName, objectName)

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
func (c *GnfdClient) BuyQuotaForBucket(operator sdk.AccAddress, bucketName string,
	quota storage_type.ReadQuota, paymentAcc sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	_, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "UpdateBucketInfo"}
	}
	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storage_type.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := c.ChainClient.HeadBucket(ctx, &queryHeadBucketRequest)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	newQuota := queryHeadBucketResponse.BucketInfo.GetReadQuota() + quota
	updateBucketMsg := storage_type.NewMsgUpdateBucketInfo(operator, bucketName, newQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "UpdateBucketInfo"}
}

func (c *GnfdClient) UpdateBucket(operator sdk.AccAddress, bucketName string,
	readQuota storage_type.ReadQuota, paymentAcc sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	_, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "UpdateBucketInfo"}
	}

	updateBucketMsg := storage_type.NewMsgUpdateBucketInfo(operator, bucketName, readQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	log.Info().Msg("get updateBucketInfo txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "UpdateBucketInfo"}
}
