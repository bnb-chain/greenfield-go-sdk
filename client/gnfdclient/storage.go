package gnfdclient

import (
	"context"
	"encoding/hex"
	"io"

	lib "github.com/bnb-chain/greenfield-common/go"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/bnb-chain/greenfield/sdk/types"
	storage_type "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

// CreateBucket get approval of creating bucket and send createBucket txn to greenfield chain
func (c *IntegratedClient) CreateBucket(ctx context.Context, bucketMeta sp.CreateBucketMeta,
	txOpts types.TxOption) (*tx.BroadcastTxResponse, error) {
	// get approval of creating bucket from sp
	signedCreateBucketMsg, err := c.SPClient.GetCreateBucketApproval(ctx, bucketMeta, sp.NewAuthInfo(false, ""))
	if err != nil {
		return nil, err
	}

	var signedMsg storage_type.MsgCreateBucket

	ModuleCdc.MustUnmarshalJSON([]byte(signedCreateBucketMsg), &signedMsg)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{&signedMsg}, &txOpts)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("get createBucket txn hash:" + resp.TxResponse.TxHash)
	return resp, nil
}

// DelBucket send DeleteBucket txn to chain
func (c *IntegratedClient) DelBucket(operator sdk.AccAddress, bucketName string, txOpts types.TxOption) (*tx.BroadcastTxResponse, error) {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return nil, err
	}
	delBucketMsg := storage_type.NewMsgDeleteBucket(operator, bucketName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delBucketMsg}, &txOpts)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("get deleteBucket txn hash:" + resp.TxResponse.TxHash)

	return resp, err
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *IntegratedClient) CreateObject(ctx context.Context, objectMeta sp.CreateObjectMeta,
	reader io.Reader, txOpts types.TxOption) (*tx.BroadcastTxResponse, error) {

	// get hash and objectSize from reader
	pieceHashRoots, size, err := lib.ComputerHash(reader, sp.SegmentSize, sp.DataShards, sp.ParityShards)
	if err != nil {
		log.Error().Msg("get hash roots fail" + err.Error())
		return nil, err
	}

	expectCheckSums := make([][]byte, 0)
	for index, hash := range pieceHashRoots {
		hashByte, err := hex.DecodeString(hash)
		if err != nil {
			return nil, err
		}
		expectCheckSums[index] = hashByte
	}
	// get approval of creating bucket from sp
	signedCreateObjectMsg, err := c.SPClient.GetCreateObjectApproval(ctx, objectMeta,
		uint64(size), expectCheckSums, sp.NewAuthInfo(false, ""))
	if err != nil {
		return nil, err
	}

	var signedMsg storage_type.MsgCreateObject
	ModuleCdc.MustUnmarshalJSON([]byte(signedCreateObjectMsg), &signedMsg)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{&signedMsg}, &txOpts)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("get createObject txn hash:" + resp.TxResponse.TxHash)
	return resp, nil
}

// DelObject send DeleteBucket txn to chain
func (c *IntegratedClient) DelObject(operator sdk.AccAddress, bucketName, objectName string,
	txOpts types.TxOption) (*tx.BroadcastTxResponse, error) {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return nil, err
	}

	if err := utils.IsValidObjectName(objectName); err != nil {
		return nil, err
	}
	delObjectMsg := storage_type.NewMsgDeleteObject(operator, bucketName, objectName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delObjectMsg}, &txOpts)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("get deleteObject txn hash:" + resp.TxResponse.TxHash)
	return resp, err
}

// CancelCreateObject send CancelCreateObject txn to chain
func (c *IntegratedClient) CancelCreateObject(operator sdk.AccAddress, bucketName,
	objectName string, txOpts types.TxOption) (*tx.BroadcastTxResponse, error) {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return nil, err
	}

	if err := utils.IsValidObjectName(objectName); err != nil {
		return nil, err
	}

	cancelCreateMsg := storage_type.NewMsgCancelCreateObject(operator, bucketName, objectName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{cancelCreateMsg}, &txOpts)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("get txn hash:" + resp.TxResponse.TxHash)
	return resp, err
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

func (c *IntegratedClient) BuyQuotaForBucket(operator sdk.AccAddress, bucketName string,
	quota storage_type.ReadQuota, paymentAcc sdk.AccAddress, txOpts types.TxOption) (*tx.BroadcastTxResponse, error) {
	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storage_type.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := c.ChainClient.HeadBucket(ctx, &queryHeadBucketRequest)
	if err != nil {
		return nil, err
	}

	newQuota := queryHeadBucketResponse.BucketInfo.GetReadQuota() + quota
	updateBucketMsg := storage_type.NewMsgUpdateBucketInfo(operator, bucketName, newQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("get updateBucketInfo txn hash:" + resp.TxResponse.TxHash)

	return resp, err
}

func (c *IntegratedClient) UpdateBucket(operator sdk.AccAddress, bucketName string,
	readQuota storage_type.ReadQuota, paymentAcc sdk.AccAddress, txOpts types.TxOption) (*tx.BroadcastTxResponse, error) {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return nil, err
	}

	updateBucketMsg := storage_type.NewMsgUpdateBucketInfo(operator, bucketName, readQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("get updateBucketInfo txn hash:" + resp.TxResponse.TxHash)

	return resp, err
}
