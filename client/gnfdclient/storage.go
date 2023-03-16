package gnfdclient

import (
	"context"
	"errors"
	"io"
	"math"

	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield/sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	aclTypes "github.com/bnb-chain/greenfield/x/permission/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
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

// CreateGroupOption  indicates the meta to construct createGroup msg
type CreateGroupOption struct {
	InitGroupMember []sdk.AccAddress
	TxOpts          *types.TxOption
}

// GroupUpdateInfo indicates the info to update group member
type GroupUpdateInfo struct {
	Group    string           // group name
	Members  []sdk.AccAddress // update chain address list
	IsRemove bool             // indicate whether to remove or add member
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

	createBucketMsg := storageTypes.NewMsgCreateBucket(km.GetAddr(), bucketName, opts.IsPublic, primaryAddr, opts.PaymentAddress, 0, nil)

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
	delBucketMsg := storageTypes.NewMsgDeleteBucket(km.GetAddr(), bucketName)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "DeleteBucket"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "DeleteBucket"}
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

	redundancyType := storageTypes.REDUNDANCY_EC_TYPE
	if opts.IsReplicaType {
		redundancyType = storageTypes.REDUNDANCY_REPLICA_TYPE
	}

	createObjectMsg := storageTypes.NewMsgCreateObject(km.GetAddr(), bucketName, objectName,
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
	delObjectMsg := storageTypes.NewMsgDeleteObject(km.GetAddr(), bucketName, objectName)

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

	cancelCreateMsg := storageTypes.NewMsgCancelCreateObject(km.GetAddr(), bucketName, objectName)

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
	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, targetQuota, paymentAcc)

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

	updateBucketMsg := storageTypes.NewMsgUpdateBucketInfo(km.GetAddr(), bucketName, readQuota, paymentAcc)

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateBucketMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "UpdateBucketInfo"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "UpdateBucketInfo"}
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
		spInfoList = append(spInfoList, info)
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

