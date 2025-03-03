package pfsense

import (
	"context"
	"fmt"
)

func (pf *Client) ExecutePHPCommand(ctx context.Context, command string, crud string) (any, error) {
	var executeErr error

	switch crud {
	case "create":
		executeErr = ErrCreateOperationFailed
		defer pf.write(&pf.mutexes.ExecutePHPCommand)()
	case "read":
		executeErr = ErrGetOperationFailed
		defer pf.read(&pf.mutexes.ExecutePHPCommand)()
	case "update":
		executeErr = ErrUpdateOperationFailed
		defer pf.write(&pf.mutexes.ExecutePHPCommand)()
	case "delete":
		executeErr = ErrDeleteOperationFailed
		defer pf.write(&pf.mutexes.ExecutePHPCommand)()
	default:
		return nil, fmt.Errorf("%w, invalid CRUD option", ErrClientValidation)
	}

	var result any
	if err := pf.executePHPCommand(ctx, command, &result); err != nil {
		return nil, fmt.Errorf("%w, %w", executeErr, err)
	}

	return result, nil
}
