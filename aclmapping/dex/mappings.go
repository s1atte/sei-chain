package acldexmapping

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkacltypes "github.com/cosmos/cosmos-sdk/types/accesscontrol"
	aclkeeper "github.com/cosmos/cosmos-sdk/x/accesscontrol/keeper"
	acltypes "github.com/cosmos/cosmos-sdk/x/accesscontrol/types"
	utils "github.com/sei-protocol/sei-chain/aclmapping/utils"
	dexmoduletypes "github.com/sei-protocol/sei-chain/x/dex/types"
)

var ErrPlaceOrdersGenerator = fmt.Errorf("invalid message received for dex module")

func GetDexDependencyGenerators() aclkeeper.DependencyGeneratorMap {
	dependencyGeneratorMap := make(aclkeeper.DependencyGeneratorMap)

	// dex place orders
	placeOrdersKey := acltypes.GenerateMessageKey(&dexmoduletypes.MsgPlaceOrders{})
	cancelOrdersKey := acltypes.GenerateMessageKey(&dexmoduletypes.MsgCancelOrders{})
	dependencyGeneratorMap[placeOrdersKey] = DexPlaceOrdersDependencyGenerator
	dependencyGeneratorMap[cancelOrdersKey] = DexCancelOrdersDependencyGenerator

	return dependencyGeneratorMap
}

func DexPlaceOrdersDependencyGenerator(keeper aclkeeper.Keeper, ctx sdk.Context, msg sdk.Msg) ([]sdkacltypes.AccessOperation, error) {
	placeOrdersMsg, ok := msg.(*dexmoduletypes.MsgPlaceOrders)
	if !ok {
		return []sdkacltypes.AccessOperation{}, ErrPlaceOrdersGenerator
	}
	// TODO: This is not final, JUST AN EXAMPLE
	return []sdkacltypes.AccessOperation{
		// validateOrder
		// transferFunds
		// getNextOrderID (this will write to KV store)
		//		as long as we do it per contract, it should be fine since thats how orderIDs are indexed
		// GetMemState.GetBlockOrders().add() - this will write to the DEX cache
		// GetMemState will just access dexcache.MemState (no locks)
		// GetBlockOrders - will call MemState.synchronizedAccess(ctx, contractAddr)
		// 		this is granularized on a contractAddr basis in SynchronizedAccess
		//			there is a timeout here so I don't think we should async access same contracts
		//			but technically, should it matter since the timeout is 5 seconds

		// I think the way it's working is that if it's accessing the contract already
		// it wil be part of the executing contract

		// Questions? What is a executingContract?

		{
			AccessType:         sdkacltypes.AccessType_READ,
			ResourceType:       sdkacltypes.ResourceType_KV_DEX,
			IdentifierTemplate: utils.GetIdentifierTemplatePerModule(utils.DEX, placeOrdersMsg.ContractAddr),
		},
		{
			AccessType:         sdkacltypes.AccessType_WRITE,
			ResourceType:       sdkacltypes.ResourceType_KV_DEX,
			IdentifierTemplate: utils.GetIdentifierTemplatePerModule(utils.DEX, placeOrdersMsg.ContractAddr),
		},

		// Last Operation should always be a commit
		{
			ResourceType:       sdkacltypes.ResourceType_ANY,
			AccessType:         sdkacltypes.AccessType_COMMIT,
			IdentifierTemplate: utils.DefaultIDTemplate,
		},
	}, nil
}

func DexCancelOrdersDependencyGenerator(keeper aclkeeper.Keeper, ctx sdk.Context, msg sdk.Msg) ([]sdkacltypes.AccessOperation, error) {
	cancelOrdersMsg, ok := msg.(*dexmoduletypes.MsgCancelOrders)
	if !ok {
		return []sdkacltypes.AccessOperation{}, ErrPlaceOrdersGenerator
	}
	// TODO: This is not final, JUST AN EXAMPLE
	return []sdkacltypes.AccessOperation{
		// GetLongAllocationForOrderID or GetShortAllocationForOrderID
		// GemMemState.GetBlockCancels will get all cancellations for a specific block
		// 	will iterate over all cancel to verify it hasn't been cancelled in a previous tx in the same block
		//	if the orderID is not already cancelled, ADD it into the MemState

		{
			AccessType:         sdkacltypes.AccessType_READ,
			ResourceType:       sdkacltypes.ResourceType_KV_DEX,
			IdentifierTemplate: utils.GetIdentifierTemplatePerModule(utils.DEX, cancelOrdersMsg.ContractAddr),
		},
		{
			AccessType:         sdkacltypes.AccessType_WRITE,
			ResourceType:       sdkacltypes.ResourceType_KV_DEX,
			IdentifierTemplate: utils.GetIdentifierTemplatePerModule(utils.DEX, cancelOrdersMsg.ContractAddr),
		},

		// Last Operation should always be a commit
		{
			ResourceType:       sdkacltypes.ResourceType_ANY,
			AccessType:         sdkacltypes.AccessType_COMMIT,
			IdentifierTemplate: utils.DefaultIDTemplate,
		},
	}, nil
}
