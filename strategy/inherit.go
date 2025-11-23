package strategy

import "github.com/pratyushtiwary/toki/config"

type InheritStrategy struct {
	Strategy
	Name string
} // noop strategy

func NewInheritStrategy(config *config.StepAuthConfig, projectStepConfig *config.ProjectStepConfig) (Strategy, *config.StepAuthConfig, error) {
	stepAuthConfig, err := projectStepConfig.GetAuthCommand(config.Name)

	if err != nil {
		return nil, nil, err
	}

	strategy, _, err := NewStrategy(stepAuthConfig.StrategyName, stepAuthConfig, nil, projectStepConfig)

	if err != nil {
		return nil, nil, err
	}

	return strategy, stepAuthConfig, nil
}
