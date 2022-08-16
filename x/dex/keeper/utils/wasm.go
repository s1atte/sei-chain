package utils

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sei-protocol/sei-chain/x/dex/keeper"
)

func CallContractSudo(sdkCtx sdk.Context, k *keeper.Keeper, contractAddr string, msg interface{}) ([]byte, error) {
	fmt.Println(contractAddr)
	contractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	fmt.Println("TIME: CallContractSudo Start", time.Now().UTC().UnixMilli())
	if err != nil {
		sdkCtx.Logger().Error(err.Error())
		return []byte{}, err
	}
	fmt.Println("TIME: Json Marshal Start", time.Now().UTC().UnixMilli())
	wasmMsg, err := json.Marshal(msg)
	fmt.Println("TIME: Json Marshal end", time.Now().UTC().UnixMilli())
	if err != nil {
		sdkCtx.Logger().Error(err.Error())
		return []byte{}, err
	}
	fmt.Println("TIME: Sudo Start", time.Now().UTC().UnixMilli())
	data, err := k.WasmKeeper.Sudo(
		sdkCtx, contractAddress, wasmMsg,
	)
	fmt.Println("TIME: Sudo end", time.Now().UTC().UnixMilli())
	if err != nil {
		sdkCtx.Logger().Error(err.Error())
		return []byte{}, err
	}
	fmt.Println("TIME: CallContractSudo end", time.Now().UTC().UnixMilli())
	return data, nil
}
