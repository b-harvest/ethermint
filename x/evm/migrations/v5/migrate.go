package v5

import (
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/types"

	v5types "github.com/evmos/ethermint/x/evm/migrations/v5/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 4 to
// version 5. Specifically, it takes the parameters that are currently stored
// in separate keys and stores them directly into the x/evm module state using
// a single params key.
func MigrateStore(
	ctx sdk.Context,
	storeService corestore.KVStoreService,
	cdc codec.BinaryCodec,
) error {
	var (
		extraEIPs   v5types.V5ExtraEIPs
		chainConfig types.ChainConfig
		params      types.Params
	)

	store := storeService.OpenKVStore(ctx)

	value, err := store.Get(types.ParamStoreKeyEVMDenom)
	if err != nil {
		return err
	}
	denom := string(value)

	extraEIPsBz, err := store.Get(types.ParamStoreKeyExtraEIPs)
	if err != nil {
		return err
	}
	cdc.MustUnmarshal(extraEIPsBz, &extraEIPs)

	// revert ExtraEIP change for Evmos testnet
	if ctx.ChainID() == "evmos_9000-4" {
		extraEIPs.EIPs = []int64{}
	}

	chainCfgBz, err := store.Get(types.ParamStoreKeyChainConfig)
	if err != nil {
		return err
	}
	cdc.MustUnmarshal(chainCfgBz, &chainConfig)

	params.EvmDenom = denom
	params.ExtraEIPs = extraEIPs.EIPs
	params.ChainConfig = chainConfig
	params.EnableCreate, _ = store.Has(types.ParamStoreKeyEnableCreate)
	params.EnableCall, _ = store.Has(types.ParamStoreKeyEnableCall)
	params.AllowUnprotectedTxs, _ = store.Has(types.ParamStoreKeyAllowUnprotectedTxs)

	_ = store.Delete(types.ParamStoreKeyChainConfig)
	_ = store.Delete(types.ParamStoreKeyExtraEIPs)
	_ = store.Delete(types.ParamStoreKeyEVMDenom)
	_ = store.Delete(types.ParamStoreKeyEnableCreate)
	_ = store.Delete(types.ParamStoreKeyEnableCall)
	_ = store.Delete(types.ParamStoreKeyAllowUnprotectedTxs)

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)

	err = store.Set(types.KeyPrefixParams, bz)
	if err != nil {
		return err
	}

	return nil
}
