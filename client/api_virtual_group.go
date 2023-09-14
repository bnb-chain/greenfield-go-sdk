package client

import (
	"context"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type IVirtualGroupClient interface {
	QueryVirtualGroupFamily(ctx context.Context, globalVirtualGroupFamilyID uint32) (*types.GlobalVirtualGroupFamily, error)
}

func (c *Client) QueryVirtualGroupFamily(ctx context.Context, globalVirtualGroupFamilyID uint32) (*types.GlobalVirtualGroupFamily, error) {
	queryResponse, err := c.chainClient.GlobalVirtualGroupFamily(ctx, &types.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: globalVirtualGroupFamilyID,
	})
	if err != nil {
		return nil, err
	}

	return queryResponse.GlobalVirtualGroupFamily, nil
}
