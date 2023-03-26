package gnfdclient

import (
	"context"
	"errors"
	"io"
	"math"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield/sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/common"
	"github.com/bnb-chain/greenfield/types/s3util"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

type Principal string

// CreateBucketOptions indicates the meta to construct createBucket msg of storage module
type CreateBucketOptions struct {
	Visibility       *storageTypes.VisibilityType
	TxOpts           *types.TxOption
	PaymentAddress   *sdk.AccAddress
	PrimarySPAddress *sdk.AccAddress
	ChargedQuota     uint64
}

// CreateObjectOptions indicates the metadata to construct `createObject` message of storage module
type CreateObjectOptions struct {
	Visibility      *storageTypes.VisibilityType
	TxOpts          *types.TxOption
	SecondarySPAccs []sdk.AccAddress
	ContentType     string
	IsReplicaType   bool // indicates whether the object use REDUNDANCY_REPLICA_TYPE
}

type DeleteObjectOption struct {
	TxOpts *types.TxOption
}

type DeleteBucketOption struct {
	TxOpts *types.TxOption
}

type CancelCreateOption struct {
	TxOpts *types.TxOption
}

type BuyQuotaOption struct {
	TxOpts *types.TxOption
}

type UpdateVisibilityOption struct {
	TxOpts *types.TxOption
}

// CreateGroupOptions  indicates the meta to construct createGroup msg
type CreateGroupOptions struct {
	InitGroupMember []sdk.AccAddress
	TxOpts          *types.TxOption
}

// UpdateGroupMemberOption indicates the info to update group member
type UpdateGroupMemberOption struct {
	TxOpts *types.TxOption
}

type LeaveGroupOption struct {
	TxOpts *types.TxOption
}

// ComputeHashOptions indicates the metadata of redundancy strategy
type ComputeHashOptions struct {
	SegmentSize  uint64
	DataShards   uint32
	ParityShards uint32
}

// ListReadRecordOption indicates the start timestamp of the return read quota record
type ListReadRecordOption struct {
	StartTimeStamp int64
}

type PutPolicyOption struct {
	TxOpts           *types.TxOption
	PolicyExpireTime *time.Time
}

type DeletePolicyOption struct {
	TxOpts *types.TxOption
}

type NewStatementOptions struct {
	StatementExpireTime *time.Time
	LimitSize           uint64
}

// CreateBucket get approval of creating bucket and send createBucket txn to greenfield chain
func (c *GnfdClient) CreateBucket(ctx context.Context, bucketName string, opts CreateBucketOptions) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	var primaryAddr sdk.AccAddress
	if opts.PrimarySPAddress != nil {
		primaryAddr = *opts.PrimarySPAddress
	} else {
		// if user has not set primarySP chain address, fetch it from chain
		primaryAddr, err = c.GetSpAddrFromEndpoint(ctx)
		if err != nil {
			return "", err
		}
	}

	var visibility storageTypes.VisibilityType
	if opts.Visibility != nil {
		visibility = *opts.Visibility
	} else {
		visibility = storageTypes.VISIBILITY_TYPE_DEFAULT
	}

	createBucketMsg := storageTypes.NewMsgCreateBucket(km.GetAddr(), bucketName,
		visibility, primaryAddr, *opts.PaymentAddress, 0, nil, opts.ChargedQuota)

	err = createBucketMsg.ValidateBasic()
	if err != nil {
		return "", err
	}
	signedMsg, err := c.SPClient.GetCreateBucketApproval(ctx, createBucketMsg, sp.NewAuthInfo(false, ""))
	if err != nil {
		return "", err
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{signedMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// DeleteBucket send DeleteBucket txn to greenfield chain and return txn hash
func (c *GnfdClient) DeleteBucket(bucketName string, opt DeleteBucketOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	delBucketMsg := storageTypes.NewMsgDeleteBucket(km.GetAddr(), bucketName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delBucketMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// GetRedundancyParams query and return the data shards, parity shards and segment size of redundancy
// configuration on chain
func (c *GnfdClient) GetRedundancyParams() (uint32, uint32, uint64, error) {
	query := storageTypes.QueryParamsRequest{}
	queryResp, err := c.ChainClient.StorageQueryClient.Params(context.Background(), &query)
	if err != nil {
		return 0, 0, 0, err
	}

	params := queryResp.Params
	return params.GetRedundantDataChunkNum(), params.GetRedundantParityChunkNum(), params.GetMaxSegmentSize(), nil
}

// ComputeHashRoots return the hash roots list and content size
func (c *GnfdClient) ComputeHashRoots(reader io.Reader) ([][]byte, int64, storageTypes.RedundancyType, error) {
	dataBlocks, parityBlocks, segSize, err := c.GetRedundancyParams()
	if err != nil {
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}

	// get hash and objectSize from reader
	return hashlib.ComputeIntegrityHash(reader, int64(segSize), int(dataBlocks), int(parityBlocks))
}

// CreateObject get approval of creating object and send createObject txn to greenfield chain
func (c *GnfdClient) CreateObject(ctx context.Context, bucketName, objectName string,
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

	km, err := c.ChainClient.GetKeyManager()
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
		contentType = sp.ContentDefault
	}

	var visibility storageTypes.VisibilityType
	if opts.Visibility != nil {
		visibility = *opts.Visibility
	} else {
		visibility = storageTypes.VISIBILITY_TYPE_INHERIT
	}

	createObjectMsg := storageTypes.NewMsgCreateObject(km.GetAddr(), bucketName, objectName,
		uint64(size), visibility, expectCheckSums, contentType, redundancyType, math.MaxUint, nil, opts.SecondarySPAccs)
	err = createObjectMsg.ValidateBasic()
	if err != nil {
		return "", err
	}

	signedCreateObjectMsg, err := c.SPClient.GetCreateObjectApproval(ctx, createObjectMsg, sp.NewAuthInfo(false, ""))
	if err != nil {
		return "", err
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{signedCreateObjectMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, err
}

// DeleteObject send DeleteBucket txn to greenfield chain and return txn hash
func (c *GnfdClient) DeleteObject(bucketName, objectName string, opt DeleteObjectOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	delObjectMsg := storageTypes.NewMsgDeleteObject(km.GetAddr(), bucketName, objectName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delObjectMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// CancelCreateObject send CancelCreateObject txn to greenfield chain
func (c *GnfdClient) CancelCreateObject(bucketName, objectName string, opt CancelCreateOption) (string, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	cancelCreateMsg := storageTypes.NewMsgCancelCreateObject(km.GetAddr(), bucketName, objectName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{cancelCreateMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// PutObject upload payload of object to storage provider
func (c *GnfdClient) PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
	reader io.Reader, opt sp.PutObjectOption,
) (res sp.UploadResult, err error) {
	return c.SPClient.PutObject(ctx, bucketName, objectName, txnHash,
		objectSize, reader, sp.NewAuthInfo(false, ""), opt)
}

// GetObject download the object from primary storage provider
func (c *GnfdClient) GetObject(ctx context.Context, bucketName, objectName string, opt sp.GetObjectOption) (io.ReadCloser, sp.ObjectInfo, error) {
	return c.SPClient.GetObject(ctx, bucketName, objectName, opt, sp.NewAuthInfo(false, ""))
}

// BuyQuotaForBucket buy the target quota of the specific bucket
// targetQuota indicates the target quota to set for the bucket
func (c *GnfdClient) BuyQuotaForBucket(ctx context.Context, bucketName string, targetQuota uint64, opt BuyQuotaOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}

	paymentAddr, err := sdk.AccAddressFromHexUnsafe(bucketInfo.PaymentAddress)
	if err != nil {
		return "", err
	}
	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, &targetQuota, paymentAddr, bucketInfo.Visibility)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// UpdateBucketVisibility update the visibilityType of bucket
func (c *GnfdClient) UpdateBucketVisibility(ctx context.Context, bucketName string,
	visibility storageTypes.VisibilityType, opt UpdateVisibilityOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return "", err
	}

	paymentAddr, err := sdk.AccAddressFromHexUnsafe(bucketInfo.PaymentAddress)
	if err != nil {
		return "", err
	}

	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, &bucketInfo.ChargedReadQuota, paymentAddr, visibility)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// GetQuotaPrice return the quota price of the SP
func (c *GnfdClient) GetQuotaPrice(ctx context.Context, SPAddress sdk.AccAddress) (uint64, error) {
	resp, err := c.ChainClient.QueryGetSpStoragePriceByTime(ctx, &spTypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    SPAddress.String(),
		Timestamp: 0,
	})
	if err != nil {
		return 0, err
	}
	return resp.SpStoragePrice.ReadPrice.BigInt().Uint64(), nil
}

// GetBucketReadQuota return quota info of bucket of current month, include chain quota, free quota and consumed quota
func (c *GnfdClient) GetBucketReadQuota(ctx context.Context, bucketName string) (sp.QuotaInfo, error) {
	return c.SPClient.GetBucketReadQuota(ctx, bucketName, sp.NewAuthInfo(false, ""))
}

// ListBucketReadRecord return read quota record info of current month
func (c *GnfdClient) ListBucketReadRecord(ctx context.Context, bucketName string, maxRecords int, opt ListReadRecordOption) (sp.QuotaRecordInfo, error) {
	return c.SPClient.ListBucketReadRecord(ctx, bucketName, maxRecords, sp.ListReadRecordOption{StartTimeStamp: opt.StartTimeStamp}, sp.NewAuthInfo(false, ""))
}

// HeadBucket query the bucketInfo on chain, return the bucket info if exists
// return err info if bucket not exist
func (c *GnfdClient) HeadBucket(ctx context.Context, bucketName string) (*storageTypes.BucketInfo, error) {
	queryHeadBucketRequest := storageTypes.QueryHeadBucketRequest{
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
func (c *GnfdClient) HeadBucketByID(ctx context.Context, bucketID string) (*storageTypes.BucketInfo, error) {
	headBucketRequest := &storageTypes.QueryHeadBucketByIdRequest{
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
func (c *GnfdClient) HeadObject(ctx context.Context, bucketName, objectName string) (*storageTypes.ObjectInfo, error) {
	queryHeadObjectRequest := storageTypes.QueryHeadObjectRequest{
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
func (c *GnfdClient) HeadObjectByID(ctx context.Context, objID string) (*storageTypes.ObjectInfo, error) {
	headObjectRequest := storageTypes.QueryHeadObjectByIdRequest{
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
func (c *GnfdClient) ListSP(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error) {
	request := &spTypes.QueryStorageProvidersRequest{}
	gnfdRep, err := c.ChainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, err
	}

	spList := gnfdRep.GetSps()
	spInfoList := make([]spTypes.StorageProvider, 0)
	for _, info := range spList {
		if isInService && info.Status != spTypes.STATUS_IN_SERVICE {
			continue
		}
		spInfoList = append(spInfoList, *info)
	}

	return spInfoList, nil
}

// GetSPInfo return the sp info  the sp chain address
func (c *GnfdClient) GetSPInfo(ctx context.Context, SPAddr sdk.AccAddress) (*spTypes.StorageProvider, error) {
	request := &spTypes.QueryStorageProviderRequest{
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
	spClientEndpoint := c.SPClient.GetURL().Host
	for _, spInfo := range spList {
		endpoint := spInfo.GetEndpoint()
		if strings.Contains(endpoint, "http") {
			s := strings.Split(endpoint, "//")
			endpoint = s[1]
		}
		if endpoint == spClientEndpoint {
			addr := spInfo.GetOperatorAddress()
			if addr == "" {
				return nil, errors.New("fail to get addr")
			}
			return sdk.MustAccAddressFromHex(spInfo.GetOperatorAddress()), nil
		}
	}
	return nil, errors.New("fail to get addr")
}

// CreateGroup create a new group on greenfield chain
// the group members can be initialized  or not
func (c *GnfdClient) CreateGroup(groupName string, opt CreateGroupOptions) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	createGroupMsg := storageTypes.NewMsgCreateGroup(km.GetAddr(), groupName, opt.InitGroupMember)

	if err = createGroupMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{createGroupMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
func (c *GnfdClient) DeleteGroup(groupName string, txOpts types.TxOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	deleteGroupMsg := storageTypes.NewMsgDeleteGroup(km.GetAddr(), groupName)
	if err = deleteGroupMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{deleteGroupMsg}, &txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// UpdateGroupMember support adding or removing members from the group and return the txn hash
func (c *GnfdClient) UpdateGroupMember(groupName string, groupOwner sdk.AccAddress,
	addMembers, removeMembers []sdk.AccAddress, opts UpdateGroupMemberOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	if groupName == "" {
		return "", errors.New("group name is empty")
	}

	if len(addMembers) == 0 && len(removeMembers) == 0 {
		return "", errors.New("no update member")
	}
	updateGroupMsg := storageTypes.NewMsgUpdateGroupMember(km.GetAddr(), groupOwner, groupName, addMembers, removeMembers)
	if err = updateGroupMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateGroupMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, nil
}

func (c *GnfdClient) LeaveGroup(groupName string, groupOwner sdk.AccAddress, opt LeaveGroupOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	leaveGroupMsg := storageTypes.NewMsgLeaveGroup(km.GetAddr(), groupOwner, groupName)
	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{leaveGroupMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, nil
}

// HeadGroup query the groupInfo on chain, return the group info if exists
// return err info if group not exist
func (c *GnfdClient) HeadGroup(ctx context.Context, groupName string, groupOwner sdk.AccAddress) (*storageTypes.GroupInfo, error) {
	headGroupRequest := storageTypes.QueryHeadGroupRequest{
		GroupOwner: groupOwner.String(),
		GroupName:  groupName,
	}

	headGroupResponse, err := c.ChainClient.HeadGroup(ctx, &headGroupRequest)
	if err != nil {
		return nil, err
	}

	return headGroupResponse.GroupInfo, nil
}

// HeadGroupMember query the group member info on chain, return true if the member exists in group
func (c *GnfdClient) HeadGroupMember(ctx context.Context, groupName string, groupOwner, headMember sdk.AccAddress) bool {
	headGroupRequest := storageTypes.QueryHeadGroupMemberRequest{
		GroupName:  groupName,
		GroupOwner: groupOwner.String(),
		Member:     headMember.String(),
	}

	_, err := c.ChainClient.HeadGroupMember(ctx, &headGroupRequest)
	return err == nil
}

// NewStatement return the statement of permission module
func NewStatement(actions []permTypes.ActionType, effect permTypes.Effect,
	resource []string, opts NewStatementOptions) permTypes.Statement {
	statement := permTypes.Statement{
		Actions:        actions,
		Effect:         effect,
		Resources:      resource,
		ExpirationTime: opts.StatementExpireTime,
	}

	if opts.LimitSize != 0 {
		statement.LimitSize = &common.UInt64Value{Value: opts.LimitSize}
	}

	return statement
}

func NewPrincipalWithAccount(principalAddr sdk.AccAddress) (Principal, error) {
	p := permTypes.NewPrincipalWithAccount(principalAddr)
	principalBytes, err := p.Marshal()
	if err != nil {
		return "", err
	}
	return Principal(principalBytes), nil
}

func NewPrincipalWithGroupId(groupId uint64) (Principal, error) {
	p := permTypes.NewPrincipalWithGroup(sdkmath.NewUint(groupId))
	principalBytes, err := p.Marshal()
	if err != nil {
		return "", err
	}
	return Principal(principalBytes), nil
}

// PutBucketPolicy apply bucket policy to the principal, return the txn hash
func (c *GnfdClient) PutBucketPolicy(bucketName string, principalStr Principal,
	statements []*permTypes.Statement, opt PutPolicyOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	resource := gnfdTypes.NewBucketGRN(bucketName)
	principal := &permTypes.Principal{}
	if err = principal.Unmarshal([]byte(principalStr)); err != nil {
		return "", err
	}

	putPolicyMsg := storageTypes.NewMsgPutPolicy(km.GetAddr(), resource.String(),
		principal, statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(putPolicyMsg, *opt.TxOpts)
}

// PutObjectPolicy apply object policy to the principal, return the txn hash
func (c *GnfdClient) PutObjectPolicy(bucketName, objectName string, principalStr Principal,
	statements []*permTypes.Statement, opt PutPolicyOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
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

	return c.sendPutPolicyTxn(putPolicyMsg, *opt.TxOpts)
}

// PutGroupPolicy apply group policy to user specified by principalAddr, the sender need to be the owner of the group
func (c *GnfdClient) PutGroupPolicy(groupName string, principalAddr sdk.AccAddress,
	statements []*permTypes.Statement, opt PutPolicyOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	sender := km.GetAddr()

	resource := gnfdTypes.NewGroupGRN(sender, groupName)
	putPolicyMsg := storageTypes.NewMsgPutPolicy(km.GetAddr(), resource.String(),
		permTypes.NewPrincipalWithAccount(principalAddr), statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(putPolicyMsg, *opt.TxOpts)
}

// sendPutPolicyTxn broadcast the putPolicy msg and return the txn hash
func (c *GnfdClient) sendPutPolicyTxn(msg *storageTypes.MsgPutPolicy, txOpts types.TxOption) (string, error) {
	if err := msg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{msg}, &txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// DeleteBucketPolicy delete the bucket policy of the principal
func (c *GnfdClient) DeleteBucketPolicy(bucketName string, principalAddr sdk.AccAddress, opt DeletePolicyOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	resource := gnfdTypes.NewBucketGRN(bucketName).String()
	principal := permTypes.NewPrincipalWithAccount(principalAddr)

	return c.sendDelPolicyTxn(km.GetAddr(), resource, principal, *opt.TxOpts)
}

func (c *GnfdClient) DeleteObjectPolicy(bucketName, objectName string, principalAddr sdk.AccAddress, opt DeletePolicyOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	principal := permTypes.NewPrincipalWithAccount(principalAddr)
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	return c.sendDelPolicyTxn(km.GetAddr(), resource.String(), principal, *opt.TxOpts)
}

// DeleteGroupPolicy  delete group policy of the principal, the sender need to be the owner of the group
func (c *GnfdClient) DeleteGroupPolicy(groupName string, principalAddr sdk.AccAddress, opt DeletePolicyOption) (string, error) {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	sender := km.GetAddr()
	resource := gnfdTypes.NewGroupGRN(sender, groupName).String()
	principal := permTypes.NewPrincipalWithAccount(principalAddr)

	return c.sendDelPolicyTxn(sender, resource, principal, *opt.TxOpts)
}

// sendDelPolicyTxn broadcast the deletePolicy msg and return the txn hash
func (c *GnfdClient) sendDelPolicyTxn(operator sdk.AccAddress, resource string, principal *permTypes.Principal, txOpts types.TxOption) (string, error) {
	delPolicyMsg := storageTypes.NewMsgDeletePolicy(operator, resource, principal)

	if err := delPolicyMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delPolicyMsg}, &txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// IsBucketPermissionAllowed check if the permission of bucket is allowed to the user
func (c *GnfdClient) IsBucketPermissionAllowed(ctx context.Context, user sdk.AccAddress,
	bucketName string, action permTypes.ActionType) (permTypes.Effect, error) {
	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   user.String(),
		BucketName: bucketName,
		ActionType: action,
	}

	verifyResp, err := c.ChainClient.VerifyPermission(ctx, &verifyReq)
	if err != nil {
		return permTypes.EFFECT_DENY, err
	}

	return verifyResp.Effect, nil
}

// IsObjectPermissionAllowed check if the permission of the object is allowed to the user
func (c *GnfdClient) IsObjectPermissionAllowed(ctx context.Context, user sdk.AccAddress,
	bucketName, objectName string, action permTypes.ActionType) (permTypes.Effect, error) {
	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   user.String(),
		BucketName: bucketName,
		ObjectName: objectName,
		ActionType: action,
	}

	verifyResp, err := c.ChainClient.VerifyPermission(ctx, &verifyReq)
	if err != nil {
		return permTypes.EFFECT_DENY, err
	}

	return verifyResp.Effect, nil
}

// GetBucketPolicy get the bucket policy info of the user specified by principalAddr
func (c *GnfdClient) GetBucketPolicy(ctx context.Context, bucketName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	queryPolicy := storageTypes.QueryPolicyForAccountRequest{
		Resource:         resource,
		PrincipalAddress: principalAddr.String(),
	}

	queryPolicyResp, err := c.ChainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// GetObjectPolicy get the object policy info of the user specified by principalAddr
func (c *GnfdClient) GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	queryPolicy := storageTypes.QueryPolicyForAccountRequest{
		Resource:         resource.String(),
		PrincipalAddress: principalAddr.String(),
	}

	queryPolicyResp, err := c.ChainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// GetBucketPolicyOfGroup get the bucket policy info of the group specified by group id
// it queries a bucket policy that grants permission to a group
func (c *GnfdClient) GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	queryPolicy := storageTypes.QueryPolicyForGroupRequest{
		Resource:         resource,
		PrincipalGroupId: sdkmath.NewUint(groupId).String(),
	}

	queryPolicyResp, err := c.ChainClient.QueryPolicyForGroup(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// GetObjectPolicyOfGroup get the object policy info of the group specified by group id
// it queries an object policy that grants permission to a group
func (c *GnfdClient) GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	queryPolicy := storageTypes.QueryPolicyForGroupRequest{
		Resource:         resource.String(),
		PrincipalGroupId: sdkmath.NewUint(groupId).String(),
	}

	queryPolicyResp, err := c.ChainClient.QueryPolicyForGroup(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}
