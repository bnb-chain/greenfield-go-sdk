package client

import (
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield/sdk/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/account"
	sdkerror "github.com/bnb-chain/greenfield-go-sdk/pkg/error"
)

type BasicOption struct {
	TxOption types.TxOption
	Account  account.Account
}

type Principal string

// CreateBucketOptions indicates the meta to construct createBucket msg of storage module
type CreateBucketOptions struct {
	BasicOption
	Visibility       storageTypes.VisibilityType
	TxOpts           *types.TxOption
	PaymentAddress   sdk.AccAddress
	PrimarySPAddress sdk.AccAddress
	ChargedQuota     uint64
}

// CreateObjectOptions indicates the metadata to construct `createObject` message of storage module
type CreateObjectOptions struct {
	BasicOption
	Visibility      storageTypes.VisibilityType
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

// ListReadRecordOptions indicates the start timestamp of the return read quota record
type ListReadRecordOptions struct {
	StartTimeStamp int64
	MaxRecords     int
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
		return sdkerror.ToInvalidArgumentResp(
			fmt.Sprintf(
				"Invalid Range : start=%d end=%d",
				start, end))
	}
	return nil
}
