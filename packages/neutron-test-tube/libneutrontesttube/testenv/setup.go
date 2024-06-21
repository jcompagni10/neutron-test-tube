package testenv

import (
	"encoding/json"
	"fmt"
	"time"

	// helpers

	// tendermint
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"

	// cosmos-sdk

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// wasmd
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	// neutron
	"github.com/neutron-org/neutron/v4/app"
	"github.com/neutron-org/neutron/v4/testutil/consumer"
	tokenfactorytypes "github.com/neutron-org/neutron/v4/x/tokenfactory/types"

	// interchain security
	consumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
)

func GenesisStateWithValSet(codec codec.Codec,
	genesisState map[string]json.RawMessage,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) app.GenesisState {

	//////////////////////
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction
	initValPowers := []abci.ValidatorUpdate{}

	for _, val := range valSet.Validators {
		pk, _ := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   math.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			MinSelfDelegation: math.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), val.Address.String(), math.LegacyOneDec()))

		// add initial validator powers so consumer InitGenesis runs correctly
		pub, _ := val.ToProto()
		initValPowers = append(initValPowers, abci.ValidatorUpdate{
			Power:  val.VotingPower,
			PubKey: pub.PubKey,
		})

	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	vals, err := tmtypes.PB2TM.ValidatorUpdates(initValPowers)
	if err != nil {
		panic("failed to get vals")
	}

	wasmGen := wasmtypes.GenesisState{
		Params: wasmtypes.Params{
			// Allow store code without gov
			CodeUploadAccess:             wasmtypes.AllowEverybody,
			InstantiateDefaultPermission: wasmtypes.AccessTypeEverybody,
		},
	}
	genesisState[wasmtypes.ModuleName] = codec.MustMarshalJSON(&wasmGen)

	consumerGenesisState := consumer.CreateMinimalConsumerTestGenesis()
	consumerGenesisState.InitialValSet = initValPowers
	consumerGenesisState.ProviderConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).
		Hash()
	consumerGenesisState.Params.Enabled = true
	genesisState[consumertypes.ModuleName] = codec.MustMarshalJSON(consumerGenesisState)

	return genesisState
}

type TestEnv struct {
	App                *app.App
	Ctx                sdk.Context
	ParamTypesRegistry ParamTypeRegistry
	ValPrivs           []*secp256k1.PrivKey
	NodeHome           string
}

// DebugAppOptions is a stub implementing AppOptions
type DebugAppOptions struct{}

// Get implements AppOptions
func (ao DebugAppOptions) Get(o string) interface{} {
	if o == server.FlagTrace {
		return true
	}
	return nil
}

func NewNeutronApp(nodeHome string) *app.App {
	var wasmOptions = []wasmkeeper.Option{}
	encoding := app.MakeEncodingConfig()
	db := dbm.NewMemDB()

	return app.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		nodeHome,
		5,
		encoding,
		DebugAppOptions{},
		wasmOptions,
		baseapp.SetChainID("neutron-1"),
	)

}

func InitChain(appInstance *app.App) (sdk.Context, secp256k1.PrivKey) {

	privVal := NewPV()
	pubKey, err := privVal.GetPubKey()
	requireNoErr(err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(
		senderPrivKey.PubKey().Address().Bytes(),
		senderPrivKey.PubKey(),
		0,
		0,
	)

	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000))),
	}

	genesisState := app.NewDefaultGenesisState(appInstance.AppCodec())

	genesisState = GenesisStateWithValSet(
		appInstance.AppCodec(),
		genesisState,
		valSet,
		[]authtypes.GenesisAccount{acc},
		balance)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	requireNoErr(err)

	appInstance.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         "neutron-1",
		},
	)

	ctx := appInstance.NewContext(false)

	return ctx, secp256k1.PrivKey{Key: privVal.PrivKey.Bytes()}
}

func (env *TestEnv) BeginNewBlock(timeIncreaseSeconds uint64) {

	newBlockTime := env.Ctx.BlockTime().Add(time.Duration(timeIncreaseSeconds) * time.Second)

	newCtx := env.Ctx.WithBlockTime(newBlockTime).WithBlockHeight(env.Ctx.BlockHeight() + 1)
	env.Ctx = newCtx

	env.App.BeginBlocker(env.Ctx)
	env.Ctx = env.App.NewContext(false)

}

func (env *TestEnv) FundValidators() {
	for _, valPriv := range env.ValPrivs {
		valAddr := sdk.AccAddress(valPriv.PubKey().Address())
		coins := sdk.NewCoins(sdk.NewInt64Coin("untrn", 9223372036854775807))
		env.FundAccount(valAddr, coins)
	}
}

func (env *TestEnv) FundAccount(accAddr sdk.AccAddress, coins sdk.Coins) {
	err := env.App.BankKeeper.MintCoins(env.Ctx, tokenfactorytypes.ModuleName, coins)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to fund account"))
	}

	err = env.App.BankKeeper.SendCoinsFromModuleToAccount(env.Ctx, tokenfactorytypes.ModuleName, accAddr, coins)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to fund account"))
	}
}

func (env *TestEnv) GetValidatorAddresses() []string {
	validators, err := env.App.ConsumerKeeper.GetAllValidators(env.Ctx)
	requireNoErr(err)
	var addresses []string
	for _, validator := range validators {
		addresses = append(addresses, validator.OperatorAddress)
	}

	return addresses
}

func (env *TestEnv) SetupDefaultValidator() {
	validators, err := env.App.ConsumerKeeper.GetAllValidators(env.Ctx)
	requireNoErr(err)
	valAddrFancy, err := validators[0].GetConsAddr()
	requireNoErr(err)
	env.setupDefaultValidatorSigningInfo(valAddrFancy)
}

func (env *TestEnv) setupDefaultValidatorSigningInfo(consAddr sdk.ConsAddress) {
	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		env.Ctx.BlockHeight(),
		0,
		time.Unix(0, 0),
		false,
		0,
	)
	env.App.SlashingKeeper.SetValidatorSigningInfo(env.Ctx, consAddr, signingInfo)
}

func (env *TestEnv) SetupParamTypes() {
	pReg := env.ParamTypesRegistry
	pReg.RegisterParamSet(&tokenfactorytypes.Params{})
}

func requireNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

func requireNoNil(name string, nilable any) {
	if nilable == nil {
		panic(fmt.Sprintf("%s must not be nil", name))
	}
}

func requierTrue(name string, b bool) {
	if !b {
		panic(fmt.Sprintf("%s must be true", name))
	}
}
