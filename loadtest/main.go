package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"

	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/sei-protocol/sei-chain/app"
	dextypes "github.com/sei-protocol/sei-chain/x/dex/types"
	tokenfactorytypes "github.com/sei-protocol/sei-chain/x/tokenfactory/types"
)

var TestConfig EncodingConfig

const (
	VortexData = "{\"position_effect\":\"Open\",\"leverage\":\"1\"}"
)

var FromMili = sdk.NewDec(1000000)

func init() {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	TestConfig = EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}
	std.RegisterLegacyAminoCodec(TestConfig.Amino)
	std.RegisterInterfaces(TestConfig.InterfaceRegistry)
	app.ModuleBasics.RegisterLegacyAminoCodec(TestConfig.Amino)
	app.ModuleBasics.RegisterInterfaces(TestConfig.InterfaceRegistry)
}

func run() {
	client := NewLoadTestClient()
	config := client.LoadTestConfig

	defer client.Close()

	if config.TxsPerBlock < config.MsgsPerTx {
		panic("Must have more TxsPerBlock than MsgsPerTx")
	}

	configString, _ := json.Marshal(config)
	fmt.Printf("Running with \n %s \n", string(configString))

	fmt.Printf("%s - Starting block prepare\n", time.Now().Format("2006-01-02T15:04:05"))
	workgroups, sendersList := client.BuildTxs()

	client.SendTxs(workgroups, sendersList)

	// Records the resulting TxHash to file
	client.WriteTxHashToFile()
	fmt.Printf("%s - Finished\n", time.Now().Format("2006-01-02T15:04:05"))
}

func generateMessage(config Config, key cryptotypes.PrivKey, msgPerTx uint64, validators []Validator) sdk.Msg {
	var msg sdk.Msg
	messageType := config.MessageType

	addr := sdk.AccAddress(key.PubKey().Address()).String()
	switch messageType {
	case "basic":
		msg = &banktypes.MsgSend{
			FromAddress: addr,
			ToAddress:   addr,
			Amount: sdk.NewCoins(sdk.Coin{
				Denom:  "usei",
				Amount: sdk.NewInt(1),
			}),
		}
	case "staking":
		msgType := config.MsgTypeDistr.Sample(messageType)

		switch msgType {
		case "delegate":
			msg = &stakingtypes.MsgDelegate{
				DelegatorAddress: addr,
				ValidatorAddress: validators[rand.Intn(len(validators))].OpperatorAddr,
				Amount:           sdk.Coin{Denom: "usei", Amount: sdk.NewInt(5)},
			}
		case "undelegate":
			msg = &stakingtypes.MsgUndelegate{
				DelegatorAddress: addr,
				ValidatorAddress: validators[rand.Intn(len(validators))].OpperatorAddr,
				Amount:           sdk.Coin{Denom: "usei", Amount: sdk.NewInt(1)},
			}
		case "begin_redelegate":
			msg = &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    addr,
				ValidatorSrcAddress: validators[rand.Intn(len(validators))].OpperatorAddr,
				ValidatorDstAddress: validators[rand.Intn(len(validators))].OpperatorAddr,
				Amount:              sdk.Coin{Denom: "usei", Amount: sdk.NewInt(1)},
			}
		default:
			panic("Unknown message type")
		}
	case "tokenfactory":
		msgType := config.MsgTypeDistr.Sample(messageType)

		// Template defined in populate_genesis_accounts.py
		denom := fmt.Sprintf("factory/%s/token", addr)

		switch msgType {
		case "mint":
			msg = &tokenfactorytypes.MsgMint{
				Sender: addr,
				Amount: sdk.Coin{Denom: denom, Amount: sdk.NewInt(1)},
			}
		case "burn":
			msg = &tokenfactorytypes.MsgBurn{
				Sender: addr,
				Amount: sdk.Coin{Denom: denom, Amount: sdk.NewInt(1)},
			}
		default:
			panic("Unknown tokenfactory message type")
		}
	case "dex":
		msgType := config.MsgTypeDistr.Sample(messageType)
		orderPlacements := []*dextypes.Order{}
		var orderType dextypes.OrderType
		if msgType == "limit" {
			orderType = dextypes.OrderType_LIMIT
		} else {
			orderType = dextypes.OrderType_MARKET
		}
		var direction dextypes.PositionDirection
		if rand.Float64() < 0.5 {
			direction = dextypes.PositionDirection_LONG
		} else {
			direction = dextypes.PositionDirection_SHORT
		}
		price := config.PriceDistr.Sample()
		quantity := config.QuantityDistr.Sample()
		contract := config.ContractDistr.Sample()
		for j := 0; j < int(msgPerTx); j++ {
			orderPlacements = append(orderPlacements, &dextypes.Order{
				Account:           addr,
				ContractAddr:      contract,
				PositionDirection: direction,
				Price:             price.Quo(FromMili),
				Quantity:          quantity.Quo(FromMili),
				PriceDenom:        "SEI",
				AssetDenom:        "ATOM",
				OrderType:         orderType,
				Data:              VortexData,
			})
		}
		amount, err := sdk.ParseCoinsNormalized(fmt.Sprintf("%d%s", price.Mul(quantity).Ceil().RoundInt64(), "usei"))
		if err != nil {
			panic(err)
		}
		msg = &dextypes.MsgPlaceOrders{
			Creator:      addr,
			Orders:       orderPlacements,
			ContractAddr: contract,
			Funds:        amount,
		}
	default:
		fmt.Printf("Unrecognized message type %s", config.MessageType)
	}
	return msg
}

func getLastHeight() int {
	out, err := exec.Command("curl", "http://localhost:26657/blockchain").Output()
	if err != nil {
		panic(err)
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(out, &dat); err != nil {
		panic(err)
	}
	height, err := strconv.Atoi(dat["last_height"].(string))
	if err != nil {
		panic(err)
	}
	return height
}

func main() {
	run()
}
