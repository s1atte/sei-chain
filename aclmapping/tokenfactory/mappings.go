package aclTokenFactorymapping

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkacltypes "github.com/cosmos/cosmos-sdk/types/accesscontrol"
	aclkeeper "github.com/cosmos/cosmos-sdk/x/accesscontrol/keeper"
	acltypes "github.com/cosmos/cosmos-sdk/x/accesscontrol/types"
	TokenFactorymoduletypes "github.com/sei-protocol/sei-chain/x/TokenFactory/types"
)

var InvalidMessageType = fmt.Errorf("Invalid message received for TokenFactory Module")

func GetTokenFactoryDependencyGenerators() aclkeeper.DependencyGeneratorMap {
	dependencyGeneratorMap := make(aclkeeper.DependencyGeneratorMap)

	// TokenFactory place orders
	MintKey := acltypes.GenerateMessageKey(&TokenFactorymoduletypes.MsgMint{})
	dependencyGeneratorMap[MintKey] = TokenFactoryMintDependencyGenerator

	return dependencyGeneratorMap
}

func TokenFactoryMintDependencyGenerator(keeper aclkeeper.Keeper, ctx sdk.Context, msg sdk.Msg) ([]sdkacltypes.AccessOperation, error) {
	MintMsg, ok := msg.(*TokenFactorymoduletypes.MsgMint)
	if !ok {
		return []sdkacltypes.AccessOperation{}, InvalidMessageType
	}

	// server.bankKeeper.GetDenomMetaData --> Pays the fee collector :( and gets Amount Denom Info - Read
	// server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	// SendCoinsFromModuleToAccount -> Other way around :) -> Deferred Withdrawls??
		// Deposit to reciever but then mass withdrawl from module account at the end
	return []sdkacltypes.AccessOperation{
		// {AccessType: sdkacltypes.AccessType_WRITE, ResourceType: sdkacltypes.ResourceType_KV, IdentifierTemplate: MintMsg.ContractAddr},
	}, nil
}
