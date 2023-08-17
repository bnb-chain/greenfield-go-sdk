package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/types/common"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateBucketOptions indicates the meta to construct createBucket msg of storage module
// PaymentAddress  indicates the HEX-encoded string of the payment address
type CreateBucketOptions struct {
	Visibility     storageTypes.VisibilityType
	TxOpts         *gnfdsdktypes.TxOption
	PaymentAddress string
	ChargedQuota   uint64
	IsAsyncMode    bool // indicate whether to create the bucket in asynchronous mode
}

type MigrateBucketOptions struct {
	DstPrimarySPID       uint32
	DstPrimarySPApproval common.Approval
	TxOpts               *gnfdsdktypes.TxOption
	IsAsyncMode          bool // indicate whether to create the bucket in asynchronous mode
}

type VoteProposalOptions struct {
	Metadata string
	TxOption gnfdsdktypes.TxOption
}

type SubmitProposalOptions struct {
	Metadata string
	TxOption gnfdsdktypes.TxOption
}

type CreateStorageProviderOptions struct {
	ReadPrice             sdk.Dec
	FreeReadQuota         uint64
	StorePrice            sdk.Dec
	ProposalDepositAmount math.Int // wei BNB
	ProposalTitle         string
	ProposalSummary       string
	ProposalMetaData      string
	TxOption              gnfdsdktypes.TxOption
}

type GrantDepositForStorageProviderOptions struct {
	Expiration *time.Time
	TxOption   gnfdsdktypes.TxOption
}

type DeleteBucketOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type UpdatePaymentOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

// UpdateBucketOptions indicates the meta to construct updateBucket msg of storage module
// PaymentAddress  indicates the HEX-encoded string of the payment address
type UpdateBucketOptions struct {
	Visibility     storageTypes.VisibilityType
	TxOpts         *gnfdsdktypes.TxOption
	PaymentAddress string
	ChargedQuota   *uint64
}

type UpdateObjectOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type CancelCreateOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type BuyQuotaOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type UpdateVisibilityOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type DeleteObjectOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type DeleteGroupOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

// CreateObjectOptions indicates the metadata to construct `createObject` message of storage module
type CreateObjectOptions struct {
	Visibility          storageTypes.VisibilityType
	TxOpts              *gnfdsdktypes.TxOption
	SecondarySPAccs     []sdk.AccAddress
	ContentType         string
	IsReplicaType       bool // indicates whether the object use REDUNDANCY_REPLICA_TYPE
	IsAsyncMode         bool // indicate whether to create the object in asynchronous mode
	IsSerialComputeMode bool // indicate whether to compute integrity hash in serial way or parallel way when creating object
}

// CreateGroupOptions  indicates the meta to construct createGroup msg
type CreateGroupOptions struct {
	Extra  string
	TxOpts *gnfdsdktypes.TxOption
}

// UpdateGroupMemberOption indicates the info to update group member
type UpdateGroupMemberOption struct {
	TxOpts         *gnfdsdktypes.TxOption
	ExpirationTime []*time.Time
}

type LeaveGroupOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

// RenewGroupMemberOption indicates the info to update group member
type RenewGroupMemberOption struct {
	TxOpts         *gnfdsdktypes.TxOption
	ExpirationTime []*time.Time
}

// ComputeHashOptions indicates the metadata of redundancy strategy
type ComputeHashOptions struct {
	SegmentSize  uint64
	DataShards   uint32
	ParityShards uint32
}

// ListReadRecordOptions indicates the start timestamp of the return read quota record
type ListReadRecordOptions struct {
	StartTimeStamp int64
	MaxRecords     int
}

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
	MaxKeys         uint64
	EndPointOptions *EndPointOptions
}

type PutPolicyOption struct {
	TxOpts           *gnfdsdktypes.TxOption
	PolicyExpireTime *time.Time
}

type DeletePolicyOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type NewStatementOptions struct {
	StatementExpireTime *time.Time
	LimitSize           uint64
}

type ApproveBucketOptions struct {
	IsPublic       bool
	PaymentAddress sdk.AccAddress
}

type ApproveObjectOptions struct {
	IsPublic        bool
	SecondarySPAccs []sdk.AccAddress
}

type PutObjectOptions struct {
	ContentType      string
	TxnHash          string
	DisableResumable bool
	PartSize         uint64
}

// GetObjectOptions contains the options of getObject
type GetObjectOptions struct {
	Range            string `url:"-" header:"Range,omitempty"` // support for downloading partial data
	SupportRecovery  bool   // support recover data from secondary SPs if primary SP not in service
	SupportResumable bool   // support resumable download. Resumable downloads refer to the capability of resuming interrupted or incomplete downloads from the point where they were paused or disrupted.
	PartSize         uint64 // indicate the resumable download's part size, download a large file in multiple parts. The part size is an integer multiple of the segment size.
}

type GetChallengeInfoOptions struct {
	Endpoint  string // indicates the endpoint of sp
	SPAddress string // indicates the HEX-encoded string of the sp address to be challenged
}

type GetSecondaryPieceOptions struct {
	Endpoint  string // indicates the endpoint of sp
	SPAddress string // indicates the HEX-encoded string of the sp address to be challenged
}

type ListGroupsOptions struct {
	SourceType      string
	Limit           int64
	Offset          int64
	EndPointOptions *EndPointOptions
}

type GroupMembersPaginationOptions struct {
	// Limit determines the number of group data records to be returned.
	// If the limit is set to 0, it will default to 50.
	// If the limit exceeds 1000, only 1000 records will be returned.
	Limit int64
	// StartAfter is used to input the user's account address for pagination purposes
	StartAfter      string
	EndPointOptions *EndPointOptions
}

type GroupsPaginationOptions struct {
	// Limit determines the number of group data records to be returned.
	// If the limit is set to 0, it will default to 50.
	// If the limit exceeds 1000, only 1000 records will be returned.
	Limit int64
	// StartAfter is used to input the group id for pagination purposes
	StartAfter string
	// Account defines the user account address
	// if account is set to "", it will default to current user address
	Account         string
	EndPointOptions *EndPointOptions
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

type EndPointOptions struct {
	Endpoint  string // indicates the endpoint of sp
	SPAddress string // indicates the HEX-encoded string of the sp address to be challenged
}

type ListBucketsOptions struct {
	// ShowRemovedObject determines whether to include buckets that have been marked as removed in the list.
	// If set to false, these buckets will be skipped.
	ShowRemovedBucket bool

	EndPointOptions *EndPointOptions
}
