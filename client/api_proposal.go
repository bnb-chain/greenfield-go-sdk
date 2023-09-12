package client

import (
	"context"
	"strconv"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govTypesV1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type Proposal interface {
	SubmitProposal(ctx context.Context, msgs []sdk.Msg, depositAmount math.Int, title, summary string, opts types.SubmitProposalOptions) (uint64, string, error)
	VoteProposal(ctx context.Context, proposalID uint64, voteOption govTypesV1.VoteOption, opts types.VoteProposalOptions) (string, error)
	GetProposal(ctx context.Context, proposalID uint64) (*govTypesV1.Proposal, error)
}

func (c *Client) SubmitProposal(ctx context.Context, msgs []sdk.Msg, depositAmount math.Int, title, summary string, opts types.SubmitProposalOptions) (uint64, string, error) {
	msgSubmitProposal, err := govTypesV1.NewMsgSubmitProposal(msgs, sdk.NewCoins(sdk.NewCoin(gnfdSdkTypes.Denom, depositAmount)), c.defaultAccount.GetAddress().String(), opts.Metadata, title, summary)
	if err != nil {
		return 0, "", err
	}
	err = msgSubmitProposal.ValidateBasic()
	if err != nil {
		return 0, "", err
	}
	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgSubmitProposal}, &opts.TxOption)
	if err != nil {
		return 0, "", err
	}
	waitCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	txResult, err := c.WaitForTx(waitCtx, txResp.TxResponse.TxHash)
	if err != nil {
		return 0, "", err
	}
	key := govTypes.AttributeKeyProposalID
	for _, event := range txResult.TxResult.Events {
		for _, attr := range event.Attributes {
			if attr.Key == key {
				proposalID, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					return 0, txResp.TxResponse.TxHash, err
				}
				return proposalID, txResp.TxResponse.TxHash, nil
			}
		}
	}
	return 0, txResp.TxResponse.TxHash, types.ErrorProposalIDNotFound
}

func (c *Client) VoteProposal(ctx context.Context, proposalID uint64, voteOption govTypesV1.VoteOption, opts types.VoteProposalOptions) (string, error) {
	msgVote := govTypesV1.NewMsgVote(c.MustGetDefaultAccount().GetAddress(), proposalID, voteOption, opts.Metadata)
	resp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgVote}, &opts.TxOption)
	if err != nil {
		return "", err
	}
	return resp.TxResponse.TxHash, nil
}

func (c *Client) GetProposal(ctx context.Context, proposalID uint64) (*govTypesV1.Proposal, error) {
	resp, err := c.chainClient.GovQueryClientV1.Proposal(ctx, &govTypesV1.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return nil, nil
	}
	return resp.Proposal, nil
}
