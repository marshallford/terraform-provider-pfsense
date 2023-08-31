package pfsense

import (
	"errors"
	"fmt"
)

var (
	ErrLoginFailed           = errors.New("login failed")
	ErrNotManagedByClient    = fmt.Errorf("not managed by %s client", clientName)
	ErrMissingField          = errors.New("missing field")
	ErrNotFound              = errors.New("not found")
	ErrUnableToParse         = errors.New("unable to parse")
	ErrUnableToScrapeHTML    = errors.New("unable to scrape HTML")
	ErrResponse              = errors.New("response error")
	ErrServerValidation      = errors.New("server validation")
	ErrGetOperationFailed    = errors.New("failed to get")
	ErrCreateOperationFailed = errors.New("failed to create")
	ErrUpdateOperationFailed = errors.New("failed to update")
	ErrDeleteOperationFailed = errors.New("failed to delete")
)
