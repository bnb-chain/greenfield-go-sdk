package gnfdclient

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	lib "github.com/bnb-chain/greenfield-common/go"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/bnb-chain/greenfield/sdk/types"
	storage_type "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

type CreateBucketOptions struct {
	IsPublic       bool
	Creator        sdk.AccAddress
	TxOpts         types.TxOption
	PaymentAddress sdk.AccAddress
}

// CreateObjectMeta indicates the meta to construct createObject msg of storage module
type CreateObjectMeta struct {
	BucketName string
	ObjectName string
	Reader     io.Reader
}

type CreateObjectOptions struct {
	IsPublic        bool
	Creator         sdk.AccAddress
	TxOpts          types.TxOption
	SecondarySPAccs []sdk.AccAddress
	ContentType     string
}

type GnfdResponse struct {
	txnHash string
	err     error
	txnType string
}

// CreateBucket get approval of creating bucket and send createBucket txn to greenfield chain
func (c *IntegratedClient) CreateBucket(ctx context.Context, bucketName string, primarySPAddress sdk.AccAddress, opts CreateBucketOptions) GnfdResponse {
	_, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreateBucket"}
	}

	var creator sdk.AccAddress
	if opts.Creator != nil {
		creator = opts.Creator
	} else if c.sender != nil {
		creator = c.sender
	} else {
		return GnfdResponse{"", errors.New("creator address fetch failed"), "CreateBucket"}
	}

	approveMeta := sp.ApproveBucketMeta{
		BucketName:       bucketName,
		IsPublic:         opts.IsPublic,
		Creator:          creator,
		PrimarySPAddress: primarySPAddress,
	}

	if opts.PaymentAddress != nil {
		approveMeta.PaymentAddress = opts.PaymentAddress
	}

	// get approval of creating bucket from sp
	signedCreateBucketMsg, err := c.SPClient.GetCreateBucketApproval(ctx, approveMeta, sp.NewAuthInfo(false, ""))
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

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{&signedMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateBucket"}
	}

	log.Info().Msg("get createBucket txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "CreateBucket"}
}

// DelBucket send DeleteBucket txn to chain
func (c *IntegratedClient) DelBucket(operator sdk.AccAddress, bucketName string, txOpts types.TxOption) GnfdResponse {
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

func (c *IntegratedClient) ComputeHash(reader io.Reader) ([]string, int64, error) {
	query := storage_type.QueryParamsRequest{}
	queryResp, err := c.ChainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return nil, 0, err
	}

	dataShards := queryResp.Params.GetRedundantDataChunkNum()
	parityShards := queryResp.Params.GetRedundantParityChunkNum()
	segSize := queryResp.Params.GetMaxSegmentSize()

	log.Info().Msg(fmt.Sprintf("get segSize %d, dataShards: %d , parityShards: %d", segSize, dataShards, parityShards))
	// get hash and objectSize from reader
	return lib.ComputerHash(reader, int64(segSize), int(dataShards), int(parityShards))
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *IntegratedClient) CreateObject(ctx context.Context, bucketName, objectName string, reader io.Reader, opts CreateObjectOptions) GnfdResponse {
	_, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreateObject"}
	}
	if reader == nil {
		return GnfdResponse{"", errors.New("fail to compute hash of payload, reader is nil"), "CreateObject"}
	}
	// compute hash root of payload
	pieceHashRoots, size, err := c.ComputeHash(reader)
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

	var creator sdk.AccAddress
	if opts.Creator != nil {
		creator = opts.Creator
	} else if c.sender != nil {
		creator = c.sender
	} else {
		return GnfdResponse{"", errors.New("creator address fetch failed"), "CreateObject"}
	}

	approveMeta := sp.ApproveObjectMeta{
		BucketName: bucketName,
		ObjectName: objectName,
		IsPublic:   opts.IsPublic,
		Creator:    creator,
	}

	if opts.SecondarySPAccs != nil {
		approveMeta.SecondarySPAccs = opts.SecondarySPAccs
	}

	if opts.ContentType != "" {
		approveMeta.ContentType = opts.ContentType
	} else {
		approveMeta.ContentType = sp.ContentDefault
	}

	// get approval of creating bucket from sp
	signedCreateObjectMsg, err := c.SPClient.GetCreateObjectApproval(ctx, approveMeta,
		uint64(size), expectCheckSums, sp.NewAuthInfo(false, ""))
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

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{&signedMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	log.Info().Msg("get createObject txn hash:" + resp.TxResponse.TxHash)
	return GnfdResponse{resp.TxResponse.TxHash, err, "CreateObject"}
}

// DelObject send DeleteBucket txn to chain
func (c *IntegratedClient) DelObject(operator sdk.AccAddress, bucketName, objectName string,
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
func (c *IntegratedClient) CancelCreateObject(operator sdk.AccAddress, bucketName,
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
func (c *IntegratedClient) UploadObject(ctx context.Context, bucketName, objectName, txnHash string,
	reader io.Reader, meta sp.ObjectMeta) (res sp.UploadResult, err error) {
	return c.SPClient.PutObject(ctx, bucketName, objectName, txnHash, reader, meta, sp.NewAuthInfo(false, ""))
}

// DownloadObject download the object from primary storage provider
func (c *IntegratedClient) DownloadObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, sp.ObjectInfo, error) {
	return c.SPClient.GetObject(ctx, bucketName, objectName, sp.DownloadOption{}, sp.NewAuthInfo(false, ""))
}

// BuyQuotaForBucket increase the quota to reach storage service of sender
func (c *IntegratedClient) BuyQuotaForBucket(operator sdk.AccAddress, bucketName string,
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

func (c *IntegratedClient) UpdateBucket(operator sdk.AccAddress, bucketName string,
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
