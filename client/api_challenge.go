package client

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	types "github.com/bnb-chain/greenfield-go-sdk/types"
)

type Challenge interface {
	ChallengeSP(ctx context.Context, info types.ChallengeInfo, authInfo types.AuthInfo) (types.ChallengeResult, error)
}

// ChallengeSP sends request to challenge and get challenge result info
func (c *client) ChallengeSP(ctx context.Context, info types.ChallengeInfo, authInfo types.AuthInfo) (types.ChallengeResult, error) {
	if info.ObjectId == "" {
		return types.ChallengeResult{}, errors.New("fail to get objectId")
	}

	if info.PieceIndex < 0 {
		return types.ChallengeResult{}, errors.New("index error, should be 0 to parityShards plus dataShards")
	}

	if info.RedundancyIndex < -1 {
		return types.ChallengeResult{}, errors.New("redundancy index error ")
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
		return types.ChallengeResult{}, err
	}

	endpoint, err := c.getSPUrlByBucket(objectInfo.BucketName)
	if err != nil {
		return types.ChallengeResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo, endpoint)
	if err != nil {
		return types.ChallengeResult{}, err
	}

	// fetch integrity hash
	integrityHash := resp.Header.Get(types.HTTPHeaderIntegrityHash)
	// fetch piece hashes
	pieceHashes := resp.Header.Get(types.HTTPHeaderPieceHash)

	if integrityHash == "" || pieceHashes == "" {
		utils.CloseResponse(resp)
		return types.ChallengeResult{}, errors.New("fail to fetch hash info")
	}

	hashList := strings.Split(pieceHashes, ",")
	// min hash num equals one segment hash plus EC dataShards, parityShards
	if len(hashList) < 1 {
		return types.ChallengeResult{}, errors.New("get piece hashes less than 1")
	}

	result := types.ChallengeResult{
		PieceData:     resp.Body,
		IntegrityHash: integrityHash,
		PiecesHash:    hashList,
	}

	return result, nil
}
