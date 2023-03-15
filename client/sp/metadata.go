package sp

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"

	storageType "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
)

type MetadataInfo struct {
	Address string
}

type Object struct {
	// object_info defines the information of the object.
	ObjectInfo *storageType.ObjectInfo `protobuf:"bytes,1,opt,name=object_info,json=objectInfo,proto3" json:"object_info,omitempty"`
	// locked_balance defines locked balance of object
	LockedBalance string `protobuf:"bytes,2,opt,name=locked_balance,json=lockedBalance,proto3" json:"locked_balance,omitempty"`
}

type Bucket struct {
	// bucket_info defines the information of the bucket.
	BucketInfo *storageType.BucketInfo `protobuf:"bytes,1,opt,name=bucket_info,json=bucketInfo,proto3" json:"bucket_info,omitempty"`
}

type ListObjectsByBucketNameResult struct {
	Objects []storageType.ObjectInfo
}

type GetUserBucketsResult struct {
	Buckets []storageType.BucketInfo
}

// ListObjectsByBucketName return object list of the specific bucket
func (c *SPClient) ListObjectsByBucketName(ctx context.Context, bucketName string, authInfo AuthInfo) (ListObjectsByBucketNameResult, error) {
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return ListObjectsByBucketNameResult{}, err
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
		log.Error().Msg("the list of user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsByBucketNameResult{}, err
	}
	defer utils.CloseResponse(resp)

	listObjectsByBucketNameResult := ListObjectsByBucketNameResult{}
	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)

	err = json.Unmarshal([]byte(buf.String()), &listObjectsByBucketNameResult)
	if err != nil {
		return ListObjectsByBucketNameResult{}, err
	}

	return listObjectsByBucketNameResult, nil
}

// GetUserBuckets get buckets for a specific user
func (c *SPClient) GetUserBuckets(ctx context.Context, metadataInfo MetadataInfo, authInfo AuthInfo) (GetUserBucketsResult, error) {
	if metadataInfo.Address == "" {
		return GetUserBucketsResult{}, errors.New("fail to get user address")
	}

	reqMeta := requestMeta{
		contentSHA256: EmptyStringSHA256,
		metadataInfo: MetadataInfo{
			Address: metadataInfo.Address,
		},
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return GetUserBucketsResult{}, err
	}
	defer utils.CloseResponse(resp)

	getUserBucketsResult := GetUserBucketsResult{}
	//unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)

	err = json.Unmarshal([]byte(buf.String()), &getUserBucketsResult)
	if err != nil {
		return GetUserBucketsResult{}, err
	}

	return getUserBucketsResult, nil
}
