package sp

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
	sdktype "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/keys"
	signer "github.com/bnb-chain/greenfield-go-sdk/keys/signer"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
)

// SPClient is a client manages communication with the inscription API.
type SPClient struct {
	endpoint  *url.URL // Parsed endpoint url provided by the user.
	client    *http.Client
	userAgent string
	host      string

	conf       *SPClientConfig
	sender     sdktype.AccAddress // sender greenfield chain address
	keyManager keys.KeyManager
	signer     *signer.MsgSigner
}

// SPClientConfig is the config info of client
type SPClientConfig struct {
	Secure           bool // use https or not
	Transport        http.RoundTripper
	RetryOpt         RetryOptions
	UploadLimitSpeed uint64
}

type Option struct {
	secure bool
}

type RetryOptions struct {
	Count      int
	Interval   time.Duration
	StatusCode []int
}

type SpClientOption interface {
	Apply(*SPClient)
}

type SpClientOptionFunc func(*SPClient)

func (f SpClientOptionFunc) Apply(client *SPClient) {
	f(client)
}

func WithKeyManager(km keys.KeyManager) SpClientOption {
	return SpClientOptionFunc(func(client *SPClient) {
		err := client.SetKeyManager(km)
		if err != nil {
			panic(err)
		}
	})
}

func WithSecure(secure bool) SpClientOption {
	return SpClientOptionFunc(func(client *SPClient) {
		client.conf.Secure = secure
	})
}

// NewSpClient returns a new greenfield client
func NewSpClient(endpoint string, opts ...SpClientOption) (*SPClient, error) {
	httpClient := &http.Client{}
	c := &SPClient{
		client:    httpClient,
		userAgent: UserAgent,
		conf: &SPClientConfig{
			RetryOpt: RetryOptions{
				Count:    3,
				Interval: time.Duration(0),
			},
		},
	}

	for _, opt := range opts {
		opt.Apply(c)
	}

	url, err := utils.GetEndpointURL(endpoint, c.conf.Secure)
	if err != nil {
		return nil, toInvalidArgumentResp(err.Error())
	}

	c.SetUrl(url)

	return c, nil
}

// SetKeyManager set the keyManager and signer of client
func (c *SPClient) SetKeyManager(keyManager keys.KeyManager) error {
	if keyManager == nil {
		return errors.New("keyManager can not be nil")
	}

	if keyManager.GetPrivKey() == nil {
		return errors.New("private key must be set")
	}

	c.keyManager = keyManager

	signer := signer.NewMsgSigner(keyManager)
	c.signer = signer

	c.sender = keyManager.GetAddr()
	return nil
}

// GetKeyManager return the keyManager object
func (c *SPClient) GetKeyManager() (keys.KeyManager, error) {
	if c.keyManager == nil {
		return nil, types.ErrorKeyManagerNotInit
	}
	return c.keyManager, nil
}

// GetMsgSigner return the signer
func (c *SPClient) GetMsgSigner() (*signer.MsgSigner, error) {
	if c.signer == nil {
		return nil, errors.New("signer is nil")
	}
	return c.signer, nil
}

// GetURL returns the URL of the S3 endpoint.
func (c *SPClient) GetURL() *url.URL {
	endpoint := *c.endpoint
	return &endpoint
}

// requestMeta - contain the metadata to construct the http request.
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
	challengeInfo    ChallengeInfo
}

// sendOptions -  options to use to send the http message
type sendOptions struct {
	method           string      // request method
	body             interface{} // request body
	disableCloseBody bool        // indicate whether to disable automatic calls to resp.Body.Close()
	txnHash          string      // the transaction hash info
	isAdminApi       bool        // indicate if it is an admin api request
}

// SetHost set host name of request
func (c *SPClient) SetHost(hostName string) {
	c.host = hostName
}

// GetHost get host name of request
func (c *SPClient) GetHost() string {
	return c.host
}

func (c *SPClient) SetUrl(url *url.URL) {
	c.endpoint = url
}

// SetAccount set client sender address
func (c *SPClient) SetAccount(addr sdktype.AccAddress) {
	c.sender = addr
}

// GetAccount get sender address info
func (c *SPClient) GetAccount() sdktype.AccAddress {
	return c.sender
}

// GetAgent get agent name
func (c *SPClient) GetAgent() string {
	return c.userAgent
}

