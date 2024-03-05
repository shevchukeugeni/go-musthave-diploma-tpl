package types

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrOrderAlreadyCreatedByUser = errors.New("order already registered by user")
var ErrOrderAlreadyCreatedByAnother = errors.New("order already registered by another user")
