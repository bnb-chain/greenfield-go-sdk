package e2e

import (
	"context"
	"cosmossdk.io/math"
	"encoding/hex"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypesV1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type ValidatorTestSuite struct {
	basesuite.BaseSuite
}

func (s *ValidatorTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *ValidatorTestSuite) Test_Validator_Operations() {

	newValAccount, _, _ := types.NewAccount("new_validator")
	newValEd25519PubKey := hex.EncodeToString(ed25519.GenPrivKey().PubKey().Bytes())
	newValidatorAddr := newValAccount.GetAddress()
	s.T().Logf("new valiadator address is %s", newValidatorAddr.String())

	// transfer some funds to the new validator
	validator0Account := s.DefaultAccount

	txHash, err := s.Client.Transfer(s.ClientContext, newValidatorAddr.String(), math.NewIntWithDecimal(1000, gnfdsdktypes.DecimalBNB), gnfdsdktypes.TxOption{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	// newVal grant gov module account for proposal execution
	s.Client.SetDefaultAccount(newValAccount)
	delegationAmount := math.NewIntWithDecimal(1, 18)

	txHash, err = s.Client.GrantDelegationForValidator(s.ClientContext, delegationAmount, nil)
	s.Require().NoError(err)
	s.T().Logf("grant auth txHash %s", txHash)

	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	description := stakingtypes.Description{Moniker: "test_new_validator"}
	rates := stakingtypes.CommissionRates{
		Rate:          stakingtypes.DefaultMinCommissionRate,
		MaxRate:       sdk.OneDec(),
		MaxChangeRate: sdk.OneDec(),
	}

	proposalID, txHash, err := s.Client.CreateValidator(s.ClientContext,
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
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	s.Client.SetDefaultAccount(validator0Account)
	_, err = s.Client.VoteProposal(s.ClientContext, proposalID, govTypesV1.OptionYes, client.VoteProposalOptions{})
	s.Require().NoError(err)

	for {
		p, err := s.Client.GetProposal(s.ClientContext, proposalID)
		s.Require().NoError(err)
		s.T().Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		} else if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_FAILED {
			s.Require().True(false)
		}
		time.Sleep(1 * time.Second)
	}
	err = s.Client.WaitForNBlocks(s.ClientContext, 1)
	s.Require().NoError(err)

	// query the new validator is present
	validators, err := s.Client.ListValidators(context.Background(), "BOND_STATUS_BONDED")
	s.Require().NoError(err)
	isPresent := false
	for _, v := range validators.Validators {
		if v.SelfDelAddress == newValidatorAddr.String() {
			isPresent = true
		}
	}
	s.Require().True(isPresent)

	// unbond
	s.Client.SetDefaultAccount(newValAccount)
	txHash, err = s.Client.Undelegate(s.ClientContext, newValidatorAddr.String(), delegationAmount, nil)
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	err = s.Client.WaitForNBlocks(s.ClientContext, 3)
	s.Require().NoError(err)

	// query the new validator which is now UNBONDING
	validators, err = s.Client.ListValidators(context.Background(), "BOND_STATUS_UNBONDING")
	s.Require().NoError(err)
	isPresent = true
	for _, v := range validators.Validators {
		s.T().Log(v)
		if v.SelfDelAddress == newValidatorAddr.String() {
			isPresent = false
		}
	}
	s.Require().False(isPresent)

	// delegate validator
	txHash, err = s.Client.DelegateValidator(s.ClientContext, newValidatorAddr.String(), delegationAmount, nil)
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	// unjain
	txHash, err = s.Client.UnJailValidator(s.ClientContext, nil)
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	// query the new validator is present
	validators, err = s.Client.ListValidators(context.Background(), "BOND_STATUS_BONDED")
	s.Require().NoError(err)
	isPresent = false
	for _, v := range validators.Validators {
		if v.SelfDelAddress == newValidatorAddr.String() {
			isPresent = true
		}
	}
	s.Require().True(isPresent)
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}
