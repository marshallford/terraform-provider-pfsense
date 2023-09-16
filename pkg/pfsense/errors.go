package pfsense

import (
	"errors"
)

var (
	ErrFailedRequest         = errors.New("failed request")
	ErrHTTPStatus            = errors.New("HTTP status")
	ErrLoginFailed           = errors.New("login failed")
	ErrNotFound              = errors.New("not found")
	ErrUnableToParse         = errors.New("unable to parse")
	ErrUnableToScrapeHTML    = errors.New("unable to scrape HTML")
	ErrClientValidation      = errors.New("client validation")
	ErrServerValidation      = errors.New("server validation")
	ErrGetOperationFailed    = errors.New("failed to get")
	ErrCreateOperationFailed = errors.New("failed to create")
	ErrUpdateOperationFailed = errors.New("failed to update")
	ErrDeleteOperationFailed = errors.New("failed to delete")
)
