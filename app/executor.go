package app

import (
	"context"
	"io"

	"cosmossdk.io/store/cachemulti"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	blockstm "github.com/crypto-org-chain/go-block-stm"
)

func DefaultTxExecutor(_ context.Context,
	blockSize int,
	ms storetypes.MultiStore,
	deliverTxWithMultiStore func(int, storetypes.MultiStore) *abci.ExecTxResult,
) ([]*abci.ExecTxResult, error) {
	results := make([]*abci.ExecTxResult, blockSize)
	for i := 0; i < blockSize; i++ {
		results[i] = deliverTxWithMultiStore(i, ms)
	}
	return evmtypes.PatchTxResponses(results), nil
}

func STMTxExecutor(stores []storetypes.StoreKey, workers int) baseapp.TxExecutor {
	return func(
		ctx context.Context,
		blockSize int,
		ms storetypes.MultiStore,
		deliverTxWithMultiStore func(int, storetypes.MultiStore) *abci.ExecTxResult,
	) ([]*abci.ExecTxResult, error) {
		if blockSize == 0 {
			return nil, nil
		}
		results := make([]*abci.ExecTxResult, blockSize)
		if err := blockstm.ExecuteBlock(
			ctx,
			blockSize,
			stores,
			stmMultiStoreWrapper{ms},
			workers,
			func(txn blockstm.TxnIndex, ms blockstm.MultiStore) {
				result := deliverTxWithMultiStore(int(txn), msWrapper{ms, stores})
				results[txn] = result
			},
		); err != nil {
			return nil, err
		}

		return evmtypes.PatchTxResponses(results), nil
	}
}

type msWrapper struct {
	blockstm.MultiStore
	storeKeys []storetypes.StoreKey
}

var _ storetypes.MultiStore = msWrapper{}

func (ms msWrapper) GetStore(key storetypes.StoreKey) storetypes.Store {
	panic("not implemented")
}

func (ms msWrapper) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return ms.MultiStore.GetKVStore(key)
}

func (ms msWrapper) CacheMultiStore() storetypes.CacheMultiStore {
	var (
		keys   = make(map[string]storetypes.StoreKey)
		stores = make(map[storetypes.StoreKey]storetypes.CacheWrapper)
	)

	for _, k := range ms.storeKeys {
		keys[k.String()] = k
		stores[k] = ms.GetKVStore(k)
	}

	return cachemulti.NewStore(nil, stores, keys, nil, nil)
}

func (ms msWrapper) CacheMultiStoreWithVersion(int64) (storetypes.CacheMultiStore, error) {
	panic("not implemented")
}

// Implements CacheWrapper.
func (ms msWrapper) CacheWrap() storetypes.CacheWrap {
	return ms.CacheMultiStore().(storetypes.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (ms msWrapper) CacheWrapWithTrace(_ io.Writer, _ storetypes.TraceContext) storetypes.CacheWrap {
	return ms.CacheWrap()
}

// LatestVersion returns the latest version in the store
func (ms msWrapper) LatestVersion() int64 {
	panic("not implemented")
}

// GetStoreType returns the type of the store.
func (ms msWrapper) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

// Implements interface MultiStore
func (ms msWrapper) SetTracer(io.Writer) storetypes.MultiStore {
	return nil
}

// Implements interface MultiStore
func (ms msWrapper) SetTracingContext(storetypes.TraceContext) storetypes.MultiStore {
	return nil
}

// Implements interface MultiStore
func (ms msWrapper) TracingEnabled() bool {
	return false
}

type stmMultiStoreWrapper struct {
	storetypes.MultiStore
}

var _ blockstm.MultiStore = stmMultiStoreWrapper{}

func (ms stmMultiStoreWrapper) GetStore(key storetypes.StoreKey) storetypes.Store {
	panic("not implementation")
}

func (ms stmMultiStoreWrapper) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return ms.MultiStore.GetKVStore(key)
}
