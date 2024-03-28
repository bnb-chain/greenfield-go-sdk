package types

import (
	"encoding/xml"

	storageType "github.com/bnb-chain/greenfield/x/storage/types"
)

// QuotaInfo indicates the quota info of bucket
type QuotaInfo struct {
	XMLName                 xml.Name `xml:"GetReadQuotaResult"`
	Version                 string   `xml:"version,attr"`
	BucketName              string   `xml:"BucketName"`
	BucketID                string   `xml:"BucketID"`                 // BucketID defines the bucket read quota value on chain
	ReadQuotaSize           uint64   `xml:"ReadQuotaSize"`            // ReadQuotaSize defines the bucket read quota value on chain
	SPFreeReadQuotaSize     uint64   `xml:"SPFreeReadQuotaSize"`      // SPFreeReadQuotaSize defines the free quota of this month
	ReadConsumedSize        uint64   `xml:"ReadConsumedSize"`         // ReadConsumedSize defines the consumed total read quota of this month
	FreeConsumedSize        uint64   `xml:"FreeConsumedSize"`         // FreeConsumedSize defines the consumed free quota
	MonthlyFreeQuota        uint64   `xml:"MonthlyFreeQuota"`         // MonthlyFreeQuota defines the consumed monthly free quota
	MonthlyFreeConsumedSize uint64   `xml:"MonthlyQuotaConsumedSize"` // MonthlyFreeConsumedSize defines the consumed monthly free quota
}

// ReadRecord indicate the download record info
type ReadRecord struct {
	XMLName            xml.Name `xml:"ReadRecord"`
	ObjectName         string   `xml:"ObjectName"`         // ObjectName The download object name
	ObjectID           string   `xml:"ObjectID"`           // ObjectID The download object id
	ReadAccountAddress string   `xml:"ReadAccountAddress"` // ReadAccountAddress The sender address of the download request
	ReadTimestampUs    int64    `xml:"ReadTimestampUs"`    // ReadTimestampUs The download time stamp
	ReadSize           uint64   `xml:"ReadSize"`           // ReadSize The download object size
}

// QuotaRecordInfo indicates the quota read record
type QuotaRecordInfo struct {
	XMLName xml.Name `xml:"GetBucketReadQuotaResult"`
	Version string   `xml:"version,attr"` // The Version defines version info
	// When using ListBucketReadRecord to list items, if the returned results do not cover all items, the NextStartTimestampUs will be returned to indicate the timestamp of the current traversal progress.
	// When you call the ListBucketReadRecord again, you can set opt.StartTimeStamp to this timestamp.
	NextStartTimestampUs int64 `xml:"NextStartTimestampUs"`
	// ReadRecords defines the result record list.
	ReadRecords []ReadRecord `xml:"ReadRecord"`
}

// UploadProgress indicates the progress info of uploading object
type UploadProgress struct {
	XMLName             xml.Name `xml:"QueryUploadProgress"`
	Version             string   `xml:"version,attr"`        // Version defines version info
	ProgressDescription string   `xml:"ProgressDescription"` // ProgressDescription defines a string message representing the upload progress.
	ErrorDescription    string   `xml:"ErrorDescription"`    // ErrorDescription defines a string message representing an upload error exception.
}

// UploadOffset indicates the offset of resumable uploading object
type UploadOffset struct {
	XMLName xml.Name `xml:"QueryResumeOffset"`
	Version string   `xml:"version,attr"` // Version defines version info
	Offset  uint64   `xml:"Offset"`       // Offset defines the offset info of resumable uploading object
}

// MigrationProgress indicates the progress info of bucket migration
type MigrationProgress struct {
	XMLName             xml.Name `xml:"QueryMigrationProgress"`
	Version             string   `xml:"version,attr"`        // Version defines version info
	ProgressDescription string   `xml:"ProgressDescription"` // ProgressDescription defines a string message representing the upload progress.
	ErrorDescription    string   `xml:"ErrorDescription"`    // ErrorDescription defines a string message representing an upload error exception.
	MigratedBytes       uint64   `xml:"MigratedBytes"`
	MigrationState      uint64   `xml:"MigrationState"`
}

// ChallengeV2Result indicates the response info of challenge v2 API
type ChallengeV2Result struct {
	XMLName         xml.Name `xml:"GetChallengeInfo"`
	Version         string   `xml:"version,attr"`
	ObjectID        string   `xml:"ObjectID"`        // ObjectID defines the object id of the challenge request
	RedundancyIndex string   `xml:"RedundancyIndex"` // RedundancyIndex defines the redundancy index of the challenge request
	PieceIndex      string   `xml:"PieceIndex"`      // PieceIndex defines the piece index of the challenge request
	IntegrityHash   string   `xml:"IntegrityHash"`   // IntegrityHash defines the integrity hash of the object
	PieceHash       string   `xml:"PieceHash"`       // PieceHash defines the return piece hashes of the object
	PieceData       string   `xml:"PieceData"`       // PieceData defines the return piece data of challenge request
}

