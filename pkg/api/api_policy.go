package api

import (
	"github.com/bnb-chain/greenfield/sdk/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sendPutPolicyTxn broadcast the putPolicy msg and return the txn hash
func (c *Client) sendPutPolicyTxn(msg *storageTypes.MsgPutPolicy, txOpts *types.TxOption) (string, error) {
	if err := msg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{msg}, txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// sendDelPolicyTxn broadcast the deletePolicy msg and return the txn hash
func (c *Client) sendDelPolicyTxn(operator sdk.AccAddress, resource string, principal *permTypes.Principal, txOpts *types.TxOption) (string, error) {
	delPolicyMsg := storageTypes.NewMsgDeletePolicy(operator, resource, principal)

	if err := delPolicyMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{delPolicyMsg}, txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}
