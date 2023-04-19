package e2e

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
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

func (s *BasicTestSuite) Test_MultiTransfer() {

	transferDetails := make([]types.TransferDetail, 0)
	totalSendAmount := math.NewInt(0)

	receiver1, _, err := types.NewAccount("receiver1")
	s.Require().NoError(err)
	receiver1Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver1Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver1.GetAddress().String(),
		Amount:    receiver1Amount,
	})

	receiver2, _, err := types.NewAccount("receiver2")
	s.Require().NoError(err)
	receiver2Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver2Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver2.GetAddress().String(),
		Amount:    receiver2Amount,
	})

	receiver3, _, err := types.NewAccount("receiver3")
	s.Require().NoError(err)
	receiver3Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver3Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver3.GetAddress().String(),
		Amount:    receiver3Amount,
	})

	s.T().Logf("totally sending %s", totalSendAmount.String())

	txHash, err := s.Client.MultiTransfer(s.ClientContext, transferDetails, nil)
	s.Require().NoError(err)
	s.T().Log(txHash)

	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	balance1, err := s.Client.GetAccountBalance(s.ClientContext, receiver1.GetAddress().String())
	s.Require().NoError(err)
	s.Assertions.Equal(receiver1Amount, balance1.Amount)

	balance2, err := s.Client.GetAccountBalance(s.ClientContext, receiver2.GetAddress().String())
	s.Require().NoError(err)
	s.Assertions.Equal(receiver2Amount, balance2.Amount)

	balance3, err := s.Client.GetAccountBalance(s.ClientContext, receiver3.GetAddress().String())
	s.Require().NoError(err)
	s.Assertions.Equal(receiver3Amount, balance3.Amount)
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
