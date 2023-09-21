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

// IChallengeClient - Client APIs for operating and querying Greenfield challenges.
type IChallengeClient interface {
	GetChallengeInfo(ctx context.Context, objectID string, pieceIndex, redundancyIndex int, opts types.GetChallengeInfoOptions) (types.ChallengeResult, error)
	SubmitChallenge(ctx context.Context, challengerAddress, spOperatorAddress, bucketName, objectName string, randomIndex bool, segmentIndex uint32, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error)
	AttestChallenge(ctx context.Context, submitterAddress, challengerAddress, spOperatorAddress string, challengeId uint64, objectId math.Uint, voteResult challengetypes.VoteResult, voteValidatorSet []uint64, VoteAggSignature []byte, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error)
	LatestAttestedChallenges(ctx context.Context, req *challengetypes.QueryLatestAttestedChallengesRequest) (*challengetypes.QueryLatestAttestedChallengesResponse, error)
	InturnAttestationSubmitter(ctx context.Context, req *challengetypes.QueryInturnAttestationSubmitterRequest) (*challengetypes.QueryInturnAttestationSubmitterResponse, error)
	ChallengeParams(ctx context.Context, req *challengetypes.QueryParamsRequest) (*challengetypes.QueryParamsResponse, error)
}

// GetChallengeInfo - Send request to storage provider, and get the integrity hash and data stored on the sp.
//
// This api is used for validators to judge whether a challenge valid or not. A validator's challenger account should be
// provided when constructing the client, otherwise the authorization will fail.
//
// - ctx: Context variables for the current API call.
//
// - objectID: The id of object being challenged.
//
// - pieceIndex: The index of the segment/piece of the object.
//
// - redundancyIndex: The redundancy index of the object, actually it also stands for which storage provider is being
// challenged and should response to this call - -1 stands for the primary storage provider.
//
// - opts: Options to define the storage provider address and its endpoint. If the storage provider address or endpoint
// is not set in the options, the storage provider endpoint will be routed by redundancyIndex.
//
// - ret1: The challenge info includes the piece data, piece hash roots and integrity hash corresponding to the accessed storage provider.
//
// - ret2: Return error when getting challenge info failed, otherwise return nil.
func (c *Client) GetChallengeInfo(ctx context.Context, objectID string, pieceIndex, redundancyIndex int, opts types.GetChallengeInfoOptions) (types.ChallengeResult, error) {
	if objectID == "" {
		return types.ChallengeResult{}, errors.New("fail to get objectId")
	}

	if pieceIndex < 0 {
		return types.ChallengeResult{}, errors.New("index error, should be 0 to parityShards plus dataShards")
	}

	var err error
	dataBlocks, parityBlocks, _, err := c.GetRedundancyParams()
	if err != nil {
		return types.ChallengeResult{}, errors.New("fail to get redundancy params:" + err.Error())
	}
	maxRedundancyIndex := int(dataBlocks+parityBlocks) - 1
	if redundancyIndex < types.PrimaryRedundancyIndex || redundancyIndex > maxRedundancyIndex {
		return types.ChallengeResult{}, fmt.Errorf("redundancy index invalid, the index should be %d to %d", types.PrimaryRedundancyIndex, maxRedundancyIndex)
	}

	reqMeta := requestMeta{
		urlRelPath:    types.ChallengeUrl,
		contentSHA256: types.EmptyStringSHA256,
		pieceInfo: types.QueryPieceInfo{
			ObjectId:        objectID,
			PieceIndex:      pieceIndex,
			RedundancyIndex: redundancyIndex,
		},
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		isAdminApi:       true,
		disableCloseBody: true,
	}

	var endpoint *url.URL
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
		objectDetail, err := c.HeadObjectByID(ctx, objectID)
		if err != nil {
			return types.ChallengeResult{}, err
		}

		if redundancyIndex == types.PrimaryRedundancyIndex {
			// get endpoint of primary sp
			endpoint, err = c.getSPUrlByBucket(objectDetail.ObjectInfo.BucketName)
			if err != nil {
				log.Error().Msg(fmt.Sprintf("route endpoint by bucket: %s failed, err: %v", objectDetail.ObjectInfo.BucketName, err))
				return types.ChallengeResult{}, err
			}
		} else {
			// get endpoint of the secondary sp
			secondarySPID := objectDetail.GlobalVirtualGroup.SecondarySpIds[redundancyIndex]
			endpoint, err = c.getSPUrlByID(secondarySPID)
			if err != nil {
				log.Error().Msg(fmt.Sprintf("route endpoint by sp address: %d failed, err: %v", secondarySPID, err))
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

// SubmitChallenge - Challenge a storage provider's data integrity for a specific data object.
//
// User can submit a challenge when he/she find his/her data is lost or tampered. A successful challenge will punish
// the storage provider and reward the challenge submitter.
//
// - ctx: Context variables for the current API call.
//
// - challengerAddress: The address of the challenger, usually it's the address of current account.
//
// - spOperatorAddress: The operator address of the storage provider to be challenged.
//
// - bucketName: The bucket name of the object to be challenged.
//
// - objectName: The name of the object to be challenged.
//
// - randomIndex: Whether the segment/piece to be challenged will be randomly generated or not.
//
// - segmentIndex: The index of the segment/piece to be challenged. If randomIndex is true, then segmentIndex will be ignored.
//
// - txOption: The options for sending the tx.
//
// - ret1: The response of Greenfield transaction.
//
// - ret2: Return error when submitting challenge tx failed, otherwise return nil.
func (c *Client) SubmitChallenge(ctx context.Context, challengerAddress, spOperatorAddress, bucketName, objectName string, randomIndex bool, segmentIndex uint32, txOption gnfdsdktypes.TxOption) (*sdk.TxResponse, error) {
	challenger, err := sdk.AccAddressFromHexUnsafe(challengerAddress)
	if err != nil {
		return nil, err
	}
	spOperator, err := sdk.AccAddressFromHexUnsafe(spOperatorAddress)
	if err != nil {
		return nil, err
	}
	msg := challengetypes.NewMsgSubmit(challenger, spOperator, bucketName, objectName, randomIndex, segmentIndex)
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return nil, err
	}
	return resp.TxResponse, nil
}

// AttestChallenge - Send the attestation result of a challenge.
//
// In-turn validator can submit the attestation when enough votes are collected.
// The attestation can include a valid challenge or is only for heartbeat purpose.
// If the challenge is valid, the related storage provider will be slashed, otherwise the storage provider will not be slashed.
//
// - ctx: Context variables for the current API call.
//
// - submitterAddress: The address of the attestation submitter, usually it's the address of current account.
//
// - challengerAddress: The address of the challenger.
//
// - spOperatorAddress: The operator address of the challenged storage provider.
//
// - challengeId: The id of the challenge to be attested.
//
// - objectId: The id of the object being challenged.
//
// - voteResult: The result of the off-chain votes from the majority validators.
//
// - voteValidatorSet: The bit set of validators who vote the voteResult for the challenge.
//
// - VoteAggSignature: The BLS aggregated signature for the attestation from validators.
//
// - txOption: The options for sending the tx.
//
// - ret1: The response of Greenfield transaction.
//
// - ret2: Return error when submitting attestation tx failed, otherwise return nil.
func (c *Client) AttestChallenge(ctx context.Context, submitterAddress, challengerAddress, spOperatorAddress string, challengeId uint64, objectId math.Uint,
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
	resp, err := c.BroadcastTx(ctx, []sdk.Msg{msg}, &txOption)
	if err != nil {
		return nil, err
	}
	return resp.TxResponse, nil
}

// LatestAttestedChallenges - Query the latest attested challenges (including heartbeat challenges).
//
// Greenfield will not keep the results of all challenges, only the latest ones will be kept and old ones will be pruned.
//
// - ctx: Context variables for the current API call.
//
// - req: The request to query latest attested challenges.
//
// - ret1: The latest attested challenges, including challenge id and attestation result.
//
// - ret2: Return error when getting latest attested challenges failed, otherwise return nil.
func (c *Client) LatestAttestedChallenges(ctx context.Context, req *challengetypes.QueryLatestAttestedChallengesRequest) (*challengetypes.QueryLatestAttestedChallengesResponse, error) {
	return c.chainClient.LatestAttestedChallenges(ctx, req)
}

// InturnAttestationSubmitter - Query the in-turn validator to submit challenge attestation.
//
// Greenfield will only allow the in-turn validator to submit challenge attestations.
//
// - ctx: Context variables for the current API call.
//
// - req: The request to query in-turn attestation submitter.
//
// - ret1: The in-turn validator information, including BLS pubkey of the validator and timeframe to submit attestations.
//
// - ret2: Return error when getting in-turn attestation submitter failed, otherwise return nil.
func (c *Client) InturnAttestationSubmitter(ctx context.Context, req *challengetypes.QueryInturnAttestationSubmitterRequest) (*challengetypes.QueryInturnAttestationSubmitterResponse, error) {
	return c.chainClient.InturnAttestationSubmitter(ctx, req)
}

// ChallengeParams - Get challenge module's parameters of Greenfield blockchain.
//
// - ctx: Context variables for the current API call.
//
// - req: The request to query challenge parameters.
//
// - ret1: The parameters of challenge module.
//
// - ret2: Return error when getting parameters failed, otherwise return nil.
func (c *Client) ChallengeParams(ctx context.Context, req *challengetypes.QueryParamsRequest) (*challengetypes.QueryParamsResponse, error) {
	return c.chainClient.ChallengeQueryClient.Params(ctx, req)
}
