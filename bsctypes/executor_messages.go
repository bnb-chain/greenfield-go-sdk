package bsctypes

import (
	"log"
	"math/big"

	"github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/gogoproto/proto"
)

type IExecutorBatchedMessage interface {
	CreatePaymentAccount(msg *types.MsgCreatePaymentAccount) *ExecutorBatchedMessage
	Deposit(msg *types.MsgDeposit) *ExecutorBatchedMessage
	DisableRefund(msg *types.MsgDisableRefund) *ExecutorBatchedMessage
	Withdraw(msg *types.MsgWithdraw) *ExecutorBatchedMessage
	MigrateBucket(msg *storagetypes.MsgMigrateBucket) *ExecutorBatchedMessage
	CancelMigrateBucket(msg *storagetypes.MsgCancelMigrateBucket) *ExecutorBatchedMessage
	UpdateBucketInfo(msg *storagetypes.MsgUpdateBucketInfo) *ExecutorBatchedMessage
	ToggleSPAsDelegatedAgent(msg *storagetypes.MsgToggleSPAsDelegatedAgent) *ExecutorBatchedMessage
	SetBucketFlowRateLimit(msg *storagetypes.MsgSetBucketFlowRateLimit) *ExecutorBatchedMessage
	CopyObject(msg *storagetypes.MsgCopyObject) *ExecutorBatchedMessage
	UpdateObjectInfo(msg *storagetypes.MsgUpdateObjectInfo) *ExecutorBatchedMessage
	UpdateGroupExtra(msg *storagetypes.MsgUpdateGroupExtra) *ExecutorBatchedMessage
	SetTag(msg *storagetypes.MsgSetTag) *ExecutorBatchedMessage
}

type ExecutorMessageUnit struct {
	MsgType  uint8
	MsgBytes []byte
}

type ExecutorBatchedMessage struct {
	Message        []*ExecutorMessageUnit
	Deployment     *Deployment
	RelayFee       *big.Int
	MinAckRelayFee *big.Int
}

// ExecutorMessages
/**
* Supported message types and its corresponding number
* 1: CreatePaymentAccount
* 2: Deposit
* 3: DisableRefund
* 4: Withdraw
* 5: MigrateBucket
* 6: CancelMigrateBucket
* 7: UpdateBucketInfo
* 8: ToggleSPAsDelegatedAgent
* 9: SetBucketFlowRateLimit
* 10: CopyObject
* 11: UpdateObjectInfo
* 12: UpdateGroupExtra
* 13: SetTag
 */
type ExecutorMessages struct {
	MsgTypes []uint8
	MsgBytes [][]byte
	RelayFee *big.Int
}

func NewExecutorBatchedMessage(deployment *Deployment, relayFee *big.Int, minAckRelayFee *big.Int) *ExecutorBatchedMessage {
	return &ExecutorBatchedMessage{
		Message:        []*ExecutorMessageUnit{},
		Deployment:     deployment,
		RelayFee:       relayFee,
		MinAckRelayFee: minAckRelayFee,
	}
}

func (e *ExecutorBatchedMessage) Build() *ExecutorMessages {
	msgTypes := make([]uint8, len(e.Message))
	msgBytes := make([][]byte, len(e.Message))
	for i, message := range e.Message {
		msgTypes[i] = message.MsgType
		msgBytes[i] = message.MsgBytes
	}
	return &ExecutorMessages{
		MsgTypes: msgTypes,
		MsgBytes: msgBytes,
		RelayFee: e.RelayFee,
	}
}

func (e *ExecutorBatchedMessage) appendMessage(msg proto.Message, msgType uint8) *ExecutorBatchedMessage {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalf("failed to marshal message: %v", err)
	}
	message := &ExecutorMessageUnit{
		MsgType:  msgType,
		MsgBytes: msgBytes,
	}
	e.Message = append(e.Message, message)
	return e
}

func (e *ExecutorBatchedMessage) CreatePaymentAccount(msg *types.MsgCreatePaymentAccount) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 1)
}

func (e *ExecutorBatchedMessage) Deposit(msg *types.MsgDeposit) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 2)
}

func (e *ExecutorBatchedMessage) DisableRefund(msg *types.MsgDisableRefund) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 3)
}

func (e *ExecutorBatchedMessage) Withdraw(msg *types.MsgWithdraw) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 4)
}

func (e *ExecutorBatchedMessage) MigrateBucket(msg *storagetypes.MsgMigrateBucket) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 5)
}

func (e *ExecutorBatchedMessage) CancelMigrateBucket(msg *storagetypes.MsgCancelMigrateBucket) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 6)
}

func (e *ExecutorBatchedMessage) UpdateBucketInfo(msg *storagetypes.MsgUpdateBucketInfo) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 7)
}

func (e *ExecutorBatchedMessage) ToggleSPAsDelegatedAgent(msg *storagetypes.MsgToggleSPAsDelegatedAgent) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 8)
}

func (e *ExecutorBatchedMessage) SetBucketFlowRateLimit(msg *storagetypes.MsgSetBucketFlowRateLimit) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 9)
}

func (e *ExecutorBatchedMessage) CopyObject(msg *storagetypes.MsgCopyObject) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 10)
}

func (e *ExecutorBatchedMessage) UpdateObjectInfo(msg *storagetypes.MsgUpdateObjectInfo) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 11)
}

func (e *ExecutorBatchedMessage) UpdateGroupExtra(msg *storagetypes.MsgUpdateGroupExtra) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 12)
}

func (e *ExecutorBatchedMessage) SetTag(msg *storagetypes.MsgSetTag) *ExecutorBatchedMessage {
	return e.appendMessage(msg, 13)
}
