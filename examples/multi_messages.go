package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/greenfield-go-sdk/bsc"
	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
)

func main() {
	account, err := bsctypes.NewBscAccountFromPrivateKey("test", bscPrivateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	client, err := bsc.New(bscRpcAddr, "qa-net", bsc.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new bsc client, %v", err)
	}

	relayFee, minAckRelayFee, err := client.GetMinAckRelayFee(context.Background())
	if err != nil {
		log.Fatalf("unable to get min ack relay fee, %v", err)
	}

	gasPrice, err := client.GetCallbackGasPrice(context.Background())
	if err != nil {
		log.Fatalf("unable to get min ack relay fee, %v", err)
	}

	// fluent interface example, executing transactions in the order specified by the user.
	messages := bsctypes.NewMessages(client.GetDeployment(), relayFee, minAckRelayFee, gasPrice)
	primarySpAddress := common.HexToAddress("0xd142052d8c0881fc7485c1270c3510bc442e05dd")
	_ = messages.CreateBucket(account.GetAddress(), &bsctypes.CreateBucketSynPackage{
		Creator:                    account.GetAddress(),
		Name:                       "cross-chain-test-1",
		Visibility:                 1,
		PaymentAddress:             account.GetAddress(),
		PrimarySpAddress:           &primarySpAddress,
		GlobalVirtualGroupFamilyId: 1,
	}).CreatePolicy(account.GetAddress(), &types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: "0x0C02787e83948e7aD29abE3a99b29c480f9F0096",
		},
		ResourceType: resource.RESOURCE_TYPE_BUCKET,
		ResourceId:   math.NewUint(179373),
		Statements: []*types.Statement{
			&types.Statement{
				Effect:  types.EFFECT_ALLOW,
				Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
			},
		},
	}).CreateGroup(account.GetAddress(), account.GetAddress(), "test-group-1")

	// transfer out money from bsc to greenfield
	_ = messages.TransferOut(account.GetAddress(), account.GetAddress(), big.NewInt(1e16))

	// Create policy
	_ = messages.CreatePolicy(account.GetAddress(), &types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: "0x0C02787e83948e7aD29abE3a99b29c480f9F0096",
		},
		ResourceType: resource.RESOURCE_TYPE_BUCKET,
		ResourceId:   math.NewUint(179373),
		Statements: []*types.Statement{
			&types.Statement{
				Effect:  types.EFFECT_ALLOW,
				Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
			},
		},
	})

	_ = messages.CreatePolicyCallBack(
		account.GetAddress(),
		&types.Policy{
			Principal: &types.Principal{
				Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
				Value: "0x2b1A0Ba5a484ef9375DF0A908096f1da6d58d86a",
			},
			ResourceType: resource.RESOURCE_TYPE_BUCKET,
			ResourceId:   math.NewUint(179373),
			Statements: []*types.Statement{
				&types.Statement{
					Effect:  types.EFFECT_ALLOW,
					Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
				},
			},
		},
		&bsctypes.ExtraData{
			AppAddress:            account.GetAddress(),
			RefundAddress:         account.GetAddress(),
			FailureHandleStrategy: 2,
		}, nil)

	//Delete Policy
	_ = messages.DeletePolicy(account.GetAddress(), big.NewInt(442))

	_ = messages.DeletePolicyCallBack(account.GetAddress(), big.NewInt(447),
		&bsctypes.ExtraData{
			AppAddress:            account.GetAddress(),
			RefundAddress:         account.GetAddress(),
			FailureHandleStrategy: 2,
		}, nil)

	// Crete group
	_ = messages.CreateGroup(account.GetAddress(), account.GetAddress(), "test-policy")
	_ = messages.CreateGroupCallBack(account.GetAddress(), account.GetAddress(), "test-callback", big.NewInt(1000000), &bsctypes.ExtraData{
		AppAddress:            account.GetAddress(),
		RefundAddress:         account.GetAddress(),
		FailureHandleStrategy: bsctypes.SkipOnFail,
	}, nil)

	// update group
	_ = messages.UpdateGroup(account.GetAddress(), &bsctypes.UpdateGroupMemberSynPackage{
		Operator:         account.GetAddress(),
		Id:               big.NewInt(398),
		OpType:           bsctypes.AddMembers,
		Members:          []common.Address{common.HexToAddress("0xcdffb8d585d954dfb7b00c76af4f173c6a8e3549")},
		MemberExpiration: []uint64{1733412997},
	})

	_ = messages.UpdateGroupCallBack(account.GetAddress(), &bsctypes.UpdateGroupMemberSynPackage{
		Operator:         account.GetAddress(),
		Id:               big.NewInt(398),
		OpType:           bsctypes.AddMembers,
		Members:          []common.Address{common.HexToAddress("0x5fa8b3f3fcd4a3ea2495e11dd5dbd399b3d8d4f8")},
		MemberExpiration: []uint64{1733412997},
	}, big.NewInt(1000000),
		&bsctypes.ExtraData{
			AppAddress:            account.GetAddress(),
			RefundAddress:         account.GetAddress(),
			FailureHandleStrategy: 0,
		}, nil)

	// Delete group
	_ = messages.DeleteGroup(account.GetAddress(), big.NewInt(398))

	_ = messages.DeleteGroupCallBack(account.GetAddress(), big.NewInt(404), big.NewInt(1000000),
		&bsctypes.ExtraData{
			AppAddress:            account.GetAddress(),
			RefundAddress:         account.GetAddress(),
			FailureHandleStrategy: 0,
		}, nil)

	// create bucket
	_ = messages.CreateBucket(account.GetAddress(), &bsctypes.CreateBucketSynPackage{
		Creator:                    account.GetAddress(),
		Name:                       "cross-chain-test-4",
		Visibility:                 1,
		PaymentAddress:             account.GetAddress(),
		PrimarySpAddress:           &primarySpAddress,
		GlobalVirtualGroupFamilyId: 1,
	})

	_ = messages.CreateBucketCallBack(
		account.GetAddress(),
		&bsctypes.CreateBucketSynPackage{
			Creator:                    account.GetAddress(),
			Name:                       "cross-chain-test-2",
			Visibility:                 1,
			PaymentAddress:             account.GetAddress(),
			PrimarySpAddress:           &primarySpAddress,
			GlobalVirtualGroupFamilyId: 1,
		},
		big.NewInt(1000000),
		&bsctypes.ExtraData{
			AppAddress:            account.GetAddress(),
			RefundAddress:         account.GetAddress(),
			FailureHandleStrategy: 0,
		}, nil)

	//delete bucket
	_ = messages.DeleteBucket(account.GetAddress(), big.NewInt(179370))

	_ = messages.DeleteBucketCallBack(account.GetAddress(), big.NewInt(179372), big.NewInt(100000),
		&bsctypes.ExtraData{
			AppAddress:            account.GetAddress(),
			RefundAddress:         account.GetAddress(),
			FailureHandleStrategy: 0,
		}, nil)

	// delele object
	_ = messages.DeleteObject(account.GetAddress(), big.NewInt(422460))

	_ = messages.DeleteObjectCallBack(account.GetAddress(), big.NewInt(422445), big.NewInt(1000000),
		&bsctypes.ExtraData{
			AppAddress:            account.GetAddress(),
			RefundAddress:         account.GetAddress(),
			FailureHandleStrategy: 0,
			CallbackData:          []byte{'1'},
		}, nil)

	tx, err := client.SendMessages(context.Background(), messages.Build())
	if err != nil {
		log.Fatalf("unable to send messages, %v", err)
	}

	time.Sleep(3 * time.Second)

	success, err := client.CheckTxStatus(context.Background(), tx)
	if err != nil {
		log.Fatalf("unable to check tx status, %v", err)
	}

	if success {
		log.Println("successfully sent the tx")
	} else {
		log.Println("failed to send the tx")
	}
}
