// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package encoding

import (
	"cosmossdk.io/simapp/params"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"

	evmv1 "github.com/evmos/ethermint/api/ethermint/evm/v1"
	enccodec "github.com/evmos/ethermint/encoding/codec"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// MakeConfig creates an EncodingConfig
func MakeConfig() ethermint.EncodingConfig {
	amn := codec.NewLegacyAmino()
	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
	}
	// evm/MsgEthereumTx
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&evmv1.MsgEthereumTx{}), evmtypes.GetSignersFromMsgEthereumTxV2)
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	if err != nil {
		panic(err)
	}
	// TODO(zsystm): We already defined custom GetSigners for MsgEthereumTx, should we remove this?
	if err := interfaceRegistry.SigningContext().Validate(); err != nil {
		panic(err)
	}
	cdc := codec.NewProtoCodec(interfaceRegistry)
	encodingConfig := ethermint.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
		TxConfig:          authtx.NewTxConfig(cdc, authtx.DefaultSignModes),
		Amino:             amn,
	}
	enccodec.RegisterLegacyAminoCodec(amn)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// TODO(zsystm): Makre sure we need this line
	// This is needed for the EIP712 txs because currently is using
	// the deprecated method legacytx.StdSignBytes
	legacytx.RegressionTestingAminoCodec = amn
	return encodingConfig
}

// MakeTestEncodingConfig creates an EncodingConfig for testing
func MakeTestEncodingConfig(modules ...module.AppModuleBasic) params.EncodingConfig {
	cdc := codec.NewLegacyAmino()

	signingOptions := signing.Options{
		AddressCodec:          address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix()},
		ValidatorAddressCodec: address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix()},
	}

	// evm/MsgEthereumTx
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&evmv1.MsgEthereumTx{}), evmtypes.GetSignersFromMsgEthereumTxV2)

	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	codec := codec.NewProtoCodec(interfaceRegistry)

	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          authtx.NewTxConfig(codec, authtx.DefaultSignModes),
		Amino:             cdc,
	}

	mb := module.NewBasicManager(modules...)

	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	mb.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	mb.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
