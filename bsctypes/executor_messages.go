package bsctypes

import (
	"log"
	"math/big"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

type IExecutorBatchedMessage interface {
	CreatePaymentAccount(msg types.MsgCreatePaymentAccount) *ExecutorBatchedMessage
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

func (e *ExecutorBatchedMessage) CreatePaymentAccount(msg *types.MsgCreatePaymentAccount) *ExecutorBatchedMessage {
	msgBytes, err := msg.Marshal()
	if err != nil {
		log.Fatalf("failed to marshal policy: %v", err)
	}
	message := &ExecutorMessageUnit{
		MsgType:  1,
		MsgBytes: msgBytes,
	}
	e.Message = append(e.Message, message)
	return e
}
