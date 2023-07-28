package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type SecretMsg struct {
	CodeHash []byte
	Msg      []byte
}

func NewSecretMsg(codeHash []byte, msg []byte) SecretMsg {
	return SecretMsg{
		CodeHash: codeHash,
		Msg:      msg,
	}
}

func (m SecretMsg) Serialize() []byte {
	return append(m.CodeHash, m.Msg...)
}

func (msg *MsgExecuteContract) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(msg.Contract); err != nil {
		return err
	}

	if !msg.SentFunds.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "sentFunds") //nolint
	}

	return nil
}

func (msg *MsgExecuteContract) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
