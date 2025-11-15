package strategy

import (
	"errors"

	"github.com/pratyushtiwary/toki/config"
)

func NewStrategy(strategyName string, params config.PipelineAuthParamsConfig, processGroupId *uintptr) (Strategy, error) {
	switch strategyName {
	case "custom":
		config, err := NewCustomStrategyConfig(params, processGroupId)

		if err != nil {
			return nil, err
		}
		return NewCustomStrategy(config)
	default:
		return nil, errors.New("invalid strategy name provided: " + strategyName)
	}
}