// ListObjectsResult indicates the result of listObjects API.
type ListObjectsResult struct {
	// Objects defines the list of object
	Objects []*ObjectMeta `xml:"Objects"`
	// KeyCount is the number of keys returned with this request
	KeyCount string `xml:"KeyCount"`
	// MaxKeys sets the maximum number of keys returned to the response
	MaxKeys string `xml:"MaxKeys"`
	// IsTruncated set to false if all of the results were returned. set to true if more keys are available to return
	IsTruncated bool `xml:"IsTruncated"`
	// NextContinuationToken is sent when is_truncated is true, which means there are more keys in the bucket that can be listed
	NextContinuationToken string `xml:"NextContinuationToken"`
	// Name defines the name of the bucket
	Name string `xml:"Name"`
	// Prefix is the prefix used during the query
	Prefix string `xml:"Prefix"`
	// Delimiter is the delimiter used during the query
	Delimiter string `xml:"Delimiter"`
	// CommonPrefixes defines a list of strings representing common prefixes. common_prefixes are those parts of object key names that fall between the specified delimiters
	CommonPrefixes []string `xml:"CommonPrefixes"`
	// ContinuationToken is the continuation token used during the query
	ContinuationToken string `xml:"ContinuationToken"`
}

// ListBucketsResult defines the response of  list bucekts response
type ListBucketsResult struct {
	// Buckets defines the list of bucket
	Buckets []*BucketMetaWithVGF `xml:"Buckets"`
}

// ListBucketsByPaymentAccountResult defines the response of list buckets by payment account
type ListBucketsByPaymentAccountResult struct {
	// Buckets defines the list of bucket
	Buckets []*BucketMeta `xml:"Buckets"`
}

// ListUserPaymentAccountsResult defines the response of list user payment accounts
type ListUserPaymentAccountsResult struct {
	// PaymentAccount defines the list of payment accounts
	PaymentAccounts []*PaymentAccounts `xml:"PaymentAccounts"`
}

// ListGroupsResult define the response of list groups
type ListGroupsResult struct {
	// Groups defines the response of group list
	Groups []*GroupMeta `json:"Groups"`
	// Count defines total groups amount
	Count int64 `xml:"Count"`
}

// GroupMembersResult indicates the response of ListGroupMembers
type GroupMembersResult struct {
	// Groups defines the response of group member list
	Groups []*GroupMembers `xml:"Groups"`
}

// GroupsResult indicates a list of group members
type GroupsResult struct {
	// Groups defines the response of group member list
	Groups []*GroupMembers `xml:"Groups"`
}

// GroupMembers indicates the group member info
type GroupMembers struct {
	// Group defines the basic group info
	Group *storageType.GroupInfo `xml:"Group"`
	// Operator defines operator address of group
	Operator string `xml:"Operator"`
	// CreateAt defines the block number when the group created
	CreateAt int64 `xml:"CreateAt"`
	// CreateTime defines the timestamp when the group created
	CreateTime int64 `xml:"CreateTime"`
	// UpdateAt defines the block number when the group updated
	UpdateAt int64 `xml:"UpdateAt"`
	// UpdateTime defines the timestamp when the group updated
	UpdateTime int64 `xml:"UpdateTime"`
	// Removed defines the group is deleted or not
	Removed bool `xml:"Removed"`
	// AccountID defines the address of account
	AccountID string `xml:"AccountId"`
	// ExpirationTime is the user expiration time for this group
	ExpirationTime string `xml:"ExpirationTime"`
}

// ObjectMeta is the structure for metadata service user object
type ObjectMeta struct {
	// ObjectInfo defines the information of the object.
	ObjectInfo *storageType.ObjectInfo `xml:"ObjectInfo"`
	// LockedBalance defines locked balance of object
	LockedBalance string `xml:"LockedBalance"`
	// Removed defines the object is deleted or not
	Removed bool `xml:"Removed"`
	// UpdateAt defines the block number when the object updated
	UpdateAt int64 `xml:"UpdateAt"`
	// DeleteAt defines the block number when the object deleted
	DeleteAt int64 `xml:"DeleteAt"`
	// DeleteReason defines the deleted reason of object
	DeleteReason string `xml:"DeleteReason"`
	// Operator defines the operator address of object
	Operator string `xml:"Operator"`
	// CreateTxHash defines the creation transaction hash of object
	CreateTxHash string `xml:"CreateTxHash"`
	// UpdateTxHash defines the update transaction hash of object
	UpdateTxHash string `xml:"UpdateTxHash"`
	// SealTxHash defines the sealed transaction hash of object
	SealTxHash string `xml:"SealTxHash"`
}

