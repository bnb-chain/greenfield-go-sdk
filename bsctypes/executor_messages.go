package bsctypes

import "math/big"

type EMessage struct {
	MsgType  uint8
	MsgBytes []byte
}

type ExecutorMessage struct {
	MsgTypes []uint8
	MsgBytes [][]byte
}

type EMessages struct {
	Message    []*EMessage
	Deployment *Deployment
	RelayFee   *big.Int
}
