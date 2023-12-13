package e2e

import (
	"bytes"
	"context"
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	types2 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
	"time"
)

type SPExitTestSuite struct {
	basesuite.BaseSuite
	ExitSP           spTypes.StorageProvider
	AnotherPrimarySP spTypes.StorageProvider
	SuccessorSP      spTypes.StorageProvider
}

func (s *SPExitTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	spList, err := s.Client.ListStorageProviders(s.ClientContext, false)
	s.Require().NoError(err)
	for _, sp := range spList {
		if sp.Endpoint != "https://sp0.greenfield.io" {
			if sp.Id == 1 {
				s.ExitSP = sp
			}
			if sp.Id == 2 {
				s.AnotherPrimarySP = sp
			}
			if sp.Id == 7 {
				s.SuccessorSP = sp
			}
		}
	}
}

func (s *SPExitTestSuite) Test_SP_Exit_VGF() {

	// User creates 1 object
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	s.T().Logf("BucketName:%s, objectName: %s", bucketName, objectName)

	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.ExitSP.OperatorAddress, types.CreateBucketOptions{})
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
	for i := 0; i < 1024*1; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	for i := 0; i < 12; i++ {
		s.T().Log("---> CreateObject and HeadObject <---")
		objectName = fmt.Sprintf("%s%d", objectName, i)
		objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
		s.Require().NoError(err)
		_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
		s.Require().NoError(err)

		objectSize := int64(buffer.Len())
		s.T().Logf("---> PutObject and GetObject, objectName:%s objectSize:%d <---", objectName, objectSize)
		err = s.Client.PutObject(s.ClientContext, bucketName, objectName, objectSize,
			bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
		s.Require().NoError(err)
		s.WaitSealObject(bucketName, objectName)
	}
	// SP0 send tx to exit
	s.T().Log(s.SP0Account.GetAddress().String())

	s.Client.SetDefaultAccount(s.SP0Account)
	msgExit := &types2.MsgStorageProviderExit{
		StorageProvider: s.SP0Account.GetAddress().String(),
	}

	tx, err := s.Client.BroadcastTx(context.Background(), []sdk.Msg{msgExit}, nil)
	s.Require().NoError(err)
	s.T().Logf("sp exit tx %v", tx)

	time.Sleep(5 * time.Second)
	// SP7 send tx to reserve SwapIn
	msgReserveSwapIn := types2.NewMsgReserveSwapIn(s.SP7Account.GetAddress(), s.ExitSP.Id, 1, 0)
	s.Client.SetDefaultAccount(s.SP7Account)
	tx, err = s.Client.BroadcastTx(context.Background(), []sdk.Msg{msgReserveSwapIn}, nil)
	s.Require().NoError(err)
	s.T().Logf("sp swapIn tx %v", tx)
}

func NewIntFromInt64WithDecimal(amount int64, decimal int64) sdkmath.Int {
	return sdk.NewInt(amount).Mul(sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimal), nil)))
}

func (s *SPExitTestSuite) Test_SP_Exit_GVG() {

	// SP1 create a GVG excluding Sp7(Id=8)
	s.Client.SetDefaultAccount(s.SP1Account)
	deposit := sdk.Coin{
		Denom:  "BNB",
		Amount: NewIntFromInt64WithDecimal(1, 18),
	}
	msgCreateGVG := &types2.MsgCreateGlobalVirtualGroup{
		StorageProvider: s.AnotherPrimarySP.OperatorAddress,
		SecondarySpIds:  []uint32{7, 3, 4, 5, 6, 1},
		Deposit:         deposit,
		FamilyId:        0,
	}

	tx, err := s.Client.BroadcastTx(context.Background(), []sdk.Msg{msgCreateGVG}, nil)
	s.Require().NoError(err)
	s.T().Logf("tx hash %s", tx.TxResponse.TxHash)
	_, err = s.Client.WaitForTx(s.ClientContext, tx.TxResponse.TxHash)
	s.Require().NoError(err)

	time.Sleep(6 * time.Second)

	s.Client.SetDefaultAccount(s.DefaultAccount)
	// User creates 1 object on SP1
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()

	s.T().Logf("BucketName:%s, objectName: %s", bucketName, objectName)

	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.AnotherPrimarySP.OperatorAddress, types.CreateBucketOptions{})
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
	for i := 0; i < 1024*1; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	for i := 0; i < 1; i++ {
		s.T().Log("---> CreateObject and HeadObject <---")
		objectName = fmt.Sprintf("%s%d", objectName, i)
		objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
		s.Require().NoError(err)
		_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
		s.Require().NoError(err)

		objectSize := int64(buffer.Len())
		s.T().Logf("---> PutObject and GetObject, objectName:%s objectSize:%d <---", objectName, objectSize)
		err = s.Client.PutObject(s.ClientContext, bucketName, objectName, objectSize,
			bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
		s.Require().NoError(err)
		s.WaitSealObject(bucketName, objectName)
	}
	// SP0 send tx to exit
	s.T().Log(s.SP0Account.GetAddress().String())

	s.Client.SetDefaultAccount(s.SP0Account)
	msgExit := &types2.MsgStorageProviderExit{
		StorageProvider: s.SP0Account.GetAddress().String(),
	}
	tx, err = s.Client.BroadcastTx(context.Background(), []sdk.Msg{msgExit}, nil)
	s.Require().NoError(err)
	s.T().Logf("sp exit tx %v", tx)

	time.Sleep(3 * time.Second)
	// SP7 send tx to reserve SwapIn SP0 secodnary in SP1's GVG
	msgReserveSwapIn := types2.NewMsgReserveSwapIn(s.SP7Account.GetAddress(), s.ExitSP.Id, 0, 1)
	s.Client.SetDefaultAccount(s.SP7Account)
	tx, err = s.Client.BroadcastTx(context.Background(), []sdk.Msg{msgReserveSwapIn}, nil)
	s.Require().NoError(err)
	s.T().Logf("sp swapIn tx %v", tx)
}

func TestSPExitTestSuite(t *testing.T) {
	suite.Run(t, new(SPExitTestSuite))
}