// CreateGroup create a new group on greenfield chain
// the group members can be inited or not
func (c *GnfdClient) CreateGroup(groupName string, opt CreateGroupOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "CreateBucket"}
	}

	createGroupMsg := storageTypes.NewMsgCreateGroup(km.GetAddr(), groupName, opt.InitGroupMember)

	if err = createGroupMsg.ValidateBasic(); err != nil {
		return GnfdResponse{"", err, "CreateGroup"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{createGroupMsg}, opt.TxOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateGroup"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "CreateGroup"}
}

// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
func (c *GnfdClient) DeleteGroup(groupName string, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteGroup"}
	}

	deleteGroupMsg := storageTypes.NewMsgDeleteGroup(km.GetAddr(), groupName)
	if err = deleteGroupMsg.ValidateBasic(); err != nil {
		return GnfdResponse{"", err, "DeleteGroup"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{deleteGroupMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateGroup"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "DeleteGroup"}
}

// UpdateGroupMember support adding or removing members from the group and return the txn hash
func (c *GnfdClient) UpdateGroupMember(info GroupUpdateInfo, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteGroup"}
	}

	var updateGroupMsg *storageTypes.MsgUpdateGroupMember
	if info.IsRemove {
		updateGroupMsg = storageTypes.NewMsgUpdateGroupMember(km.GetAddr(), info.Group, nil, info.Members)
	} else {
		updateGroupMsg = storageTypes.NewMsgUpdateGroupMember(km.GetAddr(), info.Group, info.Members, nil)
	}

	if err = updateGroupMsg.ValidateBasic(); err != nil {
		return GnfdResponse{"", err, "updateGroup"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{updateGroupMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "CreateGroup"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "DeleteGroup"}
}

// HeadGroup query the groupInfo on chain, return the group info if exists
// return err info if group not exist
func (c *GnfdClient) HeadGroup(ctx context.Context, groupName, groupOwner sdk.AccAddress) (*storageTypes.GroupInfo, error) {
	headGroupRequest := storageTypes.QueryHeadGroupRequest{
		GroupOwner: groupOwner.String(),
		GroupName:  groupName.String(),
	}

	headGroupResponse, err := c.ChainClient.HeadGroup(ctx, &headGroupRequest, nil)
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

	_, err := c.ChainClient.HeadGroupMember(ctx, &headGroupRequest, nil)
	if err != nil {
		return false
	}

	return true
}

// PutBucketPolicy apply policy to greenfield bucket
func (c *GnfdClient) PutBucketPolicy(bucketName, policy string, grantUser sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteGroup"}
	}

	statements, err := DecodeStatements(policy)
	if err != nil {
		return GnfdResponse{"", err, "PutPolicy"}
	}

	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	return c.sendPutPolicyTxn(resource, km.GetAddr(), grantUser, statements, txOpts)
}

// PutObjectPolicy apply policy to greenfield object
func (c *GnfdClient) PutObjectPolicy(bucketName, objectName, policy string, grantUser sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteGroup"}
	}

	statements, err := DecodeStatements(policy)
	if err != nil {
		return GnfdResponse{"", err, "PutPolicy"}
	}

	resource := gnfdTypes.NewObjectGRN(bucketName, objectName).String()
	return c.sendPutPolicyTxn(resource, km.GetAddr(), grantUser, statements, txOpts)
}

// PutGroupPolicy apply policy to greenfield group, the sender need to be the owner of the group
func (c *GnfdClient) PutGroupPolicy(groupName, policy string, grantUser sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteGroup"}
	}
	sender := km.GetAddr()

	statements, err := DecodeStatements(policy)
	if err != nil {
		return GnfdResponse{"", err, "PutPolicy"}
	}

	resource := gnfdTypes.NewGroupGRN(sender, groupName).String()
	return c.sendPutPolicyTxn(resource, km.GetAddr(), grantUser, statements, txOpts)
}

func DecodeStatements(policy string) ([]*aclTypes.Statement, error) {
	statements := make([]*aclTypes.Statement, 0)
	var gnfdPolicy *utils.GnfdPolicy
	err := gnfdPolicy.UnMarshal([]byte(policy))
	if err != nil {
		return nil, err
	}

	for _, s := range gnfdPolicy.Statements {
		chainActions := make([]aclTypes.ActionType, 0)

		for _, action := range s.Actions {
			chainActions = append(chainActions, utils.GetChainAction(action))
		}

		chainStatement := &aclTypes.Statement{
			Actions: chainActions,
			Effect:  utils.GetChainEffect(s.Effect),
		}
		statements = append(statements, chainStatement)
	}

	return statements, nil
}

func (c *GnfdClient) sendPutPolicyTxn(resource string, operator, grantUser sdk.AccAddress, statements []*aclTypes.Statement, txOpts types.TxOption) GnfdResponse {
	putPolicyMsg := storageTypes.NewMsgPutPolicy(operator, resource, aclTypes.NewPrincipalWithAccount(grantUser), statements)
	if err := putPolicyMsg.ValidateBasic(); err != nil {
		return GnfdResponse{"", err, "PutPolicy"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{putPolicyMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "PutPolicy"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "PutPolicy"}

}

// DeleteBucketPolicy delete the bucket policy of the grantUser
func (c *GnfdClient) DeleteBucketPolicy(bucketName string, grantUser sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteBucketPolicy"}
	}

	resource := gnfdTypes.NewBucketGRN(bucketName).String()
	principal := aclTypes.NewPrincipalWithAccount(grantUser)

	return c.sendDelPolicyTxn(km.GetAddr().String(), resource, principal, txOpts)
}

func (c *GnfdClient) DeleteObjectPolicy(bucketName, objectName string, grantUser sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteBucketPolicy"}
	}

	resource := gnfdTypes.NewObjectGRN(bucketName, objectName).String()
	principal := aclTypes.NewPrincipalWithAccount(grantUser)

	return c.sendDelPolicyTxn(km.GetAddr().String(), resource, principal, txOpts)
}

// DeleteGroupPolicy  delete group policy of the grantUser, the sender need to be the owner of the group
func (c *GnfdClient) DeleteGroupPolicy(groupName string, grantUser sdk.AccAddress, txOpts types.TxOption) GnfdResponse {
	km, err := c.ChainClient.GetKeyManager()
	if err != nil {
		return GnfdResponse{"", errors.New("key manager is nil"), "DeleteBucketPolicy"}
	}
	sender := km.GetAddr()
	resource := gnfdTypes.NewGroupGRN(sender, groupName).String()
	principal := aclTypes.NewPrincipalWithAccount(grantUser)

	return c.sendDelPolicyTxn(sender.String(), resource, principal, txOpts)
}

func (c *GnfdClient) sendDelPolicyTxn(operator string, resource string, principal *aclTypes.Principal, txOpts types.TxOption) GnfdResponse {
	delPolicyMsg := storageTypes.NewMsgDeletePolicy(operator, resource, principal)

	if err := delPolicyMsg.ValidateBasic(); err != nil {
		return GnfdResponse{"", err, "DeletePolicy"}
	}

	resp, err := c.ChainClient.BroadcastTx([]sdk.Msg{delPolicyMsg}, &txOpts)
	if err != nil {
		return GnfdResponse{"", err, "DeletePolicy"}
	}

	return GnfdResponse{resp.TxResponse.TxHash, err, "DeletePolicy"}
}

type EffectInfo struct {
	EffectInfo string // return EFFECT_ALLOW or EFFECT_DENY
	Err        error  // query error info
}

// IsBucketPermissionAllowed check if the permission of bucket is allowed to the member
func (c *GnfdClient) IsBucketPermissionAllowed(member sdk.AccAddress, bucketName string, action utils.Action) EffectInfo {
	if action.IsValid() {
		return EffectInfo{"", errors.New("invalid action")}
	}

	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   member.String(),
		BucketName: bucketName,
		ActionType: utils.GetChainAction(action),
	}
	ctx := context.Background()

	verifyResp, err := c.ChainClient.VerifyPermission(ctx, &verifyReq)
	if err != nil {
		return EffectInfo{"", err}
	}

	return EffectInfo{verifyResp.Effect.String(), nil}
}

