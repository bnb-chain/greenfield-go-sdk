package client

import (
	"context"

	"cosmossdk.io/math"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Executable interface {
	InvokeExecution(ctx context.Context, objectId math.Uint, inputObjectIds []math.Uint, maxGas math.Uint, method string, params []byte, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
	SubmitExecutionResult(ctx context.Context, taskId math.Uint, status uint32, resultDataUri string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error)
}

// InvokeExecution invokes an execution task on Greenfield
func (c *client) InvokeExecution(ctx context.Context, objectId math.Uint, inputObjectIds []math.Uint, maxGas math.Uint, method string, params []byte, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgInvoke := &storagetypes.MsgInvokeExecution{
		Operator:           c.MustGetDefaultAccount().GetAddress().String(),
		ExecutableObjectId: objectId,
		InputObjectIds:     inputObjectIds,
		MaxGas:             maxGas,
		Method:             method,
		Params:             params,
	}

	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgInvoke}, &txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

// SubmitExecutionResult submits the execution result to Greenfield
func (c *client) SubmitExecutionResult(ctx context.Context, taskId math.Uint, status uint32, resultDataUri string, txOption gnfdSdkTypes.TxOption) (*sdk.TxResponse, error) {
	msgSubmit := &storagetypes.MsgSubmitExecutionResult{
		Operator:      c.MustGetDefaultAccount().GetAddress().String(),
		TaskId:        taskId,
		Status:        status,
		ResultDataUri: resultDataUri,
	}

	txResp, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgSubmit}, &txOption)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}
