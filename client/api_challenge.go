package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
	// GetChallengeInfo return the challenge hash and data results based on the objectID and index info
	// If the sp endpoint or sp address info is not set in the GetChallengeInfoOptions, the SP endpoint will be routed by the redundancyIndex
	GetChallengeInfo(ctx context.Context, objectID string, pieceIndex, redundancyIndex int, opts types.GetChallengeInfoOptions) (types.ChallengeResult, error)
	SubmitChallenge(ctx context.Context, challengerAddress, spOperatorAddress, bucketName, objectName string, randomIndex bool, segmentIndex uint32, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error)
	AttestChallenge(ctx context.Context, submitterAddress, challengerAddress, spOperatorAddress string, challengeId uint64, objectId math.Uint, voteResult challengetypes.VoteResult, voteValidatorSet []uint64, VoteAggSignature []byte, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error)
	LatestAttestedChallenges(ctx context.Context, req *challengetypes.QueryLatestAttestedChallengesRequest) ([]uint64, error)
	InturnAttestationSubmitter(ctx context.Context, req *challengetypes.QueryInturnAttestationSubmitterRequest) (*challengetypes.QueryInturnAttestationSubmitterResponse, error)
	ChallengeParams(ctx context.Context, req *challengetypes.QueryParamsRequest) (*challengetypes.QueryParamsResponse, error)
}

// GetChallengeInfo  sends request to challenge and get challenge result info
// The challenge info includes the piece data, piece hash roots and integrity hash corresponding to the accessed SP
func (c *client) GetChallengeInfo(ctx context.Context, objectID string, pieceIndex, redundancyIndex int, opts types.GetChallengeInfoOptions) (types.ChallengeResult, error) {
	if objectID == "" {
		return types.ChallengeResult{}, errors.New("fail to get objectId")
	}

	if pieceIndex < 0 {
		return types.ChallengeResult{}, errors.New("index error, should be 0 to parityShards plus dataShards")
	}

	if redundancyIndex < types.PrimaryRedundancyIndex || redundancyIndex > types.MaxRedundancyIndex {
		return types.ChallengeResult{}, errors.New("redundancy index invalid, the index should be -1 to 5")
	}

	reqMeta := requestMeta{
		urlRelPath:    types.ChallengeUrl,
		contentSHA256: types.EmptyStringSHA256,
		challengeInfo: types.ChallengeInfo{
			objectID,
			pieceIndex,
			redundancyIndex,
		},
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		isAdminApi:       true,
		disableCloseBody: true,
	}

	var endpoint *url.URL
	var err error
	if opts.Endpoint != "" {
		var useHttps bool
		if strings.Contains(opts.Endpoint, "https") {
			useHttps = true
		} else {
			useHttps = c.secure
		}

		endpoint, err = utils.GetEndpointURL(opts.Endpoint, useHttps)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("fetch endpoint from opts %s fail:%v", opts.Endpoint, err))
			return types.ChallengeResult{}, err
		}
	} else if opts.SPAddress != "" {
		// get endpoint from sp address
		endpoint, err = c.getSPUrlByAddr(opts.SPAddress)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("route endpoint by sp address: %s failed, err: %v", opts.SPAddress, err))
			return types.ChallengeResult{}, err
		}
	} else {
		// get sp address info based on the redundancy index
		objectInfo, err := c.HeadObjectByID(ctx, objectID)
		if err != nil {
			return types.ChallengeResult{}, err
		}

		if redundancyIndex == types.PrimaryRedundancyIndex {
			// get endpoint of primary sp
			endpoint, err = c.getSPUrlByBucket(objectInfo.BucketName)
			if err != nil {
				log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %v", objectInfo.BucketName, err))
				return types.ChallengeResult{}, err
			}
		} else {
			// get endpoint of the secondary sp
			secondarySP := objectInfo.SecondarySpAddresses[redundancyIndex]
			endpoint, err = c.getSPUrlByAddr(secondarySP)
			if err != nil {
				log.Error().Msg(fmt.Sprintf("route endpoint by sp address: %s failed, err: %v", secondarySP, err))
				return types.ChallengeResult{}, err
			}
		}
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
	voteResult challengetypes.VoteResult, voteValidatorSet []uint64, VoteAggSignature []byte, txOption gnfdsdktypes.TxOption,
) (*sdk.TxResponse, error) {
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