// ListObjectsByObjectIDResponse is response type for the ListObjectsByObjectID
type ListObjectsByObjectIDResponse struct {
	// objects defines the information of a object map
	Objects map[uint64]*ObjectMeta `xml:"Objects"`
}

// ObjectAndBucketIDs is the structure for ListBucketsByBucketID & ListObjectsByObjectID request body
type ObjectAndBucketIDs struct {
	IDs []uint64 `xml:"IDs"`
}

// GlobalVirtualGroupFamily serve as a means of grouping global virtual groups.
// Each bucket must be associated with a unique global virtual group family and cannot cross families.
type GlobalVirtualGroupFamily struct {
	// Id is the identifier of the global virtual group family.
	Id uint32 `xml:"Id"`
	// PrimarySpId is the id of primary sp
	PrimarySpId uint32 `xml:"PrimarySpId"`
	// GlobalVirtualGroupIds is a list of identifiers of the global virtual groups associated with the family.
	GlobalVirtualGroupIds []uint32 `xml:"GlobalVirtualGroupIds"`
	// VirtualPaymentAddress is the payment address associated with the global virtual group family.
	VirtualPaymentAddress string `xml:"VirtualPaymentAddress"`
}

// BucketMetaWithVGF BucketMeta is the structure for metadata service user bucket
type BucketMetaWithVGF struct {
	// BucketInfo defines the information of the bucket.
	BucketInfo *storageType.BucketInfo `xml:"BucketInfo"`
	// Removed defines the bucket is deleted or not
	Removed bool `xml:"Removed"`
	// DeleteAt defines the block number when the bucket deleted.
	DeleteAt int64 `xml:"DeleteAt"`
	// DeleteReason defines the deleted reason of bucket
	DeleteReason string `xml:"DeleteReason"`
	// Operator defines the operator address of bucket
	Operator string `xml:"Operator"`
	// CreateTxHash defines the creation transaction hash of bucket
	CreateTxHash string `xml:"CreateTxHash"`
	// UpdateTxHash defines the update transaction hash of bucket
	UpdateTxHash string `xml:"UpdateTxHash"`
	// UpdateAt defines the block number when the bucket updated
	UpdateAt int64 `xml:"UpdateAt"`
	// UpdateTime defines the block number when the bucket updated
	UpdateTime int64 `xml:"UpdateTime"`
	// Vgf serve as a means of grouping global virtual groups.
	Vgf *GlobalVirtualGroupFamily `xml:"Vgf"`
	// OffChainStatus represents the status of a bucket in the off-chain storage.
	// It is used to track the current state of the bucket with respect to off-chain operations,
	// 1 means 0001 -> OffChainStatusIsLimited is true
	// 0 means 0000 -> OffChainStatusIsLimited is false
	// For an explanation of the different OffChainStatus values, please visit:https://github.com/bnb-chain/greenfield-storage-provider/blob/9d7048ad33cf51a2f7eb347e2113c5d0cc45f970/modular/blocksyncer/modules/bucket/bucket_handle.go#L40
	OffChainStatus int32 `xml:"OffChainStatus"`
}

// BucketMeta is the structure for metadata service user bucket
type BucketMeta struct {
	// BucketInfo defines the information of the bucket.
	BucketInfo *storageType.BucketInfo `xml:"BucketInfo"`
	// Removed defines the bucket is deleted or not
	Removed bool `xml:"Removed"`
	// DeleteAt defines the block number when the bucket deleted.
	DeleteAt int64 `xml:"DeleteAt"`
	// DeleteReason defines the deleted reason of bucket
	DeleteReason string `xml:"DeleteReason"`
	// Operator defines the operator address of bucket
	Operator string `xml:"Operator"`
	// CreateTxHash defines the creation transaction hash of bucket
	CreateTxHash string `xml:"CreateTxHash"`
	// UpdateTxHash defines the update transaction hash of bucket
	UpdateTxHash string `xml:"UpdateTxHash"`
	// UpdateAt defines the block number when the bucket updated
	UpdateAt int64 `xml:"UpdateAt"`
	// UpdateTime defines the block number when the bucket updated
	UpdateTime int64 `xml:"UpdateTime"`
	// OffChainStatus represents the status of a bucket in the off-chain storage.
	// It is used to track the current state of the bucket with respect to off-chain operations,
	// 1 means 0001 -> OffChainStatusIsLimited is true
	// 0 means 0000 -> OffChainStatusIsLimited is false
	// For an explanation of the different OffChainStatus values, please visit:https://github.com/bnb-chain/greenfield-storage-provider/blob/9d7048ad33cf51a2f7eb347e2113c5d0cc45f970/modular/blocksyncer/modules/bucket/bucket_handle.go#L40
	OffChainStatus int32 `xml:"OffChainStatus"`
}

