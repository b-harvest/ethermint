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
	if params.EnableCreate, err = store.Has(types.ParamStoreKeyEnableCreate); err != nil {
		return err
	}
	if params.EnableCall, err = store.Has(types.ParamStoreKeyEnableCall); err != nil {
		return err
	}
	if params.AllowUnprotectedTxs, err = store.Has(types.ParamStoreKeyAllowUnprotectedTxs); err != nil {
		return err
	}

	if err = store.Delete(types.ParamStoreKeyChainConfig); err != nil {
		return err
	}
	if err = store.Delete(types.ParamStoreKeyExtraEIPs); err != nil {
		return err
	}
	if err = store.Delete(types.ParamStoreKeyEVMDenom); err != nil {
		return err
	}
	if err = store.Delete(types.ParamStoreKeyEnableCreate); err != nil {
		return err
	}
	if err = store.Delete(types.ParamStoreKeyEnableCall); err != nil {
		return err
	}
	if err = store.Delete(types.ParamStoreKeyAllowUnprotectedTxs); err != nil {
		return err
	}

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
