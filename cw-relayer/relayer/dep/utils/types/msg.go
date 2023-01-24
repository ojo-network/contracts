package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	MaxCertificateSize = 20 * 1024
)

type PublicKey []byte

func (msg RaAuthenticate) Type() string {
	return "node-auth"
}

func (msg RaAuthenticate) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return err
	}

	if len(msg.Certificate) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Authenticating certificate cannot be empty")
	}

	if len(msg.Certificate) > MaxCertificateSize {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "certificate length too large")
	}

	return nil
}

func (msg RaAuthenticate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
