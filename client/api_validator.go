package client

import (
	"context"
	"encoding/hex"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// IValidatorClient - Client APIs for operating Greenfield validators and delegations.
type IValidatorClient interface {
	ListValidators(ctx context.Context, status string) (*stakingtypes.QueryValidatorsResponse, error)
	CreateValidator(ctx context.Context, description stakingtypes.Description, commission stakingtypes.CommissionRates,
		selfDelegation math.Int, validatorAddress string, ed25519PubKey string, selfDelAddr string, relayerAddr string, challengerAddr string, blsKey, blsProof string,
		proposalDepositAmount math.Int, proposalTitle, proposalSummary, proposalMetadata string, txOption gnfdsdktypes.TxOption) (uint64, string, error)
	EditValidator(ctx context.Context, description stakingtypes.Description, newRate *sdktypes.Dec,
		newMinSelfDelegation *math.Int, newRelayerAddr, newChallengerAddr, newBlsKey, blsProof string, txOption gnfdsdktypes.TxOption) (string, error)
	DelegateValidator(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	BeginRedelegate(ctx context.Context, validatorSrcAddr, validatorDestAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	Undelegate(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	CancelUnbondingDelegation(ctx context.Context, validatorAddr string, creationHeight int64, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	GrantDelegationForValidator(ctx context.Context, delegationAmount math.Int, txOption gnfdsdktypes.TxOption) (string, error)

	UnJailValidator(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error)
	ImpeachValidator(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error)
}

// ListValidators lists all validators (if status is empty string) or validators filtered by status.
//
// - ctx: Context variables for the current API call.
//
// - status: The status for filtering validators. It can be "BOND_STATUS_UNBONDED", "BOND_STATUS_UNBONDING" or "BOND_STATUS_BONDED".
//
// - ret1: The information of validators.
//
// - ret2: Return error when getting validators failed, otherwise return nil.
func (c *Client) ListValidators(ctx context.Context, status string) (*stakingtypes.QueryValidatorsResponse, error) {
	return c.chainClient.StakingQueryClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{Status: status})
}

// CreateValidator submits a proposal to the Greenfield for creating a validator, and it returns a proposal ID and tx hash.
//
// - ctx: Context variables for the current API call.
//
// - description: The description of the validator, including name and other information.
//
// - commission: The initial commission rates to be used for creating a validator.
//
// - selfDelegation: The amount of self delegation.
//
// - validatorAddress: The address of the validator.
//
// - ed25519PubKey: The ED25519 pubkey of the validator.
//
// - selfDelAddr: The self delegation address.
//
// - relayerAddr: The address for running off-chain relayers.
//
// - challengerAddr: The address for running off-chain challenge service.
//
// - blsKey: The BLS pubkey of the validator.
//
// - blsProof: The proof of possession of the corresponding BLS private key.
//
// - proposalDepositAmount: The amount to deposit to the proposal.
//
// - proposalTitle: The title of the proposal.
//
// - proposalSummary: The summary of the proposal.
//
// - proposalMetadata: The metadata of the proposal.
//
// - txOption: The options for sending the tx.
//
// - ret1: The id of the submitted proposal.
//
// - ret2: Transaction hash return from blockchain.
//
// - ret3: Return error when create validator tx failed, otherwise return nil.
func (c *Client) CreateValidator(ctx context.Context, description stakingtypes.Description, commission stakingtypes.CommissionRates,
	selfDelegation math.Int, validatorAddress string, ed25519PubKey string, selfDelAddr string, relayerAddr string, challengerAddr string, blsKey, blsProof string,
	proposalDepositAmount math.Int, proposalTitle, proposalSummary, proposalMetadata string, txOption gnfdsdktypes.TxOption,
) (uint64, string, error) {
	govModule, err := c.GetModuleAccountByName(ctx, govTypes.ModuleName)
	if err != nil {
		return 0, "", err
	}
	govAccountAddr := govModule.GetAddress()
	delegationCoin := sdktypes.NewCoin(gnfdsdktypes.Denom, selfDelegation)
	validator, err := sdktypes.AccAddressFromHexUnsafe(validatorAddress)
	if err != nil {
		return 0, "", err
	}
	selfDel, err := sdktypes.AccAddressFromHexUnsafe(selfDelAddr)
	if err != nil {
		return 0, "", err
	}
	relayer, err := sdktypes.AccAddressFromHexUnsafe(relayerAddr)
	if err != nil {
		return 0, "", err
	}
	challenger, err := sdktypes.AccAddressFromHexUnsafe(challengerAddr)
	if err != nil {
		return 0, "", err
	}
	pk, err := pubKeyFromHex(ed25519PubKey)
	if err != nil {
		return 0, "", err
	}
	msg, err := stakingtypes.NewMsgCreateValidator(validator, pk, delegationCoin, description, commission, selfDelegation, govAccountAddr, selfDel, relayer, challenger, blsKey, blsProof)
	if err != nil {
		return 0, "", err
	}
	if err = msg.ValidateBasic(); err != nil {
		return 0, "", err
	}
	return c.SubmitProposal(ctx, []sdktypes.Msg{msg}, proposalDepositAmount, proposalTitle, proposalSummary, types.SubmitProposalOptions{Metadata: proposalMetadata, TxOption: txOption})
}

// EditValidator edits a existing validator info.
//
// - ctx: Context variables for the current API call.
//
// - newDescription: The new description of the validator, including name and other information.
//
// - newCommissionRate: The new commission rate of the validator.
//
// - newMinSelfDelegation: The value for minimal self delegation amount
//
// - newRelayerAddr: The new address for running off-chain relayers.
//
// - newChallengerAddr: The new address for running off-chain challenge service.
//
// - newBlsKey: The new BLS pubkey of the validator.
//
// - newBlsProof: The new proof of possession of the corresponding BLS private key.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when edit validator tx failed, otherwise return nil.
func (c *Client) EditValidator(ctx context.Context, newDescription stakingtypes.Description,
	newCommissionRate *sdktypes.Dec, newMinSelfDelegation *math.Int, newRelayerAddr, newChallengerAddr, newBlsKey, newBlsProof string, txOption gnfdsdktypes.TxOption,
) (string, error) {
	relayer, err := sdktypes.AccAddressFromHexUnsafe(newRelayerAddr)
	if err != nil {
		return "", err
	}
	challenger, err := sdktypes.AccAddressFromHexUnsafe(newChallengerAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgEditValidator(c.MustGetDefaultAccount().GetAddress(), newDescription, newCommissionRate, newMinSelfDelegation, relayer, challenger, newBlsKey, newBlsProof)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// DelegateValidator makes a delegation to a validator by delegator.
//
// - ctx: Context variables for the current API call.
//
// - validatorAddr: The address of the target validator to delegate to.
//
// - amount: The amount of delegation.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when delegation tx failed, otherwise return nil.
func (c *Client) DelegateValidator(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdktypes.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgDelegate(c.MustGetDefaultAccount().GetAddress(), validator, sdktypes.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// BeginRedelegate delegates coins from a delegator and source validator to a destination validator.
//
// - ctx: Context variables for the current API call.
//
// - validatorSrcAddr: The address of the source validator to un-delegate from.
//
// - validatorDestAddr: The address of the destination validator to delegate to.
//
// - amount: The amount of re-delegation.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when re-delegation tx failed, otherwise return nil.
func (c *Client) BeginRedelegate(ctx context.Context, validatorSrcAddr, validatorDestAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
	validatorSrc, err := sdktypes.AccAddressFromHexUnsafe(validatorSrcAddr)
	if err != nil {
		return "", err
	}
	validatorDest, err := sdktypes.AccAddressFromHexUnsafe(validatorDestAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgBeginRedelegate(c.MustGetDefaultAccount().GetAddress(), validatorSrc, validatorDest, sdktypes.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// Undelegate undelegates tokens from a validator by the delegator.
//
// - ctx: Context variables for the current API call.
//
// - validatorAddr: The address of the target validator to un-delegate from.
//
// - amount: The amount of un-delegation.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when un-delegation tx failed, otherwise return nil.
func (c *Client) Undelegate(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdktypes.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgUndelegate(c.MustGetDefaultAccount().GetAddress(), validator, sdktypes.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// CancelUnbondingDelegation cancel the unbonding delegation by delegator.
//
// - ctx: Context variables for the current API call.
//
// - validatorAddr: The address of the validator to cancel the unbonding delegation.
//
// - creationHeight: The height at which the unbonding took place.
//
// - amount: The amount of un-delegation.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when cancel unbonding delegation tx failed, otherwise return nil.
func (c *Client) CancelUnbondingDelegation(ctx context.Context, validatorAddr string, creationHeight int64, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdktypes.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgCancelUnbondingDelegation(c.MustGetDefaultAccount().GetAddress(), validator, creationHeight, sdktypes.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// GrantDelegationForValidator grant the gov module for proposal execution.
//
// - ctx: Context variables for the current API call.
//
// - delegationAmount: The amount to grant.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when grant delegation tx failed, otherwise return nil.
func (c *Client) GrantDelegationForValidator(ctx context.Context, delegationAmount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
	govModule, err := c.GetModuleAccountByName(ctx, govTypes.ModuleName)
	if err != nil {
		return "", err
	}
	delegationCoin := sdktypes.NewCoin(gnfdsdktypes.Denom, delegationAmount)
	authorization, err := stakingtypes.NewStakeAuthorization([]sdktypes.AccAddress{c.MustGetDefaultAccount().GetAddress()},
		nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
		&delegationCoin)
	if err != nil {
		return "", err
	}

	msgGrant, err := authz.NewMsgGrant(c.MustGetDefaultAccount().GetAddress(),
		govModule.GetAddress(),
		authorization, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msgGrant}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// UnJailValidator unjails a validator. The default account's address will be treated the validator address to unjail.
//
// - ctx: Context variables for the current API call.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when unjail validator tx failed, otherwise return nil.
func (c *Client) UnJailValidator(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error) {
	msg := slashingtypes.NewMsgUnjail(c.MustGetDefaultAccount().GetAddress())
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// ImpeachValidator impeaches a validator.
//
// - ctx: Context variables for the current API call.
//
// - validatorAddr: The address of the validator to impeach.
//
// - txOption: The options for sending the tx.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when unjail validator tx failed, otherwise return nil.
func (c *Client) ImpeachValidator(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdktypes.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := slashingtypes.NewMsgImpeach(validator, c.MustGetDefaultAccount().GetAddress())
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func pubKeyFromHex(pk string) (cryptotypes.PubKey, error) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		return nil, err
	}
	if len(pkBytes) != ed25519.PubKeySize {
		return nil, errors.ErrInvalidPubKey
	}
	return &ed25519.PubKey{Key: pkBytes}, nil
}
