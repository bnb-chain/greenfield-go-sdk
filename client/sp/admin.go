package sp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/rs/zerolog/log"
)

const ChallengeUrl = "challenge"

// GetApproval return the signature info for the approval of preCreating resources
func (c *SPClient) GetApproval(ctx context.Context, bucketName, objectName string, authInfo AuthInfo) (string, error) {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return "", err
	}

	if objectName != "" {
		if err := utils.IsValidObjectName(objectName); err != nil {
			return "", err
		}
	}

	// set the action type
	urlVal := make(url.Values)
	if objectName != "" {
		urlVal["action"] = []string{CreateObjectAction}
	} else {
		urlVal["action"] = []string{CreateBucketAction}
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Error().Msg("get approval rejected: " + err.Error())
		return "", err
	}

	// fetch primary sp signature from sp response
	signature := resp.Header.Get(HTTPHeaderPreSignature)
	if signature == "" {
		return "", errors.New("fail to fetch pre createObject signature")
	}

	return signature, nil
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
