package config

import "errors"

func ValidateInheritStrategyConfig(config *StepAuthConfig) error {
	if len(config.Name) == 0 {
		return errors.New("no `name` provided for inherit strategy to inherit")
	}

	return nil
}
