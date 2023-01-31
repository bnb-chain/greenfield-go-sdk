package types

import "errors"

var (
	AddressMissingError   = errors.New("Address is required ")
	AddressNotValidError  = errors.New("Address is not valid ")
	OffsetOutOfRangeError = errors.New("offset out of range ")
	LimitOutOfRangeError  = errors.New("limit out of range ")
	TokenNotSupportError  = errors.New("The Token sending is not support ")
)
