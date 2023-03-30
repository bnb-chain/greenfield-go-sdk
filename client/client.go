package client

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield/sdk/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IClient interface {
	Bucket
	Object
	Group
	SP
	Payment
	Tx
}

type Bucket interface {
	CreateBucket(ctx context.Context, bucketName string, opts *CreateBucketOptions) (string, error)
	DeleteBucket(bucketName string, opt DeleteBucketOption) (string, error)
	UpdateBucketVisibility(ctx context.Context, bucketName string,
		visibility storageTypes.VisibilityType, opt UpdateVisibilityOption) (string, error)
	GetBucketReadQuota(ctx context.Context, bucketName string, authInfo AuthInfo) (QuotaInfo, error)
	ListBucketReadRecord(ctx context.Context, bucketName string, maxRecords int, opt ListReadRecordOptions, authInfo AuthInfo) (QuotaRecordInfo, error)
	HeadBucket(ctx context.Context, bucketName string) (*storageTypes.BucketInfo, error)
	HeadBucketByID(ctx context.Context, bucketID string) (*storageTypes.BucketInfo, error)
	// PutBucketPolicy apply bucket policy to the principal, return the txn hash
	PutBucketPolicy(bucketName string, principalStr Principal,
		statements []*permTypes.Statement, opt PutPolicyOption) (string, error)
	// DeleteBucketPolicy delete the bucket policy of the principal
	DeleteBucketPolicy(bucketName string, principalAddr sdk.AccAddress, opt DeletePolicyOption) (string, error)
	DeleteObjectPolicy(bucketName, objectName string, principalAddr sdk.AccAddress, opt DeletePolicyOption) (string, error)
	// GetBucketPolicy get the bucket policy info of the user specified by principalAddr
	GetBucketPolicy(ctx context.Context, bucketName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error)
	ListBuckets(ctx context.Context, userInfo UserInfo, authInfo AuthInfo) (ListBucketsResponse, error)
}

type Object interface {
	CreateObject(ctx context.Context, bucketName, objectName string,
		reader io.Reader, opts *CreateObjectOptions) (string, error)
	PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
		reader io.Reader, opt PutObjectOption) (err error)
	CancelCreateObject(bucketName, objectName string, opt CancelCreateOption) (string, error)
	GetObject(ctx context.Context, bucketName, objectName string, opt GetObjectOption) (io.ReadCloser, GetObjectResult, error)
	// HeadObject query the objectInfo on chain to check th object id, return the object info if exists
	// return err info if object not exist
	HeadObject(ctx context.Context, bucketName, objectName string) (*storageTypes.ObjectInfo, error)
	// HeadObjectByID query the objectInfo on chain by object id, return the object info if exists
	// return err info if object not exist
	HeadObjectByID(ctx context.Context, objID string) (*storageTypes.ObjectInfo, error)

	// PutObjectPolicy apply object policy to the principal, return the txn hash
	PutObjectPolicy(bucketName, objectName string, principalStr Principal,
		statements []*permTypes.Statement, opt PutPolicyOption) (string, error)

	// GetObjectPolicy get the object policy info of the user specified by principalAddr
	GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error)

	ListObjects(ctx context.Context, bucketName string, authInfo AuthInfo) (ListObjectsResponse, error)
}

type Payment interface {
	BuyQuotaForBucket(ctx context.Context, bucketName string, targetQuota uint64, opt BuyQuotaOption) (string, error)
	GetQuotaPrice(ctx context.Context, SPAddress sdk.AccAddress) (uint64, error)
	Send()
	MultiSend()
	QueryBalances()
}

type Group interface {
	// CreateGroup create a new group on greenfield chain
	// the group members can be initialized  or not
	CreateGroup(groupName string, opt CreateGroupOptions) (string, error)
	// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
	DeleteGroup(groupName string, txOpts types.TxOption) (string, error)
	// UpdateGroupMember support adding or removing members from the group and return the txn hash
	UpdateGroupMember(groupName string, groupOwner sdk.AccAddress,
		addMembers, removeMembers []sdk.AccAddress, opts UpdateGroupMemberOption) (string, error)
	LeaveGroup(groupName string, groupOwner sdk.AccAddress, opt LeaveGroupOption) (string, error)
	// HeadGroup query the groupInfo on chain, return the group info if exists
	// return err info if group not exist
	HeadGroup(ctx context.Context, groupName string, groupOwner sdk.AccAddress) (*storageTypes.GroupInfo, error)
	// HeadGroupMember query the group member info on chain, return true if the member exists in group
	HeadGroupMember(ctx context.Context, groupName string, groupOwner, headMember sdk.AccAddress) bool

	// PutGroupPolicy apply group policy to user specified by principalAddr, the sender need to be the owner of the group
	PutGroupPolicy(groupName string, principalAddr sdk.AccAddress,
		statements []*permTypes.Statement, opt PutPolicyOption) (string, error)

	// DeleteGroupPolicy  delete group policy of the principal, the sender need to be the owner of the group
	DeleteGroupPolicy(groupName string, principalAddr sdk.AccAddress, opt DeletePolicyOption) (string, error)

	// GetBucketPolicyOfGroup get the bucket policy info of the group specified by group id
	// it queries a bucket policy that grants permission to a group
	GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error)
	// GetObjectPolicyOfGroup get the object policy info of the group specified by group id
	// it queries an object policy that grants permission to a group
	GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error)
}

type SP interface {
	// ListSP return the storage provider info on chain
	// isInService indicates if only display the sp with STATUS_IN_SERVICE status
	ListSP(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error)
	// GetSPInfo return the sp info  the sp chain address
	GetSPInfo(ctx context.Context, SPAddr sdk.AccAddress) (*spTypes.StorageProvider, error)
	// GetSpAddrByEndpoint return the chain addr according to the SP endpoint
	GetSpAddrByEndpoint(ctx context.Context) (sdk.AccAddress, error)

	GetCreateBucketApproval(ctx context.Context, createBucketMsg *storageTypes.MsgCreateBucket,
		authInfo AuthInfo) (*storageTypes.MsgCreateBucket, error)

	GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject,
		authInfo AuthInfo) (*storageTypes.MsgCreateObject, error)

	ChallengeSP(ctx context.Context, info ChallengeInfo, authInfo AuthInfo) (ChallengeResult, error)

	CreateStorageProvider()
	EditStorageProvider()
	Deposit()
	GrantDeposit()
	QueryStorageProviders()
	QueryStorageProvider()
	QueryParams()
}

type Tx interface {
	BroadcastTx(txBytes []byte)
	SimulateTx()
	CreateTx()
	WaitForBlockHeight(ctx context.Context, height int64) error
	WaitForTx(ctx context.Context, hash string) error
}

// TODO(leo) check if it is needed ,  cause import cycle
func New() (IClient, error) {
	// c := api.Client{}
	return nil, nil
}
