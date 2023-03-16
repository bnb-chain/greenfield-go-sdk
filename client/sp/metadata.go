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

type UserInfo struct {
	Address string
}

// Object is the structure for user object
type Object struct {
	// object_info defines the information of the object.
	ObjectInfo *storageType.ObjectInfo `protobuf:"bytes,1,opt,name=object_info,json=objectInfo,proto3" json:"object_info,omitempty"`
	// locked_balance defines locked balance of object
	LockedBalance string `protobuf:"bytes,2,opt,name=locked_balance,json=lockedBalance,proto3" json:"locked_balance,omitempty"`
	// removed defines the object is deleted or not
	Removed bool `protobuf:"varint,3,opt,name=removed,proto3" json:"removed,omitempty"`
}

// Bucket is the structure for user bucket
type Bucket struct {
	// bucket_info defines the information of the bucket.
	BucketInfo *storageType.BucketInfo `protobuf:"bytes,1,opt,name=bucket_info,json=bucketInfo,proto3" json:"bucket_info,omitempty"`
	// removed defines the bucket is deleted or not
	Removed bool `protobuf:"varint,2,opt,name=removed,proto3" json:"removed,omitempty"`
}

type ListObjectsByBucketNameResponse struct {
	// objects defines the list of object
	Objects []*Object `protobuf:"bytes,1,rep,name=objects,proto3" json:"objects,omitempty"`
}

type ListBucketsByUserResponse struct {
	// buckets defines the list of bucket
	Buckets []*Bucket `protobuf:"bytes,1,rep,name=buckets,proto3" json:"buckets"`
}

// ListObjectsByBucketName return object list of the specific bucket
func (c *SPClient) ListObjectsByBucketName(ctx context.Context, bucketName string, authInfo AuthInfo) (ListObjectsByBucketNameResponse, error) {
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return ListObjectsByBucketNameResponse{}, err
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
		return ListObjectsByBucketNameResponse{}, err
	}
	defer utils.CloseResponse(resp)

	listObjectsByBucketNameResult := ListObjectsByBucketNameResponse{}
	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsByBucketNameResponse{}, err
	}

	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &listObjectsByBucketNameResult)
	if err != nil {
		log.Error().Msg("the list of objects in user's bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsByBucketNameResponse{}, err
	}

	return listObjectsByBucketNameResult, nil
}

// ListBucketsByUser list buckets for a specific user
func (c *SPClient) ListBucketsByUser(ctx context.Context, userInfo UserInfo, authInfo AuthInfo) (ListBucketsByUserResponse, error) {
	if userInfo.Address == "" {
		return ListBucketsByUserResponse{}, errors.New("fail to get user address")
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
		return ListBucketsByUserResponse{}, err
	}
	defer utils.CloseResponse(resp)

	getUserBucketsResult := ListBucketsByUserResponse{}
	//unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return ListBucketsByUserResponse{}, err
	}

	bufStr := buf.String()
	err = json.Unmarshal([]byte(bufStr), &getUserBucketsResult)

	if err != nil && getUserBucketsResult.Buckets == nil {
		return ListBucketsByUserResponse{}, err
	}

	return getUserBucketsResult, nil
}
