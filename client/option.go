package client

import (
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/account"
	"github.com/bnb-chain/greenfield/sdk/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BasicOption struct {
	TxOption types.TxOption
	Account  account.Account
}

type Principal string

// CreateBucketOptions indicates the meta to construct createBucket msg of storage module
type CreateBucketOptions struct {
	BasicOption
	Visibility       *storageTypes.VisibilityType
	TxOpts           *types.TxOption
	PaymentAddress   sdk.AccAddress
	PrimarySPAddress sdk.AccAddress
	ChargedQuota     uint64
}

// CreateObjectOptions indicates the metadata to construct `createObject` message of storage module
type CreateObjectOptions struct {
	BasicOption
	Visibility      *storageTypes.VisibilityType
	TxOpts          *types.TxOption
	SecondarySPAccs []sdk.AccAddress
	ContentType     string
	IsReplicaType   bool // indicates whether the object use REDUNDANCY_REPLICA_TYPE
}

type DeleteObjectOption struct {
	TxOpts *types.TxOption
}

type DeleteBucketOption struct {
	TxOpts *types.TxOption
}

type CancelCreateOption struct {
	TxOpts *types.TxOption
}

type BuyQuotaOption struct {
	TxOpts *types.TxOption
}

type UpdateVisibilityOption struct {
	TxOpts *types.TxOption
}

// CreateGroupOptions  indicates the meta to construct createGroup msg
type CreateGroupOptions struct {
	InitGroupMember []sdk.AccAddress
	TxOpts          *types.TxOption
}

// UpdateGroupMemberOption indicates the info to update group member
type UpdateGroupMemberOption struct {
	TxOpts *types.TxOption
}

type LeaveGroupOption struct {
	TxOpts *types.TxOption
}

// ComputeHashOptions indicates the metadata of redundancy strategy
type ComputeHashOptions struct {
	SegmentSize  uint64
	DataShards   uint32
	ParityShards uint32
}

// ListReadRecordOption indicates the start timestamp of the return read quota record
type ListReadRecordOption struct {
	StartTimeStamp int64
}

type PutPolicyOption struct {
	TxOpts           *types.TxOption
	PolicyExpireTime *time.Time
}

type DeletePolicyOption struct {
	TxOpts *types.TxOption
}

type NewStatementOptions struct {
	StatementExpireTime *time.Time
	LimitSize           uint64
}
