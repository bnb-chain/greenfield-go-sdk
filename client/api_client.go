package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	hashlib "github.com/bnb-chain/greenfield-common/go/hash"
	httplib "github.com/bnb-chain/greenfield-common/go/http"
	sdkerror "github.com/bnb-chain/greenfield-go-sdk/pkg/error"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	sdkclient "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Client interface {
	Basic
	Bucket
	Object
	Group
	Challenge
	Account
}

type client struct {
	// chainClients
	chainClient *sdkclient.GreenfieldClient
	// tendermintClient
	tendermintClient *sdkclient.TendermintClient
	// httpClient
	httpClient *http.Client
	// spEndpoints
	spEndpoints map[string]*url.URL

	// TODO (leo): Unused variables
	userAgent string

	host   string
	Secure bool
}

type Option struct {
	// keyManager is the manager used for generating and managing keys.
	KeyManager keys.KeyManager
	// grpcDialOption is the list of grpc dial options.
	GrpcDialOption grpc.DialOption
	// Use https or not
	Secure bool
}

// New - instantiate greenfield chain with options
func New(chainID string, grpcAddress, rpcAddress string, option *Option) (Client, error) {
	tc := sdkclient.NewTendermintClient(rpcAddress)
	cc := sdkclient.NewGreenfieldClient(grpcAddress, chainID, sdkclient.WithKeyManager(option.KeyManager), sdkclient.WithGrpcDialOption(option.GrpcDialOption))

	c := client{
		chainClient:      cc,
		tendermintClient: &tc,
		httpClient:       &http.Client{},
		userAgent:        types.UserAgent,
	}
	// fetch sp endpoints info from chain
	spInfo, err := c.GetSPAddrInfo()
	if err != nil {
		return nil, err
	}

	c.spEndpoints = spInfo
	return &c, nil
}

// getSPUrlByBucket route url of the sp from bucket name
func (c *client) getSPUrlByBucket(bucketName string) (*url.URL, error) {
	ctx := context.Background()
	bucketInfo, err := c.HeadBucket(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	primarySP := bucketInfo.GetPrimarySpAddress()
	if _, ok := c.spEndpoints[primarySP]; ok {
		return c.spEndpoints[primarySP], nil
	}
	// query sp info from chain
	newSpInfo, err := c.GetSPAddrInfo()
	if err != nil {
		return nil, err
	}

	if _, ok := newSpInfo[primarySP]; ok {
		c.spEndpoints = newSpInfo
		return newSpInfo[primarySP], nil
	} else {
		return nil, errors.New("fail to locate endpoint from bucket")
	}
}

// getSPUrlByAddr route url of the sp from sp address
func (c *client) getSPUrlByAddr(address string) (*url.URL, error) {
	if _, ok := c.spEndpoints[address]; ok {
		return c.spEndpoints[address], nil
	}
	// query sp info from chain
	newSpInfo, err := c.GetSPAddrInfo()
	if err != nil {
		return nil, err
	}

	if _, ok := newSpInfo[address]; ok {
		c.spEndpoints = newSpInfo
		return newSpInfo[address], nil
	} else {
		return nil, errors.New("fail to locate endpoint from bucket")
	}
}

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
	challengeInfo    types.ChallengeInfo
	userInfo         UserInfo
}

// SendOptions -  options to use to send the http message
type sendOptions struct {
	method           string      // request method
	body             interface{} // request body
	disableCloseBody bool        // indicate whether to disable automatic calls to resp.Body.Close()
	txnHash          string      // the transaction hash info
	isAdminApi       bool        // indicate if it is an admin api request
}

