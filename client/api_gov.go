package client

import (
	"context"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type Gov interface {
	Vote(ctx context.Context, proposalId uint64, option govv1.VoteOption, metadata string, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	SubmitProposal()
}

func (c *client) Vote(ctx context.Context, proposalId uint64, option govv1.VoteOption, metadata string, txOption *gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	vote := govv1.NewMsgVote(c.MustGetDefaultAccount().GetAddress(), proposalId, option, metadata)
	tx, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{vote}, txOption)
	if err != nil {
		return nil, err
	}
	return tx.TxResponse, nil
}
