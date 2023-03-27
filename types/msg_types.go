package types

import (
	"github.com/bnb-chain/greenfield/sdk/types"
)

type (
	MsgGrant  = types.MsgGrant
	MsgRevoke = types.MsgRevoke

	MsgSend = types.MsgSend

	MsgCreateValidator           = types.MsgCreateValidator
	MsgEditValidator             = types.MsgEditValidator
	MsgDelegate                  = types.MsgDelegate
	MsgBeginRedelegate           = types.MsgBeginRedelegate
	MsgUndelegate                = types.MsgUndelegate
	MsgCancelUnbondingDelegation = types.MsgCancelUnbondingDelegation

	MsgSetWithdrawAddress          = types.MsgSetWithdrawAddress
	MsgWithdrawDelegatorReward     = types.MsgWithdrawDelegatorReward
	MsgWithdrawValidatorCommission = types.MsgWithdrawValidatorCommission
	MsgFundCommunityPool           = types.MsgFundCommunityPool

	MsgSubmitProposal    = types.MsgSubmitProposal
	MsgExecLegacyContent = types.MsgExecLegacyContent
	MsgVote              = types.MsgVote
	MsgGovDeposit        = types.MsgGovDeposit
	MsgVoteWeighted      = types.MsgVoteWeighted

	MsgUnjail  = types.MsgUnjail
	MsgImpeach = types.MsgImpeach

	MsgGrantAllowance  = types.MsgGrantAllowance
	MsgRevokeAllowance = types.MsgRevokeAllowance

	MsgClaim = types.MsgClaim

	MsgTransferOut = types.MsgTransferOut

	MsgCreatePaymentAccount = types.MsgCreatePaymentAccount
	MsgPaymentDeposit       = types.MsgPaymentDeposit
	MsgWithdraw             = types.MsgWithdraw
	MsgDisableRefund        = types.MsgDisableRefund

	MsgCreateStorageProvider = types.MsgCreateStorageProvider
	MsgSpDeposit             = types.MsgSpDeposit
	MsgEditStorageProvider   = types.MsgEditStorageProvider

	MsgCreateBucket      = types.MsgCreateBucket
	MsgDeleteBucket      = types.MsgDeleteBucket
	MsgCreateObject      = types.MsgCreateObject
	MsgSealObject        = types.MsgSealObject
	MsgRejectSealObject  = types.MsgRejectSealObject
	MsgCopyObject        = types.MsgCopyObject
	MsgDeleteObject      = types.MsgDeleteObject
	MsgCreateGroup       = types.MsgCreateGroup
	MsgDeleteGroup       = types.MsgDeleteGroup
	MsgUpdateGroupMember = types.MsgUpdateGroupMember
	MsgLeaveGroup        = types.MsgLeaveGroup
)
