package e2e

import (
	"bytes"
	"fmt"
	"github.com/bnb-chain/greenfield/types/resource"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	greenfield_types "github.com/bnb-chain/greenfield/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
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
		if sp.Endpoint != "https://sp0.greenfield.io" && sp.Id == 1 {
			s.PrimarySP = sp
			break
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

	s.T().Logf("BucketName:%s, objectName: %s", bucketName, objectName)

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
	for i := 0; i < 1024*300; i++ {
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

	objectSize := int64(buffer.Len())
	s.T().Logf("---> PutObject and GetObject, objectName:%s objectSize:%d <---", objectName, objectSize)
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, objectSize,
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	s.WaitSealObject(bucketName, objectName)

	var updatedBuffer bytes.Buffer
	for i := 0; i < 1024*300; i++ {
		updatedBuffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	objectTx, err = s.Client.UpdateObjectContent(s.ClientContext, bucketName, objectName, bytes.NewReader(updatedBuffer.Bytes()), types.UpdateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)
	s.T().Logf("UpdateObjectContent tx hash %s", objectTx)

	objectDetail, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(true, objectDetail.ObjectInfo.IsUpdating)

	objectSize = int64(updatedBuffer.Len())
	s.T().Logf("---> PutObject, objectName:%s objectSize:%d <---", objectName, objectSize)

	err = s.PutObjectWithRetry(bucketName, objectName, objectSize,
		updatedBuffer, types.PutObjectOptions{})
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)

	s.T().Log("---> Get bucket quota <---")

	concurrentNumber := 5
	downloadCount := 5
	quota0, err := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
	s.Require().NoError(err)

	var wg sync.WaitGroup
	wg.Add(concurrentNumber)

	for i := 0; i < concurrentNumber; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < downloadCount; j++ {
				objectContent, _, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOptions{})
				if err != nil {
					fmt.Printf("error: %v", err)
					quota2, _ := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
					fmt.Printf("quota: %v", quota2)
				}
				objectBytes, err := io.ReadAll(objectContent)
				s.Require().NoError(err)
				s.Require().Equal(objectBytes, buffer.Bytes())
			}
		}()
	}
	wg.Wait()

	expectQuotaUsed := int(objectSize) * concurrentNumber * downloadCount
	quota1, err := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
	s.Require().NoError(err)
	freeQuotaConsumed := quota1.FreeConsumedSize - quota0.FreeConsumedSize
	// the consumed quota and free quota should be right
	s.Require().Equal(uint64(expectQuotaUsed), freeQuotaConsumed)

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

	s.T().Log("--->  ListObjectPolicies <---")
	objectPolicies, err := s.Client.ListObjectPolicies(s.ClientContext, objectName, bucketName, uint32(permTypes.ACTION_GET_OBJECT), types.ListObjectPoliciesOptions{})
	s.Require().NoError(err)
	s.Require().Equal(resource.RESOURCE_TYPE_OBJECT.String(), resource.ResourceType_name[objectPolicies.Policies[0].ResourceType])
	s.T().Logf("list object policies principal type:%d principal value:%s \n", objectPolicies.Policies[0].PrincipalType, objectPolicies.Policies[0].PrincipalValue)

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

	objectName2 := storageTestUtil.GenRandomObjectName()

	//var buffer bytes.Buffer
	//// Create 1MiB content where each line contains 1024 characters.
	//for i := 0; i < 1024*300; i++ {
	//	buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	//}
	err = s.Client.DelegatePutObject(s.ClientContext, bucketName, objectName2, objectSize, bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	s.WaitSealObject(bucketName, objectName2)

	var newBuffer bytes.Buffer
	for i := 0; i < 1024*300*40; i++ {
		newBuffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	newObjectSize := int64(newBuffer.Len())
	s.T().Logf("newObjectSize: %d", newObjectSize)

	err = s.Client.DelegateUpdateObjectContent(s.ClientContext, bucketName, objectName2, newObjectSize, bytes.NewReader(newBuffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

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

// UploadErrorHooker is a UploadPart hook---it will fail the 2nd segment's upload.
func UploadErrorHooker(id int) error {
	if id == 2 {
		time.Sleep(time.Second)
		return fmt.Errorf("UploadErrorHooker")
	}
	return nil
}

// DownloadErrorHooker requests hook by downloadSegment
func DownloadErrorHooker(segment int64) error {
	if segment == 2 {
		time.Sleep(time.Second)
		return fmt.Errorf("DownloadErrorHooker")
	}
	return nil
}

func (s *StorageTestSuite) createBigObjectWithoutPutObject() (bucket string, object string, objectbody bytes.Buffer) {
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
	// Create 45 MiB content, 3 segment
	for i := 0; i < 1024*1500; i++ {
		line := types.RandStr(20)
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	s.T().Logf("---> Create Bucket:%s, Object:%s <---", bucketName, objectName)

	return bucketName, objectName, buffer
}

func getTmpFilesInDirectory(directory string) ([]string, error) {
	var tmpFiles []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".temp" {
			tmpFiles = append(tmpFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(tmpFiles, func(i, j int) bool {
		fileInfoI, _ := os.Stat(tmpFiles[i])
		fileInfoJ, _ := os.Stat(tmpFiles[j])
		return fileInfoI.ModTime().After(fileInfoJ.ModTime())
	})

	return tmpFiles, nil
}

func (s *StorageTestSuite) TruncateDownloadTempFileToLessPartsize() {
	// Truncate file to less part size
	dir, err := os.Getwd()
	s.Require().NoError(err)
	files, err := getTmpFilesInDirectory(dir)
	s.Require().NoError(err)
	tempFilePath := files[0]

	file, err := os.OpenFile(tempFilePath, os.O_RDWR, 0o666)
	s.Require().NoError(err)
	defer file.Close()

	fileInfo, err := file.Stat()
	s.Require().NoError(err)
	currentSize := fileInfo.Size()
	targetSize := currentSize - 3*1024*1024

	err = file.Truncate(targetSize)
	s.T().Logf("---> Truncate file:%s to %d <---", tempFilePath, targetSize)
	s.Require().NoError(err)
}

func (s *StorageTestSuite) Test_Resumable_Upload_And_Download() {
	// 1) create big object without putobject
	bucketName, objectName, buffer := s.createBigObjectWithoutPutObject()

	s.T().Log("---> Resumable PutObject <---")
	partSize16MB := uint64(1024 * 1024 * 16)
	// 2) put a big object, the secondary segment will error, then resumable upload
	client.UploadSegmentHooker = UploadErrorHooker
	err := s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{PartSize: partSize16MB})
	s.Require().ErrorContains(err, "UploadErrorHooker")
	client.UploadSegmentHooker = client.DefaultUploadSegment

	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{PartSize: partSize16MB})
	s.Require().NoError(err)

	s.WaitSealObject(bucketName, objectName)

	// 3) FGetObjectResumable compare with FGetObject
	fileName := "test-file-" + storageTestUtil.GenRandomObjectName()
	defer os.Remove(fileName)
	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, fileName, types.GetObjectOptions{PartSize: 32 * 1024 * 1024})
	s.T().Logf("--->  object file :%s <---", fileName)
	s.Require().NoError(err)

	fGetObjectFileName := "test-file-" + storageTestUtil.GenRandomObjectName()
	defer os.Remove(fGetObjectFileName)
	s.T().Logf("--->  object file :%s <---", fGetObjectFileName)
	err = s.Client.FGetObject(s.ClientContext, bucketName, objectName, fGetObjectFileName, types.GetObjectOptions{})
	s.Require().NoError(err)

	isSame, err := types.CompareFiles(fileName, fGetObjectFileName)
	s.Require().True(isSame)
	s.Require().NoError(err)

	// 4) Resumabledownload, download a file with default checkpoint
	client.DownloadSegmentHooker = DownloadErrorHooker
	resumableDownloadFile := storageTestUtil.GenRandomObjectName()
	defer os.Remove(resumableDownloadFile)
	s.T().Logf("---> Resumable download Create newfile:%s, <---", resumableDownloadFile)

	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, resumableDownloadFile, types.GetObjectOptions{PartSize: 16 * 1024 * 1024})
	s.Require().ErrorContains(err, "DownloadErrorHooker")
	client.DownloadSegmentHooker = client.DefaultDownloadSegmentHook

	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, resumableDownloadFile, types.GetObjectOptions{PartSize: 16 * 1024 * 1024})
	s.Require().NoError(err)
	// download success, checkpoint file has been deleted

	isSame, err = types.CompareFiles(resumableDownloadFile, fGetObjectFileName)
	s.Require().True(isSame)
	s.Require().NoError(err)

	// when the downloaded file size is less than a part size
	client.DownloadSegmentHooker = DownloadErrorHooker
	resumableDownloadLessPartFile := storageTestUtil.GenRandomObjectName()
	defer os.Remove(resumableDownloadLessPartFile)
	s.T().Logf("---> Resumable download for less part size , Create newfile:%s, <---", resumableDownloadLessPartFile)

	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, resumableDownloadLessPartFile, types.GetObjectOptions{PartSize: 16 * 1024 * 1024})
	s.Require().ErrorContains(err, "DownloadErrorHooker")

	s.TruncateDownloadTempFileToLessPartsize()

	client.DownloadSegmentHooker = client.DefaultDownloadSegmentHook

	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, resumableDownloadLessPartFile, types.GetObjectOptions{PartSize: 16 * 1024 * 1024})
	s.Require().NoError(err)
	// download success, checkpoint file has been deleted

	isSame, err = types.CompareFiles(resumableDownloadLessPartFile, fGetObjectFileName)
	s.Require().True(isSame)
	s.Require().NoError(err)

	// 5) Resumabledownload, download a file with range
	s.T().Logf("--->  Resumabledownload, download a file with range <---")
	rangeOptions := types.GetObjectOptions{Range: "bytes=1000-94131999", PartSize: partSize16MB}
	resumableDownloadWithRangeFile := "test-file-" + storageTestUtil.GenRandomObjectName()
	defer os.Remove(resumableDownloadWithRangeFile)
	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, resumableDownloadWithRangeFile, rangeOptions)
	s.T().Logf("--->  object file :%s <---", resumableDownloadWithRangeFile)
	s.Require().NoError(err)

	fGetObjectWithRangeFile := "test-file-" + storageTestUtil.GenRandomObjectName()
	defer os.Remove(fGetObjectWithRangeFile)
	s.T().Logf("--->  object file :%s <---", fGetObjectWithRangeFile)
	err = s.Client.FGetObject(s.ClientContext, bucketName, objectName, fGetObjectWithRangeFile, rangeOptions)
	s.Require().NoError(err)

	isSame, err = types.CompareFiles(resumableDownloadWithRangeFile, fGetObjectWithRangeFile)
	s.Require().True(isSame)
	s.Require().NoError(err)

	// 6) Resumabledownload, download a file with range and Truncate
	s.T().Logf("--->  Resumabledownload, download a file with range and Truncate <---")
	rDownloadTruncateFile := "test-file-" + storageTestUtil.GenRandomObjectName()
	defer os.Remove(rDownloadTruncateFile)
	client.DownloadSegmentHooker = DownloadErrorHooker
	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, rDownloadTruncateFile, rangeOptions)
	s.T().Logf("--->  object file :%s <---", rDownloadTruncateFile)
	s.Require().ErrorContains(err, "DownloadErrorHooker")
	s.TruncateDownloadTempFileToLessPartsize()

	client.DownloadSegmentHooker = client.DefaultDownloadSegmentHook
	err = s.Client.FGetObjectResumable(s.ClientContext, bucketName, objectName, rDownloadTruncateFile, rangeOptions)
	s.Require().NoError(err)

	isSame, err = types.CompareFiles(rDownloadTruncateFile, fGetObjectWithRangeFile)
	s.Require().True(isSame)
	s.Require().NoError(err)
}

