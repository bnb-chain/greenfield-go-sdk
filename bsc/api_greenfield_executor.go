package bsc

import (
	"context"

	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
)

type IGreenfieldExecutorClient interface {
	Execute(ctx context.Context, message *bsctypes.MultiMessage) (bool, error)
}
