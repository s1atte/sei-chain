package aclTokenFactorymapping

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkacltypes "github.com/cosmos/cosmos-sdk/types/accesscontrol"
	aclkeeper "github.com/cosmos/cosmos-sdk/x/accesscontrol/keeper"
	acltypes "github.com/cosmos/cosmos-sdk/x/accesscontrol/types"
	utils "github.com/sei-protocol/sei-chain/aclmapping/utils"
	tokenfactorymoduletypes "github.com/sei-protocol/sei-chain/x/TokenFactory/types"
)

var InvalidMessageType = fmt.Errorf("Invalid message received for TokenFactory Module")

func GetTokenFactoryDependencyGenerators() aclkeeper.DependencyGeneratorMap {
	dependencyGeneratorMap := make(aclkeeper.DependencyGeneratorMap)

	// TokenFactory place orders
	MintKey := acltypes.GenerateMessageKey(&tokenfactorymoduletypes.MsgMint{})
	dependencyGeneratorMap[MintKey] = TokenFactoryMintDependencyGenerator

	return dependencyGeneratorMap
}

func TokenFactoryMintDependencyGenerator(keeper aclkeeper.Keeper, ctx sdk.Context, msg sdk.Msg) ([]sdkacltypes.AccessOperation, error) {
	mintMsg, ok := msg.(*tokenfactorymoduletypes.MsgMint)
	if !ok {
		return []sdkacltypes.AccessOperation{}, InvalidMessageType
	}

	// server.bankKeeper.GetDenomMetaData --> Pays the fee collector :( and gets Amount Denom Info - Read
	// server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	// 	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount)) <-- Sends coin to banker :think
	// SendCoinsFromModuleToAccount -> Other way around :) -> Deferred Withdrawls??
		// Deposit to reciever but then mass withdrawl from module account at the end
	return []sdkacltypes.AccessOperation{
		// {AccessType: sdkacltypes.AccessType_WRITE, ResourceType: sdkacltypes.ResourceType_KV, IdentifierTemplate: MintMsg.ContractAddr},
	}, nil
}

func TokenFactoryBurnDependencyGenerator(keeper aclkeeper.Keeper, ctx sdk.Context, msg sdk.Msg) ([]sdkacltypes.AccessOperation, error) {
	burnMsg, ok := msg.(*tokenfactorymoduletypes.MsgBurn)
	if !ok {
		return []sdkacltypes.AccessOperation{}, InvalidMessageType
	}

	return []sdkacltypes.AccessOperation{
		{
			AccessType:         sdkacltypes.AccessType_WRITE,
			ResourceType:       sdkacltypes.ResourceType_KV,
			IdentifierTemplate:  utils.GetIdentifierTemplatePerModule(utils.AUTH, burnMsg),
		},

		// Last Operation should always be a commit
		{
			ResourceType:       sdkacltypes.ResourceType_ANY,
			AccessType:         sdkacltypes.AccessType_COMMIT,
			IdentifierTemplate:  utils.DefaultIDTemplate,
		},	}, nil
}
