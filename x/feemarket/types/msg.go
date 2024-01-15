package types

import (
	"fmt"

	protov2 "google.golang.org/protobuf/proto"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	feemarketapi "github.com/evmos/ethermint/api/ethermint/feemarket/v1"
)

var _ sdk.Msg = &MsgUpdateParams{}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check of the provided data
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	return m.Params.Validate()
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

func GetSignersFromMsgUpdateParamsV2(msg protov2.Message) ([][]byte, error) {
	msgv1, err := GetMsgUpdateParamsFromMsgV2(msg)
	if err != nil {
		return nil, err
	}

	signers := [][]byte{}
	for _, signer := range msgv1.GetSigners() {
		signers = append(signers, signer.Bytes())
	}

	return signers, nil
}

func GetMsgUpdateParamsFromMsgV2(msg protov2.Message) (MsgUpdateParams, error) {
	msgv2, ok := msg.(*feemarketapi.MsgUpdateParams)
	if !ok {
		return MsgUpdateParams{}, fmt.Errorf("invalid x/feemarket/MsgUpdateParams msg v2: %v", msg)
	}

	msgv1 := MsgUpdateParams{
		Authority: msgv2.Authority,
	}

	return msgv1, nil
}
