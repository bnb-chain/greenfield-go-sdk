package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cosmossdk.io/math"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	challengetypes "github.com/bnb-chain/greenfield/x/challenge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	types "github.com/bnb-chain/greenfield-go-sdk/types"
)

type Challenge interface {
	GetChallengeInfo(ctx context.Context, info types.ChallengeInfo) (types.ChallengeResult, error)
	SubmitChallenge(ctx context.Context, challengerAddress, spOperatorAddress, bucketName, objectName string, randomIndex bool, segmentIndex uint32, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error)
	AttestChallenge(ctx context.Context, submitterAddress, challengerAddress, spOperatorAddress string, challengeId uint64, objectId math.Uint, voteResult challengetypes.VoteResult, voteValidatorSet []uint64, VoteAggSignature []byte, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error)
	LatestAttestedChallenges(ctx context.Context, req *challengetypes.QueryLatestAttestedChallengesRequest) ([]uint64, error)
	InturnAttestationSubmitter(ctx context.Context, req *challengetypes.QueryInturnAttestationSubmitterRequest) (*challengetypes.QueryInturnAttestationSubmitterResponse, error)
	ChallengeParams(ctx context.Context, req *challengetypes.QueryParamsRequest) (*challengetypes.QueryParamsResponse, error)
}

// GetChallengeInfo  sends request to challenge and get challenge result info
// The challenge info includes the piece data, piece hash roots and integrity hash corresponding to the accessed SP
func (c *client) GetChallengeInfo(ctx context.Context, info types.ChallengeInfo) (types.ChallengeResult, error) {
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

	bucketName := objectInfo.BucketName
	endpoint, err := c.getSPUrlByBucket(bucketName)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %s", bucketName, err.Error()))
		return types.ChallengeResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
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

// SubmitChallenge challenges the service provider data integrity, used by off-chain service greenfield-challenger.
func (c *client) SubmitChallenge(ctx context.Context, challengerAddress, spOperatorAddress, bucketName, objectName string, randomIndex bool, segmentIndex uint32, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error) {
	challenger, err := sdk.AccAddressFromHexUnsafe(challengerAddress)
	if err != nil {
		return nil, err
	}
	spOperator, err := sdk.AccAddressFromHexUnsafe(spOperatorAddress)
	if err != nil {
		return nil, err
	}
	msg := challengetypes.NewMsgSubmit(challenger, spOperator, bucketName, objectName, randomIndex, segmentIndex)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return nil, err
	}
	return resp.TxResponse, nil
}

// Attest handles user's request for attesting a challenge.
// The attestation can include a valid challenge or is only for heartbeat purpose.
// If the challenge is valid, the related storage provider will be slashed.
// For heartbeat attestation, the challenge is invalid and the storage provider will not be slashed.
func (c *client) AttestChallenge(ctx context.Context, submitterAddress, challengerAddress, spOperatorAddress string, challengeId uint64, objectId math.Uint,
	voteResult challengetypes.VoteResult, voteValidatorSet []uint64, VoteAggSignature []byte, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error) {

	submitter, err := sdk.AccAddressFromHexUnsafe(submitterAddress)
	if err != nil {
		return nil, err
	}
	if challengerAddress != "" {
		_, err = sdk.AccAddressFromHexUnsafe(challengerAddress)
		if err != nil {
			return nil, err
		}
	}
	_, err = sdk.AccAddressFromHexUnsafe(spOperatorAddress)
	if err != nil {
		return nil, err
	}

	msg := challengetypes.NewMsgAttest(submitter, challengeId, objectId, spOperatorAddress, voteResult, challengerAddress, voteValidatorSet, VoteAggSignature)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return nil, err
	}
	return resp.TxResponse, nil
}

func (c *client) LatestAttestedChallenges(ctx context.Context, req *challengetypes.QueryLatestAttestedChallengesRequest) ([]uint64, error) {
	resp, err := c.chainClient.LatestAttestedChallenges(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.ChallengeIds, nil
}

func (c *client) InturnAttestationSubmitter(ctx context.Context, req *challengetypes.QueryInturnAttestationSubmitterRequest) (*challengetypes.QueryInturnAttestationSubmitterResponse, error) {
	return c.chainClient.InturnAttestationSubmitter(ctx, req)
}

// ChallengeParams returns the on chain parameters for challenge module.
func (c *client) ChallengeParams(ctx context.Context, req *challengetypes.QueryParamsRequest) (*challengetypes.QueryParamsResponse, error) {
	return c.chainClient.ChallengeQueryClient.Params(ctx, req)
}
