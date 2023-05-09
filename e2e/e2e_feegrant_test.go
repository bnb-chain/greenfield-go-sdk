package e2e

import (
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type FeeGrantTestSuite struct {
	basesuite.BaseSuite
}

func (s *FeeGrantTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *FeeGrantTestSuite) Test_FeeGrant() {
	cli := s.Client
	t := s.T()
	ctx := s.ClientContext

	granter, _, err := types.NewAccount("granter")
	s.Require().NoError(err)
	granterAddr := granter.GetAddress().String()
	grantee, _, err := types.NewAccount("grantee")
	s.Require().NoError(err)
	granteeAddr := grantee.GetAddress().String()

	// charge granter and grantee accounts
	chargeAmount := math.NewIntWithDecimal(1, 19)
	transferDetails := make([]types.TransferDetail, 0)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: granterAddr,
		Amount:    chargeAmount,
	})
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: granteeAddr,
		Amount:    chargeAmount,
	})

	txHash, err := cli.MultiTransfer(ctx, transferDetails, types2.TxOption{})
	s.Require().NoError(err)
	_, err = cli.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	// granter grants allowance to grantee
	cli.SetDefaultAccount(granter)
	allowanceAmount := math.NewIntWithDecimal(1, 18)

	txHash, err = cli.GrantBasicAllowance(ctx, granteeAddr, allowanceAmount, nil, types2.TxOption{})
	s.Require().NoError(err)
	_, err = cli.WaitForTx(ctx, txHash)
	s.Require().NoError(err)

	// Query the allowance
	allowance, err := cli.QueryBasicAllowance(ctx, granterAddr, granteeAddr)
	s.Require().NoError(err)
	t.Log(allowance.String())

	// show grantee balance before the grantee making a tx
	granteeBalanceBefore, err := cli.GetAccountBalance(ctx, granteeAddr)
	s.Require().NoError(err)

	// grantee makes a tx and costs the fee provided by granter
	cli.SetDefaultAccount(grantee)
	txHash, err = cli.CreatePaymentAccount(ctx, granteeAddr, types2.TxOption{
		FeeGranter: granter.GetAddress(),
	})

	s.Require().NoError(err)
	_, _ = cli.WaitForTx(ctx, txHash)

	granteeBalanceAfter, err := cli.GetAccountBalance(ctx, granteeAddr)
	s.Require().NoError(err)

	// grantee balance stays still
	s.Require().Equal(granteeBalanceBefore, granteeBalanceAfter)

	// the granter revokes
	cli.SetDefaultAccount(granter)
	txHash, err = cli.RevokeAllowance(ctx, granteeAddr, types2.TxOption{})
	s.Require().NoError(err)
	_, _ = cli.WaitForTx(ctx, txHash)

	// Query the allowance
	_, err = cli.QueryBasicAllowance(ctx, granterAddr, granteeAddr)
	s.Require().Error(err)

	// transaction is failed
	cli.SetDefaultAccount(grantee)
	_, err = cli.CreatePaymentAccount(ctx, granteeAddr, types2.TxOption{
		FeeGranter: granter.GetAddress(),
	})
	s.Require().Error(err)
}

func TestFeeGrantTestSuite(t *testing.T) {
	suite.Run(t, new(FeeGrantTestSuite))
}