// newRequest constructs the http request, set url, body and headers
func (c *client) newRequest(ctx context.Context, method string, meta requestMeta,
	body interface{}, txnHash string, isAdminAPi bool, endpoint *url.URL, authInfo types.AuthInfo) (req *http.Request, err error) {
	// construct the target url
	desURL, err := c.generateURL(meta.bucketName, meta.objectName, meta.urlRelPath, meta.urlValues, isAdminAPi, endpoint)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	contentType := ""
	sha256Hex := ""
	if body != nil {
		// the body content is io.Reader type
		if ObjectReader, ok := body.(io.Reader); ok {
			reader = ObjectReader
			if meta.contentType == "" {
				contentType = types.ContentDefault
			}
		} else {
			// the body content is xml type
			content, err := xml.Marshal(body)
			if err != nil {
				return nil, err
			}
			contentType = types.ContentTypeXML
			reader = bytes.NewReader(content)
			sha256Hex = utils.CalcSHA256Hex(content)
		}
	}

	// Initialize a new HTTP request for the method.
	req, err = http.NewRequestWithContext(ctx, method, desURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// need to turn the body into ReadCloser
	if body == nil {
		req.Body = nil
	} else {
		req.Body = io.NopCloser(reader)
	}

	// set content length
	req.ContentLength = meta.contentLength

	// set txn hash header
	if txnHash != "" {
		req.Header.Set(types.HTTPHeaderTransactionHash, txnHash)
	}

	// set content type header
	if meta.contentType != "" {
		req.Header.Set(types.HTTPHeaderContentType, meta.contentType)
	} else if contentType != "" {
		req.Header.Set(types.HTTPHeaderContentType, contentType)
	} else {
		req.Header.Set(types.HTTPHeaderContentType, types.ContentDefault)
	}

	// set md5 header
	if meta.contentMD5Base64 != "" {
		req.Header[types.HTTPHeaderContentMD5] = []string{meta.contentMD5Base64}
	}

	// set sha256 header
	if meta.contentSHA256 != "" {
		req.Header[types.HTTPHeaderContentSHA256] = []string{meta.contentSHA256}
	} else {
		req.Header[types.HTTPHeaderContentSHA256] = []string{sha256Hex}
	}

	if meta.Range != "" && method == http.MethodGet {
		req.Header.Set(types.HTTPHeaderRange, meta.Range)
	}

	if isAdminAPi {
		// set challenge headers
		// if challengeInfo.ObjectId is not empty, other field should be set as well
		if meta.challengeInfo.ObjectId != "" {
			info := meta.challengeInfo
			req.Header.Set(types.HTTPHeaderObjectId, info.ObjectId)
			req.Header.Set(types.HTTPHeaderRedundancyIndex, strconv.Itoa(info.RedundancyIndex))
			req.Header.Set(types.HTTPHeaderPieceIndex, strconv.Itoa(info.PieceIndex))
		}

		if meta.TxnMsg != "" {
			req.Header.Set(types.HTTPHeaderUnsignedMsg, meta.TxnMsg)
		}

	} else {
		// set request host
		if c.host != "" {
			req.Host = c.host
		} else if req.URL.Host != "" {
			req.Host = req.URL.Host
		}
	}

	if meta.userInfo.Address != "" {
		info := meta.userInfo
		req.Header.Set(types.HTTPHeaderUserAddress, info.Address)
	}

	// set date header
	stNow := time.Now().UTC()
	req.Header.Set(types.HTTPHeaderDate, stNow.Format(types.Iso8601DateFormatSecond))

	// set user-agent
	// req.Header.Set(types.HTTPHeaderUserAgent, c.userAgent)

	// sign the total http request info when auth type v1
	err = c.signRequest(req, authInfo)
	if err != nil {
		return req, err
	}

	return
}

// doAPI call client.Do() to send request and read response from servers
func (c *client) doAPI(ctx context.Context, req *http.Request, meta requestMeta, closeBody bool) (*http.Response, error) {
	var cancel context.CancelFunc
	if closeBody {
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if urlErr, ok := err.(*url.Error); ok {
			if strings.Contains(urlErr.Err.Error(), "EOF") {
				return nil, &url.Error{
					Op:  urlErr.Op,
					URL: urlErr.URL,
					Err: errors.New("Connection closed by foreign host " + urlErr.URL + ". Retry again."),
				}
			}
		}
		return nil, err
	}
	defer func() {
		if closeBody {
			utils.CloseResponse(resp)
		}
	}()

	// construct err responses and messages
	err = sdkerror.ConstructErrResponse(resp, meta.bucketName, meta.objectName)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// sendReq sends the message via REST and handles the response
func (c *client) sendReq(ctx context.Context, metadata requestMeta, opt *sendOptions, authInfo types.AuthInfo, endpoint *url.URL) (res *http.Response, err error) {
	req, err := c.newRequest(ctx, opt.method, metadata, opt.body, opt.txnHash, opt.isAdminApi, endpoint, authInfo)
	if err != nil {
		log.Debug().Msg("new request error stop send request" + err.Error())
		return nil, err
	}

	resp, err := c.doAPI(ctx, req, metadata, !opt.disableCloseBody)
	if err != nil {
		log.Debug().Msg("do api request fail: " + err.Error())
		return nil, err
	}
	return resp, nil
}

// generateURL constructs the target request url based on the parameters
func (c *client) generateURL(bucketName string, objectName string, relativePath string,
	queryValues url.Values, isAdminApi bool, endpoint *url.URL) (*url.URL, error) {
	host := endpoint.Host
	scheme := endpoint.Scheme

	// Strip port 80 and 443
	if h, p, err := net.SplitHostPort(host); err == nil {
		if scheme == "http" && p == "80" || scheme == "https" && p == "443" {
			host = h
			if ip := net.ParseIP(h); ip != nil && ip.To16() != nil {
				host = "[" + h + "]"
			}
		}
	}

	var urlStr string
	if isAdminApi {
		prefix := types.AdminURLPrefix + types.AdminURLVersion
		urlStr = scheme + "://" + host + prefix + "/"
	} else {
		// generate s3 virtual hosted style url, consider case where ListBuckets not having a bucket name
		if utils.IsDomainNameValid(host) && bucketName != "" {
			urlStr = scheme + "://" + bucketName + "." + host + "/"
		} else {
			urlStr = scheme + "://" + host + "/"
		}

		if objectName != "" {
			urlStr += utils.EncodePath(objectName)
		}
	}

	if relativePath != "" {
		urlStr += utils.EncodePath(relativePath)
	}

	if len(queryValues) > 0 {
		urlStrNew, err := utils.AddQueryValues(urlStr, queryValues)
		if err != nil {
			return nil, err
		}
		urlStr = urlStrNew
	}

	return url.Parse(urlStr)
}

// signRequest signs the request and set authorization before send to server
func (c *client) signRequest(req *http.Request, info types.AuthInfo) error {
	var authStr []string
	if info.SignType == types.AuthV1 {
		signMsg := httplib.GetMsgToSign(req)

		km, err := c.chainClient.GetKeyManager()
		if err != nil {
			return err
		}
		// sign the request header info, generate the signature
		signature, err := km.Sign(signMsg)
		if err != nil {
			return err
		}

		authStr = []string{
			types.AuthV1 + " " + types.SignAlgorithm,
			" SignedMsg=" + hex.EncodeToString(signMsg),
			"Signature=" + hex.EncodeToString(signature),
		}

	} else if info.SignType == types.AuthV2 {
		if info.WalletSignStr == "" {
			return errors.New("wallet signature can not be empty with auth v2 type")
		}
		// wallet should use same sign algorithm
		authStr = []string{
			types.AuthV2 + " " + types.SignAlgorithm,
			" Signature=" + info.WalletSignStr,
		}
	} else {
		return errors.New("sign type error")
	}

	// set auth header
	req.Header.Set(types.HTTPHeaderAuthorization, strings.Join(authStr, ", "))

	return nil
}

// GetPieceHashRoots returns primary pieces, secondary piece Hash roots list and the object size
// It is used for generate meta of object on the chain
func (c *client) GetPieceHashRoots(reader io.Reader, segSize int64,
	dataShards, parityShards int) ([]byte, [][]byte, int64, storageTypes.RedundancyType, error) {
	pieceHashRoots, size, redundancyType, err := hashlib.ComputeIntegrityHash(reader, segSize, dataShards, parityShards)
	if err != nil {
		return nil, nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}

	return pieceHashRoots[0], pieceHashRoots[1:], size, redundancyType, nil
}

// sendPutPolicyTxn broadcast the putPolicy msg and return the txn hash
func (c *client) sendPutPolicyTxn(msg *storageTypes.MsgPutPolicy, txOpts *gnfdSdkTypes.TxOption) (string, error) {
	if err := msg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{msg}, txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// sendDelPolicyTxn broadcast the deletePolicy msg and return the txn hash
func (c *client) sendDelPolicyTxn(operator sdk.AccAddress, resource string, principal *permTypes.Principal, txOpts *gnfdSdkTypes.TxOption) (string, error) {
	delPolicyMsg := storageTypes.NewMsgDeletePolicy(operator, resource, principal)

	if err := delPolicyMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{delPolicyMsg}, txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}
