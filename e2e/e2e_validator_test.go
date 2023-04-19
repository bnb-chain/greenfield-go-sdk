package e2e

import (
	"context"
	"cosmossdk.io/math"
	"encoding/hex"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypesV1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Validator_Operations(t *testing.T) {

	newValAccount, _, _ := types.NewAccount("new_validator")
	newValEd25519PubKey := hex.EncodeToString(ed25519.GenPrivKey().PubKey().Bytes())
	newValidatorAddr := newValAccount.GetAddress()
	t.Logf("new valiadator address is %s", newValidatorAddr.String())

	ctx := context.Background()
	cli, err := client.New(ChainID, Endpoint, client.Option{DefaultAccount: newValAccount})
	assert.NoError(t, err)

	// transfer fund from validator0 -> new Validator so that the new validator can create proposal
	mnemonic := ParseValidatorMnemonic(0)
	validator0Account, err := types.NewAccountFromMnemonic("test", mnemonic)
	assert.NoError(t, err)
	cli.SetDefaultAccount(validator0Account)
	txHash, err := cli.Transfer(ctx, newValidatorAddr.String(), math.NewIntWithDecimal(1000, gnfdsdktypes.DecimalBNB), gnfdsdktypes.TxOption{})
	assert.NoError(t, err)

	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	// newVal grant gov module account for proposal execution
	cli.SetDefaultAccount(newValAccount)
	delegationAmount := math.NewIntWithDecimal(1, 18)

	txHash, err = cli.GrantDelegationForValidator(ctx, delegationAmount, nil)
	assert.NoError(t, err)
	t.Logf("grant auth txHash %s", txHash)

	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	description := stakingtypes.Description{Moniker: "test_new_validator"}
	rates := stakingtypes.CommissionRates{
		Rate:          stakingtypes.DefaultMinCommissionRate,
		MaxRate:       sdk.OneDec(),
		MaxChangeRate: sdk.OneDec(),
	}

	proposalID, txHash, err := cli.CreateValidator(ctx,
		description,
		rates,
		delegationAmount,
		newValidatorAddr.String(),
		newValEd25519PubKey,
		newValAccount.GetAddress().String(),
		"0xA4A2957E858529FFABBBb483D1D704378a9fca6b",
		"0x4038993E087832D84e2Ac855d27f6b0b2EEc1907",
		"a5e140ee80a0ff1552a954701f599622adf029916f55b3157a649e16086a0669900f784d03bff79e69eb8eb7ccfd77d8",
		math.NewIntWithDecimal(1, 18),
		"",
		gnfdsdktypes.TxOption{},
	)
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	cli.SetDefaultAccount(validator0Account)
	_, err = cli.VoteProposal(ctx, proposalID, govTypesV1.OptionYes, client.VoteProposalOptions{})
	assert.NoError(t, err)

	for {
		p, err := cli.GetProposal(ctx, proposalID)
		assert.NoError(t, err)
		t.Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		} else if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_FAILED {
			assert.True(t, false)
		}
		time.Sleep(1 * time.Second)
	}
	err = cli.WaitForNBlocks(ctx, 1)
	assert.NoError(t, err)

	// query the new validator is present
	validators, err := cli.ListValidators(context.Background(), "BOND_STATUS_BONDED")
	assert.NoError(t, err)
	isPresent := false
	for _, v := range validators.Validators {
		if v.SelfDelAddress == newValidatorAddr.String() {
			isPresent = true
		}
	}
	assert.True(t, isPresent)

	// unbond
	cli.SetDefaultAccount(newValAccount)
	txHash, err = cli.Undelegate(ctx, newValidatorAddr.String(), delegationAmount, nil)
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	// query the new validator which is now unbonded
	validators, err = cli.ListValidators(context.Background(), "BOND_STATUS_UNBONDED")
	assert.NoError(t, err)
	isPresent = true
	for _, v := range validators.Validators {
		if v.SelfDelAddress == newValidatorAddr.String() {
			isPresent = false
		}
	}

	// delegate validator
	txHash, err = cli.DelegateValidator(ctx, newValidatorAddr.String(), delegationAmount, nil)
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	// unjain
	txHash, err = cli.UnJailValidator(ctx, nil)
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	// query the new validator is present
	validators, err = cli.ListValidators(context.Background(), "BOND_STATUS_BONDED")
	assert.NoError(t, err)
	isPresent = false
	for _, v := range validators.Validators {
		if v.SelfDelAddress == newValidatorAddr.String() {
			isPresent = true
		}
	}
	assert.True(t, isPresent)
}
