package strategy

import (
	"errors"

	"github.com/pratyushtiwary/toki/config"
)

func NewStrategy(strategyName string, stepConfig *config.StepAuthConfig, processGroupId *uintptr, projectStepConfig *config.ProjectStepConfig) (Strategy, *config.StepAuthConfig, error) {
	switch strategyName {
	case "custom":
		err := config.ValidateCustomStrategyConfig(stepConfig, processGroupId)

		if err != nil {
			return nil, nil, err
		}
		return NewCustomStrategy(stepConfig)
	case "curl_cookie":
		err := config.ValidateCurlCookieStrategyConfig(stepConfig)

		if err != nil {
			return nil, nil, err
		}

		return NewCurlCookieStrategy(stepConfig)
	case "inherit":
		err := config.ValidateInheritStrategyConfig(stepConfig)

		if err != nil {
			return nil, nil, err
		}

		return NewInheritStrategy(stepConfig, projectStepConfig)
	default:
		return nil, nil, errors.New("invalid strategy name provided: " + strategyName)
	}
}
