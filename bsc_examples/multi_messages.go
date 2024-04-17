package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/bsc"
	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
)

func main() {
	account, err := bsctypes.NewBscAccountFromPrivateKey("barry", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	client, err := bsc.New(rpcAddr, "qa-net", bsc.Option{DefaultAccount: account})
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

	messages := bsctypes.NewMessages(client.GetDeployment(), relayFee, minAckRelayFee, gasPrice)

	// success case
	// Create Group
	//_ = messages.CreateGroup(account.GetAddress(), account.GetAddress(), "barry-test")
	//_ = messages.CreateGroupCallBack(account.GetAddress(), account.GetAddress(), "barry-test-callback", big.NewInt(1000000), &bsctypes.ExtraData{
	//	AppAddress:            account.GetAddress(),
	//	RefundAddress:         account.GetAddress(),
	//	FailureHandleStrategy: bsctypes.SkipOnFail,
	//	CallbackData:          []byte{'t'},
	//}, nil)

	//paymentAddress := common.HexToAddress("0xd4a205968691416b982685602d663567ede960d5")
	//	primarySpAddress := common.HexToAddress("0xd4a205968691416b982685602d663567ede960d5")
	//	_ = messages.CreateBucket(account.GetAddress(), &bsctypes.CreateBucketSynPackage{
	//		Creator:                        account.GetAddress(),
	//		Name:                           "barry-cross-chain-test",
	//		Visibility:                     0,
	//		PaymentAddress:                 &paymentAddress,
	//		PrimarySpAddress:               &primarySpAddress,
	//		PrimarySpApprovalExpiredHeight: 0,
	//		PrimarySpSignature:             []byte{'t'},
	//		ChargedReadQuota:               0,
	//		ExtraData:                      []byte{'t'},
	//	})

	//paymentAddress := common.HexToAddress("0xd4a205968691416b982685602d663567ede960d5")
	//primarySpAddress := common.HexToAddress("0xd4a205968691416b982685602d663567ede960d5")
	//_ = messages.CreateBucketCallBack(
	//	account.GetAddress(),
	//	&bsctypes.CreateBucketSynPackage{
	//		Creator:                        account.GetAddress(),
	//		Name:                           "barry-cross-chain-test",
	//		Visibility:                     0,
	//		PaymentAddress:                 &paymentAddress,
	//		PrimarySpAddress:               &primarySpAddress,
	//		PrimarySpApprovalExpiredHeight: 0,
	//		PrimarySpSignature:             []byte{'t'},
	//		ChargedReadQuota:               0,
	//		ExtraData:                      []byte{'t'},
	//	}, big.NewInt(1000000),
	//	&bsctypes.ExtraData{
	//		AppAddress:            account.GetAddress(),
	//		RefundAddress:         account.GetAddress(),
	//		FailureHandleStrategy: 0,
	//		CallbackData:          []byte{'1'},
	//	}, nil)

	//_ = messages.DeleteBucket(account.GetAddress(), big.NewInt(1111))

	//	_ = messages.DeleteBucketCallBack(account.GetAddress(), big.NewInt(1111), big.NewInt(1000000),
	//		&bsctypes.ExtraData{
	//			AppAddress:            account.GetAddress(),
	//			RefundAddress:         account.GetAddress(),
	//			FailureHandleStrategy: 0,
	//			CallbackData:          []byte{'1'},
	//		}, nil)

	//_ = messages.DeleteObject(account.GetAddress(), big.NewInt(1111))
	//
	//_ = messages.DeleteObjectCallBack(account.GetAddress(), big.NewInt(1111), big.NewInt(1000000),
	//	&bsctypes.ExtraData{
	//		AppAddress:            account.GetAddress(),
	//		RefundAddress:         account.GetAddress(),
	//		FailureHandleStrategy: 0,
	//		CallbackData:          []byte{'1'},
	//	}, nil)

	//_ = messages.DeleteGroup(account.GetAddress(), big.NewInt(1111))

	//_ = messages.DeleteGroupCallBack(account.GetAddress(), big.NewInt(1111), big.NewInt(1000000),
	//	&bsctypes.ExtraData{
	//		AppAddress:            account.GetAddress(),
	//		RefundAddress:         account.GetAddress(),
	//		FailureHandleStrategy: 0,
	//		CallbackData:          []byte{'1'},
	//	}, nil)

	//_ = messages.UpdateGroup(account.GetAddress(), &bsctypes.UpdateGroupMemberSynPackage{
	//	Operator:         account.GetAddress(),
	//	Id:               big.NewInt(1),
	//	OpType:           bsctypes.AddMembers,
	//	Members:          []common.Address{*account.GetAddress()},
	//	ExtraData:        []byte{'1'},
	//	MemberExpiration: []uint64{1111111111},
	//})

	//_ = messages.UpdateGroupCallBack(account.GetAddress(), &bsctypes.UpdateGroupMemberSynPackage{
	//	Operator:         account.GetAddress(),
	//	Id:               big.NewInt(1),
	//	OpType:           bsctypes.AddMembers,
	//	Members:          []common.Address{*account.GetAddress()},
	//	ExtraData:        []byte{'1'},
	//	MemberExpiration: []uint64{1111111111},
	//}, big.NewInt(1000000),
	//	&bsctypes.ExtraData{
	//		AppAddress:            account.GetAddress(),
	//		RefundAddress:         account.GetAddress(),
	//		FailureHandleStrategy: 0,
	//		CallbackData:          []byte{'1'},
	//	}, nil)

	//_ = messages.CreatePolicy(account.GetAddress(), &types.Policy{
	//	Id:             types.Uint{},
	//	Principal:      nil,
	//	ResourceType:   0,
	//	ResourceId:     types.Uint{},
	//	Statements:     nil,
	//	ExpirationTime: nil,
	//})
	//
	//_ = messages.CreatePolicyCallBack(
	//	account.GetAddress(),
	//	&types.Policy{
	//		Id:             types.Uint{},
	//		Principal:      nil,
	//		ResourceType:   0,
	//		ResourceId:     types.Uint{},
	//		Statements:     nil,
	//		ExpirationTime: nil,
	//	},
	//	&bsctypes.ExtraData{
	//		AppAddress:            account.GetAddress(),
	//		RefundAddress:         account.GetAddress(),
	//		FailureHandleStrategy: 0,
	//		CallbackData:          []byte{'1'},
	//	}, &bsctypes.RelayFeeOption{AckRelayFee: minAckRelayFee})

	//_ = messages.DeletePolicy(account.GetAddress(), big.NewInt(1111))

	//_ = messages.DeletePolicyCallBack(account.GetAddress(), big.NewInt(1111),
	//	&bsctypes.ExtraData{
	//		AppAddress:            account.GetAddress(),
	//		RefundAddress:         account.GetAddress(),
	//		FailureHandleStrategy: 0,
	//		CallbackData:          []byte{'1'},
	//	}, nil)

	// success test cases
	//to := common.HexToAddress("0xe0523429ea945ced7bd3b170fd8dbe797778049b")
	//_ = messages.TransferOut(account.GetAddress(), &to, big.NewInt(1111))

	tx, err := client.SendMessages(context.Background(), messages.Build())
	if err != nil {
		log.Fatalf("unable to send messages, %v", err)
	}

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
