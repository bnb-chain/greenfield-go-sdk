package e2e

import (
	"bytes"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"io"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/suite"
)

type StorageTestSuite struct {
	basesuite.BaseSuite
	PrimarySP spTypes.StorageProvider
}

func (s *StorageTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	spList, err := s.Client.ListStorageProviders(s.ClientContext, false)
	s.Require().NoError(err)
	for _, sp := range spList {
		if sp.Endpoint != "https://sp0.greenfield.io" {
			s.PrimarySP = sp
		}
	}
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (s *StorageTestSuite) Test_Bucket() {
	bucketName := storageTestUtil.GenRandomBucketName()

	chargedQuota := uint64(100)
	s.T().Log("---> CreateBucket and HeadBucket <---")
	opts := types.CreateBucketOptions{ChargedQuota: chargedQuota}
	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, opts)
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
		s.Require().Equal(bucketInfo.ChargedReadQuota, chargedQuota)
	}

	s.T().Log("--->  UpdateBucket <---")
	updateBucketTx, err := s.Client.UpdateBucketVisibility(s.ClientContext, bucketName,
		storageTypes.VISIBILITY_TYPE_PUBLIC_READ, types.UpdateVisibilityOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, updateBucketTx)
	s.Require().NoError(err)

	s.T().Log("---> BuyQuotaForBucket <---")
	targetQuota := uint64(300)
	buyQuotaTx, err := s.Client.BuyQuotaForBucket(s.ClientContext, bucketName, targetQuota, types.BuyQuotaOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, buyQuotaTx)
	s.Require().NoError(err)

	s.T().Log("---> Query Quota info <---")
	quota, err := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
	s.Require().NoError(err)
	s.Require().Equal(quota.ReadQuotaSize, targetQuota)

	s.T().Log("---> PutBucketPolicy <---")
	principal, _, err := types.NewAccount("principal")
	s.Require().NoError(err)

	principalStr, err := utils.NewPrincipalWithAccount(principal.GetAddress())
	s.Require().NoError(err)
	statements := []*permTypes.Statement{
		{
			Effect: permTypes.EFFECT_ALLOW,
			Actions: []permTypes.ActionType{
				permTypes.ACTION_UPDATE_BUCKET_INFO,
				permTypes.ACTION_DELETE_BUCKET,
				permTypes.ACTION_CREATE_OBJECT,
			},
			Resources:      []string{},
			ExpirationTime: nil,
			LimitSize:      nil,
		},
	}
	policy, err := s.Client.PutBucketPolicy(s.ClientContext, bucketName, principalStr, statements, types.PutPolicyOption{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, policy)
	s.Require().NoError(err)

	s.T().Log("---> GetBucketPolicy <---")
	bucketPolicy, err := s.Client.GetBucketPolicy(s.ClientContext, bucketName, principal.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("get bucket policy:%s\n", bucketPolicy.String())

	s.T().Log("---> DeleteBucketPolicy <---")
	deleteBucketPolicy, err := s.Client.DeleteBucketPolicy(s.ClientContext, bucketName, principalStr, types.DeletePolicyOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, deleteBucketPolicy)
	s.Require().NoError(err)
	_, err = s.Client.GetBucketPolicy(s.ClientContext, bucketName, principal.GetAddress().String())
	s.Require().Error(err)

	s.T().Log("--->  DeleteBucket <---")
	delBucket, err := s.Client.DeleteBucket(s.ClientContext, bucketName, types.DeleteBucketOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, delBucket)
	s.Require().NoError(err)

	_, err = s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().Error(err)
}

func (s *StorageTestSuite) Test_Object() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

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
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890`
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024*100; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectInfo, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectInfo.ObjectName, objectName)
	s.Require().Equal(objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	s.T().Log("---> PutObject and GetObject <---")
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	time.Sleep(10 * time.Second)
	objectInfo, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")
	}

	ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOption{})
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(info.ObjectName, objectName)
		objectBytes, err := io.ReadAll(ior)
		s.Require().NoError(err)
		s.Require().Equal(objectBytes, buffer.Bytes())
	}

	s.T().Log("---> PutObjectPolicy <---")
	principal, _, err := types.NewAccount("principal")
	s.Require().NoError(err)
	principalWithAccount, err := utils.NewPrincipalWithAccount(principal.GetAddress())
	s.Require().NoError(err)
	statements := []*permTypes.Statement{
		{
			Effect: permTypes.EFFECT_ALLOW,
			Actions: []permTypes.ActionType{
				permTypes.ACTION_GET_OBJECT,
			},
			Resources:      nil,
			ExpirationTime: nil,
			LimitSize:      nil,
		},
	}
	policy, err := s.Client.PutObjectPolicy(s.ClientContext, bucketName, objectName, principalWithAccount, statements, types.PutPolicyOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, policy)
	s.Require().NoError(err)

	s.T().Log("--->  GetObjectPolicy <---")
	objectPolicy, err := s.Client.GetObjectPolicy(s.ClientContext, bucketName, objectName, principal.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("get object policy:%s\n", objectPolicy.String())

	s.T().Log("---> DeleteObjectPolicy <---")

	principalStr, err := utils.NewPrincipalWithAccount(principal.GetAddress())
	s.Require().NoError(err)
	deleteObjectPolicy, err := s.Client.DeleteObjectPolicy(s.ClientContext, bucketName, objectName, principalStr, types.DeletePolicyOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, deleteObjectPolicy)
	s.Require().NoError(err)

	s.T().Log("---> DeleteObject <---")
	deleteObject, err := s.Client.DeleteObject(s.ClientContext, bucketName, objectName, types.DeleteObjectOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, deleteObject)
	s.Require().NoError(err)
	_, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().Error(err)
}

func (s *StorageTestSuite) Test_Group() {
	groupName := storageTestUtil.GenRandomGroupName()

	groupOwner := s.DefaultAccount.GetAddress()
	s.T().Log("---> CreateGroup and HeadGroup <---")
	_, err := s.Client.CreateGroup(s.ClientContext, groupName, types.CreateGroupOptions{})
	s.Require().NoError(err)
	s.T().Logf("create GroupName: %s", groupName)

	time.Sleep(5 * time.Second)
	headResult, err := s.Client.HeadGroup(s.ClientContext, groupName, groupOwner.String())
	s.Require().NoError(err)
	s.Require().Equal(groupName, headResult.GroupName)

	s.T().Log("---> Update GroupMember <---")
	addAccount, _, err := types.NewAccount("member1")
	s.Require().NoError(err)
	updateMember := addAccount.GetAddress().String()
	updateMembers := []string{updateMember}
	txnHash, err := s.Client.UpdateGroupMember(s.ClientContext, groupName, groupOwner.String(), updateMembers, nil, types.UpdateGroupMemberOption{})
	s.T().Logf("add groupMember: %s", updateMembers[0])
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txnHash)
	s.Require().NoError(err)

	// head added member
	exist := s.Client.HeadGroupMember(s.ClientContext, groupName, groupOwner.String(), updateMember)
	s.Require().Equal(true, exist)
	if exist {
		s.T().Logf("header groupMember: %s , exist", updateMembers[0])
	}

	// remove groupMember
	txnHash, err = s.Client.UpdateGroupMember(s.ClientContext, groupName, groupOwner.String(), nil, updateMembers, types.UpdateGroupMemberOption{})
	s.T().Logf("remove groupMember: %s", updateMembers[0])
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txnHash)
	s.Require().NoError(err)

	// head removed member
	exist = s.Client.HeadGroupMember(s.ClientContext, groupName, groupOwner.String(), updateMember)
	s.Require().Equal(false, exist)
	if !exist {
		s.T().Logf("header groupMember: %s , not exist", updateMembers[0])
	}

	s.T().Log("---> Set Group Permission<---")
	grantUser, _, err := types.NewAccount("member2")
	s.Require().NoError(err)

	resp, err := s.Client.Transfer(s.ClientContext, grantUser.GetAddress().String(), math.NewIntWithDecimal(1, types2.DecimalBNB), types2.TxOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, resp)
	s.Require().NoError(err)

	statement := utils.NewStatement([]permTypes.ActionType{permTypes.ACTION_UPDATE_GROUP_MEMBER},
		permTypes.EFFECT_ALLOW, nil, types.NewStatementOptions{})

	// put group policy to another user
	txnHash, err = s.Client.PutGroupPolicy(s.ClientContext, groupName, grantUser.GetAddress().String(),
		[]*permTypes.Statement{&statement}, types.PutPolicyOption{})
	s.Require().NoError(err)

	s.T().Logf("put group policy to user %s", grantUser.GetAddress().String())
	_, err = s.Client.WaitForTx(s.ClientContext, txnHash)
	s.Require().NoError(err)
	// use this user to update group
	s.Client.SetDefaultAccount(grantUser)
	s.Require().NoError(err)

	// check permission, add back the member by grantClient
	updateHash, err := s.Client.UpdateGroupMember(s.ClientContext, groupName, groupOwner.String(), updateMembers,
		nil, types.UpdateGroupMemberOption{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, updateHash)
	s.Require().NoError(err)

	s.Client.SetDefaultAccount(s.DefaultAccount)
	// head removed member
	exist = s.Client.HeadGroupMember(s.ClientContext, groupName, groupOwner.String(), updateMember)
	s.Require().Equal(true, exist)
	if exist {
		s.T().Logf("header groupMember: %s , exist", updateMembers[0])
	}
}

func (s *StorageTestSuite) Test_DownloadFileWithCpChoiceOptions() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, types.CreateBucketOptions{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
	}

	s.T().Logf("---> Create Bucket:%s, Object:%s <---", bucketName, objectName)

	var buffer bytes.Buffer
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024*1200; i++ {
		line := types.RandStr(20)
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectInfo, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectInfo.ObjectName, objectName)
	s.Require().Equal(objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	s.T().Log("---> PutObject and GetObject <---")
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	time.Sleep(10 * time.Second)
	objectInfo, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")
	}

	fileName := "test-file-" + storageTestUtil.GenRandomObjectName()
	cp := types.CheckpointConfig{
		IsEnable: true,
		DirPath:  "./",
	}
	err = s.Client.GetObjectResumable(s.ClientContext, bucketName, objectName, types.GetObjectResumableOption{CpConfig: cp}, fileName)
	s.T().Logf("--->  object file :%s <---", fileName)
	s.T().Logf("--->  GetObjectResumable error:%s <---", err)
	s.Require().NoError(err)

	fGetObjectFileName := "test-file-" + storageTestUtil.GenRandomObjectName()
	s.T().Logf("--->  object file :%s <---", fGetObjectFileName)
	err = s.Client.FGetObject(s.ClientContext, bucketName, objectName, fGetObjectFileName, types.GetObjectOption{})
	s.T().Logf("--->  GetObjectResumable error:%s <---", err)
	s.Require().NoError(err)

	isSame, err := types.CompareFiles(fileName, fGetObjectFileName)
	s.Require().True(isSame)
	s.Require().NoError(err)
}

// DownErrorHooker requests hook by downloadSegment
func DownErrorHooker(segment types.SegmentPiece) error {
	if segment.Index == 1 {
		time.Sleep(time.Second)
		return fmt.Errorf("ErrorHooker")
	}
	return nil
}

// TestDownloadRoutineWithRecovery multi-routine resumable download
func (s *StorageTestSuite) TestDownloadRoutineWithRecovery() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, types.CreateBucketOptions{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
	}

	s.T().Logf("---> Create Bucket:%s, Object:%s <---", bucketName, objectName)

	var buffer bytes.Buffer
	// Create 18 MiB content where each line contains 1024 characters.
	for i := 0; i < 1024*20; i++ {
		line := types.RandStr(1024)
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectInfo, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectInfo.ObjectName, objectName)
	s.Require().Equal(objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	s.T().Log("---> PutObject and GetObject <---")
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	time.Sleep(10 * time.Second)
	objectInfo, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")
	}

	// Download a file with default checkpoint
	client.DownloadSegmentHooker = DownErrorHooker
	newFile := storageTestUtil.GenRandomObjectName()

	s.T().Logf("---> Create newfile:%s, <---", newFile)

	cp := types.CheckpointConfig{
		IsEnable: true,
		FilePath: newFile + ".cp",
	}
	err = s.Client.GetObjectResumable(s.ClientContext, bucketName, objectName, types.GetObjectResumableOption{CpConfig: cp}, newFile)
	s.Require().ErrorContains(err, "ErrorHooker")
	client.DownloadSegmentHooker = client.DefaultDownloadSegmentHook

	// Check
	dcp := types.GetObjectCheckpoint{}
	err = dcp.Load(newFile + ".cp")
	s.Require().NoError(err)
	s.Require().Equal(dcp.Magic, types.DownloadCpMagic)
	s.Require().Equal(dcp.FilePath, newFile)
	s.Require().Equal(dcp.ObjStat.Size, int64(21155840))
	s.Require().Equal(dcp.Object, objectName)
	s.Require().Equal(len(dcp.Segments), 2)
	s.Require().Equal(len(dcp.TodoSegments()), 1)
	// TODO(chris): add MD5 LastModified Etag

	err = s.Client.GetObjectResumable(s.ClientContext, bucketName, objectName, types.GetObjectResumableOption{CpConfig: cp}, newFile)
	s.Require().NoError(err)
	//download success, checkpoint file has been deleted
	err = dcp.Load(newFile + ".cp")
	s.Require().ErrorContains(err, "no such file or directory")

	fGetObjectFileName := "test-file-" + storageTestUtil.GenRandomObjectName()
	s.T().Logf("--->  object file :%s <---", fGetObjectFileName)
	err = s.Client.FGetObject(s.ClientContext, bucketName, objectName, fGetObjectFileName, types.GetObjectOption{})
	s.T().Logf("--->  GetObjectResumable error:%s <---", err)
	s.Require().NoError(err)

	isSame, err := types.CompareFiles(newFile, fGetObjectFileName)
	s.Require().True(isSame)
	s.Require().NoError(err)

}
