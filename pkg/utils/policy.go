package utils

import (
	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/bnb-chain/greenfield/types/common"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewStatement return the statement of permission module
func NewStatement(actions []permTypes.ActionType, effect permTypes.Effect,
	resource []string, opts types.NewStatementOptions) permTypes.Statement {
	statement := permTypes.Statement{
		Actions:        actions,
		Effect:         effect,
		Resources:      resource,
		ExpirationTime: opts.StatementExpireTime,
	}

	if opts.LimitSize != 0 {
		statement.LimitSize = &common.UInt64Value{Value: opts.LimitSize}
	}

	return statement
}

func NewPrincipalWithAccount(principalAddr sdk.AccAddress) (types.Principal, error) {
	p := permTypes.NewPrincipalWithAccount(principalAddr)
	principalBytes, err := p.Marshal()
	if err != nil {
		return "", err
	}
	return types.Principal(principalBytes), nil
}

func NewPrincipalWithGroupId(groupId uint64) (types.Principal, error) {
	p := permTypes.NewPrincipalWithGroup(sdkmath.NewUint(groupId))
	principalBytes, err := p.Marshal()
	if err != nil {
		return "", err
	}
	return types.Principal(principalBytes), nil
}
