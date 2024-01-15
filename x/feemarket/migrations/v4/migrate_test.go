package v4_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/encoding"
	v4 "github.com/evmos/ethermint/x/feemarket/migrations/v4"
	"github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspaceEmpty() mockSubspace {
	return mockSubspace{}
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeTestEncodingConfig()
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	storeService := runtime.NewKVStoreService(storeKey)

	legacySubspaceEmpty := newMockSubspaceEmpty()
	require.Error(t, v4.MigrateStore(ctx, storeService, legacySubspaceEmpty, cdc))

	legacySubspace := newMockSubspace(types.DefaultParams())
	require.NoError(t, v4.MigrateStore(ctx, storeService, legacySubspace, cdc))

	kvStore := storeService.OpenKVStore(ctx)
	paramsBz, err := kvStore.Get(types.ParamsKey)
	require.NoError(t, err)
	var params types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	require.Equal(t, params, legacySubspace.ps)
}
