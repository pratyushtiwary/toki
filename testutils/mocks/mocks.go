package test_mock_utils

import (
	"errors"

	"github.com/pratyushtiwary/toki/process"
)

func MockNewCommand(authCommand process.ProcessInterface) func() {
	var originalNewCommand = process.NewCommand
	process.NewCommand = func(execCommand string, parentProcessGroup *uintptr) (process.ProcessInterface, error) {
		if authCommand == nil {
			return nil, errors.New("test error")
		}
		return authCommand, nil
	}
	var resetNewCommand = func() {
		process.NewCommand = originalNewCommand
	}

	return resetNewCommand
}
