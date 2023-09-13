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

	log.Printf("add group member: %s to group: %s successfully \n", groupMember, groupName)

	// head group member
	memIsExist := cli.HeadGroupMember(ctx, groupName, creator.GetAddress().String(), groupMember)
	if !memIsExist {
		log.Fatalf("head group member %s fail \n", groupMember)
	}

	log.Printf(" head member %s exist \n", groupMember)

	// list groups
	groups, err := cli.ListGroup(ctx, "e", "t", types.ListGroupsOptions{SourceType: "SOURCE_TYPE_ORIGIN", Limit: 10, Offset: 0, Endpoint: httpsAddr, SPAddress: ""})
	log.Println("list groups result:")
	for _, group := range groups.Groups {
		log.Printf("name: %s, source type: %s\n", group.Group.GroupName, group.Group.SourceType)
	}

	// get group members
	groupMembers, err := cli.ListGroupMembers(ctx, 10, types.GroupMembersPaginationOptions{
		Limit:      10,
		StartAfter: "",
		Endpoint:   httpsAddr,
		SPAddress:  "",
	})
	log.Println("list groups result:")
	for _, group := range groupMembers.Groups {
		log.Printf("name: %s, source type: %s\n", group.Group.GroupName, group.Group.SourceType)
	}

	// get user groups
	userGroups, err := cli.ListGroupsByAccount(ctx, types.GroupsPaginationOptions{
		StartAfter: "",
		Account:    "0x6a45de47a2cd53084b4793fca7c1e706b9f54ed1",
		Endpoint:   httpsAddr,
		SPAddress:  "",
	})
	log.Println("list groups result:")
	for _, group := range userGroups.Groups {
		log.Printf("name: %s, source type: %s\n", group.Group.GroupName, group.Group.SourceType)
	}

	// get group members
	ownedGroups, err := cli.ListGroupsByOwner(ctx, types.GroupsOwnerPaginationOptions{
		StartAfter: "",
		Owner:      "0x6a45de47a2cd53084b4793fca7c1e706b9f54ed1",
		Endpoint:   httpsAddr,
		SPAddress:  "",
	})
	log.Println("list groups result:")
	for _, group := range ownedGroups.Groups {
		log.Printf("name: %s, source type: %s\n", group.Group.GroupName, group.Group.SourceType)
	}

	// delete group
	delTx, err := cli.DeleteGroup(ctx, groupName, types.DeleteGroupOption{})
	handleErr(err, "DeleteGroup")
	_, err = cli.WaitForTx(ctx, delTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	log.Printf("group: %s has been deleted\n", groupName)
}
