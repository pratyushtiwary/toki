package config

import (
	"errors"
)

type CustomStrategyConfig struct {
	StrategyConfig
	Expiry             *int `json:"expiry,omitempty" yaml:"expiry,omitempty"`
	RefreshAfter       *int `json:"refresh_after,omitempty" yaml:"refresh_after,omitempty"`
	parentProcessGroup *uintptr
}

func (cSC *CustomStrategyConfig) GetParentProcessGroup() *uintptr {
	return cSC.parentProcessGroup
}

func ValidateCustomStrategyConfig(config *StepAuthConfig, processGroupId *uintptr) error {
	expiry := config.Params.Expiry
	refreshAfter := config.Params.RefreshAfter

	if len(config.Params.Command) == 0 {
		return errors.New("custom strategy: command is missing")
	}

	if expiry == nil {
		return errors.New("custom strategy: expiry is missing")
	}

	if refreshAfter == nil {
		return errors.New("custom strategy: refresh_after is missing")
	}

	if *expiry < 0 {
		return errors.New("custom strategy: invalid value provided for expire, make sure value is positive")
	}

	if *refreshAfter < 0 {
		return errors.New("custom strategy: invalid value provided for refresh_after, make sure value is positive")
	}

	if *refreshAfter >= *expiry {
		return errors.New("custom strategy: refresh_after should be less than expiry")
	}

	return nil
}
