package e2e

import (
	"bytes"
	"cosmossdk.io/math"
	"encoding/hex"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	govTypesV1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
	"time"

	types3 "github.com/bnb-chain/greenfield/x/sp/types"
)

type BucketMigrateTestSuite struct {
	basesuite.BaseSuite
	PrimarySP spTypes.StorageProvider

	SPList []spTypes.StorageProvider

	// destSP config
	OperatorAcc *types.Account
	FundingAcc  *types.Account
	SealAcc     *types.Account
	ApprovalAcc *types.Account
	GcAcc       *types.Account
	BlsAcc      *types.Account
}

func (s *BucketMigrateTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	spList, err := s.Client.ListStorageProviders(s.ClientContext, false)
	s.Require().NoError(err)
	for _, sp := range spList {
		if sp.Endpoint != "https://sp0.greenfield.io" {
			s.PrimarySP = sp
		}
	}
	s.SPList = spList
}

func TestBucketMigrateTestSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(BucketMigrateTestSuite))
}

func (s *BucketMigrateTestSuite) CreateStorageProvider() {
	var err error

	s.OperatorAcc, _, err = types.NewAccount("operator")
	s.Require().NoError(err)
	s.FundingAcc, _, err = types.NewAccount("funding")
	s.Require().NoError(err)
	s.SealAcc, _, err = types.NewAccount("seal")
	s.Require().NoError(err)
	s.ApprovalAcc, _, err = types.NewAccount("approval")
	s.Require().NoError(err)
	s.GcAcc, _, err = types.NewAccount("gc")
	s.Require().NoError(err)
	s.BlsAcc, _, err = types.NewBlsAccount("bls")
	s.Require().NoError(err)
	s.T().Logf("FundingAddr: %s, sealAddr: %s, approvalAddr: %s, operatorAddr: %s, gcAddr: %s, blsPubKey: %s",
		s.FundingAcc.GetAddress().String(),
		s.SealAcc.GetAddress().String(),
		s.ApprovalAcc.GetAddress().String(),
		s.OperatorAcc.GetAddress().String(),
		s.GcAcc.GetAddress().String(),
		s.BlsAcc.GetKeyManager().PubKey().String(),
	)

	txHash, err := s.Client.Transfer(s.ClientContext, s.FundingAcc.GetAddress().String(), math.NewIntWithDecimal(10001, types2.DecimalBNB), types2.TxOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	fundingBalance, err := s.Client.GetAccountBalance(s.ClientContext, s.FundingAcc.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("funding validatorAccount balance: %s", fundingBalance.String())

	txHash, err = s.Client.Transfer(s.ClientContext, s.SealAcc.GetAddress().String(), math.NewIntWithDecimal(1, types2.DecimalBNB), types2.TxOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	sealBalance, err := s.Client.GetAccountBalance(s.ClientContext, s.SealAcc.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("seal validatorAccount balance: %s", sealBalance.String())

	txHash, err = s.Client.Transfer(s.ClientContext, s.OperatorAcc.GetAddress().String(), math.NewIntWithDecimal(1000, types2.DecimalBNB), types2.TxOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	operatorBalance, err := s.Client.GetAccountBalance(s.ClientContext, s.OperatorAcc.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("operator validatorAccount balance: %s", operatorBalance.String())

	s.Client.SetDefaultAccount(s.FundingAcc)
	txHash, err = s.Client.GrantDepositForStorageProvider(s.ClientContext, s.OperatorAcc.GetAddress().String(), math.NewIntWithDecimal(10000, types2.DecimalBNB), types.GrantDepositForStorageProviderOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	s.Client.SetDefaultAccount(s.OperatorAcc)
	proposalID, txHash, err := s.Client.CreateStorageProvider(s.ClientContext, s.FundingAcc.GetAddress().String(), s.SealAcc.GetAddress().String(), s.ApprovalAcc.GetAddress().String(), s.GcAcc.GetAddress().String(),
		hex.EncodeToString(s.BlsAcc.GetKeyManager().PubKey().Bytes()),
		"https://sp0.greenfield.io",
		math.NewIntWithDecimal(10000, types2.DecimalBNB),
		types3.Description{Moniker: "test"},
		types.CreateStorageProviderOptions{ProposalMetaData: "create", ProposalTitle: "test", ProposalSummary: "test"})
	s.Require().NoError(err)

	createTx, err := s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	s.T().Log(createTx.Logs.String())

	for {
		p, err := s.Client.GetProposal(s.ClientContext, proposalID)
		s.T().Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		s.Require().NoError(err)
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
			break
		}
		time.Sleep(1 * time.Second)
	}

	s.Client.SetDefaultAccount(s.DefaultAccount)
	voteTxHash, err := s.Client.VoteProposal(s.ClientContext, proposalID, govTypesV1.OptionYes, types.VoteProposalOptions{})
	s.Require().NoError(err)

	tx, err := s.Client.WaitForTx(s.ClientContext, voteTxHash)
	s.Require().NoError(err)
	s.T().Logf("VoteTx: %s", tx.TxHash)

	for {
		p, err := s.Client.GetProposal(s.ClientContext, proposalID)
		s.T().Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		s.Require().NoError(err)
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		} else if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_FAILED {
			s.Require().True(false)
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *BucketMigrateTestSuite) Test_Bucket_Migrate_Object() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	// 1) create bucket and object in srcSP
	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, types.CreateBucketOptions{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
	}

	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,123456789012`
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024*30; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	s.T().Log("---> PutObject and GetObject <---")
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	time.Sleep(20 * time.Second)
	objectDetail, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")
	}

	ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOptions{})
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(info.ObjectName, objectName)
		objectBytes, err := io.ReadAll(ior)
		s.Require().NoError(err)
		s.Require().Equal(objectBytes, buffer.Bytes())
	}

	//2) create destSP
	//s.CreateStorageProvider()
	//var destSP *spTypes.StorageProvider
	//destSP, err = s.Client.GetStorageProviderInfo(s.ClientContext, s.OperatorAcc.GetAddress())
	//s.Require().NoError(err)

	//3) migrate bucket
	// TODO : determine destSP bucketInfo.GetGlobalVirtualGroupFamilyId()

	//sp, err := s.Client.GetStorageProviderInfo(s.ClientContext, sdk.AccAddress(s.PrimarySP.OperatorAddress))
	//s.Require().NoError(err)
	////sp.

	destSP := s.SPList[7]
	txhash, err := s.Client.MigrateBucket(s.ClientContext, bucketName, types.MigrateBucketOptions{TxOpts: nil, DstPrimarySPID: destSP.GetId(), IsAsyncMode: false})
	s.Require().NoError(err)

	tx, err := s.Client.WaitForTx(s.ClientContext, txhash)
	s.Require().NoError(err)
	s.T().Logf("VoteTx: %s", tx.TxHash)

	// TODO check migration whether success
	//for {
	//	p, err := s.Client.GetProposal(s.ClientContext, proposalID)
	//	s.T().Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
	//	s.Require().NoError(err)
	//	if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_PASSED {
	//		break
	//	} else if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_FAILED {
	//		s.Require().True(false)
	//	}
	//	time.Sleep(1 * time.Second)
	//}
}
