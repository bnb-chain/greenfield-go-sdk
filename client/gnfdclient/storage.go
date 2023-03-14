package gnfdclient

import (
	"context"
	"errors"
	"io"
	"math"

	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/bnb-chain/greenfield/sdk/types"
	spType "github.com/bnb-chain/greenfield/x/sp/types"
	storageType "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

// CreateBucketOptions indicates the meta to construct createBucket msg of storage module
type CreateBucketOptions struct {
	IsPublic         bool
	TxOpts           *types.TxOption
	PaymentAddress   sdk.AccAddress
	PrimarySPAddress sdk.AccAddress
}

// CreateObjectOptions indicates the meta to construct createObject msg of storage module
type CreateObjectOptions struct {
	IsPublic        bool
	TxOpts          *types.TxOption
	SecondarySPAccs []sdk.AccAddress
	ContentType     string
	IsReplicaType   bool // indicates whether the object use REDUNDANCY_REPLICA_TYPE
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
func (c *GnfdClient) CreateBucket(ctx context.Context, bucketName string, opts CreateBucketOptions) GnfdResponse {
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreateBucket"}
	}
	var primaryAddr sdk.AccAddress
	if opts.PrimarySPAddress != nil {
		primaryAddr = opts.PrimarySPAddress
	} else {
		// if user has not set primarySP chain address, fetch it from chain
		primaryAddr, err = c.GetSpAddrFromEndpoint(ctx)
		if err != nil {
			return GnfdResponse{"", err, "CreateBucket"}
		}
	}

	createBucketMsg := storageType.NewMsgCreateBucket(km.GetAddr(), bucketName, opts.IsPublic, primaryAddr, opts.PaymentAddress, 0, nil)

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

// GetRedundancyParams query and return the data shards, parity shards and segment size of redundancy
// configuration on chain
func (c *GnfdClient) GetRedundancyParams() (uint32, uint32, uint64, error) {
	query := storageType.QueryParamsRequest{}
	queryResp, err := c.ChainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return 0, 0, 0, err
	}

	params := queryResp.Params
	return params.GetRedundantDataChunkNum(), params.GetRedundantParityChunkNum(), params.GetMaxSegmentSize(), nil
}

// ComputeHashRoots return the hash roots list and content size
func (c *GnfdClient) ComputeHashRoots(reader io.Reader) ([][]byte, int64, error) {
	dataBlocks, parityBlocks, segSize, err := c.GetRedundancyParams()
	if err != nil {
		return nil, 0, err
	}

	// get hash and objectSize from reader
	return hashlib.ComputeIntegrityHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks))
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
	expectCheckSums, size, err := c.ComputeHashRoots(reader)
	if err != nil {
		return GnfdResponse{"", err, "CreateObject"}
	}

	var contentType string
	if opts.ContentType != "" {
		contentType = opts.ContentType
	} else {
		contentType = sp.ContentDefault
	}

	redundancyType := storageType.REDUNDANCY_EC_TYPE
	if opts.IsReplicaType {
		redundancyType = storageType.REDUNDANCY_REPLICA_TYPE
	}

	createObjectMsg := storageType.NewMsgCreateObject(km.GetAddr(), bucketName, objectName,
		uint64(size), opts.IsPublic, expectCheckSums, contentType, redundancyType, math.MaxUint, nil, opts.SecondarySPAccs)
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

	return GnfdResponse{resp.TxResponse.TxHash, err, "CancelCreateObject"}
}

// PutObject upload payload of object to storage provider
func (c *GnfdClient) PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
	reader io.Reader, opt sp.UploadOption) (res sp.UploadResult, err error) {
	return c.SPClient.PutObject(ctx, bucketName, objectName, txnHash,
		objectSize, reader, sp.NewAuthInfo(false, ""), opt)
}

// GetObject download the object from primary storage provider
func (c *GnfdClient) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, sp.ObjectInfo, error) {
	return c.SPClient.GetObject(ctx, bucketName, objectName, sp.DownloadOption{}, sp.NewAuthInfo(false, ""))
}

// BuyQuotaForBucket buy the target quota of the specific bucket
// targetQuota indicates the target quota to set for the bucket
func (c *GnfdClient) BuyQuotaForBucket(bucketName string,
	targetQuota uint64, paymentAcc sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "UpdateBucketInfo"}
	}
	updateBucketMsg := storageType.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, targetQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "UpdateBucketInfo"}
}