// newRequest construct the http request, set url, body and headers
func (c *SPClient) newRequest(ctx context.Context,
	method string, meta requestMeta, body interface{}, txnHash string, isAdminAPi bool, authInfo AuthInfo) (req *http.Request, err error) {
	// construct the target url
	desURL, err := c.GenerateURL(meta.bucketName, meta.objectName, meta.urlRelPath, meta.urlValues, isAdminAPi)
	log.Debug().Msg("new request Url:" + desURL.String())
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
				contentType = ContentDefault
			}
		} else {
			// the body content is xml type
			content, err := xml.Marshal(body)
			if err != nil {
				return nil, err
			}
			contentType = contentTypeXML
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
		req.Header.Set(HTTPHeaderTransactionHash, txnHash)
	}

	// set content type header
	if meta.contentType != "" {
		req.Header.Set(HTTPHeaderContentType, meta.contentType)
	} else if contentType != "" {
		req.Header.Set(HTTPHeaderContentType, contentType)
	} else {
		req.Header.Set(HTTPHeaderContentType, ContentDefault)
	}

	// set md5 header
	if meta.contentMD5Base64 != "" {
		req.Header[HTTPHeaderContentMD5] = []string{meta.contentMD5Base64}
	}

	// set sha256 header
	if meta.contentSHA256 != "" {
		req.Header[HTTPHeaderContentSHA256] = []string{meta.contentSHA256}
	} else {
		req.Header[HTTPHeaderContentSHA256] = []string{sha256Hex}
	}

	if meta.Range != "" && method == http.MethodGet {
		req.Header.Set(HTTPHeaderRange, meta.Range)
	}

	if isAdminAPi {
		// set challenge headers
		// if challengeInfo.ObjectId is not empty, other field should be set as well
		if meta.challengeInfo.ObjectId != "" {
			info := meta.challengeInfo
			req.Header.Set(HTTPHeaderObjectId, info.ObjectId)
			req.Header.Set(HTTPHeaderRedundancyIndex, strconv.Itoa(info.RedundancyIndex))
			req.Header.Set(HTTPHeaderPieceIndex, strconv.Itoa(info.PieceIndex))
		}

		if meta.TxnMsg != "" {
			req.Header.Set(HTTPHeaderUnsignedMsg, meta.TxnMsg)
		}

	} else {
		// set request host
		if c.host != "" {
			req.Host = c.host
		} else if req.URL.Host != "" {
			req.Host = req.URL.Host
		}
	}

	// set date header
	stNow := time.Now().UTC()
	req.Header.Set(HTTPHeaderDate, stNow.Format(iso8601DateFormatSecond))

	// set user-agent
	req.Header.Set(HTTPHeaderUserAgent, c.userAgent)

	// sign the total http request info when auth type v1
	err = c.SignRequest(req, authInfo)
	if err != nil {
		return req, err
	}

	return
}

// doAPI call client.Do() to send request and read response from servers
func (c *SPClient) doAPI(ctx context.Context, req *http.Request, meta requestMeta, closeBody bool) (*http.Response, error) {
	var cancel context.CancelFunc
	if closeBody {
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
	}
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
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
	err = constructErrResponse(resp, meta.bucketName, meta.objectName)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// sendReq new restful request, send the message and handle the response
func (c *SPClient) sendReq(ctx context.Context, metadata requestMeta, opt *sendOptions, authInfo AuthInfo) (res *http.Response, err error) {
	req, err := c.newRequest(ctx, opt.method, metadata, opt.body, opt.txnHash, opt.isAdminApi, authInfo)
	if err != nil {
		log.Error().Msg("new request error stop send request" + err.Error())
		return nil, err
	}

	resp, err := c.doAPI(ctx, req, metadata, !opt.disableCloseBody)
	if err != nil {
		log.Error().Msg("do api request fail: " + err.Error())
		return nil, err
	}
	return resp, nil
}

// GenerateURL construct the target request url based on the parameters
func (c *SPClient) GenerateURL(bucketName string, objectName string, relativePath string,
	queryValues url.Values, isAdminApi bool) (*url.URL, error) {
	host := c.endpoint.Host
	scheme := c.endpoint.Scheme

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
		prefix := AdminURLPrefix + AdminURLVersion
		urlStr = scheme + "://" + host + prefix + "/"
	} else {
		if bucketName == "" {
			err := errors.New("no BucketName in path")
			return nil, err
		}

		// generate s3 virtual hosted style url
		if utils.IsDomainNameValid(host) {
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

// SignRequest sign the request and set authorization before send to server
func (c *SPClient) SignRequest(req *http.Request, info AuthInfo) error {
	var authStr []string
	if info.SignType == AuthV1 {
		signMsg := httplib.GetMsgToSign(req)

		if c.signer == nil {
			return errors.New("signer can not be nil with auth v1 type")
		}

		// sign the request header info, generate the signature
		signature, _, err := c.signer.Sign(signMsg)
		if err != nil {
			return err
		}

		authStr = []string{
			AuthV1 + " " + SignAlgorithm,
			" SignedMsg=" + hex.EncodeToString(signMsg),
			"Signature=" + hex.EncodeToString(signature),
		}

	} else if info.SignType == AuthV2 {
		if info.WalletSignStr == "" {
			return errors.New("wallet signature can not be empty with auth v2 type")
		}
		// wallet should use same sign algorithm
		authStr = []string{
			AuthV2 + " " + SignAlgorithm,
			" Signature=" + info.WalletSignStr,
		}
	} else {
		return errors.New("sign type error")
	}

	// set auth header
	req.Header.Set(HTTPHeaderAuthorization, strings.Join(authStr, ", "))

	return nil
}

// GetPieceHashRoots return primary pieces Hash and secondary piece Hash roots list and object size
// It is used for generate meta of object on the chain
func (c *SPClient) GetPieceHashRoots(reader io.Reader, segSize int64, dataShards, parityShards int) (string, []string, int64, error) {
	pieceHashRoots, size, err := hashlib.ComputerHash(reader, segSize, dataShards, parityShards)
	if err != nil {
		log.Error().Msg("get hash roots fail" + err.Error())
		return "", nil, 0, err
	}

	return pieceHashRoots[0], pieceHashRoots[1:], size, nil
}
