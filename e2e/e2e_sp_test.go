package e2e

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	types3 "github.com/bnb-chain/greenfield/x/sp/types"
	govTypesV1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_CreateStorageProvider(t *testing.T) {
	mnemonic := ParseValidatorMnemonic(0)
	validatorAccount, err := types.NewAccountFromMnemonic("test", mnemonic)

	operatorAcc, _, _ := types.NewAccount("operator")
	fundingAcc, _, _ := types.NewAccount("funding")
	sealAcc, _, _ := types.NewAccount("seal")
	approvalAcc, _, _ := types.NewAccount("approval")
	gcAcc, _, _ := types.NewAccount("gc")

	t.Logf("FundingAddr: %s, sealAddr: %s, approvalAddr: %s, operatpr: %s", fundingAcc.GetAddress().String(), sealAcc.GetAddress().String(), approvalAcc.GetAddress().String(), operatorAcc.GetAddress().String())

	assert.NoError(t, err)
	ctx := context.Background()
	cli, err := client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: validatorAccount,
		GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	assert.NoError(t, err)

	txHash, err := cli.Transfer(ctx, fundingAcc.GetAddress().String(), math.NewIntWithDecimal(10001, types2.DecimalBNB), types2.TxOption{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)
	fundingBalance, err := cli.GetAccountBalance(ctx, fundingAcc.GetAddress().String())
	assert.NoError(t, err)
	t.Logf("funding validatorAccount balance: %s", fundingBalance.String())

	txHash, err = cli.Transfer(ctx, sealAcc.GetAddress().String(), math.NewIntWithDecimal(1, types2.DecimalBNB), types2.TxOption{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)
	sealBalance, err := cli.GetAccountBalance(ctx, sealAcc.GetAddress().String())
	assert.NoError(t, err)
	t.Logf("seal validatorAccount balance: %s", sealBalance.String())

	txHash, err = cli.Transfer(ctx, operatorAcc.GetAddress().String(), math.NewIntWithDecimal(1000, types2.DecimalBNB), types2.TxOption{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)
	operatorBalance, err := cli.GetAccountBalance(ctx, operatorAcc.GetAddress().String())
	assert.NoError(t, err)
	t.Logf("operator validatorAccount balance: %s", operatorBalance.String())

	cli.SetDefaultAccount(fundingAcc)
	txHash, err = cli.GrantDepositForStorageProvider(ctx, operatorAcc.GetAddress().String(), math.NewIntWithDecimal(10000, types2.DecimalBNB), client.GrantDepositForStorageProviderOptions{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	cli.SetDefaultAccount(operatorAcc)
	proposalID, txHash, err := cli.CreateStorageProvider(ctx, fundingAcc.GetAddress().String(), sealAcc.GetAddress().String(), approvalAcc.GetAddress().String(), gcAcc.GetAddress().String(),
		"https://sp0.greenfield.io",
		math.NewIntWithDecimal(10000, types2.DecimalBNB),
		types3.Description{Moniker: "test"},
		client.CreateStorageProviderOptions{ProposalMetaData: "create"})
	assert.NoError(t, err)

	createTx, err := cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)
	t.Log(createTx.Logs.String())

	for {
		p, err := cli.GetProposal(ctx, proposalID)
		t.Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		assert.NoError(t, err)
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
			break
		}
		time.Sleep(1 * time.Second)
	}

	cli.SetDefaultAccount(validatorAccount)
	voteTxHash, err := cli.VoteProposal(ctx, proposalID, govTypesV1.OptionYes, client.VoteProposalOptions{})
	assert.NoError(t, err)

	tx, err := cli.WaitForTx(ctx, voteTxHash)
	assert.NoError(t, err)
	t.Logf("VoteTx: %s", tx.TxHash)

	for {
		p, err := cli.GetProposal(ctx, proposalID)
		t.Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		assert.NoError(t, err)
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		} else if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_FAILED {
			assert.True(t, false)
		}
		time.Sleep(1 * time.Second)
	}

}