// UpdateBucket update the bucket read quota on chain
func (c *GnfdClient) UpdateBucket(bucketName string,
	readQuota uint64, paymentAcc sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
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

	return GnfdResponse{resp.TxResponse.TxHash, err, "UpdateBucketInfo"}
}

// HeadBucket query the bucketInfo on chain to check the bucketId
// return err info if bucket not exist
func (c *GnfdClient) HeadBucket(ctx context.Context, bucketName string) (*storageType.BucketInfo, error) {
	queryHeadBucketRequest := storageType.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := c.ChainClient.HeadBucket(ctx, &queryHeadBucketRequest)
	if err != nil {
		return nil, err
	}

	return queryHeadBucketResponse.BucketInfo, nil
}

// HeadBucketByID query the bucketInfo on chain by bucketId, return the bucket info if exists
// return err info if bucket not exist
func (c *GnfdClient) HeadBucketByID(ctx context.Context, bucketID string) (*storageType.BucketInfo, error) {
	headBucketRequest := &storageType.QueryHeadBucketByIdRequest{
		BucketId: bucketID,
	}

	headBucketResponse, err := c.ChainClient.HeadBucketById(ctx, headBucketRequest, nil)
	if err != nil {
		return nil, err
	}

	return headBucketResponse.BucketInfo, nil
}

// HeadObject query the objectInfo on chain to check th object id, return the object info if exists
// return err info if object not exist
func (c *GnfdClient) HeadObject(ctx context.Context, bucketName, objectName string) (*storageType.ObjectInfo, error) {
	queryHeadObjectRequest := storageType.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := c.ChainClient.HeadObject(ctx, &queryHeadObjectRequest)
	if err != nil {
		return nil, err
	}

	return queryHeadObjectResponse.ObjectInfo, nil
}

// HeadObjectByID query the objectInfo on chain by object id, return the object info if exists
// return err info if object not exist
func (c *GnfdClient) HeadObjectByID(ctx context.Context, objID string) (*storageType.ObjectInfo, error) {
	headObjectRequest := storageType.QueryHeadObjectByIdRequest{
		ObjectId: objID,
	}
	queryHeadObjectResponse, err := c.ChainClient.HeadObjectById(ctx, &headObjectRequest, nil)
	if err != nil {
		return nil, err
	}

	return queryHeadObjectResponse.ObjectInfo, nil
}

// ListSP return the storage provider info on chain
// isInService indicates if only display the sp with STATUS_IN_SERVICE status
func (c *GnfdClient) ListSP(ctx context.Context, isInService bool) ([]spType.StorageProvider, error) {
	request := &spType.QueryStorageProvidersRequest{}
	gnfdRep, err := c.ChainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, err
	}

	spList := gnfdRep.GetSps()
	spInfoList := make([]spType.StorageProvider, 0)
	for _, info := range spList {
		if isInService && info.Status != spType.STATUS_IN_SERVICE {
			continue
		}
		spInfoList = append(spInfoList, info)
	}

	return spInfoList, nil
}

// GetSPInfo return the sp info  the sp chain address
func (c *GnfdClient) GetSPInfo(ctx context.Context, SPAddr sdk.AccAddress) (*spType.StorageProvider, error) {
	request := &spType.QueryStorageProviderRequest{
		SpAddress: SPAddr.String(),
	}

	gnfdRep, err := c.ChainClient.StorageProvider(ctx, request)
	if err != nil {
		return nil, err
	}

	return gnfdRep.StorageProvider, nil
}

// GetSpAddrFromEndpoint return the chain addr according to the SP endpoint
func (c *GnfdClient) GetSpAddrFromEndpoint(ctx context.Context) (sdk.AccAddress, error) {
	spList, err := c.ListSP(ctx, false)
	if err != nil {
		return nil, err
	}
	for _, spInfo := range spList {
		if spInfo.GetEndpoint() == c.SPClient.GetURL().Host {
			addr := spInfo.GetOperatorAddress()
			if addr == "" {
				return nil, errors.New("fail to get addr")
			}
			return sdk.MustAccAddressFromHex(spInfo.GetOperatorAddress()), nil
		}
	}
	return nil, errors.New("fail to get addr")
}
