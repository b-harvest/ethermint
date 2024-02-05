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
			EnhanceCustomCommand: true, // We provide custom Storage and Code in client/cli/tx.go
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get the evm params",
					Long:      "Get the evm parameter values.",
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
			EnhanceCustomCommand: true, // We provide custom RawTx in client/cli/tx.go
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
