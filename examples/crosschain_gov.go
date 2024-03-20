package main

import (
	"context"
	"log"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func main() {
	account, _ := types.NewAccountFromPrivateKey("proposer", privateKey)
	cli, _ := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	ctx := context.Background()

	// The example below is for parameter change, each proposal can have only 1 msg, either change parameter or upgrade contract
	proposalID, txHash, err := cli.SubmitProposal(
		ctx,
		[]sdk.Msg{parameterChange()}, // or upgradeContract() for upgrading contract
		math.NewIntWithDecimal(1000, gnfdSdkTypes.DecimalBNB), // deposit, various from different env
		"Change BSC contract parameter",
		"Change BSC contract parameter",
		types.SubmitProposalOptions{TxOpts: gnfdSdkTypes.TxOption{}},
	)
	if err != nil {
		log.Fatalf("unable to submit proposal , %v", err)
	}
	cli.WaitForTx(ctx, txHash)

	// Have validators to vote for the proposal
	// there should be enough validators to vote for the proposal
	validatorPrivKey := "0x..."
	validatorAcct, _ := types.NewAccountFromPrivateKey("validator", validatorPrivKey)
	cli.SetDefaultAccount(validatorAcct)
	voteTxHash, err := cli.VoteProposal(ctx, proposalID, govv1.OptionYes, types.VoteProposalOptions{})
	cli.WaitForTx(ctx, voteTxHash)
}

// Suppose we want to modify a parameter of contract 0x40eC91B82D7aCAA065d54B08D751505D479b0E43, fillin CrossChainParamsChange as below
// note: Values if a slice of hex representation of the value you want to modify to(must not include 0x prefix), Targets defines the target contract address(es).
func parameterChange() sdk.Msg {
	govAcctAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	msgUpdateParams := &govv1.MsgUpdateCrossChainParams{
		Authority: govAcctAddress,
		Params: govv1.CrossChainParamsChange{
			Key:     "batchSizeForOracle",                                                         // The parameter name.
			Values:  []string{"0000000000000000000000000000000000000000000000000000000000000034"}, // the value in hex format. The length might vary depend on the exact parameter you want to change.
			Targets: []string{"0x40eC91B82D7aCAA065d54B08D751505D479b0E43"},                       // the contract's address
		},
		DestChainId: 97, // Dest BSC chain ID
	}
	return msgUpdateParams
}

// Suppose the current bucketHub contract is 0x111568F484E4b8759a3aeC6aF11EA17BC18479A8, objectHub 0x2F0cf555a0E1dAE8CDacef66D8244E49Ee72Ad2D, grouphub 0x40eC91B82D7aCAA065d54B08D751505D479b0E43.
// respectively, we want to upgrade to 0x82CDc0BDb92Af93F301332Ed05F4F844c7c74FD6, 0xd00137EABe7CC9434EA70Cde29f9DB5f65a335f7, 0xc11bFABfFE9e1A4A1557f1494cb74Cc86AB69441.
// fill this the MsgUpdateCrossChainParams as below
func upgradeContract() sdk.Msg {
	govAcctAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	msgUpdateParams := &govv1.MsgUpdateCrossChainParams{
		Authority: govAcctAddress,
		Params: govv1.CrossChainParamsChange{
			Key:     "upgrade", // The key specify the purpose of the governance. It must be ""upgrade" for upgrade contract
			Values:  []string{"0x82CDc0BDb92Af93F301332Ed05F4F844c7c74FD6", "0xd00137EABe7CC9434EA70Cde29f9DB5f65a335f7", "0xc11bFABfFE9e1A4A1557f1494cb74Cc86AB69441"},
			Targets: []string{"0x111568F484E4b8759a3aeC6aF11EA17BC18479A8", "0x2F0cf555a0E1dAE8CDacef66D8244E49Ee72Ad2D", "0x40eC91B82D7aCAA065d54B08D751505D479b0E43"},
		},
		DestChainId: 97, // Dest BSC chain ID
	}
	return msgUpdateParams
}
