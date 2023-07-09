package e2e

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/suite"

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
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")

	ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOptions{})
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(info.ObjectName, objectName)
		objectBytes, err := io.ReadAll(ior)
		s.Require().NoError(err)
		s.Require().Equal(objectBytes, buffer.Bytes())
	}

	// selete a storage provider to miragte
	sps, err := s.Client.ListStorageProviders(s.ClientContext, true)
	s.Require().NoError(err)

	spIDs := make(map[uint32]bool)
	spIDs[objectDetail.GlobalVirtualGroup.PrimarySpId] = true
	for _, id := range objectDetail.GlobalVirtualGroup.SecondarySpIds {
		spIDs[id] = true
	}
	s.Require().Equal(len(spIDs), 7)

	var destSP *types3.StorageProvider
	for _, sp := range sps {
		_, exist := spIDs[sp.Id]
		if !exist {
			destSP = &sp
			break
		}
	}
	s.Require().NotNil(destSP)

	// send migrate bucket transaction
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