func (s *StorageTestSuite) Test_Upload_Object_With_Tampering_Content() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	s.T().Logf("BucketName:%s, objectName: %s", bucketName, objectName)

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
	for i := 0; i < 1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	var tamperingBuffer bytes.Buffer
	tamperingLine := `0987654321,0987654321,0987654321,0987654321,0987654321,0987654321,0987654321,0987654321,098765432112`
	for i := 0; i < 1024; i++ {
		tamperingBuffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, tamperingLine))
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

	objectSize := int64(tamperingBuffer.Len())
	s.T().Logf("---> PutObject and GetObject, objectName:%s objectSize:%d <---", objectName, objectSize)
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, objectSize,
		bytes.NewReader(tamperingBuffer.Bytes()), types.PutObjectOptions{})
	s.Require().Error(err)

	time.Sleep(20 * time.Second)

	// Object should not be sealed
	objectDetail, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")
}

func (s *StorageTestSuite) Test_Group_with_Tag() {
	//create group with tag
	groupName := storageTestUtil.GenRandomGroupName()

	groupOwner := s.DefaultAccount.GetAddress()
	s.T().Log("---> CreateGroup and HeadGroup <---")

	var tags storageTypes.ResourceTags
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key2", Value: "value2"})

	_, err := s.Client.CreateGroup(s.ClientContext, groupName, types.CreateGroupOptions{Tags: &tags})
	s.Require().NoError(err)
	s.T().Logf("create GroupName: %s", groupName)

	time.Sleep(5 * time.Second)
	headResult, err := s.Client.HeadGroup(s.ClientContext, groupName, groupOwner.String())
	s.Require().NoError(err)
	s.Require().Equal(groupName, headResult.GroupName)
	s.Require().Equal(tags, *headResult.Tags)
}