// ListBucketsByBucketIDResponse is response type for the ListBucketsByBucketID
type ListBucketsByBucketIDResponse struct {
	// Buckets defines the information of a bucket map
	Buckets map[uint64]*BucketMeta `xml:"Buckets"`
}

// ListGroupsByGroupIDResponse is response type for the ListGroupsByGroupID
type ListGroupsByGroupIDResponse struct {
	// Groups defines the information of a group map
	Groups map[uint64]*GroupMeta `xml:"Groups"`
}

// GroupMeta is the structure for group information
type GroupMeta struct {
	// Group defines the basic group info
	Group *storageType.GroupInfo `xml:"Group"`
	// NumberOfMembers defines how many members in this group
	NumberOfMembers int64 `xml:"NumberOfMembers"`
	// Operator defines operator address of group
	Operator string `xml:"Operator"`
	// CreateAt defines the block number when the group created
	CreateAt int64 `xml:"CreateAt"`
	// CreateTime defines the timestamp when the group created
	CreateTime int64 `xml:"CreateTime"`
	// UpdateAt defines the block number when the group updated
	UpdateAt int64 `xml:"UpdateAt"`
	// UpdateTime defines the timestamp when the group updated
	UpdateTime int64 `xml:"UpdateTime"`
	// Removed defines the group is deleted or not
	Removed bool `xml:"Removed"`
}

type PaymentAccounts struct {
	// refundable defines the payment account is refundable or not
	PaymentAccount *PaymentAccount `xml:"PaymentAccount"`
	// stream_records defines stream payment records of a stream account
	StreamRecord *StreamRecord `xml:"StreamRecord"`
}

// StreamRecord defines Record of a stream account
type StreamRecord struct {
	// Account address
	Account string `xml:"Account"`
	// CrudTimestamp defines latest update timestamp of the stream record
	CrudTimestamp int64 `xml:"CrudTimestamp"`
	// NetflowRate defines The per-second rate that an account's balance is changing.
	// It is the sum of the account's inbound and outbound flow rates.
	NetflowRate int64 `xml:"NetflowRate"`
	// StaticBalance defines The balance of the stream account at the latest CRUD timestamp.
	StaticBalance int64 `xml:"StaticBalance"`
	// BufferBalance defines reserved balance of the stream account
	// If the netflow rate is negative, the reserved balance is `netflow_rate * reserve_time`
	BufferBalance int64 `xml:"BufferBalance"`
	// LockBalance defines the locked balance of the stream account after it puts a new object and before the object is sealed
	LockBalance int64 `xml:"LockBalance"`
	// Status defines the status of the stream account
	Status int32 `xml:"Status"`
	// SettleTimestamp defines the unix timestamp when the stream account will be settled
	SettleTimestamp int64 `xml:"SettleTimestamp"`
	// OutFlowCount defines the count of its out flows
	OutFlowCount uint64 `xml:"OutFlowCount"`
	// FrozenNetflowRate defines the frozen netflow rate, which is used when resuming stream account
	FrozenNetflowRate int64 `xml:"FrozenNetflowRate"`
}

// PaymentAccount defines payment account info
type PaymentAccount struct {
	// Address defines the address of payment account
	Address string `xml:"Address"`
	// Owner defines the owner of this payment account
	Owner string `xml:"Owner"`
	// Refundable defines the payment account is refundable or not
	Refundable bool `xml:"Refundable"`
	// UpdateAt defines the update block height of this payment account
	UpdateAt int64 `xml:"UpdateAt"`
	// UpdateTime defines the update time of this payment account
	UpdateTime int64 `xml:"UpdateTime"`
}

// ListObjectPoliciesResponse define the response of list object policies
type ListObjectPoliciesResponse struct {
	Policies []*PolicyMeta `xml:"Policies"`
}

// PolicyMeta defines the policy info
type PolicyMeta struct {
	// PrincipalType defines the type of principal
	PrincipalType int32 `xml:"PrincipalType"`
	// PrincipalValue defines the value of principal
	PrincipalValue string `xml:"PrincipalValue"`
	// ResourceType defines the type of resource that grants permission for
	ResourceType int32 `xml:"ResourceType"`
	// ResourceId defines the bucket/object/group id of the resource that grants permission for
	ResourceId string `xml:"ResourceId"`
	// CreateTimestamp defines the create time of permission
	CreateTimestamp int64 `xml:"CreateTimestamp"`
	// UpdateTimestamp defines the update time of permission
	UpdateTimestamp int64 `xml:"UpdateTimestamp"`
	// ExpirationTime defines the expiration time of permission
	ExpirationTime int64 `xml:"ExpirationTime"`
}
