package client

import (
	"encoding/xml"
	"io"

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

// GetObjectResult contains the metadata of download objects
type GetObjectResult struct {
	ObjectName  string
	ContentType string
	Size        int64
}

// AuthInfo is the authorization info of requests
type AuthInfo struct {
	SignType      string // if using wallet sign, set authV2
	WalletSignStr string
}

type UserInfo struct {
	Address string
}

type ListObjectsResponse struct {
	// objects defines the list of object
	Objects []*ObjectMeta `json:"objects"`
}

type ListBucketsResponse struct {
	// buckets defines the list of bucket
	Buckets []*BucketMeta `json:"buckets"`
}

// ObjectMeta is the structure for metadata service user object
type ObjectMeta struct {
	// object_info defines the information of the object.
	ObjectInfo *GetObjectResult `json:"object_info"`
	// locked_balance defines locked balance of object
	LockedBalance string `json:"locked_balance"`
	// removed defines the object is deleted or not
	Removed bool `json:"removed"`
}

// BucketMeta is the structure for metadata service user bucket
type BucketMeta struct {
	// bucket_info defines the information of the bucket.
	BucketInfo *BucketInfo `protobuf:"bytes,1,opt,name=bucket_info,json=bucketInfo,proto3" json:"bucket_info"`
	// removed defines the bucket is deleted or not
	Removed bool `protobuf:"varint,2,opt,name=removed,proto3" json:"removed"`
}

// GetObjectResult differ from GetObjectResult in greenfield as it adds uint64/int64 unmarshal guide in json part
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
	// is_public define the highest permissions for object. When the object is public, everyone can access it.
	IsPublic bool `json:"is_public"`
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
	// is_public define the highest permissions for bucket. When the bucket is public, everyone can get storage objects in it.
	IsPublic bool `json:"is_public"`
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
	// read_quota defines the traffic quota for read in bytes per month, add "string" in json part for correct unmarshal
	ReadQuota uint64 `json:"read_quota,string"`
	// billing info of the bucket
	BillingInfo storageType.BillingInfo `json:"billing_info"`
}

// ChallengeInfo indicates the challenge object info
// RedundancyIndex if it is primary sp, the value should be -1，
// else it indicates the index of secondary sp
type ChallengeInfo struct {
	ObjectId        string
	PieceIndex      int
	RedundancyIndex int
}

// ChallengeResult indicates the challenge hash and data results
type ChallengeResult struct {
	PieceData     io.ReadCloser
	IntegrityHash string
	PiecesHash    []string
}
