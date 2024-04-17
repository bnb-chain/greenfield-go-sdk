package bsctypes

import (
	"log"
	"math/big"
	"strings"

	permissiontype "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	bsccommon "github.com/bnb-chain/greenfield-go-sdk/common"
)

type IMessages interface {
	CreateBucket(sender *common.Address, synPkg *CreateBucketSynPackage) *Messages
	CreateBucketCallBack(sender *common.Address, synPkg *CreateBucketSynPackage, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages
	DeleteBucket(sender *common.Address, id *big.Int) *Messages
	DeleteBucketCallBack(sender *common.Address, id *big.Int, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages
	DeleteObject(sender *common.Address, id *big.Int) *Messages
	DeleteObjectCallBack(sender *common.Address, id *big.Int, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages
	CreateGroup(sender *common.Address, owner *common.Address, name string) *Messages
	CreateGroupCallBack(sender *common.Address, owner *common.Address, name string, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages
	DeleteGroup(sender *common.Address, id *big.Int) *Messages
	DeleteGroupCallBack(sender *common.Address, id *big.Int, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages
	UpdateGroup(sender *common.Address, synPkg *UpdateGroupMemberSynPackage) *Messages
	UpdateGroupCallBack(sender *common.Address, synPkg *UpdateGroupMemberSynPackage, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages
	CreatePolicy(sender *common.Address, policy *permissiontype.Policy) *Messages
	CreatePolicyCallBack(sender *common.Address, policy *permissiontype.Policy, extraData *ExtraData, opt *RelayFeeOption) *Messages
	DeletePolicy(sender *common.Address, id *big.Int) *Messages
	DeletePolicyCallBack(sender *common.Address, id *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages
	TransferOut(sender *common.Address, recipient *common.Address, amount *big.Int) *Messages
}

type Message struct {
	Target *common.Address
	Data   []byte
	Value  *big.Int
}

type MultiMessage struct {
	Targets []common.Address
	Data    [][]byte
	Values  []*big.Int
}

type Messages struct {
	Message          []*Message
	Deployment       *Deployment
	RelayFee         *big.Int
	MinAckRelayFee   *big.Int
	CallbackGasPrice *big.Int
}

func NewMessages(deployment *Deployment, relayFee *big.Int, minAckRelayFee *big.Int, callbackGasPrice *big.Int) *Messages {
	return &Messages{
		Message:          []*Message{},
		Deployment:       deployment,
		RelayFee:         relayFee,
		MinAckRelayFee:   minAckRelayFee,
		CallbackGasPrice: callbackGasPrice,
	}
}

func (m *Messages) Build() *MultiMessage {
	targets := make([]common.Address, len(m.Message))
	data := make([][]byte, len(m.Message))
	values := make([]*big.Int, len(m.Message))
	for i, message := range m.Message {
		targets[i] = *message.Target
		data[i] = message.Data
		values[i] = message.Value
	}
	return &MultiMessage{
		Targets: targets,
		Data:    data,
		Values:  values,
	}
}

func (m *Messages) CreateBucket(sender *common.Address, synPkg *CreateBucketSynPackage) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.BucketHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.BucketABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareCreateBucket", sender, synPkg)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) CreateBucketCallBack(sender *common.Address, synPkg *CreateBucketSynPackage, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		if opt.AckRelayFee.Cmp(m.MinAckRelayFee) < 0 {
			log.Fatalf("opt.AckRelayFee can't be smaller than MinAckRelayFee")
		}
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)

	address := common.HexToAddress(m.Deployment.BucketHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.BucketABI))
	if err != nil {
		log.Fatalf("failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareCreateBucket0", sender, synPkg, callbackGasLimit, extraData)
	if err != nil {
		log.Fatalf("failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeleteBucket(sender *common.Address, id *big.Int) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.BucketHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.BucketABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeleteBucket", sender, id)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeleteBucketCallBack(sender *common.Address, id *big.Int, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		if opt.AckRelayFee.Cmp(m.MinAckRelayFee) < 0 {
			log.Fatalf("opt.AckRelayFee can't be smaller than MinAckRelayFee")
		}
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)

	address := common.HexToAddress(m.Deployment.BucketHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.BucketABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeleteBucket0", sender, id, callbackGasLimit, extraData)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeleteObject(sender *common.Address, id *big.Int) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.ObjectHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.ObjectABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeleteObject", sender, id)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeleteObjectCallBack(sender *common.Address, id *big.Int, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		if opt.AckRelayFee.Cmp(m.MinAckRelayFee) < 0 {
			log.Fatalf("opt.AckRelayFee can't be smaller than MinAckRelayFee")
		}
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)

	address := common.HexToAddress(m.Deployment.ObjectHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.ObjectABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeleteObject0", sender, id, callbackGasLimit, extraData)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) CreateGroup(sender *common.Address, owner *common.Address, name string) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.GroupHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.GroupABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareCreateGroup0", sender, owner, name)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) GetCallBackFee() {

}

func (m *Messages) CreateGroupCallBack(sender *common.Address, owner *common.Address, name string, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	// fee = relayFee + minAckRelayFee + callbackGasLimit * callbackGasPrice
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)
	if callbackGasLimit != nil {
		callbackGasCost := new(big.Int).Mul(callbackGasLimit, m.CallbackGasPrice)
		fee.Add(fee, callbackGasCost)
	}

	address := common.HexToAddress(m.Deployment.GroupHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.GroupABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareCreateGroup", sender, owner, name, callbackGasLimit, extraData)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeleteGroup(sender *common.Address, id *big.Int) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.GroupHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.GroupABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeleteGroup", sender, id)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeleteGroupCallBack(sender *common.Address, id *big.Int, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		if opt.AckRelayFee.Cmp(m.MinAckRelayFee) < 0 {
			log.Fatalf("opt.AckRelayFee can't be smaller than MinAckRelayFee")
		}
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)

	address := common.HexToAddress(m.Deployment.GroupHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.GroupABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeleteGroup0", sender, id, callbackGasLimit, extraData)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) UpdateGroup(sender *common.Address, synPkg *UpdateGroupMemberSynPackage) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.GroupHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.GroupABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareUpdateGroup0", sender, synPkg)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) UpdateGroupCallBack(sender *common.Address, synPkg *UpdateGroupMemberSynPackage, callbackGasLimit *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		if opt.AckRelayFee.Cmp(m.MinAckRelayFee) < 0 {
			log.Fatalf("opt.AckRelayFee can't be smaller than MinAckRelayFee")
		}
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)

	address := common.HexToAddress(m.Deployment.GroupHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.GroupABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareUpdateGroup", sender, synPkg, callbackGasLimit, extraData)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) CreatePolicy(sender *common.Address, policy *permissiontype.Policy) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.PermissionHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.PermissionABI))
	if err != nil {
		log.Fatalf("failed to parse contract ABI: %v", err)
	}

	data, err := policy.Marshal()
	if err != nil {
		log.Fatalf("failed to marshal policy: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareCreatePolicy", sender, data)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) CreatePolicyCallBack(sender *common.Address, policy *permissiontype.Policy, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		if opt.AckRelayFee.Cmp(m.MinAckRelayFee) < 0 {
			log.Fatalf("opt.AckRelayFee can't be smaller than MinAckRelayFee")
		}
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)

	address := common.HexToAddress(m.Deployment.PermissionHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.PermissionABI))
	if err != nil {
		log.Fatalf("failed to parse contract ABI: %v", err)
	}

	data, err := policy.Marshal()
	if err != nil {
		log.Fatalf("failed to marshal policy: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareCreatePolicy0", sender, data, extraData)
	if err != nil {
		log.Fatalf("failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeletePolicy(sender *common.Address, id *big.Int) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)

	address := common.HexToAddress(m.Deployment.PermissionHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.PermissionABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeletePolicy", sender, id)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) DeletePolicyCallBack(sender *common.Address, id *big.Int, extraData *ExtraData, opt *RelayFeeOption) *Messages {
	fee := new(big.Int)
	ackFee := m.MinAckRelayFee
	if opt != nil && opt.AckRelayFee != nil {
		if opt.AckRelayFee.Cmp(m.MinAckRelayFee) < 0 {
			log.Fatalf("opt.AckRelayFee can't be smaller than MinAckRelayFee")
		}
		ackFee = opt.AckRelayFee
	}
	fee.Add(m.RelayFee, ackFee)

	address := common.HexToAddress(m.Deployment.PermissionHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.PermissionABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareDeletePolicy0", sender, id, extraData)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}

func (m *Messages) TransferOut(sender *common.Address, recipient *common.Address, amount *big.Int) *Messages {
	fee := new(big.Int)
	fee.Add(m.RelayFee, m.MinAckRelayFee)
	fee.Add(fee, amount)
	address := common.HexToAddress(m.Deployment.TokenHub)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.TokenABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("prepareTransferOut", sender, recipient, amount)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	message := &Message{
		Target: &address,
		Data:   packedData,
		Value:  fee,
	}
	m.Message = append(m.Message, message)
	return m
}
