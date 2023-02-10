package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/bnb-chain/gnfd-go-sdk/keys"
	signer "github.com/bnb-chain/gnfd-go-sdk/keys/signer"
	"github.com/bnb-chain/gnfd-go-sdk/utils"
)

const (
	HTTPHeaderAuthorization = "Authorization"
	SignAlgorithm           = "ECDSA-secp256k1"
	AuthV1                  = "authTypeV1"
	AuthV2                  = "authTypeV2"
)

// AuthInfo is the authorization info of requests
type AuthInfo struct {
	SignType      string // if using wallet sign, set authV2
	WalletSignStr string
}

// NewAuthInfo return the AuthInfo which need to pass to api
// useWalletSign indicate whether you need use wallet to sign
// signStr indicate the wallet signature or jwt token
func NewAuthInfo(useWalletSign bool, signStr string) AuthInfo {
	if !useWalletSign {
		return AuthInfo{
			SignType:      AuthV1,
			WalletSignStr: "",
		}
	} else {
		return AuthInfo{
			SignType:      AuthV2,
			WalletSignStr: signStr,
		}
	}
}

// getCanonicalHeaders generate a list of request headers with their values
func getCanonicalHeaders(req *http.Request) string {
	var content bytes.Buffer
	var containHostHeader bool
	sortHeaders := getSortedHeaders(req)
	headerMap := make(map[string][]string)
	for key, data := range req.Header {
		headerMap[strings.ToLower(key)] = data
	}

	for _, header := range sortHeaders {
		content.WriteString(strings.ToLower(header))
		content.WriteByte(':')

		if header != "host" {
			for i, v := range headerMap[header] {
				if i > 0 {
					content.WriteByte(',')
				}
				trimVal := strings.Join(strings.Fields(v), " ")
				content.WriteString(trimVal)
			}
			content.WriteByte('\n')
		} else {
			containHostHeader = true
			content.WriteString(getHostInfo(req))
			content.WriteByte('\n')
		}
	}

	if !containHostHeader {
		content.WriteString(getHostInfo(req))
		content.WriteByte('\n')
	}

	return content.String()
}

// getSignedHeaders return the sorted header array
func getSortedHeaders(req *http.Request) []string {
	var signHeaders []string
	for k := range req.Header {
		headerKey := http.CanonicalHeaderKey(k)
		if headerKey != HTTPHeaderAuthorization && headerKey != "User-Agent" &&
			headerKey != "Accept-Encoding" && headerKey != "Content-Length" {
			signHeaders = append(signHeaders, strings.ToLower(k))
		}
	}
	sort.Strings(signHeaders)
	return signHeaders
}

// getSignedHeaders return the alphabetically sorted, semicolon-separated list of lowercase request header names.
func getSignedHeaders(req *http.Request) string {
	return strings.Join(getSortedHeaders(req), ";")
}

// getCanonicalRequest generate the canonicalRequest base on aws s3 sign without payload hash.
// https://docs.aws.amazon.com/general/latest/gr/create-signed-request.html#create-canonical-request
func GetCanonicalRequest(req *http.Request) string {
	req.URL.RawQuery = strings.ReplaceAll(req.URL.Query().Encode(), "+", "%20")
	canonicalRequest := strings.Join([]string{
		req.Method,
		utils.EncodePath(req.URL.Path),
		req.URL.RawQuery,
		getCanonicalHeaders(req),
		getSignedHeaders(req),
	}, "\n")

	return canonicalRequest
}

// GetMsgToSign generate the msg bytes from canonicalRequest to sign
func GetMsgToSign(req *http.Request) []byte {
	signBytes := calcSHA256([]byte(GetCanonicalRequest(req)))
	return crypto.Keccak256(signBytes)
}

// SignRequest sign the request and set authorization before send to server
func SignRequest(req *http.Request, keyManager keys.KeyManager, info AuthInfo) error {
	var signature []byte
	var err error
	var authStr []string
	if info.SignType == AuthV1 {
		if keyManager.GetPrivKey() == nil {
			return errors.New("private key must be set when using sign v1 mode")
		}
		signMsg := GetMsgToSign(req)
		// sign the request header info, generate the signature
		signer := signer.NewMsgSigner(keyManager)
		signature, _, err = signer.Sign(signMsg)
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
			return errors.New("wallet signature can not be empty when using sign v2 types")
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

func calcSHA256(msg []byte) (sum []byte) {
	h := sha256.New()
	h.Write(msg)
	sum = h.Sum(nil)
	return
}

// getHostInfo returns host header from the request
func getHostInfo(req *http.Request) string {
	host := req.Header.Get("host")
	if host != "" {
		return host
	}
	if req.Host != "" {
		return req.Host
	}
	return req.URL.Host
}
