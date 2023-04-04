package types

import (
	"fmt"
	"time"

	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateBucketOptions indicates the meta to construct createBucket msg of storage module
type CreateBucketOptions struct {
	Visibility       storageTypes.VisibilityType
	TxOpts           *gnfdsdktypes.TxOption
	PaymentAddress   sdk.AccAddress
	PrimarySPAddress sdk.AccAddress
	ChargedQuota     uint64
}

type DeleteBucketOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type UpdatePaymentOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type UpdateBucketOption struct {
	Visibility     storageTypes.VisibilityType
	TxOpts         *gnfdsdktypes.TxOption
	PaymentAddress sdk.AccAddress
	ChargedQuota   *uint64
}

type CancelCreateOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type BuyQuotaOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type UpdateVisibilityOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type DeleteObjectOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type DeleteGroupOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

// CreateObjectOptions indicates the metadata to construct `createObject` message of storage module
type CreateObjectOptions struct {
	Visibility      storageTypes.VisibilityType
	TxOpts          *gnfdsdktypes.TxOption
	SecondarySPAccs []sdk.AccAddress
	ContentType     string
	IsReplicaType   bool // indicates whether the object use REDUNDANCY_REPLICA_TYPE
}

// CreateGroupOptions  indicates the meta to construct createGroup msg
type CreateGroupOptions struct {
	InitGroupMember []sdk.AccAddress
	TxOpts          *gnfdsdktypes.TxOption
}

// UpdateGroupMemberOption indicates the info to update group member
type UpdateGroupMemberOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type LeaveGroupOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

// ComputeHashOptions indicates the metadata of redundancy strategy
type ComputeHashOptions struct {
	SegmentSize  uint64
	DataShards   uint32
	ParityShards uint32
}

// ListReadRecordOptions indicates the start timestamp of the return read quota record
type ListReadRecordOptions struct {
	StartTimeStamp int64
	MaxRecords     int
}

type PutPolicyOption struct {
	TxOpts           *gnfdsdktypes.TxOption
	PolicyExpireTime *time.Time
}

type DeletePolicyOption struct {
	TxOpts *gnfdsdktypes.TxOption
}

type NewStatementOptions struct {
	StatementExpireTime *time.Time
	LimitSize           uint64
}

type ApproveBucketOptions struct {
	IsPublic       bool
	PaymentAddress sdk.AccAddress
}

type ApproveObjectOptions struct {
	IsPublic        bool
	SecondarySPAccs []sdk.AccAddress
}

type PutObjectOption struct {
	ContentType string
}

// GetObjectOption contains the options of getObject
type GetObjectOption struct {
	Range string `url:"-" header:"Range,omitempty"` // support for downloading partial data
}

func (o *GetObjectOption) SetRange(start, end int64) error {
	switch {
	case 0 < start && end == 0:
		// `bytes=N-`.
		o.Range = fmt.Sprintf("bytes=%d-", start)
	case 0 <= start && start <= end:
		// `bytes=N-M`
		o.Range = fmt.Sprintf("bytes=%d-%d", start, end)
	default:
		return ToInvalidArgumentResp(
			fmt.Sprintf(
				"Invalid Range : start=%d end=%d",
				start, end))
	}
	return nil
}
