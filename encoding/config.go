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
	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"

	evmv1 "github.com/evmos/ethermint/api/ethermint/evm/v1"
	feemarketv1 "github.com/evmos/ethermint/api/ethermint/feemarket/v1"
	enccodec "github.com/evmos/ethermint/encoding/codec"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
)

// MakeConfig creates an EncodingConfig for testing
func MakeConfig(mb module.BasicManager) params.EncodingConfig {
	cdc := amino.NewLegacyAmino()

	signingOptions := signing.Options{
		AddressCodec:          address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix()},
		ValidatorAddressCodec: address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix()},
	}

	// evm/MsgEthereumTx, evm/MsgUpdateParams, feemarket/MsgUpdateParams // TODO(dudong2)
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&evmv1.MsgEthereumTx{}), func(msg protov2.Message) ([][]byte, error) {
		msgEthereumTx, ok := msg.(*evmv1.MsgEthereumTx)
		if !ok {
			return nil, nil
		}

		if len(msgEthereumTx.From) > 0 {
			return [][]byte{common.HexToAddress(msgEthereumTx.From).Bytes()}, nil
		}

		var dataAny *types.Any
		var err error
		switch msgEthereumTx.Data.TypeUrl {
		case "/ethermint.evm.v1.LegacyTx":
			legacyTx := evmtypes.LegacyTx{}
			evmtypes.ModuleCdc.MustUnmarshal(msgEthereumTx.Data.Value, &legacyTx)
			dataAny, err = evmtypes.PackTxData(&legacyTx)
			if err != nil {
				return nil, err
			}
		case "/ethermint.evm.v1.DynamicFeeTx":
			dynamicFeeTx := evmtypes.DynamicFeeTx{}
			evmtypes.ModuleCdc.MustUnmarshal(msgEthereumTx.Data.Value, &dynamicFeeTx)
			dataAny, err = evmtypes.PackTxData(&dynamicFeeTx)
			if err != nil {
				return nil, err
			}
		case "/ethermint.evm.v1.AccessListTx":
			accessListTx := evmtypes.AccessListTx{}
			evmtypes.ModuleCdc.MustUnmarshal(msgEthereumTx.Data.Value, &accessListTx)
			dataAny, err = evmtypes.PackTxData(&accessListTx)
			if err != nil {
				return nil, err
			}
		}
		t := evmtypes.MsgEthereumTx{Data: dataAny}
		t.Hash = t.AsTransaction().Hash().Hex()

		signers := [][]byte{}
		for _, signer := range t.GetSigners() {
			signers = append(signers, signer.Bytes())
		}

		return signers, nil
	})

	signingOptions.DefineCustomGetSigners(protov2.MessageName(&evmv1.MsgUpdateParams{}), func(msg protov2.Message) ([][]byte, error) {
		msgUpdateParams, ok := msg.(*evmv1.MsgUpdateParams)
		if !ok {
			return nil, nil
		}

		t := evmtypes.MsgUpdateParams{
			Authority: msgUpdateParams.Authority,
		}

		signers := [][]byte{}
		for _, signer := range t.GetSigners() {
			signers = append(signers, signer.Bytes())
		}

		return signers, nil
	})

	signingOptions.DefineCustomGetSigners(protov2.MessageName(&feemarketv1.MsgUpdateParams{}), func(msg protov2.Message) ([][]byte, error) {
		msgUpdateParams, ok := msg.(*feemarketv1.MsgUpdateParams)
		if !ok {
			return nil, nil
		}

		t := evmtypes.MsgUpdateParams{
			Authority: msgUpdateParams.Authority,
		}

		signers := [][]byte{}
		for _, signer := range t.GetSigners() {
			signers = append(signers, signer.Bytes())
		}

		return signers, nil
	})

	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	codec := amino.NewProtoCodec(interfaceRegistry)

	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
		Amino:             cdc,
	}

	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	mb.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	mb.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
