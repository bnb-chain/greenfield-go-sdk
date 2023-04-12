package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

func Test_Storage(t *testing.T) {
	bucketName := "test-bucket"
	objectName := "test-object"

	mnemonic := ParseValidatorMnemonic(0)
	account, err := types.NewAccountFromMnemonic("test", mnemonic)
	assert.NoError(t, err)
	cli, err := client.New(ChainID, GrpcAddress, client.Option{
		DefaultAccount: account,
		GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials()),
		Host:           bucketName + ".gnfd.nodereal.com",
	})
	assert.NoError(t, err)
	ctx := context.Background()

	spList, err := cli.ListSP(ctx, false)
	assert.NoError(t, err)
	primarySp := spList[0].GetOperator()

	chargedQuota := uint64(100)
	t.Log("---> CreateBucket and HeadBucket <---")
	opts := types.CreateBucketOptions{ChargedQuota: chargedQuota}
	bucketTx, err := cli.CreateBucket(ctx, bucketName, primarySp, opts)
	assert.NoError(t, err)

	_, err = cli.WaitForTx(ctx, bucketTx)
	assert.NoError(t, err)

	bucketInfo, err := cli.HeadBucket(ctx, bucketName)
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
		assert.Equal(t, bucketInfo.ChargedReadQuota, chargedQuota)
	}

	t.Log("--->  UpdateBucket <---")
	updateBucketTx, err := cli.UpdateBucketVisibility(ctx, bucketName,
		storageTypes.VISIBILITY_TYPE_PUBLIC_READ, types.UpdateVisibilityOption{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, updateBucketTx)
	assert.NoError(t, err)

	t.Log("---> BuyQuotaForBucket <---")
	targetQuota := uint64(300)
	buyQuotaTx, err := cli.BuyQuotaForBucket(ctx, bucketName, targetQuota, types.BuyQuotaOption{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, buyQuotaTx)
	assert.NoError(t, err)

	t.Log("---> BuyQuotaForBucket <---")
	quota, err := cli.GetBucketReadQuota(ctx, bucketName)
	assert.NoError(t, err)
	assert.Equal(t, quota.ReadQuotaSize, targetQuota)

	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890`
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024*100; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	t.Log("---> CreateObject and HeadObject <---")
	objectTx, err := cli.CreateObject(ctx, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, objectTx)
	assert.NoError(t, err)

	time.Sleep(5 * time.Second)
	objectInfo, err := cli.HeadObject(ctx, bucketName, objectName)
	assert.NoError(t, err)
	assert.Equal(t, objectInfo.ObjectName, objectName)
	assert.Equal(t, objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")

	t.Log("---> PutObject and GetObject <---")
	// put Object
	err = cli.PutObject(ctx, bucketName, objectName, objectTx, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOption{})
	assert.NoError(t, err)

	time.Sleep(10 * time.Second)
	objectInfo, err = cli.HeadObject(ctx, bucketName, objectName)
	assert.NoError(t, err)
	assert.Equal(t, objectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")

	// GetObject
	ior, info, err := cli.GetObject(ctx, bucketName, objectName, types.GetObjectOption{})
	assert.NoError(t, err)
	assert.Equal(t, info.ObjectName, objectName)
	assert.Equal(t, info.Size, len(buffer.Bytes()))
	objectBytes, err := io.ReadAll(ior)
	assert.NoError(t, err)
	assert.Equal(t, objectBytes, buffer.Bytes())
}

func Test_Group(t *testing.T) {
	groupName := "test-group"
	mnemonic := ParseValidatorMnemonic(0)
	account, err := types.NewAccountFromMnemonic("test", mnemonic)
	assert.NoError(t, err)
	cli, err := client.New(ChainID, GrpcAddress, client.Option{
		DefaultAccount: account,
		GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials()),
	})
	assert.NoError(t, err)
	ctx := context.Background()

	groupOwner := account.GetAddress()
	t.Log("---> CreateGroup and HeadGroup <---")
	_, err = cli.CreateGroup(ctx, groupName, types.CreateGroupOptions{})
	assert.NoError(t, err)
	t.Logf("create GroupName: %s", groupName)

	time.Sleep(5 * time.Second)
	headResult, err := cli.HeadGroup(ctx, groupName, groupOwner)
	assert.NoError(t, err)
	assert.Equal(t, groupName, headResult.GroupName)

	t.Log("---> Update GroupMember <---")
	mnemonic = ParseValidatorMnemonic(1)
	addAccount, err := types.NewAccountFromMnemonic("test1", mnemonic)
	assert.NoError(t, err)
	updateMember := addAccount.GetAddress()
	updateMembers := []sdk.AccAddress{updateMember}
	txnHash, err := cli.UpdateGroupMember(ctx, groupName, groupOwner, updateMembers, nil, types.UpdateGroupMemberOption{})
	t.Logf("add groupMember: %s", updateMember.String())
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txnHash)
	assert.NoError(t, err)

	// head added member
	exist := cli.HeadGroupMember(ctx, groupName, groupOwner, updateMember)
	assert.Equal(t, true, exist)
	if exist {
		t.Logf("header groupMember: %s , exist", updateMember.String())
	}

	// remove groupMember
	txnHash, err = cli.UpdateGroupMember(ctx, groupName, groupOwner, nil, updateMembers, types.UpdateGroupMemberOption{})
	t.Logf("remove groupMember: %s", updateMember.String())
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txnHash)
	assert.NoError(t, err)

	// head removed member
	exist = cli.HeadGroupMember(ctx, groupName, groupOwner, updateMember)
	assert.Equal(t, false, exist)
	if !exist {
		t.Logf("header groupMember: %s , not exist", updateMember.String())
	}

	t.Log("---> Set Group Permission<---")
	mnemonic = ParseValidatorMnemonic(2)
	grantUser, err := types.NewAccountFromMnemonic("test2", mnemonic)
	assert.NoError(t, err)
	statement := utils.NewStatement([]permTypes.ActionType{permTypes.ACTION_UPDATE_GROUP_MEMBER},
		permTypes.EFFECT_ALLOW, nil, types.NewStatementOptions{})

	// put group policy to another user
	txnHash, err = cli.PutGroupPolicy(ctx, groupName, grantUser.GetAddress(),
		[]*permTypes.Statement{&statement}, types.PutPolicyOption{})
	assert.NoError(t, err)

	t.Logf("put group policy to user %s", grantUser.GetAddress().String())
	_, err = cli.WaitForTx(ctx, txnHash)
	assert.NoError(t, err)
	// use this user to update group
	grantClient, err := client.New(ChainID, GrpcAddress, client.Option{
		DefaultAccount: grantUser,
		GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials()),
	})
	assert.NoError(t, err)

	// check permission, add back the member by grantClient
	txnHash, err = grantClient.UpdateGroupMember(ctx, groupName, groupOwner, updateMembers,
		nil, types.UpdateGroupMemberOption{})
	assert.NoError(t, err)
	_, err = cli.WaitForTx(ctx, txnHash)
	assert.NoError(t, err)

	// head removed member
	exist = cli.HeadGroupMember(ctx, groupName, groupOwner, updateMember)
	assert.Equal(t, true, exist)
	if exist {
		t.Logf("header groupMember: %s , exist", updateMember.String())
	}

}