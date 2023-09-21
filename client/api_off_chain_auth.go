package client

import (
	"bytes"
	"crypto/subtle"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	httplib "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/blake2b"
)

// IAuthClient - Client APIs for register Greenfield off chain auth keys and make signatures.
type IAuthClient interface {
	RegisterEDDSAPublicKey(spAddress string, spEndpoint string) (string, error)
	GetNextNonce(spEndpoint string) (string, error)
	OffChainAuthSign(unsignBytes []byte) string
}

// OffChainAuthSign - generate EdDSA private key according to a preconfigured seed and then make the signature for given []byte
//
// - ret1: The signature made by EdDSA private key of the Client.
func (c *Client) OffChainAuthSign(unsignBytes []byte) string {
	sk, _ := generateEddsaPrivateKey(c.offChainAuthOption.Seed)
	hFunc := mimc.NewMiMC()
	sig, _ := sk.Sign(unsignBytes, hFunc)
	authString := fmt.Sprintf("%s,Signature=%v", httplib.Gnfd1Eddsa, hex.EncodeToString(sig))
	return authString
}

// requestNonceResp is the structure for off chain auth nonce response
type requestNonceResp struct {
	CurrentNonce     int32  `xml:"CurrentNonce"`
	NextNonce        int32  `xml:"NextNonce"`
	CurrentPublicKey string `xml:"CurrentPublicKey"`
	ExpiryDate       int64  `xml:"ExpiryDate"`
}

// GetNextNonce get the nonce value by giving user account and domain
//
// - ret1: The next nonce value for the Client if it needs to register a new EdDSA public key
//
// - ret2: Return error when getting next nonce failed, otherwise return nil.
func (c *Client) GetNextNonce(spEndpoint string) (string, error) {
	header := make(map[string]string)
	header["X-Gnfd-User-Address"] = c.defaultAccount.GetAddress().String()
	header["X-Gnfd-App-Domain"] = c.offChainAuthOption.Domain

	response, err := httpGetWithHeader(spEndpoint+"/auth/request_nonce", header)
	if err != nil {
		return "0", err
	}
	authNonce := requestNonceResp{}
	// decode the xml content from response body
	err = xml.NewDecoder(bytes.NewBufferString(response)).Decode(&authNonce)
	if err != nil {
		return "0", err
	}
	return strconv.Itoa(int(authNonce.NextNonce)), nil
}

const (
	unsignedContentTemplate string = `%s wants you to sign in with your BNB Greenfield account:
%s

Register your identity public key %s

URI: %s
Version: 1
Chain ID: 5600
Issued At: %s
Expiration Time: %s
Resources:
- SP %s (name: SP_001) with nonce: %s`
)

// RegisterEDDSAPublicKey register EdDSA public key of this client for the given sp address and spEndpoint
//
// To enable EdDSA authentication, you need to config OffChainAuthOption for the client.
// The overall register process could be referred to https://docs.bnbchain.org/greenfield-docs/docs/guide/storage-provider/modules/authenticator#workflow.
//
// The EdDSA registering process is typically used in a website, e.g. https://dcellar.io,
// which obtains a user's signature via a wallet and then posts the user's EdDSA public key to a sp.
//
// Here we also provide an SDK method to implement this process, because sometimes you might want to test if a given SP provides correct EdDSA authentication or not.
// It also helps if you want implement it on a non-browser environment.
//
// - ret1: The register result when invoking SP UpdateUserPublicKey API.
//
// - ret2: Return error when registering failed, otherwise return nil.
func (c *Client) RegisterEDDSAPublicKey(spAddress string, spEndpoint string) (string, error) {
	appDomain := c.offChainAuthOption.Domain
	eddsaSeed := c.offChainAuthOption.Seed
	nextNonce, err := c.GetNextNonce(spEndpoint)
	if err != nil {
		return "", err
	}
	// get the EDDSA private and public key
	userEddsaPublicKeyStr := getEddsaCompressedPublicKey(eddsaSeed)
	log.Info().Msg("userEddsaPublicKeyStr is " + userEddsaPublicKeyStr)

	IssueDate := time.Now().Format(time.RFC3339)
	// ExpiryDate format := "2023-06-27T06:35:24Z"
	ExpiryDate := time.Now().Add(time.Hour * 24).Format(time.RFC3339)

	unSignedContent := fmt.Sprintf(unsignedContentTemplate, appDomain, c.defaultAccount.GetAddress().String(), userEddsaPublicKeyStr, appDomain, IssueDate, ExpiryDate, spAddress, nextNonce)

	unSignedContentHash := accounts.TextHash([]byte(unSignedContent))
	sig, _ := c.defaultAccount.GetKeyManager().Sign(unSignedContentHash)
	authString := fmt.Sprintf("%s,SignedMsg=%s,Signature=%s", httplib.Gnfd1EthPersonalSign, unSignedContent, hexutil.Encode(sig))
	authString = strings.ReplaceAll(authString, "\n", "\\n")
	headers := make(map[string]string)
	headers["x-gnfd-app-domain"] = appDomain
	headers["x-gnfd-app-reg-nonce"] = nextNonce
	headers["x-gnfd-app-reg-public-key"] = userEddsaPublicKeyStr
	headers["X-Gnfd-Expiry-Timestamp"] = ExpiryDate
	headers["authorization"] = authString
	headers["origin"] = appDomain
	headers["x-gnfd-user-address"] = c.defaultAccount.GetAddress().String()
	jsonResult, error1 := httpPostWithHeader(spEndpoint+"/auth/update_key", "{}", headers)

	return jsonResult, error1
}