// IsObjectPermissionAllowed check if the permission of the object is allowed to the member
func (c *GnfdClient) IsObjectPermissionAllowed(member sdk.AccAddress, bucketName, objectName string, action utils.Action) EffectInfo {
	if action.IsValid() {
		return EffectInfo{"", errors.New("invalid action")}
	}

	verifyReq := storageTypes.QueryVerifyPermissionRequest{
		Operator:   member.String(),
		BucketName: bucketName,
		ObjectName: objectName,
		ActionType: utils.GetChainAction(action),
	}
	ctx := context.Background()

	verifyResp, err := c.ChainClient.VerifyPermission(ctx, &verifyReq)
	if err != nil {
		return EffectInfo{"", err}
	}

	return EffectInfo{verifyResp.Effect.String(), nil}
}

// GetBucketPolicy get the policy info of the bucket resource
func (c *GnfdClient) GetBucketPolicy(bucketName string, principalAddress sdk.AccAddress) (*aclTypes.Policy, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	queryPolicy := storageTypes.QueryPolicyForAccountRequest{Resource: resource,
		PrincipalAddress: principalAddress.String()}

	ctx := context.Background()
	queryPolicyResp, err := c.ChainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// GetObjectPolicy get the policy info of the object resource
func (c *GnfdClient) GetObjectPolicy(bucketName, objectName string, principalAddress sdk.AccAddress) (*aclTypes.Policy, error) {
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName).String()

	queryPolicy := storageTypes.QueryPolicyForAccountRequest{Resource: resource,
		PrincipalAddress: principalAddress.String()}

	ctx := context.Background()
	queryPolicyResp, err := c.ChainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// GetGroupPolicy get the policy info of the  resource
func (c *GnfdClient) GetGroupPolicy(bucketName string, principalGroupId string) (string, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	queryPolicy := storageTypes.QueryPolicyForGroupRequest{Resource: resource,
		PrincipalGroupId: principalGroupId}

	ctx := context.Background()
	queryPolicyResp, err := c.ChainClient.QueryPolicyForGroup(ctx, &queryPolicy)
	if err != nil {
		return "", err
	}

	return queryPolicyResp.Policy.String(), nil
}
