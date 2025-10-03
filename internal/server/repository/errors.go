package repository

import "errors"

// ErrVersionConflict indicates optimistic lock failure on update.
var ErrVersionConflict = errors.New("version conflict")
