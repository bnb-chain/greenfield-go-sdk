package e2e

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/stretchr/testify/suite"
)

type BasicTestSuite struct {
	basesuite.BaseSuite
}

func (s *BasicTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *BasicTestSuite) Test_Basic() {
	_, _, err := s.Client.GetNodeInfo(s.ClientContext)
	s.Require().NoError(err)

	latestBlock, err := s.Client.GetLatestBlock(s.ClientContext)
	s.Require().NoError(err)
	fmt.Println(latestBlock.String())

	heightBefore := latestBlock.Header.Height
	err = s.Client.WaitForBlockHeight(s.ClientContext, heightBefore+10)
	s.Require().NoError(err)
	height, err := s.Client.GetLatestBlockHeight(s.ClientContext)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(height, heightBefore+10)

	syncing, err := s.Client.GetSyncing(s.ClientContext)
	s.Require().NoError(err)
	s.Require().False(syncing)

	blockByHeight, err := s.Client.GetBlockByHeight(s.ClientContext, heightBefore)
	s.Require().NoError(err)
	s.Require().Equal(blockByHeight.GetHeader(), latestBlock.GetHeader())
}

func (s *BasicTestSuite) Test_Account() {
	balance, err := s.Client.GetAccountBalance(s.ClientContext, s.DefaultAccount.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("Balance: %s", balance.String())

	account1, _, err := types.NewAccount("test2")
	s.Require().NoError(err)
	transferTxHash, err := s.Client.Transfer(s.ClientContext, account1.GetAddress().String(), math.NewIntFromUint64(1), types2.TxOption{})
	s.Require().NoError(err)
	s.T().Logf("Transfer response: %s", transferTxHash)

	waitForTx, err := s.Client.WaitForTx(s.ClientContext, transferTxHash)
	s.Require().NoError(err)
	s.T().Logf("Wair for tx: %s", waitForTx.String())

	balance, err = s.Client.GetAccountBalance(s.ClientContext, account1.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("Balance: %s", balance.String())
	s.Require().True(balance.Amount.Equal(math.NewInt(1)))

	acc, err := s.Client.GetAccount(s.ClientContext, account1.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("Acc: %s", acc.String())
	s.Require().Equal(acc.GetAddress(), account1.GetAddress())
	s.Require().Equal(acc.GetSequence(), uint64(0))

	txHash, err := s.Client.CreatePaymentAccount(s.ClientContext, s.DefaultAccount.GetAddress().String(), &types2.TxOption{})
	s.Require().NoError(err)
	s.T().Logf("Acc: %s", txHash)
	waitForTx, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	s.T().Logf("Wair for tx: %s", waitForTx.String())

	paymentAccountsByOwner, err := s.Client.GetPaymentAccountsByOwner(s.ClientContext, s.DefaultAccount.GetAddress().String())
	s.Require().NoError(err)
	s.Require().Equal(len(paymentAccountsByOwner), 1)
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}

func Test_Payment(t *testing.T) {
	mnemonic := ParseValidatorMnemonic(0)
	account, err := types.NewAccountFromMnemonic("test", mnemonic)
	assert.NoError(t, err)
	cli, err := client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: account,
		GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	assert.NoError(t, err)
	ctx := context.Background()

	txHash, err := cli.CreatePaymentAccount(ctx, account.GetAddress().String(), &types2.TxOption{})
	assert.NoError(t, err)
	t.Logf("Acc: %s", txHash)
	waitForTx, err := cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)
	t.Logf("Wair for tx: %s", waitForTx.String())

	paymentAccountsByOwner, err := cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())
	assert.NoError(t, err)
	assert.Equal(t, len(paymentAccountsByOwner), 1)

	// deposit
	paymentAddr := paymentAccountsByOwner[0].Addr
	depositAmount := math.NewIntFromUint64(100)
	depositTxHash, err := cli.Deposit(ctx, paymentAddr, depositAmount, nil)
	assert.NoError(t, err)
	t.Logf("deposit tx: %s", depositTxHash)
	waitForTx, err = cli.WaitForTx(ctx, depositTxHash)
	assert.NoError(t, err)
	t.Logf("Wair for tx: %s", waitForTx.String())

	// get stream record
	streamRecord, err := cli.GetStreamRecord(ctx, paymentAddr)
	assert.NoError(t, err)
	assert.Equal(t, streamRecord.StaticBalance.String(), depositAmount.String())

	// withdraw
	withdrawAmount := math.NewIntFromUint64(50)
	withdrawTxHash, err := cli.Withdraw(ctx, paymentAddr, withdrawAmount, nil)
	assert.NoError(t, err)
	t.Logf("withdraw tx: %s", withdrawTxHash)
	waitForTx, err = cli.WaitForTx(ctx, withdrawTxHash)
	assert.NoError(t, err)
	t.Logf("Wair for tx: %s", waitForTx.String())
	streamRecordAfterWithdraw, err := cli.GetStreamRecord(ctx, paymentAddr)
	assert.NoError(t, err)
	assert.Equal(t, streamRecordAfterWithdraw.StaticBalance.String(), depositAmount.Sub(withdrawAmount).String())

	// disable refund
	assert.True(t, paymentAccountsByOwner[0].Refundable)
	disableRefundTxHash, err := cli.DisableRefund(ctx, paymentAddr, nil)
	assert.NoError(t, err)
	t.Logf("disable refund tx: %s", disableRefundTxHash)
	waitForTx, err = cli.WaitForTx(ctx, disableRefundTxHash)
	assert.NoError(t, err)
	t.Logf("Wair for tx: %s", waitForTx.String())
	paymentAccountsByOwner, err = cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())
	assert.NoError(t, err)
	assert.False(t, paymentAccountsByOwner[0].Refundable)
}
