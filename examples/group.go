package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// it is the example of basic group SDKs usage
func main() {
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()

	// create group
	groupTx, err := cli.CreateGroup(ctx, groupName, types.CreateGroupOptions{})
	handleErr(err, "CreateGroup")
	_, err = cli.WaitForTx(ctx, groupTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	log.Printf("create group %s successfully \n", groupName)

	// head group
	creator, err := cli.GetDefaultAccount()
	handleErr(err, "GetDefaultAccount")
	groupInfo, err := cli.HeadGroup(ctx, groupName, creator.GetAddress().String())
	handleErr(err, "HeadGroup")
	log.Println("head group info:", groupInfo.String())

	_, err = sdk.AccAddressFromHexUnsafe(groupMember)
	if err != nil {
		log.Fatalln("the group member is invalid")
	}
	// update group member
	updateTx, err := cli.UpdateGroupMember(ctx, groupName, creator.GetAddress().String(), []string{groupMember}, []string{},
		types.UpdateGroupMemberOption{})
	handleErr(err, "UpdateGroupMember")
	_, err = cli.WaitForTx(ctx, updateTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	// head group member
	memIsExist := cli.HeadGroupMember(ctx, groupName, creator.GetAddress().String(), groupMember)
	if !memIsExist {
		log.Fatalf("head group member %s fail \n", groupMember)
	}

	// delete group
	delTx, err := cli.DeleteGroup(ctx, groupName, types.DeleteGroupOption{})
	handleErr(err, "DeleteGroup")
	_, err = cli.WaitForTx(ctx, delTx)
	if err != nil {
		log.Fatalln("txn fail")
	}
}
