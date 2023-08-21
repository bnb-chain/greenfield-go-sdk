package e2e

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type BucketMigrateTestSuite struct {
	basesuite.BaseSuite
	PrimarySP spTypes.StorageProvider
}

func (s *BucketMigrateTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	spList, err := s.Client.ListStorageProviders(s.ClientContext, false)
	s.Require().NoError(err)
	for _, sp := range spList {
		if sp.Endpoint != "https://sp0.greenfield.io" {
			s.PrimarySP = sp
			break
		}
	}
}

func TestBucketMigrateTestSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(BucketMigrateTestSuite))
}

func (s *BucketMigrateTestSuite) CreateObjects(bucketName string, count int) ([]*types.ObjectDetail, []bytes.Buffer, error) {
	var (
		objectNames   []string
		contentBuffer []bytes.Buffer
		objectDetails []*types.ObjectDetail
	)

	// create object
	for i := 0; i < count; i++ {
		var buffer bytes.Buffer
		line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,123456789012`
		// Create 1MiB content where each line contains 1024 characters.
		for n := 0; n < 1024*3; n++ {
			buffer.WriteString(fmt.Sprintf("[%05d] %s\n", n, line))
		}
		objectName := storageTestUtil.GenRandomObjectName()
		s.T().Logf("---> CreateObject and HeadObject, bucketname:%s, objectname:%s <---", bucketName, objectName)
		objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{Visibility: storageTypes.VISIBILITY_TYPE_PUBLIC_READ})
		s.Require().NoError(err)
		_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
		s.Require().NoError(err)

		objectNames = append(objectNames, objectName)
		contentBuffer = append(contentBuffer, buffer)
	}

	// head object
	time.Sleep(5 * time.Second)
	for _, objectName := range objectNames {
		objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
		s.Require().NoError(err)
		s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
		s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")
	}

	s.T().Log("---> PutObject and GetObject <---")
	for idx, objectName := range objectNames {
		buffer := contentBuffer[idx]
		err := s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
			bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
		s.Require().NoError(err)
	}

	time.Sleep(20 * time.Second)
	// seal object
	for idx, objectName := range objectNames {
		objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
		s.Require().NoError(err)
		s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")

		ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOptions{})
		s.Require().NoError(err)
		if err == nil {
			s.Require().Equal(info.ObjectName, objectName)
			objectBytes, err := io.ReadAll(ior)
			s.Require().NoError(err)
			s.Require().Equal(objectBytes, contentBuffer[idx].Bytes())
		}
		objectDetails = append(objectDetails, objectDetail)
	}

	return objectDetails, contentBuffer, nil
}

func (s *BucketMigrateTestSuite) MustCreateBucket(visibility storageTypes.VisibilityType) (string, *storageTypes.BucketInfo) {
	bucketName := storageTestUtil.GenRandomBucketName()
	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, types.CreateBucketOptions{Visibility: visibility})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, visibility)
	}

	s.T().Logf("success to create a new bucket: %s", bucketInfo)

	return bucketName, bucketInfo
}

func (s *BucketMigrateTestSuite) SelectDestSP(objectDetail *types.ObjectDetail) *spTypes.StorageProvider {
	sps, err := s.Client.ListStorageProviders(s.ClientContext, true)
	s.Require().NoError(err)

	spIDs := make(map[uint32]bool)
	spIDs[objectDetail.GlobalVirtualGroup.PrimarySpId] = true
	for _, id := range objectDetail.GlobalVirtualGroup.SecondarySpIds {
		spIDs[id] = true
	}
	s.Require().Equal(len(spIDs), 7)

	var destSP *spTypes.StorageProvider
	for _, sp := range sps {
		_, exist := spIDs[sp.Id]
		if !exist {
			destSP = &sp
			break
		}
	}
	s.Require().NotNil(destSP)

	return destSP
}

func (s *BucketMigrateTestSuite) waitUntilBucketMigrateFinish(bucketName string, destSP *spTypes.StorageProvider) *storageTypes.BucketInfo {
	var (
		bucketInfo *storageTypes.BucketInfo
		err        error
	)

	// wait 5 minutes
	for i := 0; i < 100; i++ {
		bucketInfo, err = s.Client.HeadBucket(s.ClientContext, bucketName)
		s.T().Logf("HeadBucket: %s", bucketInfo)
		s.Require().NoError(err)
		if bucketInfo.BucketStatus != storageTypes.BUCKET_STATUS_MIGRATING {
			break
		}
		time.Sleep(3 * time.Second)
	}

	family, err := s.Client.QueryVirtualGroupFamily(s.ClientContext, bucketInfo.GlobalVirtualGroupFamilyId)
	s.Require().NoError(err)
	s.Require().Equal(family.PrimarySpId, destSP.GetId())

	return bucketInfo
}

// test only one object's case
func (s *BucketMigrateTestSuite) Test_Bucket_Migrate_Simple_Case() {

	// 1) create bucket and object in srcSP
	bucketName, _ := s.MustCreateBucket(storageTypes.VISIBILITY_TYPE_PUBLIC_READ)

	// test only one object's case
	objectDetails, contentBuffer, err := s.CreateObjects(bucketName, 1)
	s.Require().NoError(err)

	objectDetail := objectDetails[0]
	buffer := contentBuffer[0]

	// selete a storage provider to miragte
	destSP := s.SelectDestSP(objectDetail)

	s.T().Logf(":Migrate Bucket DstPrimarySPID %d", destSP.GetId())

	// normal no conflict send migrate bucket transaction
	txhash, err := s.Client.MigrateBucket(s.ClientContext, bucketName, types.MigrateBucketOptions{TxOpts: nil, DstPrimarySPID: destSP.GetId(), IsAsyncMode: false})
	s.Require().NoError(err)

	s.T().Logf("MigrateBucket : %s", txhash)
	s.waitUntilBucketMigrateFinish(bucketName, destSP)

	ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectDetail.ObjectInfo.ObjectName, types.GetObjectOptions{})
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(info.ObjectName, objectDetail.ObjectInfo.ObjectName)
		objectBytes, err := io.ReadAll(ior)
		s.Require().NoError(err)
		s.Require().Equal(objectBytes, buffer.Bytes())
	}
	s.CheckChallenge(uint32(objectDetail.ObjectInfo.Id.Uint64()))
}

// test only conflict sp's case
func (s *BucketMigrateTestSuite) Test_Bucket_Migrate_Simple_Conflict_Case() {
	// 1) create bucket and object in srcSP
	bucketName, _ := s.MustCreateBucket(storageTypes.VISIBILITY_TYPE_PUBLIC_READ)

	// test only one object's case
	objectDetails, contentBuffer, err := s.CreateObjects(bucketName, 1)
	s.Require().NoError(err)

	objectDetail := objectDetails[0]
	buffer := contentBuffer[0]

	// selete a storage provider to miragte
	sps, err := s.Client.ListStorageProviders(s.ClientContext, true)
	s.Require().NoError(err)

	spIDs := make(map[uint32]bool)
	spIDs[objectDetail.GlobalVirtualGroup.PrimarySpId] = true
	for _, id := range objectDetail.GlobalVirtualGroup.SecondarySpIds {
		spIDs[id] = true
	}
	s.Require().Equal(len(spIDs), 7)

	var destSP *spTypes.StorageProvider
	for _, sp := range sps {
		_, exist := spIDs[sp.Id]
		if !exist {
			destSP = &sp
			break
		}
	}
	s.Require().NotNil(destSP)

	// migrate bucket with conflict
	conflictSPID := objectDetail.GlobalVirtualGroup.SecondarySpIds[0]
	s.T().Logf(":Migrate Bucket DstPrimarySPID %d", conflictSPID)

	txhash, err := s.Client.MigrateBucket(s.ClientContext, bucketName, types.MigrateBucketOptions{TxOpts: nil, DstPrimarySPID: conflictSPID, IsAsyncMode: false})
	s.Require().NoError(err)

	s.T().Logf("MigrateBucket : %s", txhash)

	var bucketInfo *storageTypes.BucketInfo

	for {
		bucketInfo, err = s.Client.HeadBucket(s.ClientContext, bucketName)
		s.T().Logf("HeadBucket: %s", bucketInfo)
		s.Require().NoError(err)
		if bucketInfo.BucketStatus != storageTypes.BUCKET_STATUS_MIGRATING {
			break
		}
		time.Sleep(3 * time.Second)
	}

	family, err := s.Client.QueryVirtualGroupFamily(s.ClientContext, bucketInfo.GlobalVirtualGroupFamilyId)
	s.Require().NoError(err)
	s.Require().Equal(family.PrimarySpId, conflictSPID)
	ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectDetail.ObjectInfo.ObjectName, types.GetObjectOptions{})
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(info.ObjectName, objectDetail.ObjectInfo.ObjectName)
		objectBytes, err := io.ReadAll(ior)
		s.Require().NoError(err)
		s.Require().Equal(objectBytes, buffer.Bytes())
	}
	s.CheckChallenge(uint32(objectDetail.ObjectInfo.Id.Uint64()))
}

// test empty bucket case
func (s *BucketMigrateTestSuite) Test_Empty_Bucket_Migrate_Simple_Case() {
	// 1) create bucket and object in srcSP
	bucketName, bucketInfo := s.MustCreateBucket(storageTypes.VISIBILITY_TYPE_PUBLIC_READ)

	s.T().Logf("CreateBucket : %s", bucketInfo)
	virtualGroupFamily, err := s.Client.QueryVirtualGroupFamily(s.ClientContext, bucketInfo.GetGlobalVirtualGroupFamilyId())
	s.Require().NoError(err)
	s.T().Logf("virtualGroupFamily : %s", virtualGroupFamily)

	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PUBLIC_READ)
	}

	time.Sleep(5 * time.Second)
	// selete a storage provider to miragte
	sps, err := s.Client.ListStorageProviders(s.ClientContext, true)
	s.Require().NoError(err)

	var destSP *spTypes.StorageProvider
	for _, sp := range sps {
		if sp.GetId() != virtualGroupFamily.GetPrimarySpId() {
			destSP = &sp
			break
		}
	}
	s.Require().NotNil(destSP)

	s.T().Logf(":Migrate Bucket DstPrimarySPID %s", destSP.String())

	// normal no conflict send migrate bucket transaction
	txhash, err := s.Client.MigrateBucket(s.ClientContext, bucketName, types.MigrateBucketOptions{TxOpts: nil, DstPrimarySPID: destSP.GetId(), IsAsyncMode: false})
	s.Require().NoError(err)

	s.T().Logf("MigrateBucket : %s", txhash)

	for {
		bucketInfo, err = s.Client.HeadBucket(s.ClientContext, bucketName)
		s.T().Logf("HeadBucket: %s", bucketInfo)
		s.Require().NoError(err)
		if bucketInfo.BucketStatus != storageTypes.BUCKET_STATUS_MIGRATING {
			break
		}
		time.Sleep(3 * time.Second)
	}

	family, err := s.Client.QueryVirtualGroupFamily(s.ClientContext, bucketInfo.GlobalVirtualGroupFamilyId)
	s.Require().NoError(err)
	s.Require().Equal(family.PrimarySpId, destSP.GetId())
}

func (s *BucketMigrateTestSuite) CheckChallenge(objectId uint32) bool {
	i := objectId
	infos, err := s.Client.HeadObjectByID(context.Background(), fmt.Sprintf("%d", i))
	s.Require().NoError(err)
	if infos.ObjectInfo.ObjectStatus == storageTypes.OBJECT_STATUS_SEALED {
		reader, _, err := s.Client.GetObject(context.Background(), infos.ObjectInfo.BucketName, infos.ObjectInfo.ObjectName, types.GetObjectOptions{})
		s.NoError(err, fmt.Sprintf("%d", i), infos.ObjectInfo.BucketName, infos.ObjectInfo.ObjectName)
		_, err = io.ReadAll(reader)
		s.NoError(err, fmt.Sprintf("%d", i), infos.ObjectInfo.BucketName, infos.ObjectInfo.ObjectName)
		for j := -1; j < 6; j++ {
			s.T().Logf("====challenge %v,%v,=====", i, j)
			_, errPk := s.Client.GetChallengeInfo(context.Background(), infos.ObjectInfo.Id.String(), 0, j, types.GetChallengeInfoOptions{})
			s.NoError(errPk, infos.ObjectInfo.BucketName, infos.ObjectInfo.ObjectName, i, j)
			if errPk != nil {
				s.T().Errorf(infos.ObjectInfo.BucketName, infos.ObjectInfo.ObjectName, i, j)
			}
		}
	}

	return true
}
