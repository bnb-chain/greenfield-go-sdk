package client

import (
	"bytes"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/blake2b"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type OffChainAuth interface {
	RegisterEDDSAPublicKey(spEndpoint string, appDomain string, eddsaSeed string) (*storageTypes.MsgCreateObject, error)
	OffChainAuthSign() string
}

func (c *client) OffChainAuthSign() string {
	sk, _ := GenerateEddsaPrivateKey(c.offChainAuthOption.Seed)
	unSignedMsg := fmt.Sprintf("InvokeSPAPI_%v", time.Now().Add(time.Minute*4).UnixMilli())
	hFunc := mimc.NewMiMC()
	sig, _ := sk.Sign([]byte(unSignedMsg), hFunc)
	authString := fmt.Sprintf("OffChainAuth EDDSA,SignedMsg=%v,Signature=%v", unSignedMsg, hex.EncodeToString(sig))
	return authString
}

// AuthNonce is the structure for off chain auth nonce response
type AuthNonce struct {
	CurrentNonce     int    `json:"current_nonce"`
	CurrentPublicKey string `json:"current_public_key"`
	ExpiryDate       int64  `json:"expiry_date"`
	NextNonce        int    `json:"next_nonce"`
}

// getNonce
func (c *client) GetNextNonce(spEndpoint string) (string, error) {
	header := make(map[string]string)
	header["X-Gnfd-User-Address"] = c.defaultAccount.GetAddress().String()
	header["X-Gnfd-App-Domain"] = c.offChainAuthOption.Domain

	response, err := HttpGetWithHeader(spEndpoint+"/auth/request_nonce", header)
	if err != nil {
		return "0", err
	}
	authNonce := AuthNonce{}
	err = json.Unmarshal([]byte(response), &authNonce)
	return strconv.Itoa(authNonce.NextNonce), nil
}

const (
	UnsignedContentTemplate string = `%s wants you to sign in with your BNB Greenfield account:
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

func (c *client) RegisterEDDSAPublicKey(spAddress string, spEndpoint string) (string, error) {
	appDomain := c.offChainAuthOption.Domain
	eddsaSeed := c.offChainAuthOption.Seed
	nextNonce, err := c.GetNextNonce(spEndpoint)
	if err != nil {
		return "", err
	}
	// get the EDDSA private and public key
	userEddsaPublicKeyStr := GetEddsaCompressedPublicKey(eddsaSeed)
	log.Println("userEddsaPublicKeyStr is %s", userEddsaPublicKeyStr)

	IssueDate := time.Now().Format(time.RFC3339)
	// ExpiryDate := "2023-06-27T06:35:24Z"
	ExpiryDate := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	log.Println("ExpiryDate:", ExpiryDate)
	unSignedContent := fmt.Sprintf(UnsignedContentTemplate, appDomain, c.defaultAccount.GetAddress().String(), userEddsaPublicKeyStr, appDomain, IssueDate, ExpiryDate, spAddress, nextNonce)

	unSignedContentHash := accounts.TextHash([]byte(unSignedContent))
	sig, _ := c.defaultAccount.GetKeyManager().Sign(unSignedContentHash)
	authString := fmt.Sprintf("PersonalSign ECDSA-secp256k1,SignedMsg=%s,Signature=%s", unSignedContent, hexutil.Encode(sig))
	authString = strings.ReplaceAll(authString, "\n", "\\n")
	headers := make(map[string]string)
	headers["x-gnfd-app-domain"] = appDomain
	headers["x-gnfd-app-reg-nonce"] = nextNonce
	headers["x-gnfd-app-reg-public-key"] = userEddsaPublicKeyStr
	headers["x-gnfd-app-reg-expiry-date"] = ExpiryDate
	headers["authorization"] = authString
	headers["origin"] = appDomain
	headers["x-gnfd-user-address"] = c.defaultAccount.GetAddress().String()
	jsonResult, error1 := HttpPostWithHeader(spEndpoint+"/auth/update_key", "{}", headers)
	log.Println("error1:", error1)
	log.Println("jsonResult:", jsonResult)
	return jsonResult, error1
}

func HttpGetWithHeader(url string, header map[string]string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("get error")
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("get error")
		return "", err
	}
	if (nil != resp) && (nil != resp.Body) {
		defer resp.Body.Close()
	}
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		log.Println("ioutil read error")
	}
	return string(body), err
}

func HttpPostWithHeader(url string, jsonStr string, header map[string]string) (string, error) {
	var json = []byte(jsonStr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range header {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("post error")
		return "", err
	}
	if (nil != resp) && (nil != resp.Body) {
		defer resp.Body.Close()
	}
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		log.Println("ioutil read error")
	}
	return string(body), err
}

func GetEddsaCompressedPublicKey(seed string) string {

	sk, err := GenerateEddsaPrivateKey(seed)
	if err != nil {
		return err.Error()
	}
	var buf bytes.Buffer
	buf.Write(sk.PublicKey.Bytes())
	return hex.EncodeToString(buf.Bytes())
}

type (
	PrivateKey = eddsa.PrivateKey
)

// GenerateEddsaPrivateKey: generate eddsa private key
func GenerateEddsaPrivateKey(seed string) (sk *PrivateKey, err error) {
	buf := make([]byte, 32)
	copy(buf, seed)
	reader := bytes.NewReader(buf)
	sk, err = GenerateKey(reader)
	return sk, err
}

const (
	sizeFr = fr.Bytes
)

type PublicKey = eddsa.PublicKey

func GenerateKey(r io.Reader) (*PrivateKey, error) {

	c := twistededwards.GetEdwardsCurve()

	var (
		randSrc = make([]byte, 32)
		scalar  = make([]byte, 32)
		pub     PublicKey
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

	var sk = &PrivateKey{}
	// make sure sk is not nil

	_, err = sk.SetBytes(res[:])

	return sk, err
}
