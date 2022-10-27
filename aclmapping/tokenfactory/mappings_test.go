package aclTokenFactorymapping

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkacltypes "github.com/cosmos/cosmos-sdk/types/accesscontrol"
	acltypes "github.com/cosmos/cosmos-sdk/x/accesscontrol/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	oracletypes "github.com/sei-protocol/sei-chain/x/oracle/types"
	tokenfactorytypes "github.com/sei-protocol/sei-chain/x/tokenfactory/types"
	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)


func cacheTxContext(ctx sdk.Context) (sdk.Context, sdk.CacheMultiStore) {
	ms := ctx.MultiStore()
	msCache := ms.CacheMultiStore()
	return ctx.WithMultiStore(msCache), msCache
}

func TestMsgBurnValidation(t *testing.T) {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := secp256k1.GenPrivKey()
	addr2 := sdk.AccAddress(priv2.PubKey().Address())
	coins     := sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}

	tests := []struct {
		name          string
		expectedError error
		msg           *types.MsgSend
		dynamicDep 	  bool
	}{
		{
			name:          "default send",
			msg:           types.NewMsgSend(addr1, addr2, coins),
			expectedError: nil,
			dynamicDep: true,
		},
		{
			name:          "dont check synchronous",
			msg:           types.NewMsgSend(addr1, addr2, coins),
			expectedError: nil,
			dynamicDep: false,
		},
	}

	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}
	acc2 := &authtypes.BaseAccount{
		Address: addr2.String(),
	}
	accs := authtypes.GenesisAccounts{acc1, acc2}
	balances := []types.Balance{
		{
			Address: addr1.String(),
			Coins:   coins,
		},
		{
			Address: addr2.String(),
			Coins:   coins,
		},
	}

	app := simapp.SetupWithGenesisAccounts(accs, balances...)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	handler := bank.NewHandler(app.BankKeeper)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handlerCtx, cms := cacheTxContext(ctx)
			_, err := handler(handlerCtx, tc.msg)

			depdenencies , _ := TokenFactoryBurnDependencyGenerator(app.AccessControlKeeper, handlerCtx, tc.msg)

			if !tc.dynamicDep {
				depdenencies = sdkacltypes.SynchronousAccessOps()
			}

			if tc.expectedError != nil {
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
			missing := (sdkacltypes.ValidateAccessOperations(depdenencies, cms.GetEvents()))
			require.Empty(t, missing)
		})
	}
}

func TestGeneratorInvalidMessageTypes(t *testing.T) {
	accs := authtypes.GenesisAccounts{}
	balances := []types.Balance{}

	app := simapp.SetupWithGenesisAccounts(accs, balances...)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	oracleVote := oracletypes.MsgAggregateExchangeRateVote{
		ExchangeRates: "1usei",
		Feeder:        "test",
		Validator:     "validator",
	}

	_, err := TokenFactoryBurnDependencyGenerator(app.AccessControlKeeper, ctx, &oracleVote)
	require.Error(t, err)
}

func TestMsgBeginBankSendGenerator(t *testing.T) {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())

	accs := authtypes.GenesisAccounts{}
	balances := []types.Balance{}

	app := simapp.SetupWithGenesisAccounts(accs, balances...)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	sendMsg := tokenfactorytypes.MsgBurn{
		Sender: addr1.String(),
		Amount: sdk.NewInt64Coin("usei", 10),
	}

	accessOps, err := TokenFactoryBurnDependencyGenerator(app.AccessControlKeeper, ctx, &sendMsg)
	require.NoError(t, err)
	err = acltypes.ValidateAccessOps(accessOps)
	require.NoError(t, err)
}
