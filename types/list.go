package types

import (
	"encoding/xml"

	storageType "github.com/bnb-chain/greenfield/x/storage/types"
)

// QuotaInfo indicates the quota info of bucket
type QuotaInfo struct {
	XMLName             xml.Name `xml:"GetReadQuotaResult"`
	Version             string   `xml:"version,attr"`
	BucketName          string   `xml:"BucketName"`
	BucketID            string   `xml:"BucketID"`            // BucketID defines the bucket read quota value on chain
	ReadQuotaSize       uint64   `xml:"ReadQuotaSize"`       // ReadQuotaSize defines the bucket read quota value on chain
	SPFreeReadQuotaSize uint64   `xml:"SPFreeReadQuotaSize"` // SPFreeReadQuotaSize defines the free quota of this month
	ReadConsumedSize    uint64   `xml:"ReadConsumedSize"`    // ReadConsumedSize defines the consumed total read quota of this month
	FreeConsumedSize    uint64   `xml:"FreeConsumedSize"`    // FreeConsumedSize defines the consumed free quota
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

type ListObjectsResult struct {
	// objects defines the list of object
	Objects               []*ObjectMeta `xml:"Objects"`
	KeyCount              string        `xml:"KeyCount"`
	MaxKeys               string        `xml:"MaxKeys"`
	IsTruncated           bool          `xml:"IsTruncated"`
	NextContinuationToken string        `xml:"NextContinuationToken"`
	Name                  string        `xml:"Name"`
	Prefix                string        `xml:"Prefix"`
	Delimiter             string        `xml:"Delimiter"`
	CommonPrefixes        []string      `xml:"CommonPrefixes"`
	ContinuationToken     string        `xml:"ContinuationToken"`
}

type ListBucketsResult struct {
	// buckets defines the list of bucket
	Buckets []*BucketMetaWithVGF `xml:"Buckets"`
}

type ListBucketsByPaymentAccountResult struct {
	// buckets defines the list of bucket
	Buckets []*BucketMeta `xml:"Buckets"`
}

type ListUserPaymentAccountsResult struct {
	// PaymentAccount defines the list of payment accounts
	PaymentAccounts []*PaymentAccounts `xml:"PaymentAccounts"`
}

type ListGroupsResult struct {
	// groups defines the response of group list
	Groups []*GroupMeta `json:"Groups"`
	// count defines total groups amount
	Count int64 `xml:"Count"`
}

type GroupMembersResult struct {
	// groups defines the response of group member list
	Groups []*GroupMembers `xml:"Groups"`
}

type GroupsResult struct {
	// groups defines the response of group member list
	Groups []*GroupMembers `xml:"Groups"`
}

type GroupMembers struct {
	// group defines the basic group info
	Group *GroupInfo `xml:"Group"`
	// operator defines operator address of group
	Operator string `xml:"Operator"`
	// create_at defines the block number when the group created
	CreateAt int64 `xml:"CreateAt"`
	// create_time defines the timestamp when the group created
	CreateTime int64 `xml:"CreateTime"`
	// update_at defines the block number when the group updated
	UpdateAt int64 `xml:"UpdateAt"`
	// update_time defines the timestamp when the group updated
	UpdateTime int64 `xml:"UpdateTime"`
	// removed defines the group is deleted or not
	Removed bool `xml:"Removed"`
	// the address of account
	AccountID string `xml:"AccountId"`
	// ExpirationTime is the user expiration time for this group
	ExpirationTime string `xml:"ExpirationTime"`
}

// ObjectMeta is the structure for metadata service user object
type ObjectMeta struct {
	// object_info defines the information of the object.
	ObjectInfo *ObjectInfo `xml:"ObjectInfo"`
	// locked_balance defines locked balance of object
	LockedBalance string `xml:"LockedBalance"`
	// removed defines the object is deleted or not
	Removed bool `xml:"Removed"`
	// update_at defines the block number when the object updated
	UpdateAt int64 `xml:"UpdateAt"`
	// delete_at defines the block number when the object deleted
	DeleteAt int64 `xml:"DeleteAt"`
	// delete_reason defines the deleted reason of object
	DeleteReason string `xml:"DeleteReason"`
	// operator defines the operator address of object
	Operator string `xml:"Operator"`
	// create_tx_hash defines the creation transaction hash of object
	CreateTxHash string `xml:"CreateTxHash"`
	// update_tx_hash defines the update transaction hash of object
	UpdateTxHash string `xml:"UpdateTxHash"`
	// seal_tx_hash defines the sealed transaction hash of object
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
	// id is the identifier of the global virtual group family.
	Id uint32 `xml:"Id"`
	// primary_sp_id is the id of primary sp
	PrimarySpId uint32 `xml:"PrimarySpId"`
	// global_virtual_group_ids is a list of identifiers of the global virtual groups associated with the family.
	GlobalVirtualGroupIds []uint32 `xml:"GlobalVirtualGroupIds"`
	// virtual_payment_address is the payment address associated with the global virtual group family.
	VirtualPaymentAddress string `xml:"VirtualPaymentAddress"`
}

// BucketMetaWithVGF BucketMeta is the structure for metadata service user bucket
type BucketMetaWithVGF struct {
	// bucket_info defines the information of the bucket.
	BucketInfo *BucketInfo `xml:"BucketInfo"`
	// removed defines the bucket is deleted or not
	Removed bool `xml:"Removed"`
	// delete_at defines the block number when the bucket deleted.
	DeleteAt int64 `xml:"DeleteAt"`
	// delete_reason defines the deleted reason of bucket
	DeleteReason string `xml:"DeleteReason"`
	// operator defines the operator address of bucket
	Operator string `xml:"Operator"`
	// create_tx_hash defines the creation transaction hash of bucket
	CreateTxHash string `xml:"CreateTxHash"`
	// update_tx_hash defines the update transaction hash of bucket
	UpdateTxHash string `xml:"UpdateTxHash"`
	// update_at defines the block number when the bucket updated
	UpdateAt int64 `xml:"UpdateAt"`
	// update_time defines the block number when the bucket updated
	UpdateTime int64 `xml:"UpdateTime"`
	// GlobalVirtualGroupFamily serve as a means of grouping global virtual groups.
	Vgf *GlobalVirtualGroupFamily `xml:"Vgf"`
}

// BucketMeta is the structure for metadata service user bucket
type BucketMeta struct {
	// bucket_info defines the information of the bucket.
	BucketInfo *BucketInfo `xml:"BucketInfo"`
	// removed defines the bucket is deleted or not
	Removed bool `xml:"Removed"`
	// delete_at defines the block number when the bucket deleted.
	DeleteAt int64 `xml:"DeleteAt"`
	// delete_reason defines the deleted reason of bucket
	DeleteReason string `xml:"DeleteReason"`
	// operator defines the operator address of bucket
	Operator string `xml:"Operator"`
	// create_tx_hash defines the creation transaction hash of bucket
	CreateTxHash string `xml:"CreateTxHash"`
	// update_tx_hash defines the update transaction hash of bucket
	UpdateTxHash string `xml:"UpdateTxHash"`
	// update_at defines the block number when the bucket updated
	UpdateAt int64 `xml:"UpdateAt"`
	// update_time defines the block number when the bucket updated
	UpdateTime int64 `xml:"UpdateTime"`
}

// ObjectInfo differ from ObjectInfo in greenfield as it adds uint64/int64 unmarshal guide in json part
type ObjectInfo struct {
	Owner string `xml:"Owner"`
	// bucket_name is the name of the bucket
	BucketName string `xml:"BucketName"`
	// object_name is the name of object
	ObjectName string `xml:"ObjectName"`
	// id is the unique identifier of object
	Id                  uint64 `xml:"Id"`
	LocalVirtualGroupId uint32 `xml:"LocalVirtualGroupId"`
	// payloadSize is the total size of the object payload
	PayloadSize uint64 `xml:"PayloadSize"`
	// visibility defines the highest permissions for object. When an object is public, everyone can access it.
	Visibility storageType.VisibilityType `xml:"Visibility"`
	// content_type define the format of the object which should be a standard MIME type.
	ContentType string `xml:"ContentType"`
	// create_at define the block number when the object created
	CreateAt int64 `xml:"CreateAt"`
	// object_status define the upload status of the object.
	ObjectStatus storageType.ObjectStatus `xml:"ObjectStatus"`
	// redundancy_type define the type of the redundancy which can be multi-replication or EC.
	RedundancyType storageType.RedundancyType `xml:"RedundancyType"`
	// source_type define the source of the object.
	SourceType storageType.SourceType `xml:"SourceType"`
	// checksums define the root hash of the pieces which stored in a SP.
	Checksums [][]byte `xml:"Checksums"`
}

// BucketInfo differ from BucketInfo in greenfield as it adds uint64/int64 unmarshal guide in json part
type BucketInfo struct {
	// owner is the account address of bucket creator, it is also the bucket owner.
	Owner string `xml:"Owner"`
	// bucket_name is a globally unique name of bucket
	BucketName string `xml:"BucketName"`
	// visibility defines the highest permissions for bucket. When a bucket is public, everyone can get storage objects in it.
	Visibility storageType.VisibilityType `xml:"Visibility"`
	// id is the unique identification for bucket.
	Id uint64 `xml:"Id"`
	// source_type defines which chain the user should send the bucket management transactions to
	SourceType storageType.SourceType `xml:"SourceType"`
	// create_at define the block number when the bucket created
	CreateAt int64 `xml:"CreateAt"`
	// payment_address is the address of the payment account
	PaymentAddress string `xml:"PaymentAddress"`
	// primary_sp_id is the unique id of the primary sp. Objects belongs to this bucket will never
	// leave this SP, unless you explicitly shift them to another SP.
	PrimarySpId uint32 `xml:"PrimarySpId"`
	// global_virtual_group_family_id defines the unique id of gvg family
	GlobalVirtualGroupFamilyId uint32 `xml:"GlobalVirtualGroupFamilyId"`
	// charged_read_quota defines the traffic quota for read in bytes per month.
	// The available read data for each user is the sum of the free read data provided by SP and
	// the ChargeReadQuota specified here.
	ChargedReadQuota uint64 `xml:"ChargedReadQuota"`
	// bucket_status define the status of the bucket.
	BucketStatus storageType.BucketStatus `xml:"BucketStatus"`
}

// ListBucketsByBucketIDResponse is response type for the ListBucketsByBucketID
type ListBucketsByBucketIDResponse struct {
	// buckets defines the information of a bucket map
	Buckets map[uint64]*BucketMeta `xml:"Buckets"`
}

// ListGroupsByGroupIDResponse is response type for the ListGroupsByGroupID
type ListGroupsByGroupIDResponse struct {
	// Groups defines the information of a group map
	Groups map[uint64]*GroupMeta `xml:"Groups"`
}

// GroupMeta is the structure for group information
type GroupMeta struct {
	// group defines the basic group info
	Group *GroupInfo `xml:"Group"`
	// NumberOfMembers defines how many members in this group
	NumberOfMembers int64 `xml:"NumberOfMembers"`
	// operator defines operator address of group
	Operator string `xml:"Operator"`
	// create_at defines the block number when the group created
	CreateAt int64 `xml:"CreateAt"`
	// create_time defines the timestamp when the group created
	CreateTime int64 `xml:"CreateTime"`
	// update_at defines the block number when the group updated
	UpdateAt int64 `xml:"UpdateAt"`
	// update_time defines the timestamp when the group updated
	UpdateTime int64 `xml:"UpdateTime"`
	// removed defines the group is deleted or not
	Removed bool `xml:"Removed"`
}

// GroupInfo differ from GroupInfo in greenfield as it adds uint64/int64 unmarshal guide in json part
type GroupInfo struct {
	// owner is the owner of the group. It can not changed once it created.
	Owner string `xml:"Owner"`
	// group_name is the name of group which is unique under an account.
	GroupName string `xml:"GroupName"`
	// source_type
	SourceType storageType.SourceType `xml:"SourceType"`
	// id is the unique identifier of group
	Id uint64 `xml:"Id"`
	// extra is used to store extra info for the group
	Extra string `xml:"Extra"`
}

type PaymentAccounts struct {
	// refundable defines the payment account is refundable or not
	PaymentAccount *PaymentAccount `xml:"PaymentAccount"`
	// stream_records defines stream payment records of a stream account
	StreamRecord *StreamRecord `xml:"StreamRecord"`
}

// StreamRecord defines Record of a stream account
type StreamRecord struct {
	// account address
	Account string `xml:"Account"`
	// latest update timestamp of the stream record
	CrudTimestamp int64 `xml:"CrudTimestamp"`
	// The per-second rate that an account's balance is changing.
	// It is the sum of the account's inbound and outbound flow rates.
	NetflowRate int64 `xml:"NetflowRate"`
	// The balance of the stream account at the latest CRUD timestamp.
	StaticBalance int64 `xml:"StaticBalance"`
	// reserved balance of the stream account
	// If the netflow rate is negative, the reserved balance is `netflow_rate * reserve_time`
	BufferBalance int64 `xml:"BufferBalance"`
	// the locked balance of the stream account after it puts a new object and before the object is sealed
	LockBalance int64 `xml:"LockBalance"`
	// the status of the stream account
	Status int32 `xml:"Status"`
	// the unix timestamp when the stream account will be settled
	SettleTimestamp int64 `xml:"SettleTimestamp"`
	// the count of its out flows
	OutFlowCount uint64 `xml:"OutFlowCount"`
	// the frozen netflow rate, which is used when resuming stream account
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

type ListObjectPoliciesResponse struct {
	Policies []*PolicyMeta `xml:"Policies"`
}

type PolicyMeta struct {
	// principal_type defines the type of principal
	PrincipalType int32 `xml:"PrincipalType"`
	// principal_value defines the value of principal
	PrincipalValue string `xml:"PrincipalValue"`
	// resource_type defines the type of resource that grants permission for
	ResourceType int32 `xml:"ResourceType"`
	// resource_id defines the bucket/object/group id of the resource that grants permission for
	ResourceId string `xml:"ResourceId"`
	// create_timestamp defines the create time of permission
	CreateTimestamp int64 `xml:"CreateTimestamp"`
	// update_timestamp defines the update time of permission
	UpdateTimestamp int64 `xml:"UpdateTimestamp"`
	// expiration_time defines the expiration time of permission
	ExpirationTime int64 `xml:"ExpirationTime"`
}
