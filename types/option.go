package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SetTagsOptions struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// CreateBucketOptions indicates the metadata to construct `CreateBucket` msg of storage module.
type CreateBucketOptions struct {
	Visibility     storageTypes.VisibilityType // Visibility defines the bucket public status.
	TxOpts         *gnfdsdktypes.TxOption      // TxOpts defines the options to customize a transaction.
	PaymentAddress string                      // PaymentAddress indicates the HEX-encoded string of the payment address.
	ChargedQuota   uint64                      // ChargedQuota defines the read data that users are charged for, measured in bytes.
	IsAsyncMode    bool                        // indicate whether to create the bucket in asynchronous mode.
	Tags           *storageTypes.ResourceTags  // set tags when creating bucket
}

// MigrateBucketOptions indicates the metadata to construct `MigrateBucket` msg of storage module.
type MigrateBucketOptions struct {
	TxOpts      *gnfdsdktypes.TxOption
	IsAsyncMode bool // indicate whether to create the bucket in asynchronous mode
}

// CancelMigrateBucketOptions indicates the metadata to construct `CancelMigrateBucket` msg of storage module.
type CancelMigrateBucketOptions struct {
	TxOpts      *gnfdsdktypes.TxOption
	IsAsyncMode bool // indicate whether to create the bucket in asynchronous mode
}

