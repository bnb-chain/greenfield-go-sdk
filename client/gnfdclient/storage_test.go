package gnfdclient

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gnfd_client "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	spType "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/stretchr/testify/require"
)

var NewPrivateKeyManager = keys.NewPrivateKeyManager
var WithGrpcDialOption = gnfd_client.WithGrpcDialOption
var WithKeyManager = gnfd_client.WithKeyManager

func TestCreateBucket(t *testing.T) {
	// keyManager, err := keys.NewPrivateKeyManager("92b4cdd49090cab5f3b6bf334021486ac3d3de865e02d1bced72f95f933c15e8")
	keyManager, err := keys.NewPrivateKeyManager("6547492644d0136f76ef65e3bd04a77d079ed38028f747700c6c6063564d7032")
	require.NoError(t, err)

	fmt.Println("addr:", keyManager.GetAddr().String())
	// dev
	grpcAddr := "gnfd-dev-grpc.qa.bnbchain.world:9090"
	chainId := "greenfield_7971-1"
	endpoint := "gf-sp-a-bk.dev.nodereal.cc"

	client2, err := NewGnfdClient(grpcAddr, chainId, endpoint, keyManager, false,
		WithKeyManager(keyManager),
		WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))

	ctx := context.Background()
	var resp GnfdResponse

	request := &spType.QueryStorageProvidersRequest{}
	gnfdRep, err := client2.ChainClient.StorageProviders(ctx, request)
	require.NoError(t, err)

	var opAddress sdk.AccAddress
	for _, sp := range gnfdRep.GetSps() {
		fmt.Println("sp list:", sp.String())
		if sp.Description.Moniker == "sp0" {
			opAddress = sp.GetOperator()
		}
	}
	fmt.Println("resp:", resp)

	bucketName := "testxxx22343"
	resp = client2.CreateBucket(ctx, bucketName, CreateBucketOptions{PrimarySPAddress: opAddress})
	require.NoError(t, resp.Err)

}

func TestPutPolicy(t *testing.T) {
	keyManager1, err := keys.NewPrivateKeyManager("92b4cdd49090cab5f3b6bf334021486ac3d3de865e02d1bced72f95f933c15e8")
	//keyManager1, err := keys.NewPrivateKeyManager("6547492644d0136f76ef65e3bd04a77d079ed38028f747700c6c6063564d7032")
	require.NoError(t, err)
	//user1
	keyManager2, err := keys.NewPrivateKeyManager("10a19eb168b3e3ddae3c3ff06c2538d2ce559fe39b70a4b4dc0e03c60b3cbba6")
	require.NoError(t, err)
	user2 := keyManager2.GetAddr()

	// qa
	grpcAddr := "gnfd-grpc-plaintext.qa.bnbchain.world:9090"
	chainId := "greenfield_9000-1741"
	endpoint := "gf-sp-a.bk.nodereal.cc"

	client, err := NewGnfdClient(grpcAddr, chainId, endpoint, keyManager1, false,
		WithKeyManager(keyManager1),
		WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	ctx := context.Background()
	var resp GnfdResponse

	request := &spType.QueryStorageProvidersRequest{}
	gnfdRep, err := client.ChainClient.StorageProviders(ctx, request)
	require.NoError(t, err)

	var opAddress sdk.AccAddress
	for _, sp := range gnfdRep.GetSps() {
		if sp.Description.Moniker == "sp0" {
			opAddress = sp.GetOperator()
		}
	}

	bucketName := GenRandomBucketName()

	resp = client.CreateBucket(ctx, bucketName, CreateBucketOptions{PrimarySPAddress: opAddress})
	require.NoError(t, resp.Err)

	time.Sleep(10 * time.Second)
	// head bucket
	_, err = client.HeadBucket(ctx, bucketName)
	fmt.Println("bucekt name:", bucketName)
	require.NoError(t, err)
	// verify permission
	permInfo := client.IsBucketPermissionAllowed(user2, bucketName, utils.ListObjectAction)
	require.NoError(t, permInfo.Err)
	fmt.Println("permission:", permInfo.EffectInfo)

	statement := utils.NewStatement(utils.AllowEffect, []utils.Action{utils.ListObjectAction})
	policy := utils.GnfdPolicy{
		Statements: []utils.GnfdStatement{statement},
	}
	policyByte, err := policy.MarshalJSON()
	require.NoError(t, err)
	// put bucket policy
	resp = client.PutBucketPolicy(bucketName, string(policyByte), user2, types.TxOption{})
	require.NoError(t, resp.Err)

	time.Sleep(10 * time.Second)
	// verify permission should be ALLOW
	permInfo = client.IsBucketPermissionAllowed(user2, bucketName, utils.ListObjectAction)
	require.NoError(t, permInfo.Err)
	fmt.Println("permission:", permInfo.EffectInfo)
}

// GenRandomBucketName generate random bucket name.
func GenRandomBucketName() string {
	return randString(rand.Intn(10) + 3)
}

var mtx sync.Mutex

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyz01234569"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randString(n int) string {
	mtx.Lock()
	src := rand.NewSource(time.Now().UnixNano())
	time.Sleep(1 * time.Millisecond)
	mtx.Unlock()

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
