package strategy

import (
	"testing"

	"github.com/pratyushtiwary/toki/test"
	"github.com/stretchr/testify/assert"
)

func TestInheritStrategyHappyPath(t *testing.T) {
	inheritStrategyConfig, projectConfig := test.GetFirstStep(
		t,
		"inheritPipelineConfig.json",
		stringPtr("projectConfig.json"),
	)
	authConfig := inheritStrategyConfig.AuthCommands[0]

	projectStepConfig, err := projectConfig.GetStep(authConfig.Name)

	assert.Nil(t, err)

	inheritStrategy, _config, err := NewInheritStrategy(authConfig, projectStepConfig)

	assert.Nil(t, err)
	assert.IsType(t, &CustomStrategy{}, inheritStrategy)
	assert.Equal(t, _config.StrategyName, projectStepConfig.AuthCommands[0].StrategyName)
}

func TestInheritStrategyWithInvalidInheritName(t *testing.T) {
	inheritStrategyConfig, projectConfig := test.GetFirstStep(
		t,
		"inheritPipelineConfig.json",
		stringPtr("projectConfig.json"),
	)
	authConfig := inheritStrategyConfig.AuthCommands[0]

	projectStepConfig, err := projectConfig.GetStep(authConfig.Name)

	assert.Nil(t, err)

	authConfig.Name = "invalid"

	inheritStrategy, _config, err := NewInheritStrategy(authConfig, projectStepConfig)

	_, expectedError := projectStepConfig.GetAuthCommand("invalid")

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedError.Error())
	assert.Nil(t, inheritStrategy)
	assert.Nil(t, _config)
}

func TestInheritStrategyWithInvalidProjectStepConfig(t *testing.T) {
	inheritStrategyConfig, projectConfig := test.GetFirstStep(
		t,
		"inheritPipelineConfig.json",
		stringPtr("projectConfig.json"),
	)
	authConfig := inheritStrategyConfig.AuthCommands[0]

	projectStepConfig, err := projectConfig.GetStep(authConfig.Name)

	assert.Nil(t, err)

	projectStepConfig.AuthCommands[0].Params.Command = ""

	inheritStrategy, _config, err := NewInheritStrategy(authConfig, projectStepConfig)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "custom strategy: command is missing")
	assert.Nil(t, inheritStrategy)
	assert.Nil(t, _config)
}
