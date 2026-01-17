package config

import (
	"errors"
	"os"
	"testing"

	test_path_utils "github.com/pratyushtiwary/toki/testutils/path"
	"github.com/stretchr/testify/assert"
)

func TestNewPipelineConfigHappyPath(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)

	// without project config - json
	pipelineConfig, err := NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.json"), nil, nil)

	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)
	assert.Nil(t, pipelineConfig[0].projectStepConfig)

	// with project config - json
	pipelineConfig, err = NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.json"), projectConfig, nil)
	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)
	assert.NotNil(t, pipelineConfig[0].projectStepConfig)

	// without project config - yaml
	pipelineConfig, err = NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.yml"), nil, nil)

	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)
	assert.Nil(t, pipelineConfig[0].projectStepConfig)

	// with project config - yaml
	pipelineConfig, err = NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.yml"), projectConfig, nil)
	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)
	assert.NotNil(t, pipelineConfig[0].projectStepConfig)
}

func TestNewPipelineConfigInvalidFilepath(t *testing.T) {
	pipelineConfig, err := NewPipelineConfig("test.xyz", nil, nil)

	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))
	assert.Nil(t, pipelineConfig)
}

func TestNewPipelineConfigInvalidYaml(t *testing.T) {
	pipelineConfig, err := NewPipelineConfig(test_path_utils.GetTestFilePath("brokenInheritPipelineConfig.yml"), nil, nil)

	assert.NotNil(t, err)
	assert.Nil(t, pipelineConfig)
}

func TestNewPipelineConfigInvalidJson(t *testing.T) {
	pipelineConfig, err := NewPipelineConfig(test_path_utils.GetTestFilePath("brokenInheritPipelineConfig.yml"), nil, nil)

	assert.NotNil(t, err)
	assert.Nil(t, pipelineConfig)
}

func TestPipelineStepConfigMergeWithProjectConfig(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)

	projectStepConfig, err := projectConfig.GetStep("test")

	assert.Nil(t, err)

	pipelineConfig, err := NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.json"), nil, nil)

	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)

	pipelineStepConfig := pipelineConfig[0]

	assert.Nil(t, pipelineStepConfig.MergeWithProjectConfig(projectConfig))

	pipelineStepConfig.Use = "test2"
	pipelineStepConfig.Command = ""
	assert.Equal(t, pipelineStepConfig.MergeWithProjectConfig(projectConfig).Error(), "`test2` step doesn't exist in project config")

	pipelineStepConfig.Use = "test"
	pipelineStepConfig.Command = ""
	assert.Nil(t, pipelineStepConfig.MergeWithProjectConfig(projectConfig))

	pipelineStepConfig.AuthCommands = make([]*StepAuthConfig, 0)
	assert.Nil(t, pipelineStepConfig.MergeWithProjectConfig(projectConfig))
	assert.Equal(t, pipelineStepConfig.Command, projectStepConfig.Command)
	assert.Len(t, pipelineStepConfig.AuthCommands, len(projectStepConfig.AuthCommands))
	assert.Equal(t, pipelineStepConfig.AuthCommands[0], projectStepConfig.AuthCommands[0])
}

func TestPipelineStepConfigValidatePipelineStepConfig(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)

	pipelineConfig, err := NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.json"), nil, nil)

	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)

	pipelineStepConfig := pipelineConfig[0]

	assert.Nil(t, pipelineStepConfig.ValidatePipelineStepConfig(nil))
	pipelineStepConfig.Use = ""
	pipelineStepConfig.Command = ""

	assert.Equal(t, pipelineStepConfig.ValidatePipelineStepConfig(nil).Error(), "either of `use` or `command` needs to be present in the pipeline step")

	pipelineStepConfig.Use = "test"
	pipelineStepConfig.Command = ""
	assert.Equal(t, pipelineStepConfig.ValidatePipelineStepConfig(nil).Error(), "`command` needs to be present since project config is not supplied and this step uses `use`")

	pipelineStepConfig.Use = "test2"
	pipelineStepConfig.Command = ""
	assert.Equal(t, pipelineStepConfig.ValidatePipelineStepConfig(projectConfig).Error(), "`test2` step doesn't exist in project config")
}

func TestStepAuthConfigIsEqual(t *testing.T) {
	pipelineConfig, err := NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.json"), nil, nil)

	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)

	pipelineStepConfig := pipelineConfig[0]
	stepAuthConfig := pipelineStepConfig.AuthCommands[0]
	equalStepAuthConfig := *stepAuthConfig
	differentStepAuthConfig := *stepAuthConfig

	assert.True(t, stepAuthConfig.IsEqual(equalStepAuthConfig))

	differentStepAuthConfig.StrategyName = "test"
	assert.False(t, stepAuthConfig.IsEqual(differentStepAuthConfig))

	differentStepAuthConfig.StrategyName = stepAuthConfig.StrategyName
	differentStepAuthConfig.Params.Command = "test"
	assert.False(t, stepAuthConfig.IsEqual(differentStepAuthConfig))
}

func TestGetProjectStepConfig(t *testing.T) {
	projectConfig, err := NewProjectConfig(test_path_utils.GetTestFilePath("projectConfig.json"))

	assert.Nil(t, err)

	pipelineConfig, err := NewPipelineConfig(test_path_utils.GetTestFilePath("inheritPipelineConfig.json"), projectConfig, nil)

	assert.Nil(t, err)
	assert.Len(t, pipelineConfig, 1)

	pipelineStepConfig := pipelineConfig[0]

	expectedProjectStepConfig, _ := projectConfig.GetStep("test")

	assert.Equal(t, pipelineStepConfig.GetProjectStepConfig(), expectedProjectStepConfig)
}
