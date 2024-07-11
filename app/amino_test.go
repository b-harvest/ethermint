package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-proto/rapidproto"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/signing/aminojson"
	signing_testutil "cosmossdk.io/x/tx/signing/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	evmv1 "github.com/evmos/ethermint/api/ethermint/evm/v1"
	feemarketv1 "github.com/evmos/ethermint/api/ethermint/feemarket/v1"
	"github.com/evmos/ethermint/x/evm"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/ethermint/x/feemarket"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

// TestAminoJSON_Equivalence tests that x/tx/Encoder encoding is equivalent to the legacy Encoder encoding.
// A custom generator is used to generate random messages that are then encoded using both encoders.  The custom
// generator only supports proto.Message (which implement the protoreflect API) so in order to test legacy gogo types
// we end up with a workflow as follows:
//
// 1. Generate a random protobuf proto.Message using the custom generator
// 2. Marshal the proto.Message to protobuf binary bytes
// 3. Unmarshal the protobuf bytes to a gogoproto.Message
// 4. Marshal the gogoproto.Message to amino JSON bytes
// 5. Marshal the proto.Message to amino JSON bytes
// 6. Compare the amino JSON bytes from steps 4 and 5
//
// In order for step 3 to work certain restrictions on the data generated in step 1 must be enforced and are described
// by the mutation of genOpts passed to the generator.
func TestAminoJSON_Equivalence(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(
		evm.AppModuleBasic{},
		feemarket.AppModuleBasic{})
	legacytx.RegressionTestingAminoCodec = encCfg.Amino
	aj := aminojson.NewEncoder(aminojson.EncoderOptions{DoNotSortFields: true})

	GenOpts := rapidproto.GeneratorOptions{
		Resolver:  protoregistry.GlobalTypes,
		FieldMaps: []rapidproto.FieldMapper{GeneratorFieldMapper},
	}

	testedMsgs := []GeneratedType{
		// evm
		// the type of evmtypes.MsgEthereumTx.Size_ is float64 that is not supported at aminojson encoder,
		// so comment MsgEthereumTx test.
		// GenType(&evmtypes.MsgEthereumTx{}, &evmv1.MsgEthereumTx{}, GenOpts.WithDisallowNil()),

		// evm
		GenType(&evmtypes.MsgUpdateParams{}, &evmv1.MsgUpdateParams{}, GenOpts.WithDisallowNil()),
		GenType(&evmtypes.Params{}, &evmv1.Params{}, GenOpts.WithDisallowNil()),

		// feemarket
		GenType(&feemarkettypes.MsgUpdateParams{}, &feemarketv1.MsgUpdateParams{}, GenOpts.WithDisallowNil()),
		GenType(&feemarkettypes.Params{}, &feemarketv1.Params{}, GenOpts.WithDisallowNil()),
	}

	for _, tt := range testedMsgs {
		desc := tt.Pulsar.ProtoReflect().Descriptor()
		name := string(desc.FullName())
		t.Run(name, func(t *testing.T) {
			gen := rapidproto.MessageGenerator(tt.Pulsar, tt.Opts)
			fmt.Printf("testing %s\n", tt.Pulsar.ProtoReflect().Descriptor().FullName())
			rapid.Check(t, func(t *rapid.T) {
				// uncomment to debug; catch a panic and inspect application state
				// defer func() {
				// 	if r := recover(); r != nil {
				// 		// fmt.Printf("Panic: %+v\n", r)
				// 		t.FailNow()
				// 	}
				// }()

				msg := gen.Draw(t, "msg")
				postFixPulsarMessage(msg)

				gogo := tt.Gogo
				sanity := tt.Pulsar

				protoBz, err := proto.Marshal(msg)
				require.NoError(t, err)

				err = proto.Unmarshal(protoBz, sanity)
				require.NoError(t, err)

				err = encCfg.Codec.Unmarshal(protoBz, gogo)
				require.NoError(t, err)

				legacyAminoJSON, err := encCfg.Amino.MarshalJSON(gogo)
				require.NoError(t, err)
				aminoJSON, err := aj.Marshal(msg)
				require.NoError(t, err)
				require.Equal(t, string(legacyAminoJSON), string(aminoJSON))

				// test amino json signer handler equivalence
				if !proto.HasExtension(desc.Options(), msgv1.E_Signer) {
					// not signable
					return
				}

				handlerOptions := signing_testutil.HandlerArgumentOptions{
					ChainID:       "test-chain",
					Memo:          "sometestmemo",
					Msg:           tt.Pulsar,
					AccNum:        1,
					AccSeq:        2,
					SignerAddress: "signerAddress",
					Fee: &txv1beta1.Fee{
						Amount: []*v1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
					},
				}

				signerData, txData, err := signing_testutil.MakeHandlerArguments(handlerOptions)
				require.NoError(t, err)

				handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
				signBz, err := handler.GetSignBytes(context.Background(), signerData, txData)
				require.NoError(t, err)

				legacyHandler := tx.NewSignModeLegacyAminoJSONHandler()
				txBuilder := encCfg.TxConfig.NewTxBuilder()
				require.NoError(t, txBuilder.SetMsgs([]types.Msg{tt.Gogo}...))
				txBuilder.SetMemo(handlerOptions.Memo)
				txBuilder.SetFeeAmount(types.Coins{types.NewInt64Coin("uatom", 1000)})
				theTx := txBuilder.GetTx()

				legacySigningData := signing.SignerData{
					ChainID:       handlerOptions.ChainID,
					Address:       handlerOptions.SignerAddress,
					AccountNumber: handlerOptions.AccNum,
					Sequence:      handlerOptions.AccSeq,
				}
				legacySignBz, err := legacyHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					legacySigningData, theTx)
				require.NoError(t, err)
				require.Equal(t, string(legacySignBz), string(signBz))
			})
		})
	}
}

func postFixPulsarMessage(msg proto.Message) {
	if m, ok := msg.(*authapi.ModuleAccount); ok {
		if m.BaseAccount == nil {
			m.BaseAccount = &authapi.BaseAccount{}
		}
		_, _, bz := testdata.KeyTestPubAddr()
		// always set address to a valid bech32 address
		text, _ := bech32.ConvertAndEncode("cosmos", bz)
		m.BaseAccount.Address = text

		// see negative test
		if len(m.Permissions) == 0 {
			m.Permissions = nil
		}
	}

	if m, ok := msg.(*evmv1.MsgUpdateParams); ok {
		m.Params.ChainConfig.CancunBlock = "10"
	}

	if m, ok := msg.(*evmv1.Params); ok {
		m.ChainConfig.CancunBlock = "10"
	}

	if m, ok := msg.(*feemarketv1.MsgUpdateParams); ok {
		m.Params.BaseFee = "10"
	}

	if m, ok := msg.(*feemarketv1.Params); ok {
		m.BaseFee = "10"
	}
}
