package types

import (
	"fmt"
	"math/big"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/types"
)

var _ evmtypes.LegacyParams = &V4Params{}

var (
	// DefaultEVMDenom defines the default EVM denomination on Ethermint
	DefaultEVMDenom = types.AttoPhoton
	// DefaultAllowUnprotectedTxs rejects all unprotected txs (i.e false)
	DefaultAllowUnprotectedTxs = false
	DefaultEnableCreate        = true
	DefaultEnableCall          = true
	DefaultExtraEIPs           = V4ExtraEIPs{AvailableExtraEIPs}
)

// Parameter keys
var (
	ParamStoreKeyEVMDenom            = []byte("EVMDenom")
	ParamStoreKeyEnableCreate        = []byte("EnableCreate")
	ParamStoreKeyEnableCall          = []byte("EnableCall")
	ParamStoreKeyExtraEIPs           = []byte("EnableExtraEIPs")
	ParamStoreKeyChainConfig         = []byte("ChainConfig")
	ParamStoreKeyAllowUnprotectedTxs = []byte("AllowUnprotectedTxs")

	// AvailableExtraEIPs define the list of all EIPs that can be enabled by the
	// EVM interpreter. These EIPs are applied in order and can override the
	// instruction sets from the latest hard fork enabled by the ChainConfig. For
	// more info check:
	// https://github.com/ethereum/go-ethereum/blob/master/core/vm/interpreter.go#L97
	AvailableExtraEIPs = []int64{1344, 1884, 2200, 2929, 3198, 3529}
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&V4Params{})
}

// NewParams creates a new Params instance
func NewParams(evmDenom string, allowUnprotectedTxs, enableCreate, enableCall bool, config V4ChainConfig, extraEIPs V4ExtraEIPs) V4Params {
	return V4Params{
		EvmDenom:            evmDenom,
		AllowUnprotectedTxs: allowUnprotectedTxs,
		EnableCreate:        enableCreate,
		EnableCall:          enableCall,
		V4ExtraEIPs:         extraEIPs,
		V4ChainConfig:       config,
	}
}

// DefaultParams returns default evm parameters
// ExtraEIPs is empty to prevent overriding the latest hard fork instruction set
func DefaultParams() V4Params {
	return V4Params{
		EvmDenom:            DefaultEVMDenom,
		EnableCreate:        DefaultEnableCreate,
		EnableCall:          DefaultEnableCall,
		V4ChainConfig:       DefaultChainConfig(),
		V4ExtraEIPs:         DefaultExtraEIPs,
		AllowUnprotectedTxs: DefaultAllowUnprotectedTxs,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *V4Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEVMDenom, &p.EvmDenom, validateEVMDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCreate, &p.EnableCreate, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCall, &p.EnableCall, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyExtraEIPs, &p.V4ExtraEIPs, validateEIPs),
		paramtypes.NewParamSetPair(ParamStoreKeyChainConfig, &p.V4ChainConfig, validateChainConfig),
		paramtypes.NewParamSetPair(ParamStoreKeyAllowUnprotectedTxs, &p.AllowUnprotectedTxs, validateBool),
	}
}

// Validate performs basic validation on evm parameters.
func (p V4Params) Validate() error {
	if err := sdk.ValidateDenom(p.EvmDenom); err != nil {
		return err
	}

	if err := validateEIPs(p.V4ExtraEIPs); err != nil {
		return err
	}

	return p.V4ChainConfig.Validate()
}

// EIPs returns the ExtraEIPS as a int slice
func (p V4Params) EIPs() []int {
	eips := make([]int, len(p.V4ExtraEIPs.EIPs))
	for i, eip := range p.V4ExtraEIPs.EIPs {
		eips[i] = int(eip)
	}
	return eips
}

func validateEVMDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter EVM denom type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateEIPs(i interface{}) error {
	eips, ok := i.(V4ExtraEIPs)
	if !ok {
		return fmt.Errorf("invalid EIP slice type: %T", i)
	}

	for _, eip := range eips.EIPs {
		if !vm.ValidEip(int(eip)) {
			return fmt.Errorf("EIP %d is not activateable, valid EIPS are: %s", eip, vm.ActivateableEips())
		}
	}

	return nil
}

func validateChainConfig(i interface{}) error {
	cfg, ok := i.(V4ChainConfig)
	if !ok {
		return fmt.Errorf("invalid chain config type: %T", i)
	}

	return cfg.Validate()
}

// IsLondon returns if london hardfork is enabled.
func IsLondon(ethConfig *params.ChainConfig, height int64) bool {
	return ethConfig.IsLondon(big.NewInt(height))
}
