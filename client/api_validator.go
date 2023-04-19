package client

import (
	"context"
	"cosmossdk.io/math"
	"encoding/hex"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Staking interface {
	// ListValidators lists all validators (if status is empty string) or validators filtered by status.
	// status:
	//  "BOND_STATUS_UNSPECIFIED",
	//  "BOND_STATUS_UNBONDED",
	//  "BOND_STATUS_UNBONDING",
	//	"BOND_STATUS_BONDED",
	ListValidators(ctx context.Context, status string) (*stakingtypes.QueryValidatorsResponse, error)
	// CreateValidator submits a proposal to create a validator to the greenfield blockchain, and it returns a proposal ID and tx hash.
	CreateValidator(ctx context.Context, description stakingtypes.Description, commission stakingtypes.CommissionRates,
		selfDelegation math.Int, validatorAddress string, pubKey string, selfDelAddr string, relayerAddr string, challengerAddr string, blsKey string,
		proposalDepositAmount math.Int, proposalMetadata string, txOption gnfdsdktypes.TxOption) (uint64, string, error)
	EditValidator(ctx context.Context, description stakingtypes.Description, newRate *sdk.Dec,
		newMinSelfDelegation *math.Int, newRelayerAddr, newChallengerAddr, newBlsKey string, txOption *gnfdsdktypes.TxOption) (string, error)
	DelegateValidator(ctx context.Context, validatorAddr string, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error)
	BeginRedelegate(ctx context.Context, validatorSrcAddr, validatorDestAddr string, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error)
	Undelegate(ctx context.Context, validatorAddr string, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error)
	CancelUnbondingDelegation(ctx context.Context, validatorAddr string, creationHeight int64, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error)
	// GrantDelegationForValidator grant the gov module for proposal execution
	GrantDelegationForValidator(ctx context.Context, delegationAmount math.Int, txOption *gnfdsdktypes.TxOption) (string, error)

	UnJailValidator(ctx context.Context, txOption *gnfdsdktypes.TxOption) (string, error)
	ImpeachValidator(ctx context.Context, validatorAddr string, txOption *gnfdsdktypes.TxOption) (string, error)
}

func (c *client) ListValidators(ctx context.Context, status string) (*stakingtypes.QueryValidatorsResponse, error) {
	return c.chainClient.StakingQueryClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{Status: status})
}

func (c *client) CreateValidator(ctx context.Context, description stakingtypes.Description, commission stakingtypes.CommissionRates,
	selfDelegation math.Int, validatorAddress string, pubKey string, selfDelAddr string, relayerAddr string, challengerAddr string, blsKey string,
	proposalDepositAmount math.Int, proposalMetadata string, txOption gnfdsdktypes.TxOption) (uint64, string, error) {

	govModule, err := c.GetModuleAccountByName(ctx, govTypes.ModuleName)
	if err != nil {
		return 0, "", err
	}
	govAccountAddr := govModule.GetAddress()
	delegationCoin := types.NewCoin(gnfdsdktypes.Denom, selfDelegation)
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddress)
	if err != nil {
		return 0, "", err
	}
	selfDel, err := sdk.AccAddressFromHexUnsafe(selfDelAddr)
	if err != nil {
		return 0, "", err
	}
	relayer, err := sdk.AccAddressFromHexUnsafe(relayerAddr)
	if err != nil {
		return 0, "", err
	}
	challenger, err := sdk.AccAddressFromHexUnsafe(challengerAddr)
	if err != nil {
		return 0, "", err
	}
	pk, err := pubKeyFromHex(pubKey)
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
	return c.SubmitProposal(ctx, []sdk.Msg{msg}, proposalDepositAmount, SubmitProposalOptions{Metadata: proposalMetadata, TxOption: txOption})
}

func (c *client) EditValidator(ctx context.Context, description stakingtypes.Description,
	newRate *sdk.Dec, newMinSelfDelegation *math.Int, newRelayerAddr, newChallengerAddr, newBlsKey string, txOption *gnfdsdktypes.TxOption) (string, error) {
	relayer, err := sdk.AccAddressFromHexUnsafe(newRelayerAddr)
	if err != nil {
		return "", err
	}
	challenger, err := sdk.AccAddressFromHexUnsafe(newChallengerAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgEditValidator(c.MustGetDefaultAccount().GetAddress(), description, newRate, newMinSelfDelegation, relayer, challenger, newBlsKey)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) DelegateValidator(ctx context.Context, validatorAddr string, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgDelegate(c.MustGetDefaultAccount().GetAddress(), validator, types.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) BeginRedelegate(ctx context.Context, validatorSrcAddr, validatorDestAddr string, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error) {
	validatorSrc, err := sdk.AccAddressFromHexUnsafe(validatorSrcAddr)
	if err != nil {
		return "", err
	}
	validatorDest, err := sdk.AccAddressFromHexUnsafe(validatorDestAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgBeginRedelegate(c.MustGetDefaultAccount().GetAddress(), validatorSrc, validatorDest, types.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) Undelegate(ctx context.Context, validatorAddr string, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgUndelegate(c.MustGetDefaultAccount().GetAddress(), validator, types.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) CancelUnbondingDelegation(ctx context.Context, validatorAddr string, creationHeight int64, amount math.Int, txOption *gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := stakingtypes.NewMsgCancelUnbondingDelegation(c.MustGetDefaultAccount().GetAddress(), validator, creationHeight, types.NewCoin(gnfdsdktypes.Denom, amount))
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) GrantDelegationForValidator(ctx context.Context, delegationAmount math.Int, txOption *gnfdsdktypes.TxOption) (string, error) {
	govModule, err := c.GetModuleAccountByName(ctx, govTypes.ModuleName)
	if err != nil {
		return "", err
	}
	delegationCoin := types.NewCoin(gnfdsdktypes.Denom, delegationAmount)
	authorization, err := stakingtypes.NewStakeAuthorization([]sdk.AccAddress{c.MustGetDefaultAccount().GetAddress()},
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

	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgGrant}, txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) UnJailValidator(ctx context.Context, txOption *gnfdsdktypes.TxOption) (string, error) {
	msg := slashingtypes.NewMsgUnjail(c.MustGetDefaultAccount().GetAddress())
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *client) ImpeachValidator(ctx context.Context, validatorAddr string, txOption *gnfdsdktypes.TxOption) (string, error) {
	validator, err := sdk.AccAddressFromHexUnsafe(validatorAddr)
	if err != nil {
		return "", err
	}
	msg := slashingtypes.NewMsgImpeach(validator, c.MustGetDefaultAccount().GetAddress())
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msg}, txOption)
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
