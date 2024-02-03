package feemarket

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	feemarketv1 "github.com/evmos/ethermint/api/ethermint/feemarket/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              feemarketv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get the fee market params",
					Long:      "Get the fee market parameter values.",
				},
				{
					RpcMethod: "BaseFee",
					Use:       "base-fee",
					Short:     "Get the base fee amount at a given block height",
					Long: `Get the base fee amount at a given block height.
If the height is not provided, it will use the latest height from context.`,
				},
				{
					RpcMethod: "BlockGas",
					Use:       "block-gas",
					Short:     "Get the block gas used at a given block height",
					Long: `Get the block gas used at a given block height.
If the height is not provided, it will use the latest height from context`,
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              feemarketv1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
		},
	}
}
