package strategy

import (
	"testing"

	"github.com/pratyushtiwary/toki/test"
	"github.com/stretchr/testify/assert"
)

func stringPtr(s string) *string {
	return &s
}

func TestRegistryCustomStrategyHappyPath(t *testing.T) {
	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	processor, config, err := NewStrategy(authConfig.StrategyName, authConfig, nil, nil)

	assert.Equal(t, authConfig, config)
	assert.Nil(t, err)
	assert.IsType(t, &CustomStrategy{}, processor)
}

func TestRegistryCustomStrategyInvalidConfig(t *testing.T) {
	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	authConfig.Params.Command = ""

	processor, config, err := NewStrategy(authConfig.StrategyName, authConfig, nil, nil)

	assert.Nil(t, processor)
	assert.Nil(t, config)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "custom strategy: command is missing")
}

func TestRegistryInheritStrategyHappyPath(t *testing.T) {
	inheritStrategyConfig, projectConfig := test.GetFirstStep(
		t,
		"inheritPipelineConfig.json",
		stringPtr("projectConfig.json"),
	)
	authConfig := inheritStrategyConfig.AuthCommands[0]

	projectConfigStep, err := projectConfig.GetStep(authConfig.Name)
	assert.Nil(t, err)

	processor, config, err := NewStrategy(authConfig.StrategyName, authConfig, nil, projectConfigStep)

	assert.Equal(t, config.StrategyName, "custom")
	assert.Nil(t, err)
	assert.IsType(t, &CustomStrategy{}, processor) // InheritStrategy does a recursive call on registry, so we should be getting instance of `inherited` strategy and not inherit strategy itself
}

func TestRegistryInheritStrategyInvalidConfig(t *testing.T) {
	inheritStrategyConfig, projectConfig := test.GetFirstStep(
		t,
		"inheritPipelineConfig.json",
		stringPtr("projectConfig.json"),
	)
	authConfig := inheritStrategyConfig.AuthCommands[0]

	projectConfigStep, err := projectConfig.GetStep(authConfig.Name)
	assert.Nil(t, err)

	authConfig.Name = ""
	processor, config, err := NewStrategy(authConfig.StrategyName, authConfig, nil, projectConfigStep)

	assert.Nil(t, processor)
	assert.Nil(t, config)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "no `name` provided for inherit strategy to inherit")
}

func TestRegistryInvalidStrategy(t *testing.T) {
	invalidStrategyConfig, _ := test.GetFirstStep(
		t,
		"unsupportedStrategyPipelineConfig.json",
		nil,
	)
	authConfig := invalidStrategyConfig.AuthCommands[0]

	processor, config, err := NewStrategy(authConfig.StrategyName, authConfig, nil, nil)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid strategy name provided: "+authConfig.StrategyName)
	assert.Nil(t, processor)
	assert.Nil(t, config)
}
