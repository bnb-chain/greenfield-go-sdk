package sp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/bnb-chain/greenfield/types/s3util"
	"github.com/rs/zerolog/log"
)

type PutObjectOption struct {
	ContentType string
}

// PutObject supports the second stage of uploading the object to bucket.
// txnHash should be the str which hex.encoding from txn hash bytes
func (c *SPClient) PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
	reader io.Reader, authInfo AuthInfo, opt PutObjectOption,
) (err error) {
	if txnHash == "" {
		return errors.New("txn hash empty")
	}

	if objectSize <= 0 {
		return errors.New("object size not set")
	}

	var contentType string
	if opt.ContentType != "" {
		contentType = opt.ContentType
	} else {
		contentType = ContentDefault
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: EmptyStringSHA256,
		contentLength: objectSize,
		contentType:   contentType,
	}

	sendOpt := sendOptions{
		method:  http.MethodPut,
		body:    reader,
		txnHash: txnHash,
	}

	_, err = c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Printf("upload payload the object failed: %s \n", err.Error())
		return err
	}

	return nil
}

// FPutObject supports uploading object from local file
func (c *SPClient) FPutObject(ctx context.Context, bucketName, objectName,
	filePath, txnHash, contentType string, authInfo AuthInfo,
) (err error) {
	fReader, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		return err
	}
	defer fReader.Close()

	// Save the file stat.
	stat, err := fReader.Stat()
	if err != nil {
		return err
	}

	return c.PutObject(ctx, bucketName, objectName, txnHash, stat.Size(), fReader, authInfo, PutObjectOption{ContentType: contentType})
}

// ObjectInfo contains the metadata of downloaded objects
type ObjectInfo struct {
	ObjectName  string
	ContentType string
	Size        int64
}

// GetObjectOption contains the options of getObject
type GetObjectOption struct {
	Range string `url:"-" header:"Range,omitempty"` // support for downloading partial data
}

func (o *GetObjectOption) SetRange(start, end int64) error {
	switch {
	case 0 < start && end == 0:
		// `bytes=N-`.
		o.Range = fmt.Sprintf("bytes=%d-", start)
	case 0 <= start && start <= end:
		// `bytes=N-M`
		o.Range = fmt.Sprintf("bytes=%d-%d", start, end)
	default:
		return toInvalidArgumentResp(
			fmt.Sprintf(
				"Invalid Range : start=%d end=%d",
				start, end))
	}
	return nil
}

// GetObject download s3 object payload and return the related object info
func (c *SPClient) GetObject(ctx context.Context, bucketName, objectName string,
	opts GetObjectOption, authInfo AuthInfo) (io.ReadCloser, ObjectInfo, error) {
	if err := s3util.CheckValidBucketName(bucketName); err != nil {
		return nil, ObjectInfo{}, err
	}

	if err := s3util.CheckValidObjectName(objectName); err != nil {
		return nil, ObjectInfo{}, err
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: EmptyStringSHA256,
	}

	if opts.Range != "" {
		reqMeta.Range = opts.Range
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		return nil, ObjectInfo{}, err
	}

	ObjInfo, err := getObjInfo(objectName, resp.Header)
	if err != nil {
		utils.CloseResponse(resp)
		return nil, ObjectInfo{}, err
	}

	return resp.Body, ObjInfo, nil
}

// FGetObject download s3 object payload adn write the object content into local file specified by filePath
func (c *SPClient) FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts GetObjectOption, authinfo AuthInfo) error {
	// Verify if destination already exists.
	st, err := os.Stat(filePath)
	if err == nil {
		// If the destination exists and is a directory.
		if st.IsDir() {
			return errors.New("fileName is a directory.")
		}
	}

	// If file exist, open it in append mode
	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o660)
	if err != nil {
		return err
	}

	body, _, err := c.GetObject(ctx, bucketName, objectName, opts, authinfo)
	if err != nil {
		log.Printf("download object:%s fail %s \n", objectName, err.Error())
		return err
	}
	defer body.Close()

	_, err = io.Copy(fd, body)
	fd.Close()
	if err != nil {
		return err
	}

	return nil
}

// getObjInfo generates objectInfo base on the response http header content
func getObjInfo(objectName string, h http.Header) (ObjectInfo, error) {
	// Parse content length is exists
	var size int64 = -1
	var err error
	contentLength := h.Get(HTTPHeaderContentLength)
	if contentLength != "" {
		size, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return ObjectInfo{}, ErrResponse{
				Code:    "InternalError",
				Message: fmt.Sprintf("Content-Length parse error %v", err),
			}
		}
	}

	// fetch content type
	contentType := strings.TrimSpace(h.Get("Content-Type"))
	if contentType == "" {
		contentType = ContentDefault
	}

	return ObjectInfo{
		ObjectName:  objectName,
		ContentType: contentType,
		Size:        size,
	}, nil
}
