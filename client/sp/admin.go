package sp

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
	storage_type "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
)

const ChallengeUrl = "challenge"

// ApproveBucketMeta indicates the core meta to construct createBucket msg of storage module
type ApproveBucketMeta struct {
	BucketName       string
	PrimarySPAddress sdk.AccAddress
}

type ApproveBucketOptions struct {
	IsPublic       bool
	PaymentAddress sdk.AccAddress
}

// ApproveObjectMeta indicates the meta to construct createObject msgof storage module
type ApproveObjectMeta struct {
	BucketName  string
	ObjectName  string
	ContentType string
}

type ApproveObjectOptions struct {
	IsPublic        bool
	SecondarySPAccs []sdk.AccAddress
}

// GetCreateBucketApproval return the signature info for the approval of preCreating resources
func (c *SPClient) GetCreateBucketApproval(ctx context.Context, createBucketMsg *storage_type.MsgCreateBucket, authInfo AuthInfo) (*storage_type.MsgCreateBucket, error) {
	unsignedBytes := createBucketMsg.GetSignBytes()

	// set the action type
	urlVal := make(url.Values)
	urlVal["action"] = []string{CreateBucketAction}

	reqMeta := requestMeta{
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: EmptyStringSHA256,
		TxnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("get approval rejected: " + err.Error())
		return nil, err
	}

	// fetch primary signed msg from sp response
	signedRawMsg := resp.Header.Get(HTTPHeaderSignedMsg)
	if signedRawMsg == "" {
		return nil, errors.New("fail to fetch pre createObject signature")
	}

	signedMsgBytes, err := hex.DecodeString(signedRawMsg)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("signedMsgBytes: " + signedRawMsg)

	var signedMsg storage_type.MsgCreateBucket
	storage_type.ModuleCdc.MustUnmarshalJSON(signedMsgBytes, &signedMsg)

	return &signedMsg, nil
}

// GetCreateObjectApproval return the signature info for the approval of preCreating resources
func (c *SPClient) GetCreateObjectApproval(ctx context.Context, createObjectMsg *storage_type.MsgCreateObject, authInfo AuthInfo) (*storage_type.MsgCreateObject, error) {
	unsignedBytes := createObjectMsg.GetSignBytes()

	// set the action type
	urlVal := make(url.Values)
	urlVal["action"] = []string{CreateObjectAction}

	reqMeta := requestMeta{
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: EmptyStringSHA256,
		TxnMsg:        hex.EncodeToString(unsignedBytes),
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("get approval rejected: " + err.Error())
		return nil, err
	}

	// fetch primary signed msg from sp response
	signedRawMsg := resp.Header.Get(HTTPHeaderSignedMsg)
	if signedRawMsg == "" {
		return nil, errors.New("fail to fetch pre createObject signature")
	}

	signedMsgBytes, err := hex.DecodeString(signedRawMsg)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("signedMsgBytes: " + signedRawMsg)

	var signedMsg storage_type.MsgCreateObject
	storage_type.ModuleCdc.MustUnmarshalJSON(signedMsgBytes, &signedMsg)

	return &signedMsg, nil
}

// ChallengeInfo indicates the challenge object info
// RedundancyIndex if it is primary sp, the value should be -1，
// else it indicates the index of secondary sp
type ChallengeInfo struct {
	ObjectId        string
	PieceIndex      int
	RedundancyIndex int
}

// ChallengeResult indicates the challenge hash and data results
type ChallengeResult struct {
	PieceData     io.ReadCloser
	IntegrityHash string
	PiecesHash    []string
}

// ChallengeSP send request to challenge and get challenge result info
func (c *SPClient) ChallengeSP(ctx context.Context, info ChallengeInfo, authInfo AuthInfo) (ChallengeResult, error) {
	if info.ObjectId == "" {
		return ChallengeResult{}, errors.New("fail to get objectId")
	}

	if info.PieceIndex < 0 {
		return ChallengeResult{}, errors.New("index error, should be 0 to parityShards plus dataShards")
	}

	if info.RedundancyIndex < -1 {
		return ChallengeResult{}, errors.New("redundancy index error ")
	}

	reqMeta := requestMeta{
		urlRelPath:    ChallengeUrl,
		contentSHA256: EmptyStringSHA256,
		challengeInfo: info,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		isAdminApi:       true,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("get challenge result fail: " + err.Error())
		return ChallengeResult{}, err
	}

	// fetch integrity hash
	integrityHash := resp.Header.Get(HTTPHeaderIntegrityHash)
	// fetch piece hashes
	pieceHashes := resp.Header.Get(HTTPHeaderPieceHash)

	if integrityHash == "" || pieceHashes == "" {
		utils.CloseResponse(resp)
		return ChallengeResult{}, errors.New("fail to fetch hash info")
	}

	hashList := strings.Split(pieceHashes, ",")
	// min hash num equals one segment hash plus EC dataShards, parityShards
	if len(hashList) < 1 {
		return ChallengeResult{}, errors.New("get piece hashes less than 1")
	}

	result := ChallengeResult{
		PieceData:     resp.Body,
		IntegrityHash: integrityHash,
		PiecesHash:    hashList,
	}

	return result, nil
}
