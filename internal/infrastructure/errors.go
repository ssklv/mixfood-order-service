package infrastructure

import "errors"

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrFailedToBeginTx = errors.New("failed to start database transaction")
)
