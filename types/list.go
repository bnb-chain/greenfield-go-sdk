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
	BucketID            string   `xml:"BucketID"`
	ReadQuotaSize       uint64   `xml:"ReadQuotaSize"`       // the bucket read quota value on chain
	SPFreeReadQuotaSize uint64   `xml:"SPFreeReadQuotaSize"` // the free quota of this month
	ReadConsumedSize    uint64   `xml:"ReadConsumedSize"`    // the consumed read quota of this month
}

type ReadRecord struct {
	XMLName            xml.Name `xml:"ReadRecord"`
	ObjectName         string   `xml:"ObjectName"`
	ObjectID           string   `xml:"ObjectID"`
	ReadAccountAddress string   `xml:"ReadAccountAddress"`
	ReadTimestampUs    int64    `xml:"ReadTimestampUs"`
	ReadSize           uint64   `xml:"ReadSize"`
}

// QuotaRecordInfo indicates the quota read record
type QuotaRecordInfo struct {
	XMLName              xml.Name     `xml:"GetBucketReadQuotaResult"`
	Version              string       `xml:"version,attr"`
	NextStartTimestampUs int64        `xml:"NextStartTimestampUs"`
	ReadRecords          []ReadRecord `xml:"ReadRecord"`
}

// UploadProgress indicates the progress info of uploading object
type UploadProgress struct {
	XMLName             xml.Name `xml:"QueryUploadProgress"`
	Version             string   `xml:"version,attr"`
	ProgressDescription string   `xml:"ProgressDescription"`
}

type ListObjectsResult struct {
	// objects defines the list of object
	Objects               []*ObjectMeta `json:"objects"`
	KeyCount              string        `json:"key_count"`
	MaxKeys               string        `json:"max_keys"`
	IsTruncated           bool          `json:"is_truncated"`
	NextContinuationToken string        `json:"next_continuation_token"`
	Name                  string        `json:"name"`
	Prefix                string        `json:"prefix"`
	Delimiter             string        `json:"delimiter"`
	CommonPrefixes        []string      `json:"common_prefixes"`
	ContinuationToken     string        `json:"continuation_token"`
}

type ListBucketsResult struct {
	// buckets defines the list of bucket
	Buckets []*BucketMeta `json:"buckets"`
}

// ObjectMeta is the structure for metadata service user object
type ObjectMeta struct {
	// object_info defines the information of the object.
	ObjectInfo *ObjectInfo `json:"object_info"`
	// locked_balance defines locked balance of object
	LockedBalance string `json:"locked_balance"`
	// removed defines the object is deleted or not
	Removed bool `json:"removed"`
	// update_at defines the block number when the object updated
	UpdateAt int64 `json:"update_at,string"`
	// delete_at defines the block number when the object deleted
	DeleteAt int64 `json:"delete_at,string"`
	// delete_reason defines the deleted reason of object
	DeleteReason string `json:"delete_reason"`
	// operator defines the operator address of object
	Operator string `json:"operator"`
	// create_tx_hash defines the creation transaction hash of object
	CreateTxHash string `json:"create_tx_hash"`
	// update_tx_hash defines the update transaction hash of object
	UpdateTxHash string `json:"update_tx_hash"`
	// seal_tx_hash defines the sealed transaction hash of object
	SealTxHash string `json:"seal_tx_hash"`
}

// BucketMeta is the structure for metadata service user bucket
type BucketMeta struct {
	// bucket_info defines the information of the bucket.
	BucketInfo *BucketInfo `json:"bucket_info"`
	// removed defines the bucket is deleted or not
	Removed bool `json:"removed"`
	// delete_at defines the block number when the bucket deleted.
	DeleteAt int64 `json:"delete_at,string"`
	// delete_reason defines the deleted reason of bucket
	DeleteReason string `json:"delete_reason"`
	// operator defines the operator address of bucket
	Operator string `json:"operator"`
	// create_tx_hash defines the creation transaction hash of bucket
	CreateTxHash string `json:"create_tx_hash"`
	// update_tx_hash defines the update transaction hash of bucket
	UpdateTxHash string `json:"update_tx_hash"`
	// update_at defines the block number when the bucket updated
	UpdateAt int64 `json:"update_at,string"`
	// update_time defines the block number when the bucket updated
	UpdateTime int64 `json:"update_time,string"`
}

// ObjectInfo differ from ObjectInfo in greenfield as it adds uint64/int64 unmarshal guide in json part
type ObjectInfo struct {
	Owner string `json:"owner"`
	// bucket_name is the name of the bucket
	BucketName string `json:"bucket_name"`
	// object_name is the name of object
	ObjectName string `json:"object_name"`
	// id is the unique identifier of object
	Id storageType.Uint `json:"id"`
	// payloadSize is the total size of the object payload
	PayloadSize uint64 `json:"payload_size,string"`
	// visibility defines the highest permissions for object. When an object is public, everyone can access it.
	Visibility storageType.VisibilityType `json:"visibility"`
	// content_type define the format of the object which should be a standard MIME type.
	ContentType string `json:"content_type"`
	// create_at define the block number when the object created
	CreateAt int64 `json:"create_at,string"`
	// object_status define the upload status of the object.
	ObjectStatus storageType.ObjectStatus `json:"object_status"`
	// redundancy_type define the type of the redundancy which can be multi-replication or EC.
	RedundancyType storageType.RedundancyType `json:"redundancy_type"`
	// source_type define the source of the object.
	SourceType storageType.SourceType `json:"source_type"`
	// checksums define the root hash of the pieces which stored in a SP.
	Checksums [][]byte `json:"checksums" traits:"omit"`
	// secondary_sp_addresses define the addresses of secondary_sps
	SecondarySpAddresses []string `json:"secondary_sp_addresses"`
}

// BucketInfo differ from BucketInfo in greenfield as it adds uint64/int64 unmarshal guide in json part
type BucketInfo struct {
	// owner is the account address of bucket creator, it is also the bucket owner.
	Owner string `json:"owner"`
	// bucket_name is a globally unique name of bucket
	BucketName string `json:"bucket_name"`
	// visibility defines the highest permissions for bucket. When a bucket is public, everyone can get storage objects in it.
	Visibility storageType.VisibilityType `json:"visibility"`
	// id is the unique identification for bucket.
	Id storageType.Uint `json:"id"`
	// source_type defines which chain the user should send the bucket management transactions to
	SourceType storageType.SourceType `json:"source_type"`
	// create_at define the block number when the bucket created, add "string" in json part for correct unmarshal
	CreateAt int64 `json:"create_at,string"`
	// payment_address is the address of the payment account
	PaymentAddress string `json:"payment_address"`
	// primary_sp_address is the address of the primary sp. Objects belongs to this bucket will never
	// leave this SP, unless you explicitly shift them to another SP.
	PrimarySpAddress string `json:"primary_sp_address"`
	// charged_read_quota defines the traffic quota for read in bytes per month.
	// The available read data for each user is the sum of the free read data provided by SP and
	// the ChargeReadQuota specified here.
	ChargedReadQuota uint64 `json:"charged_read_quota,string"`
	// billing info of the bucket
	BillingInfo storageType.BillingInfo `json:"billing_info"`
	// bucket_status define the status of the bucket.
	BucketStatus storageType.BucketStatus `json:"bucket_status"`
}
