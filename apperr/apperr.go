package apperr

import "errors"

var (
    ErrInsufficientStock = errors.New("INSUFFICIENT_STOCK")
    ErrNotFound          = errors.New("ITEM_NOT_FOUND")
    ErrValidation        = errors.New("VALIDATION_ERROR")
)
