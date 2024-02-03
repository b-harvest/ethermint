package evm

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	evmv1 "github.com/evmos/ethermint/api/ethermint/evm/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              evmv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get the evm params",
					Long:      "Get the evm parameter values.",
				},
				{
					RpcMethod:      "Code",
					Use:            "code ADDRESS",
					Short:          "Gets code from an account",
					Long:           "Gets code from an account. If the height is not provided, it will use the latest height from context.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "Storage",
					Use:            "storage ADDRESS KEY",
					Short:          "Gets storage for an account with a given key and height",
					Long:           "Gets storage for an account with a given key and height. If the height is not provided, it will use the latest height from context.", //nolint:lll
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}, {ProtoField: "key"}},
				},
				{
					RpcMethod: "Account",
					Skip:      true,
				},
				{
					RpcMethod: "CosmosAccount",
					Skip:      true,
				},
				{
					RpcMethod: "ValidatorAccount",
					Skip:      true,
				},
				{
					RpcMethod: "Balance",
					Skip:      true,
				},
				{
					RpcMethod: "EthCall",
					Skip:      true,
				},
				{
					RpcMethod: "EstimateGas",
					Skip:      true,
				},
				{
					RpcMethod: "TraceTx",
					Skip:      true,
				},
				{
					RpcMethod: "TraceBlock",
					Skip:      true,
				},
				{
					RpcMethod: "BaseFee",
					Skip:      true,
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              evmv1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // We provide costom RawTx in client/cli/tx.go
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod: "EthereumTx",
					Skip:      true,
				},
			},
		},
	}
}
