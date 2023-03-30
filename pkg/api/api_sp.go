package api

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/types"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storage_type "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (c Client) CreateStorageProvider() {}
func (c Client) EditStorageProvider()   {}
func (c Client) Deposit()               {}
func (c Client) GrantDeposit()          {}
func (c Client) QueryStorageProviders() {}
func (c Client) QueryStorageProvider()  {}
func (c Client) QueryParams()           {}

// GetCreateBucketApproval returns the signature info for the approval of preCreating resources
func (c *Client) GetCreateBucketApproval(ctx context.Context, createBucketMsg *storage_type.MsgCreateBucket,
	authInfo AuthInfo) (*storage_type.MsgCreateBucket, error) {
	unsignedBytes := createBucketMsg.GetSignBytes()

	// set the action type
	urlVal := make(url.Values)
	urlVal["action"] = []string{types.CreateBucketAction}

	reqMeta := requestMeta{
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: types.EmptyStringSHA256,
		TxnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	endpoint, err := c.getSPUrlFromBucket(createBucketMsg.BucketName)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return nil, err
	}

	// fetch primary signed msg from sp response
	signedRawMsg := resp.Header.Get(types.HTTPHeaderSignedMsg)
	if signedRawMsg == "" {
		return nil, errors.New("fail to fetch pre createObject signature")
	}

	signedMsgBytes, err := hex.DecodeString(signedRawMsg)
	if err != nil {
		return nil, err
	}

	var signedMsg storage_type.MsgCreateBucket
	storage_type.ModuleCdc.MustUnmarshalJSON(signedMsgBytes, &signedMsg)

	return &signedMsg, nil
}

// GetCreateObjectApproval returns the signature info for the approval of preCreating resources
func (c *Client) GetCreateObjectApproval(ctx context.Context, createObjectMsg *storage_type.MsgCreateObject,
	authInfo AuthInfo) (*storage_type.MsgCreateObject, error) {
	unsignedBytes := createObjectMsg.GetSignBytes()

	// set the action type
	urlVal := make(url.Values)
	urlVal["action"] = []string{types.CreateObjectAction}

	reqMeta := requestMeta{
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: types.EmptyStringSHA256,
		TxnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	endpoint, err := c.getSPUrlFromBucket(createObjectMsg.BucketName)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return nil, err
	}

	// fetch primary signed msg from sp response
	signedRawMsg := resp.Header.Get(types.HTTPHeaderSignedMsg)
	if signedRawMsg == "" {
		return nil, errors.New("fail to fetch pre createObject signature")
	}

	signedMsgBytes, err := hex.DecodeString(signedRawMsg)
	if err != nil {
		return nil, err
	}

	var signedMsg storage_type.MsgCreateObject
	storage_type.ModuleCdc.MustUnmarshalJSON(signedMsgBytes, &signedMsg)

	return &signedMsg, nil
}

// ChallengeSP sends request to challenge and get challenge result info
func (c *Client) ChallengeSP(ctx context.Context, info client.ChallengeInfo, authInfo AuthInfo) (client.ChallengeResult, error) {
	if info.ObjectId == "" {
		return client.ChallengeResult{}, errors.New("fail to get objectId")
	}

	if info.PieceIndex < 0 {
		return client.ChallengeResult{}, errors.New("index error, should be 0 to parityShards plus dataShards")
	}

	if info.RedundancyIndex < -1 {
		return client.ChallengeResult{}, errors.New("redundancy index error ")
	}

	reqMeta := requestMeta{
		urlRelPath:    types.ChallengeUrl,
		contentSHA256: types.EmptyStringSHA256,
		challengeInfo: info,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		isAdminApi:       true,
		disableCloseBody: true,
	}

	objectInfo, err := c.HeadObjectByID(ctx, info.ObjectId)
	if err != nil {
		return client.ChallengeResult{}, err
	}

	endpoint, err := c.getSPUrlFromBucket(objectInfo.BucketName)
	if err != nil {
		return client.ChallengeResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return client.ChallengeResult{}, err
	}

	// fetch integrity hash
	integrityHash := resp.Header.Get(types.HTTPHeaderIntegrityHash)
	// fetch piece hashes
	pieceHashes := resp.Header.Get(types.HTTPHeaderPieceHash)

	if integrityHash == "" || pieceHashes == "" {
		utils.CloseResponse(resp)
		return client.ChallengeResult{}, errors.New("fail to fetch hash info")
	}

	hashList := strings.Split(pieceHashes, ",")
	// min hash num equals one segment hash plus EC dataShards, parityShards
	if len(hashList) < 1 {
		return client.ChallengeResult{}, errors.New("get piece hashes less than 1")
	}

	result := client.ChallengeResult{
		PieceData:     resp.Body,
		IntegrityHash: integrityHash,
		PiecesHash:    hashList,
	}

	return result, nil
}

func (c *Client) GetSPAddrInfo() (map[string]*url.URL, error) {
	ctx := context.Background()
	spInfo := make(map[string]*url.URL, 0)
	request := &spTypes.QueryStorageProvidersRequest{}
	gnfdRep, err := c.chainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, err
	}
	spList := gnfdRep.GetSps()
	for _, info := range spList {
		endpoint := info.Endpoint
		urlInfo, urlErr := utils.GetEndpointURL(endpoint, c.Secure)
		if urlErr != nil {
			return nil, urlErr
		}
		spInfo[info.GetOperator().String()] = urlInfo
	}

	return spInfo, nil
}

// ListSP return the storage provider info on chain
// isInService indicates if only display the sp with STATUS_IN_SERVICE status
func (c *Client) ListSP(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error) {
	request := &spTypes.QueryStorageProvidersRequest{}
	gnfdRep, err := c.chainClient.StorageProviders(ctx, request)
	if err != nil {
		return nil, err
	}

	spList := gnfdRep.GetSps()
	spInfoList := make([]spTypes.StorageProvider, 0)
	for _, info := range spList {
		if isInService && info.Status != spTypes.STATUS_IN_SERVICE {
			continue
		}
		spInfoList = append(spInfoList, *info)
	}

	return spInfoList, nil
}

// GetSpAddrFromEndpoint return the chain addr according to the SP endpoint
func (c *Client) GetSpAddrFromEndpoint(ctx context.Context, spEndpoint string) (sdk.AccAddress, error) {
	spList, err := c.ListSP(ctx, false)
	if err != nil {
		return nil, err
	}

	if strings.Contains(spEndpoint, "http") {
		s := strings.Split(spEndpoint, "//")
		spEndpoint = s[1]
	}

	for _, spInfo := range spList {
		endpoint := spInfo.GetEndpoint()
		if strings.Contains(endpoint, "http") {
			s := strings.Split(endpoint, "//")
			endpoint = s[1]
		}
		if endpoint == spEndpoint {
			addr := spInfo.GetOperatorAddress()
			if addr == "" {
				return nil, errors.New("fail to get addr")
			}
			return sdk.MustAccAddressFromHex(spInfo.GetOperatorAddress()), nil
		}
	}
	return nil, errors.New("fail to get addr")
}
