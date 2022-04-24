package helpers

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

func UniqueViolationError(err error) error {
	if driverErr, ok := err.(*pq.Error); ok {
		if pgerrcode.UniqueViolation == driverErr.Code {
			return errors.New("not uniquie url")
		}
	}
	return err
}