func (s *StorageTestSuite) Test_CreateGroup_And_Set_Tag() {
	//create group with tag
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

	grn := greenfield_types.NewGroupGRN(groupOwner, groupName)
	var tags storageTypes.ResourceTags
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key2", Value: "value2"})

	_, err = s.Client.SetTag(s.ClientContext, grn.String(), tags, types.SetTagsOptions{})
	s.Require().NoError(err)
	s.T().Logf("set tag: %v for group %s", tags, groupName)

	time.Sleep(5 * time.Second)
	headResult, err = s.Client.HeadGroup(s.ClientContext, groupName, groupOwner.String())
	s.Require().NoError(err)
	s.Require().Equal(groupName, headResult.GroupName)
	s.Require().Equal(tags, *headResult.Tags)
}

func (s *StorageTestSuite) Test_Bucket_with_Tag() {
	bucketName := storageTestUtil.GenRandomBucketName()
	var tags storageTypes.ResourceTags
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key2", Value: "value2"})

	chargedQuota := uint64(100)
	s.T().Log("---> CreateBucket and HeadBucket <---")
	opts := types.CreateBucketOptions{ChargedQuota: chargedQuota, Tags: &tags}

	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, opts)
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
		s.Require().Equal(bucketInfo.ChargedReadQuota, chargedQuota)
		s.Require().Equal(tags, *bucketInfo.Tags)
	}
}

