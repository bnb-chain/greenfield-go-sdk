package sp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield/types/s3util"

	"github.com/bnb-chain/greenfield-go-sdk/types"
)

type UserInfo struct {
	Address string
}

type ListObjectsResponse struct {
	// objects defines the list of object
	Objects []*types.Object `json:"objects"`
}

type ListBucketsResponse struct {
	// buckets defines the list of bucket
	Buckets []*types.Bucket `json:"buckets"`
}

// ListObjects return object list of the specific bucket
func (c *SPClient) ListObjects(ctx context.Context, bucketName string, authInfo AuthInfo) (ListObjectsResponse, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
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