func httpGetWithHeader(url string, header map[string]string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if (nil != resp) && (nil != resp.Body) {
		defer resp.Body.Close()
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", err
	}
	return string(body), err
}

func httpPostWithHeader(url string, jsonStr string, header map[string]string) (string, error) {
	json := []byte(jsonStr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range header {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if (nil != resp) && (nil != resp.Body) {
		defer resp.Body.Close()
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", readErr
	}
	return string(body), err
}

func getEddsaCompressedPublicKey(seed string) string {
	sk, err := generateEddsaPrivateKey(seed)
	if err != nil {
		return err.Error()
	}
	var buf bytes.Buffer
	buf.Write(sk.PublicKey.Bytes())
	return hex.EncodeToString(buf.Bytes())
}

// generateEddsaPrivateKey: generate eddsa private key
func generateEddsaPrivateKey(seed string) (sk *eddsa.PrivateKey, err error) {
	buf := make([]byte, 32)
	copy(buf, seed)
	reader := bytes.NewReader(buf)
	sk, err = generateKey(reader)
	return sk, err
}

const (
	sizeFr = fr.Bytes
)

func generateKey(r io.Reader) (*eddsa.PrivateKey, error) {
	c := twistededwards.GetEdwardsCurve()

	var (
		randSrc = make([]byte, 32)
		scalar  = make([]byte, 32)
		pub     eddsa.PublicKey
	)

	// hash(h) = private_key || random_source, on 32 bytes each
	seed := make([]byte, 32)
	_, err := r.Read(seed)
	if err != nil {
		return nil, err
	}
	h := blake2b.Sum512(seed[:])
	for i := 0; i < 32; i++ {
		randSrc[i] = h[i+32]
	}

	// prune the key
	// https://tools.ietf.org/html/rfc8032#section-5.1.5, key generation

	h[0] &= 0xF8
	h[31] &= 0x7F
	h[31] |= 0x40

	// 0xFC = 1111 1100
	// convert 256 bits to 254 bits supporting bn254 curve

	h[31] &= 0xFC

	// reverse first bytes because setBytes interpret stream as big endian
	// but in eddsa specs s is the first 32 bytes in little endian
	for i, j := 0, sizeFr-1; i < sizeFr; i, j = i+1, j-1 {
		scalar[i] = h[j]
	}

	a := new(big.Int).SetBytes(scalar[:])
	for i := 253; i < 256; i++ {
		a.SetBit(a, i, 0)
	}

	copy(scalar[:], a.FillBytes(make([]byte, 32)))

	var bscalar big.Int
	bscalar.SetBytes(scalar[:])
	pub.A.ScalarMul(&c.Base, &bscalar)

	var res [sizeFr * 3]byte
	pubkBin := pub.A.Bytes()
	subtle.ConstantTimeCopy(1, res[:sizeFr], pubkBin[:])
	subtle.ConstantTimeCopy(1, res[sizeFr:2*sizeFr], scalar[:])
	subtle.ConstantTimeCopy(1, res[2*sizeFr:], randSrc[:])

	sk := &eddsa.PrivateKey{}
	// make sure sk is not nil

	_, err = sk.SetBytes(res[:])

	return sk, err
}
