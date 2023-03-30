package api

import (
	"net/url"

	"github.com/bnb-chain/greenfield-go-sdk/client"
)

// requestMeta - contains the metadata to construct the http request.
type requestMeta struct {
	bucketName       string
	objectName       string
	urlRelPath       string     // relative path of url
	urlValues        url.Values // url values to be added into url
	Range            string
	ApproveAction    string
	TxnMsg           string
	SignType         string
	contentType      string
	contentLength    int64
	contentMD5Base64 string // base64 encoded md5sum
	contentSHA256    string // hex encoded sha256sum
	challengeInfo    client.ChallengeInfo
	userInfo         client.UserInfo
}

// AuthInfo is the authorization info of requests
type AuthInfo struct {
	SignType      string // if using wallet sign, set authV2
	WalletSignStr string
}

// SendOptions -  options to use to send the http message
type sendOptions struct {
	method           string      // request method
	body             interface{} // request body
	disableCloseBody bool        // indicate whether to disable automatic calls to resp.Body.Close()
	txnHash          string      // the transaction hash info
	isAdminApi       bool        // indicate if it is an admin api request
}

// ObjectInfo contains the metadata of downloaded objects
type ObjectInfo struct {
	ObjectName  string
	ContentType string
	Size        int64
}
