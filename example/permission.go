package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
)

func testPermission(cli client.Client, bucketName, objectName string) {
	if len(principal) < 42 {
		log.Println("please set principal if you need run permission test")
		return
	}
	// put bucket policy
	bucketActions := []permTypes.ActionType{
		permTypes.ACTION_UPDATE_BUCKET_INFO,
		permTypes.ACTION_DELETE_BUCKET,
	}
	ctx := context.Background()
	statements := utils.NewStatement(bucketActions, permTypes.EFFECT_ALLOW, nil, types.NewStatementOptions{})
	policyTx, err := cli.PutBucketPolicy(ctx, bucketName, principal, []*permTypes.Statement{&statements},
		types.PutPolicyOption{})
	HandleErr(err, "PutBucketPolicy")
	_, err = cli.WaitForTx(ctx, policyTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	// get bucket policy
	policyInfo, err := cli.GetBucketPolicy(ctx, bucketName, principal)
	HandleErr(err, "GetBucketPolicy")
	log.Printf("bucket: %s policy info:%s\n", bucketName, policyInfo.String())

	// verify permission
	effect, err := cli.IsBucketPermissionAllowed(ctx, principal, bucketName, permTypes.ACTION_DELETE_BUCKET)
	HandleErr(err, "IsBucketPermissionAllowed")

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
	HandleErr(err, "PutObjectPolicy")
	_, err = cli.WaitForTx(ctx, policyTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	// verify permission
	effect, err = cli.IsObjectPermissionAllowed(ctx, principal, bucketName, objectName, permTypes.ACTION_DELETE_OBJECT)
	HandleErr(err, "IsObjectPermissionAllowed")

	if effect != permTypes.EFFECT_ALLOW {
		log.Fatalln("permission not allowed to:", principal)
	}
}
