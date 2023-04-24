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

type Validator interface {
	// ListValidators lists all validators (if status is empty string) or validators filtered by status.
	// status:
	//  "BOND_STATUS_UNBONDED",
	//  "BOND_STATUS_UNBONDING",
	//	"BOND_STATUS_BONDED",
	ListValidators(ctx context.Context, status string) (*stakingtypes.QueryValidatorsResponse, error)

	CreateValidator(ctx context.Context, description stakingtypes.Description, commission stakingtypes.CommissionRates,
		selfDelegation math.Int, validatorAddress string, ed25519PubKey string, selfDelAddr string, relayerAddr string, challengerAddr string, blsKey string,
		proposalDepositAmount math.Int, title, summary, proposalMetadata string, txOption gnfdsdktypes.TxOption) (uint64, string, error)
	EditValidator(ctx context.Context, description stakingtypes.Description, newRate *sdktypes.Dec,
		newMinSelfDelegation *math.Int, newRelayerAddr, newChallengerAddr, newBlsKey string, txOption gnfdsdktypes.TxOption) (string, error)
	DelegateValidator(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	BeginRedelegate(ctx context.Context, validatorSrcAddr, validatorDestAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	Undelegate(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	CancelUnbondingDelegation(ctx context.Context, validatorAddr string, creationHeight int64, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error)
	GrantDelegationForValidator(ctx context.Context, delegationAmount math.Int, txOption gnfdsdktypes.TxOption) (string, error)

	UnJailValidator(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error)
	ImpeachValidator(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error)
}

func (c *client) ListValidators(ctx context.Context, status string) (*stakingtypes.QueryValidatorsResponse, error) {
	return c.chainClient.StakingQueryClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{Status: status})
}

// CreateValidator submits a proposal to the Greenfield for creating a validator, and it returns a proposal ID and tx hash.
func (c *client) CreateValidator(ctx context.Context, description stakingtypes.Description, commission stakingtypes.CommissionRates,
	selfDelegation math.Int, validatorAddress string, ed25519PubKey string, selfDelAddr string, relayerAddr string, challengerAddr string, blsKey string,
	proposalDepositAmount math.Int, title, summary, proposalMetadata string, txOption gnfdsdktypes.TxOption) (uint64, string, error) {

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
	msg, err := stakingtypes.NewMsgCreateValidator(validator, pk, delegationCoin, description, commission, selfDelegation, govAccountAddr, selfDel, relayer, challenger, blsKey)
	if err != nil {
		return 0, "", err
	}
	if err = msg.ValidateBasic(); err != nil {
		return 0, "", err
	}
	return c.SubmitProposal(ctx, []sdktypes.Msg{msg}, proposalDepositAmount, title, summary, types.SubmitProposalOptions{Metadata: proposalMetadata, TxOption: txOption})
}

// EditValidator edits a existing validator info.
func (c *client) EditValidator(ctx context.Context, description stakingtypes.Description,
	newRate *sdktypes.Dec, newMinSelfDelegation *math.Int, newRelayerAddr, newChallengerAddr, newBlsKey string, txOption gnfdsdktypes.TxOption) (string, error) {
	relayer, err := sdktypes.AccAddressFromHexUnsafe(newRelayerAddr)
	if err != nil {
		return "", err
	}
	challenger, err := sdktypes.AccAddressFromHexUnsafe(newChallengerAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgEditValidator(c.MustGetDefaultAccount().GetAddress(), description, newRate, newMinSelfDelegation, relayer, challenger, newBlsKey)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// DelegateValidator makes a delegation to a validator by delegator.
func (c *client) DelegateValidator(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
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

// BeginRedelegate delegates coins from a delegator and source validator to a destination validator
func (c *client) BeginRedelegate(ctx context.Context, validatorSrcAddr, validatorDestAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
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
func (c *client) Undelegate(ctx context.Context, validatorAddr string, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
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

// CancelUnbondingDelegation cancel the unbonding delegation by delegator
func (c *client) CancelUnbondingDelegation(ctx context.Context, validatorAddr string, creationHeight int64, amount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
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

// GrantDelegationForValidator grant the gov module for proposal execution
func (c *client) GrantDelegationForValidator(ctx context.Context, delegationAmount math.Int, txOption gnfdsdktypes.TxOption) (string, error) {
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

// UnJailValidator unjails the validator
func (c *client) UnJailValidator(ctx context.Context, txOption gnfdsdktypes.TxOption) (string, error) {
	msg := slashingtypes.NewMsgUnjail(c.MustGetDefaultAccount().GetAddress())
	resp, err := c.chainClient.BroadcastTx(ctx, []sdktypes.Msg{msg}, &txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

// ImpeachValidator impeaches a validator
func (c *client) ImpeachValidator(ctx context.Context, validatorAddr string, txOption gnfdsdktypes.TxOption) (string, error) {
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
