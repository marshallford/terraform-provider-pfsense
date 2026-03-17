package pfsense

import (
	"context"
	"fmt"
)

type ExecutePHPCommand struct{}

func (ExecutePHPCommand) Privileges() Privileges {
	return Privileges{
		Create: []string{PrivDiagnosticsCommand}}
}

func (pf *Client) ExecutePHPCommand(ctx context.Context, command string, write bool) (any, error) {
	if write {
		defer pf.write(&pf.mutexes.ExecutePHPCommand)()
	} else {
		defer pf.read(&pf.mutexes.ExecutePHPCommand)()
	}

	var result any
	if err := pf.executePHPCommand(ctx, command, &result); err != nil {
		return nil, fmt.Errorf("%w, %w", ErrExecOperationFailed, err)
	}

	return result, nil
}