func (s *StorageTestSuite) Test_CreateBucket_And_Set_Tag() {
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

	// set tag
	grn := greenfield_types.NewBucketGRN(bucketName)
	var tags storageTypes.ResourceTags
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key2", Value: "value2"})

	_, err = s.Client.SetTag(s.ClientContext, grn.String(), tags, types.SetTagsOptions{})
	s.Require().NoError(err)
	s.T().Logf("set tag: %v for bucket %s", tags, bucketName)

	time.Sleep(5 * time.Second)
	bucketInfo, err = s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(tags, *bucketInfo.Tags)
	}
}

func (s *StorageTestSuite) Test_Object_with_Tag() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	s.T().Logf("BucketName:%s, objectName: %s", bucketName, objectName)

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
	for i := 0; i < 1024*300; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	var tags storageTypes.ResourceTags
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key2", Value: "value2"})
	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()),
		types.CreateObjectOptions{Tags: &tags})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")
	s.Require().Equal(tags, *objectDetail.ObjectInfo.Tags)
}

func (s *StorageTestSuite) Test_Object_And_Set_Tag() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	s.T().Logf("BucketName:%s, objectName: %s", bucketName, objectName)

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
	for i := 0; i < 1024*300; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()),
		types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	// set tag
	grn := greenfield_types.NewObjectGRN(bucketName, objectName)
	var tags storageTypes.ResourceTags
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storageTypes.ResourceTags_Tag{Key: "key2", Value: "value2"})

	_, err = s.Client.SetTag(s.ClientContext, grn.String(), tags, types.SetTagsOptions{})
	s.Require().NoError(err)
	s.T().Logf("set tag: %v for object %s", tags, objectName)

	time.Sleep(5 * time.Second)
	objectDetail, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(tags, *objectDetail.ObjectInfo.Tags)
	}
}

func (s *StorageTestSuite) Test_Get_Object_With_ForcedSpEndpoint() {
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	s.T().Logf("BucketName:%s, objectName: %s", bucketName, objectName)

	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, types.CreateBucketOptions{Visibility: storageTypes.VISIBILITY_TYPE_PUBLIC_READ})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PUBLIC_READ)
	}

	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,123456789012`
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024*300; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{Visibility: storageTypes.VISIBILITY_TYPE_PUBLIC_READ})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)
	objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	objectSize := int64(buffer.Len())
	s.T().Logf("---> PutObject and GetObject, objectName:%s objectSize:%d <---", objectName, objectSize)
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, objectSize,
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	s.WaitSealObject(bucketName, objectName)

	s.T().Log("---> client.New with ForceToUseSpecifiedSpEndpointForDownloadOnly option param filled <---")
	origClient := s.Client
	s.Client, err = client.New(basesuite.ChainID, basesuite.Endpoint, client.Option{
		DefaultAccount: s.DefaultAccount,
		ForceToUseSpecifiedSpEndpointForDownloadOnly: s.PrimarySP.Endpoint,
	})
	s.Require().NoError(err)

	s.T().Log("---> get object with ForceToUseSpecifiedSpEndpointForDownloadOnly <---")
	objectContent, _, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOptions{})
	if err != nil {
		fmt.Printf("error: %v", err)
		quota2, _ := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
		fmt.Printf("quota: %v", quota2)
	}
	objectBytes, err := io.ReadAll(objectContent)
	s.Require().NoError(err)
	s.Require().Equal(objectBytes, buffer.Bytes())
	s.Require().NoError(err)

	s.T().Log("---> restore client without ForceToUseSpecifiedSpEndpointForDownloadOnly option param <---")
	s.Client = origClient
}

func (s *StorageTestSuite) PutObjectWithRetry(bucketName, objectName string, objectSize int64, buffer bytes.Buffer, option types.PutObjectOptions) error {
	var err error
	for retry := 0; retry < 5; retry++ {
		err = s.Client.PutObject(s.ClientContext, bucketName, objectName, objectSize,
			bytes.NewReader(buffer.Bytes()), option)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	return err
}
