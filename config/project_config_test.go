package config

import (
	"errors"
	"os"
	"testing"

	test_path_utils "github.com/pratyushtiwary/toki/testutils/path"
	"github.com/stretchr/testify/assert"
)

func TestProjectConfigJsonHappyPath(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)
	assert.IsType(t, &ProjectConfig{}, projectConfig)

	projectConfigStep, err := projectConfig.GetStep("test")

	assert.Nil(t, err)
	assert.Equal(t, projectConfigStep.Command, "echo 1")
	assert.Equal(t, len(projectConfigStep.AuthCommands), 1)
	assert.Equal(t, projectConfigStep.AuthCommands[0].Name, "test")
	assert.Equal(t, projectConfigStep.AuthCommands[0].StrategyName, "custom")
}

func TestProjectConfigInvalidJson(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("brokenProjectConfig.json"))

	assert.NotNil(t, err)
	assert.Nil(t, projectConfig)
}

func TestProjectConfigYamlHappyPath(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.yml"))

	assert.Nil(t, err)
	assert.IsType(t, &ProjectConfig{}, projectConfig)

	projectConfigStep, err := projectConfig.GetStep("test")

	assert.Nil(t, err)
	assert.Equal(t, projectConfigStep.Command, "echo 1")
	assert.Equal(t, len(projectConfigStep.AuthCommands), 1)
	assert.Equal(t, projectConfigStep.AuthCommands[0].Name, "test")
	assert.Equal(t, projectConfigStep.AuthCommands[0].StrategyName, "custom")
}

func TestProjectConfigInvalidYaml(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("brokenProjectConfig.yml"))

	assert.NotNil(t, err)
	assert.Nil(t, projectConfig)
}

func TestProjectConfigEmptyPath(t *testing.T) {
	projectConfig, err := NewProjectConfig("")

	assert.Nil(t, err)
	assert.Nil(t, projectConfig)
}
func TestProjectConfigInvalidFilepath(t *testing.T) {
	projectConfig, err := NewProjectConfig("test.xyz")

	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))
	assert.Nil(t, projectConfig)
}

func TestProjectGetStepMissingStep(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)
	assert.IsType(t, &ProjectConfig{}, projectConfig)

	projectStepConfig, err := projectConfig.GetStep("xyz")
	expectedError := "`xyz` step doesn't exist in project config"

	assert.NotNil(t, err)
	assert.Nil(t, projectStepConfig)
	assert.Equal(t, err.Error(), expectedError)
}

func TestValidateProjectStepConfigEmptyCommand(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)
	assert.IsType(t, &ProjectConfig{}, projectConfig)

	projectStepConfig, err := projectConfig.GetStep("test")

	projectStepConfig.Command = ""
	assert.Nil(t, err)
	assert.Equal(t, projectStepConfig.ValidateProjectStepConfig().Error(), "`command` needs to be present in project step")
}
func TestAddAuthCommandEmptyName(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)
	assert.IsType(t, &ProjectConfig{}, projectConfig)

	projectStepConfig, err := projectConfig.GetStep("test")

	assert.Nil(t, err)

	err = projectStepConfig.AddAuthCommand(&StepAuthConfig{})

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "`name` missing from auth config")
}
func TestAddStepCommandEmptyName(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)
	assert.IsType(t, &ProjectConfig{}, projectConfig)

	authCommands := make([]*StepAuthConfig, 0)
	authCommands = append(authCommands, &StepAuthConfig{})

	err = projectConfig.AddStep(&ProjectStepConfig{
		Name: "test",
		StepConfig: StepConfig{
			AuthCommands: authCommands,
		},
	})

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "`name` missing from auth config")
}
