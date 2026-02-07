package test_config_utils

import (
	"testing"

	"github.com/pratyushtiwary/toki/config"
	test_path_utils "github.com/pratyushtiwary/toki/testutils/path"
	"github.com/stretchr/testify/assert"
)

func ParseTestConfig(configFile string) []*config.PipelineStepConfig {
	pipelineConfig, err := config.NewPipelineConfig(test_path_utils.GetTestFilePath(configFile), nil, nil)

	if err != nil {
		panic(err)
	}

	return pipelineConfig
}

func ParseTestConfigWithProjectConfig(configFile string, projectConfigFile string) ([]*config.PipelineStepConfig, config.ProjectConfigInterface) {
	projectConfig, err := config.NewProjectConfig(test_path_utils.GetTestFilePath(projectConfigFile))

	if err != nil {
		panic(err)
	}

	pipelineConfig, err := config.NewPipelineConfig(test_path_utils.GetTestFilePath(configFile), projectConfig, nil)

	if err != nil {
		panic(err)
	}

	return pipelineConfig, projectConfig
}

func GetNStep(t *testing.T, pipelineConfigPath string, projectConfigPath *string, n int) (*config.PipelineStepConfig, config.ProjectConfigInterface) {
	var pipelineConfig []*config.PipelineStepConfig
	var projectConfig config.ProjectConfigInterface
	if projectConfigPath != nil {
		pipelineConfig, projectConfig = ParseTestConfigWithProjectConfig(pipelineConfigPath, *projectConfigPath)
		assert.NotNil(t, projectConfig)
		assert.IsType(t, &config.ProjectConfig{}, projectConfig)
	} else {
		pipelineConfig = ParseTestConfig(pipelineConfigPath)
		assert.NotNil(t, pipelineConfig)
	}

	pipelineStep := pipelineConfig[n]

	return pipelineStep, projectConfig
}

func GetFirstStep(t *testing.T, pipelineConfigPath string, projectConfigPath *string) (*config.PipelineStepConfig, config.ProjectConfigInterface) {
	return GetNStep(t, pipelineConfigPath, projectConfigPath, 0)
}
