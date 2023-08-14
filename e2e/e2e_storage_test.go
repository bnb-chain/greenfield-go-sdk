package e2e

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
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
			break
		}
	}
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
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
	for i := 0; i < 1024*30; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
	s.Require().NoError(err)

	time.Sleep(10 * time.Second)
	var contentLen uint64
	objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")
	fmt.Println("content length:", objectDetail.ObjectInfo.PayloadSize, "buf len", buffer.Len())
	contentLen = objectDetail.ObjectInfo.PayloadSize

	s.T().Logf("---> PutObject and GetObject, objectName:%s objectSize:%d <---", objectName, int64(buffer.Len()))
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)

	s.waitSealObject(bucketName, objectName)

	concurrentNumber := 5
	downloadCount := 5
	quota0, err := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
	s.Require().NoError(err)
	fmt.Println("get quota;", quota0)

	for i := 0; i < concurrentNumber; i++ {
		for j := 0; j < downloadCount; j++ {
			objectContent, _, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOptions{})
			if err != nil {
				fmt.Printf("error: %v", err)
				quota2, _ := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
				//	s.NoError(err, quota2)
				fmt.Printf("quota: %v", quota2)
			}
			objectBytes, err := io.ReadAll(objectContent)
			s.Require().NoError(err)
			s.Require().Equal(objectBytes, buffer.Bytes())
		}
	}

	expectQuotaUsed := int(contentLen) * concurrentNumber * downloadCount
	fmt.Println("expect quota:", expectQuotaUsed)
	quota1, err := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
	s.Require().NoError(err)
	consumedQuota := quota1.ReadConsumedSize - quota0.ReadConsumedSize
	fmt.Println("actual quota:", consumedQuota)
	s.Require().Equal(uint64(expectQuotaUsed), consumedQuota)

	var buffer2 bytes.Buffer
	line2 := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,123456789012`
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024*17; i++ {
		buffer2.WriteString(fmt.Sprintf("[%05d] %s\n", i, line2))
	}

	s.T().Log("---> CreateObject and HeadObject <---")
	objectName2 := objectName + "xx1"

	objectTx2, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName2, bytes.NewReader(buffer2.Bytes()), types.CreateObjectOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, objectTx2)
	s.Require().NoError(err)

	time.Sleep(10 * time.Second)
	var contentLen2 uint64
	objectDetail2, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName2)
	s.Require().NoError(err)
	s.Require().Equal(objectDetail2.ObjectInfo.ObjectName, objectName2)
	s.Require().Equal(objectDetail2.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")
	fmt.Println("content length:", objectDetail2.ObjectInfo.PayloadSize, "buf len", buffer2.Len())
	contentLen2 = objectDetail2.ObjectInfo.PayloadSize

	s.T().Logf("---> PutObject and GetObject, objectName:%s objectSize:%d <---", objectName2, int64(buffer2.Len()))
	err = s.Client.PutObject(s.ClientContext, bucketName, objectName2, int64(buffer2.Len()),
		bytes.NewReader(buffer2.Bytes()), types.PutObjectOptions{})
	s.Require().NoError(err)
	s.waitSealObject(bucketName, objectName2)

	quota0, err = s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
	s.Require().NoError(err)

	var wg sync.WaitGroup
	wg.Add(concurrentNumber + 40)
	downloadTime1 := 0
	downloadTime2 := 0
	for j := 0; j < concurrentNumber+40; j++ {
		if j%2 == 0 {
			go func() {
				downloadTime1++
				defer wg.Done()
				for i := 0; i < 5; i++ {
					objectContent, _, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, types.GetObjectOptions{})
					if err != nil {
						fmt.Printf("error: %v", err)
						quota2, _ := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
						//	s.NoError(err, quota2)
						fmt.Printf("quota: %v", quota2)
					}
					objectBytes, err := io.ReadAll(objectContent)
					s.Require().NoError(err)
					s.Require().Equal(objectBytes, buffer.Bytes())
				}
			}()
		} else {
			go func() {
				downloadTime2++
				defer wg.Done()
				for i := 0; i < 5; i++ {
					objectContent, _, err := s.Client.GetObject(s.ClientContext, bucketName, objectName2, types.GetObjectOptions{})
					if err != nil {
						fmt.Printf("error: %v", err)
						quota2, _ := s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
						//	s.NoError(err, quota2)
						fmt.Printf("quota: %v", quota2)
					}
					objectBytes2, err := io.ReadAll(objectContent)
					s.Require().NoError(err)
					s.Require().Equal(objectBytes2, buffer2.Bytes())
				}
			}()
		}
	}
	wg.Wait()

	concurrentOne := (concurrentNumber + 40 + 1) / 2
	fmt.Println("concurrentOne", concurrentOne, "download time,", downloadTime1, downloadTime2)
	expectQuotaUsed = int(contentLen)*(downloadTime1)*downloadCount + int(contentLen2)*(downloadTime2)*downloadCount

	fmt.Println("expect quota:", expectQuotaUsed)
	quota1, err = s.Client.GetBucketReadQuota(s.ClientContext, bucketName)
	fmt.Println("Get quota:", quota1.ReadConsumedSize, "free :", quota1.SPFreeReadQuotaSize)
	s.Require().NoError(err)
	consumedQuota = quota1.ReadConsumedSize - quota0.ReadConsumedSize
	fmt.Println("actual quota:", consumedQuota)

	s.Require().Equal(uint64(expectQuotaUsed), consumedQuota)

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

func (s *StorageTestSuite) waitSealObject(bucketName string, objectName string) {
	startCheckTime := time.Now()
	var (
		objectDetail *types.ObjectDetail
		err          error
	)

	// wait 300s
	for i := 0; i < 100; i++ {
		objectDetail, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
		s.Require().NoError(err)
		if objectDetail.ObjectInfo.GetObjectStatus() == storageTypes.OBJECT_STATUS_SEALED {
			break
		}
		time.Sleep(3 * time.Second)
	}

	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")
	s.T().Logf("---> Wait Seal Object cost %d ms, <---", time.Since(startCheckTime).Milliseconds())
}
