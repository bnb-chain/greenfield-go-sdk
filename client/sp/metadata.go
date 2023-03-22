package sp

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
	storageType "github.com/bnb-chain/greenfield/x/storage/types"
)

type UserInfo struct {
	Address string
}

// Object is the structure for user object
type Object struct {
	// object_info defines the information of the object.
	ObjectInfo *ObjectInfoSDK `json:"object_info,omitempty"`
	// locked_balance defines locked balance of object
	LockedBalance string `json:"locked_balance,omitempty"`
	// removed defines the object is deleted or not
	Removed bool `json:"removed,omitempty"`
}

// ObjectInfoSDK differ from ObjectInfo in greenfield as it adds uint64/int64 unmarshal guide in json part
type ObjectInfoSDK struct {
	Owner string `json:"owner,omitempty"`
	// bucket_name is the name of the bucket
	BucketName string `json:"bucket_name,omitempty"`
	// object_name is the name of object
	ObjectName string `json:"object_name,omitempty"`
	// id is the unique identifier of object
	Id storageType.Uint `json:"id"`
	// payloadSize is the total size of the object payload
	PayloadSize uint64 `json:"payload_size,string,omitempty"`
	// is_public define the highest permissions for object. When the object is public, everyone can access it.
	IsPublic bool `json:"is_public,omitempty"`
	// content_type define the format of the object which should be a standard MIME type.
	ContentType string `json:"content_type,omitempty"`
	// create_at define the block number when the object created
	CreateAt int64 `json:"create_at,string,omitempty"`
	// object_status define the upload status of the object.
	ObjectStatus storageType.ObjectStatus `json:"object_status,omitempty"`
	// redundancy_type define the type of the redundancy which can be multi-replication or EC.
	RedundancyType storageType.RedundancyType `json:"redundancy_type,omitempty"`
	// source_type define the source of the object.
	SourceType storageType.SourceType `json:"source_type,omitempty"`
	// checksums define the root hash of the pieces which stored in a SP.
	Checksums [][]byte `json:"checksums,omitempty" traits:"omit"`
	// secondary_sp_addresses define the addresses of secondary_sps
	SecondarySpAddresses []string `json:"secondary_sp_addresses,omitempty"`
}

// Bucket is the structure for user bucket
type Bucket struct {
	// bucket_info defines the information of the bucket.
	BucketInfo *BucketInfoSDK `protobuf:"bytes,1,opt,name=bucket_info,json=bucketInfo,proto3" json:"bucket_info,omitempty"`
	// removed defines the bucket is deleted or not
	Removed bool `protobuf:"varint,2,opt,name=removed,proto3" json:"removed,omitempty"`
}

// BucketInfoSDK differ from BucketInfo in greenfield as it adds uint64/int64 unmarshal guide in json part
type BucketInfoSDK struct {
	// owner is the account address of bucket creator, it is also the bucket owner.
	Owner string `json:"owner,omitempty"`
	// bucket_name is a globally unique name of bucket
	BucketName string `json:"bucket_name,omitempty"`
	// is_public define the highest permissions for bucket. When the bucket is public, everyone can get storage objects in it.
	IsPublic bool `json:"is_public,omitempty"`
	// id is the unique identification for bucket.
	Id storageType.Uint `json:"id"`
	// source_type defines which chain the user should send the bucket management transactions to
	SourceType storageType.SourceType `json:"source_type,omitempty"`
	// create_at define the block number when the bucket created, add "string" in json part for correct unmarshal
	CreateAt int64 `json:"create_at,string,omitempty"`
	// payment_address is the address of the payment account
	PaymentAddress string `json:"payment_address,omitempty"`
	// primary_sp_address is the address of the primary sp. Objects belongs to this bucket will never
	// leave this SP, unless you explicitly shift them to another SP.
	PrimarySpAddress string `json:"primary_sp_address,omitempty"`
	// read_quota defines the traffic quota for read in bytes per month, add "string" in json part for correct unmarshal
	ReadQuota uint64 `json:"read_quota,string,omitempty"`
	// billing info of the bucket
	BillingInfo storageType.BillingInfo `json:"billing_info"`
}

type ListObjectsResponse struct {
	// objects defines the list of object
	Objects []*Object `json:"objects,omitempty"`
}

type ListBucketsResponse struct {
	// buckets defines the list of bucket
	Buckets []*Bucket `json:"buckets"`
}

// ListObjects return object list of the specific bucket
func (c *SPClient) ListObjects(ctx context.Context, bucketName string, authInfo AuthInfo) (ListObjectsResponse, error) {
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return ListObjectsResponse{}, err
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		contentSHA256: EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsResponse{}, err
	}
	defer utils.CloseResponse(resp)

	ListObjectsResult := ListObjectsResponse{}
	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsResponse{}, err
	}

	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &ListObjectsResult)
	//TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil && ListObjectsResult.Objects == nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsResponse{}, err
	}

	return ListObjectsResult, nil
}

// ListBuckets list buckets for the owner
func (c *SPClient) ListBuckets(ctx context.Context, userInfo UserInfo, authInfo AuthInfo) (ListBucketsResponse, error) {
	if userInfo.Address == "" {
		return ListBucketsResponse{}, errors.New("fail to get user address")
	}

	reqMeta := requestMeta{
		contentSHA256: EmptyStringSHA256,
		userInfo: UserInfo{
			Address: userInfo.Address,
		},
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return ListBucketsResponse{}, err
	}
	defer utils.CloseResponse(resp)

	listBucketsResult := ListBucketsResponse{}
	//unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return ListBucketsResponse{}, err
	}

	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &listBucketsResult)

	//TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil && listBucketsResult.Buckets == nil {
		return ListBucketsResponse{}, err
	}

	return listBucketsResult, nil
}
