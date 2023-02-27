package sp

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
)

type ListObjectsResult struct {
	objectList []ObjectInfo
	prefix     string
}

// ListObjects return object name list of the specific bucket
func (c *SPClient) ListObjects(ctx context.Context, bucketName, objectPrefix string, maxkeys int, authInfo AuthInfo) (ListObjectsResult, error) {
	if err := utils.VerifyBucketName(bucketName); err != nil {
		return ListObjectsResult{}, err
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		contentSHA256: EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}
	urlValues := make(url.Values)
	urlValues.Set("prefix", objectPrefix)
	if maxkeys > 0 {
		urlValues.Set("max-keys", fmt.Sprintf("%d", maxkeys))
	}
	reqMeta.urlValues = urlValues

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("listObjects of bucket:" + bucketName + " failed: " + err.Error())
		return ListObjectsResult{}, err
	}
	defer utils.CloseResponse(resp)

	listObjectsResult := ListObjectsResult{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&listObjectsResult)
	if err != nil {
		return ListObjectsResult{}, err
	}

	return listObjectsResult, nil
}