// VoteProposalOptions indicates the metadata to construct `VoteProposal` msg.
type VoteProposalOptions struct {
	Metadata string                // Metadata defines the metadata to be submitted along with the vote.
	TxOpts   gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// SubmitProposalOptions indicates the metadata to construct `SubmitProposal` msg.
type SubmitProposalOptions struct {
	Metadata string                // metadata efines the metadata to be submitted along with the proposal.
	TxOpts   gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// CreateStorageProviderOptions indicates the metadata to construct `CreateStorageProvider` msg.
type CreateStorageProviderOptions struct {
	ReadPrice             sdk.Dec  // ReadPrice defines the storage provider's read price, in bnb wei per charge byte.
	FreeReadQuota         uint64   // FreeReadQuota defines the free read quota of the SP.
	StorePrice            sdk.Dec  // StorePrice defines the store price of the SP, in bnb wei per charge byte.
	ProposalDepositAmount math.Int // ProposalDepositAmount defines the amount needed for a proposal.
	ProposalTitle         string
	ProposalSummary       string
	ProposalMetaData      string
	TxOpts                gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// GrantDepositForStorageProviderOptions indicates the metadata to construct `Grant` msg.
type GrantDepositForStorageProviderOptions struct {
	Expiration *time.Time            // Expiration defines the expiration time of grant.
	TxOpts     gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// DeleteBucketOption indicates the metadata to construct `DeleteBucket` msg.
type DeleteBucketOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// UpdatePaymentOption indicates the metadata to construct `UpdateBucketInfo` msg.
type UpdatePaymentOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// UpdateBucketOptions indicates the metadata to construct `UpdateBucketInfo` msg of storage module.
type UpdateBucketOptions struct {
	Visibility     storageTypes.VisibilityType // Visibility defines the bucket public status.
	TxOpts         *gnfdsdktypes.TxOption      // TxOpts defines the options to customize a transaction.
	PaymentAddress string                      // PaymentAddress defines the HEX-encoded string of the payment address.
	ChargedQuota   *uint64                     // ChargedQuota defines the read data that users are charged for, measured in bytes.
}

// UpdateObjectOption indicates the metadata to construct `UpdateObjectInfo` msg of storage module.
type UpdateObjectOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// CancelUpdateObjectOption indicates the metadata to construct `CancelUpdateObjectContent` msg of storage module.
type CancelUpdateObjectOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// CancelCreateOption indicates the metadata to construct `CancelCreateObject` msg of storage module.
type CancelCreateOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// BuyQuotaOption indicates the metadata to construct `UpdateBucketInfo` msg of storage module.
type BuyQuotaOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// UpdateVisibilityOption indicates the metadata to construct `UpdateBucketInfo` msg of storage module.
type UpdateVisibilityOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// DeleteObjectOption indicates the metadata to construct `DeleteObject` msg of storage module.
type DeleteObjectOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// DeleteGroupOption indicates the metadata to construct `DeleteGroup` msg of storage module.
type DeleteGroupOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// CreateObjectOptions - indicates the metadata to construct `createObject` message of storage module.
type CreateObjectOptions struct {
	Visibility          storageTypes.VisibilityType // Visibility defines the bucket public status.
	TxOpts              *gnfdsdktypes.TxOption      // TxOpts defines the options to customize a transaction.
	SecondarySPAccs     []sdk.AccAddress            // SecondarySPAccs indicates a list of secondary Storage Provider's addresses.
	ContentType         string                      // ContentType defines the content type of object.
	IsReplicaType       bool                        // IsReplicaType indicates whether the object uses REDUNDANCY_REPLICA_TYPE.
	IsAsyncMode         bool                        // IsAsyncMode indicate whether to create the object in asynchronous mode.
	IsSerialComputeMode bool                        // IsSerialComputeMode indicate whether to compute integrity hash in serial way or parallel way when creating an object.
	Tags                *storageTypes.ResourceTags  // set tags when creating bucket
}

// UpdateObjectOptions - indicates the metadata to construct `updateObjectContent` message of storage module.
type UpdateObjectOptions struct {
	TxOpts              *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
	SecondarySPAccs     []sdk.AccAddress       // SecondarySPAccs indicates a list of secondary Storage Provider's addresses.
	ContentType         string                 // ContentType defines the content type of object.
	IsReplicaType       bool                   // IsReplicaType indicates whether the object uses REDUNDANCY_REPLICA_TYPE.
	IsAsyncMode         bool                   // IsAsyncMode indicate whether to update the object in asynchronous mode.
	IsSerialComputeMode bool                   // IsSerialComputeMode indicate whether to compute integrity hash in serial way or parallel way when creating an object.
}

// CreateGroupOptions indicates the metadata to construct `CreateGroup` msg.
type CreateGroupOptions struct {
	Extra  string                     // Extra defines the extra meta for a group.
	TxOpts *gnfdsdktypes.TxOption     // TxOpts defines the options to customize a transaction.
	Tags   *storageTypes.ResourceTags // set tags when creating bucket
}

// UpdateGroupMemberOption indicates the metadata to construct `UpdateGroupMembers` msg.
type UpdateGroupMemberOption struct {
	TxOpts         *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
	ExpirationTime []*time.Time           // ExpirationTime defines a list of expiration time for each group member to be updated.
}

// LeaveGroupOption indicates the metadata to construct `LeaveGroup` msg of storage module.
type LeaveGroupOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

// RenewGroupMemberOption indicates the metadata to construct `RenewGroupMember` msg of storage module.
type RenewGroupMemberOption struct {
	TxOpts         *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
	ExpirationTime []*time.Time           // ExpirationTime defines a list of expiration time for each group member to be updated.
}

// ComputeHashOptions indicates the metadata of redundancy strategy.
type ComputeHashOptions struct {
	SegmentSize  uint64
	DataShards   uint32
	ParityShards uint32
}

// ListReadRecordOptions contains the options for `ListBucketReadRecord` API.
type ListReadRecordOptions struct {
	StartTimeStamp int64 // StartTimeStamp indicates the start timestamp of the return read quota record.
	MaxRecords     int
}

// ListObjectsOptions contains the options for `ListObjects` API.
type ListObjectsOptions struct {
	// ShowRemovedObject determines whether to include objects that have been marked as removed in the list.
	// If set to false, these objects will be skipped.
	ShowRemovedObject bool

	// StartAfter defines the starting object name for the listing of objects.
	// The listing will start from the next object after the one named in this attribute.
	StartAfter string

	// ContinuationToken is the token returned from a previous list objects request to indicate where
	// in the list of objects to resume the listing. This is used for pagination.
	ContinuationToken string

	// Delimiter is a character that is used to group keys.
	// All keys that contain the same string between the prefix and the first occurrence of the delimiter
	// are grouped under a single result element in common prefixes.
	// It is used for grouping keys, currently only '/' is supported.
	Delimiter string

	// Prefix limits the response to keys that begin with the specified prefix.
	// You can use prefixes to separate a bucket into different sets of keys in a way similar to how a file
	// system uses folders.
	Prefix string

	// MaxKeys defines the maximum number of keys returned to the response body.
	// If not specified, the default value is 50.
	// The maximum limit for returning objects is 1000
	MaxKeys   uint64
	Endpoint  string // indicates the endpoint of sp.
	SPAddress string // indicates the HEX-encoded string of the sp address to be challenged.
}

// PutPolicyOption indicates the metadata to construct `PutPolicy` msg of storage module.
type PutPolicyOption struct {
	TxOpts           *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
	PolicyExpireTime *time.Time             // PolicyExpireTime defines the expiration timestamp of policy.
}

// DeletePolicyOption indicates the metadata to construct `DeletePolicy` msg of storage module.
type DeletePolicyOption struct {
	TxOpts *gnfdsdktypes.TxOption // TxOpts defines the options to customize a transaction.
}

type NewStatementOptions struct {
	StatementExpireTime *time.Time
	LimitSize           uint64
}

// PutObjectOptions indicates the options for uploading an object to Storage Provider.
type PutObjectOptions struct {
	ContentType      string // ContentType indicates the content type of object.
	TxnHash          string // TxnHash indicates the transaction hash creating the object meta on chain.
	DisableResumable bool   // DisableResumable indicates whether upload the object to Storage Provider via resumable upload.
	PartSize         uint64
	Delegated        bool
	IsUpdate         bool
	Visibility       storageTypes.VisibilityType
}

// GetObjectOptions contains the options for `GetObject` API.
type GetObjectOptions struct {
	Range            string `url:"-" header:"Range,omitempty"` // Range support for downloading partial data.
	SupportResumable bool   // SupportResumable support resumable download. Resumable downloads refer to the capability of resuming interrupted or incomplete downloads from the point where they were paused or disrupted.
	PartSize         uint64 // PartSize indicate the resumable download's part size, download a large file in multiple parts. The part size is an integer multiple of the segment size.
}

// GetChallengeInfoOptions contains the options for querying challenge data.
type GetChallengeInfoOptions struct {
	Endpoint     string // Endpoint indicates the endpoint of sp
	SPAddress    string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
	UseV2version bool   // UseV2version indicates whether using of the v2 version get-challenge API
}

// GetSecondaryPieceOptions contains the options for `GetSecondaryPiece` API.
type GetSecondaryPieceOptions struct {
	Endpoint  string // Endpoint indicates the endpoint of sp.
	SPAddress string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

// ListGroupsOptions contains the options for `ListGroups` API.
type ListGroupsOptions struct {
	SourceType string // SourceType indicates the source type of group.
	Limit      int64
	Offset     int64
	Endpoint   string // Endpoint indicates the endpoint of sp.
	SPAddress  string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

// GroupMembersPaginationOptions contains the options for `ListGroupMembers` API.
type GroupMembersPaginationOptions struct {
	// Limit determines the number of group data records to be returned.
	// If the limit is set to 0, it will default to 50.
	// If the limit exceeds 1000, only 1000 records will be returned.
	Limit      int64
	StartAfter string // StartAfter is used to input the user's account address for pagination purposes.
	Endpoint   string // indicates the endpoint of sp.
	SPAddress  string // indicates the HEX-encoded string of the sp address to be challenged.
}

// GroupsOwnerPaginationOptions contains the options for `ListGroupsByOwner` API.
type GroupsOwnerPaginationOptions struct {
	// Limit determines the number of group data records to be returned.
	// If the limit is set to 0, it will default to 50.
	// If the limit exceeds 1000, only 1000 records will be returned.
	Limit      int64
	StartAfter string // StartAfter is used to input the group id for pagination purposes.
	Owner      string // Owner defines the owner account address of groups, if owner is set to "", it will default to current user address.
	Endpoint   string // Endpoint indicates the endpoint of sp.
	SPAddress  string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

// GroupsPaginationOptions contains the options for `ListGroupsByAccount` API.
type GroupsPaginationOptions struct {
	// Limit determines the number of group data records to be returned.
	// If the limit is set to 0, it will default to 50.
	// If the limit exceeds 1000, only 1000 records will be returned.
	Limit      int64
	StartAfter string // StartAfter is used to input the group id for pagination purposes.
	Account    string // Account defines the user account address, if it is set to "", it will default to the current user address.
	Endpoint   string // Endpoint indicates the endpoint of sp.
	SPAddress  string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

func (o *GetObjectOptions) SetRange(start, end int64) error {
	switch {
	case 0 < start && end == 0:
		// `bytes=N-`.
		o.Range = fmt.Sprintf("bytes=%d-", start)
	case 0 <= start && start <= end:
		// `bytes=N-M`
		o.Range = fmt.Sprintf("bytes=%d-%d", start, end)
	default:
		return ToInvalidArgumentResp(
			fmt.Sprintf(
				"Invalid Range : start=%d end=%d",
				start, end))
	}
	return nil
}

// EndPointOptions contains the options for querying a specified SP.
type EndPointOptions struct {
	Endpoint  string // Endpoint indicates the endpoint of sp.
	SPAddress string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

// ListBucketsOptions contains the options for `ListBuckets` API.
type ListBucketsOptions struct {
	ShowRemovedBucket bool   // ShowRemovedBucket determines whether to include buckets that have been marked as removed in the list. If set to false, these buckets will be skipped.
	Account           string // Account defines the user account address, if it is set to "", it will default to the current user address.
	Endpoint          string // Endpoint indicates the endpoint of sp.
	SPAddress         string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

// ListBucketsByPaymentAccountOptions contains the options for `ListBucketsByPaymentAccount` API.
type ListBucketsByPaymentAccountOptions struct {
	Endpoint  string // Endpoint indicates the endpoint of sp.
	SPAddress string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

// ListUserPaymentAccountsOptions contains the options for `ListUserPaymentAccounts` API.
type ListUserPaymentAccountsOptions struct {
	Account   string // Account defines the user account address, if it is set to "", it will default to the current user address.
	Endpoint  string // Endpoint indicates the endpoint of sp.
	SPAddress string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}

// ListObjectPoliciesOptions contains the options for `ListObjectPolicies` API.
type ListObjectPoliciesOptions struct {
	// Limit determines the number of policies data records to be returned.
	// If the limit is set to 0, it will default to 50.
	// If the limit exceeds 1000, only 1000 records will be returned.
	Limit      int64
	StartAfter string // StartAfter is used to input the policy id for pagination purposes.
	Endpoint   string // Endpoint indicates the endpoint of sp.
	SPAddress  string // SPAddress indicates the HEX-encoded string of the sp address to be challenged.
}
