package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// it is the example of basic permission SDKs usage
// the storage example need to run before permission examples to make sure the resources has been created
func main() {
	// you need to set the principal address in config.go to run this examples
	if len(principal) < 42 {
		log.Println("please set principal if you need run permission test")
		return
	}
	granteeAddr, err := sdk.AccAddressFromHexUnsafe(principal)
	if err != nil {
		log.Fatalf("principal addr invalid %v", err)
	}

	principalStr, err := utils.NewPrincipalWithAccount(granteeAddr)
	if err != nil {
		log.Fatalf("fail to generate marshaled principal %v", err)
	}

	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}

	// put bucket policy
	bucketActions := []permTypes.ActionType{
		permTypes.ACTION_UPDATE_BUCKET_INFO,
		permTypes.ACTION_DELETE_BUCKET,
	}
	ctx := context.Background()
	statements := utils.NewStatement(bucketActions, permTypes.EFFECT_ALLOW, nil, types.NewStatementOptions{})

	policyTx, err := cli.PutBucketPolicy(ctx, bucketName, principalStr, []*permTypes.Statement{&statements},
		types.PutPolicyOption{})
	handleErr(err, "PutBucketPolicy")
	_, err = cli.WaitForTx(ctx, policyTx)
	if err != nil {
		log.Fatalln("txn fail")
	}
	log.Printf("put bucket %s policy sucessfully, principal is: %s.\n", bucketName, principal)

	// get bucket policy
	policyInfo, err := cli.GetBucketPolicy(ctx, bucketName, principal)
	handleErr(err, "GetBucketPolicy")
	log.Printf("bucket: %s policy info:%s\n", bucketName, policyInfo.String())

	// delete bucket policy
	policyTx, err = cli.DeleteBucketPolicy(ctx, bucketName, principalStr, types.DeletePolicyOption{})
	handleErr(err, "DeleteBucketPolicy")

	// verify permission
	effect, err := cli.IsBucketPermissionAllowed(ctx, principal, bucketName, permTypes.ACTION_DELETE_BUCKET)
	handleErr(err, "IsBucketPermissionAllowed")

	if effect != permTypes.EFFECT_ALLOW {
		log.Fatalln("permission not allowed to:", principal)
	}

	// put object policy
	objectActions := []permTypes.ActionType{
		permTypes.ACTION_DELETE_OBJECT,
		permTypes.ACTION_GET_OBJECT,
	}
	statements = utils.NewStatement(objectActions, permTypes.EFFECT_ALLOW, nil, types.NewStatementOptions{})
	policyTx, err = cli.PutObjectPolicy(ctx, bucketName, objectName, principal, []*permTypes.Statement{&statements},
		types.PutPolicyOption{})
	handleErr(err, "PutObjectPolicy")
	_, err = cli.WaitForTx(ctx, policyTx)
	if err != nil {
		log.Fatalln("txn fail")
	}
	log.Printf("put object: %s policy sucessfully, principal is: %s.\n", objectName, principal)

	// verify permission
	effect, err = cli.IsObjectPermissionAllowed(ctx, principal, bucketName, objectName, permTypes.ACTION_DELETE_OBJECT)
	handleErr(err, "IsObjectPermissionAllowed")

	if effect != permTypes.EFFECT_ALLOW {
		log.Fatalln("permission not allowed to:", principal)
	}

	// delete object permission
	policyTx, err = cli.DeleteObjectPolicy(ctx, bucketName, objectName, principalStr, types.DeletePolicyOption{})
	handleErr(err, "DeleteObjectPolicy")
}
