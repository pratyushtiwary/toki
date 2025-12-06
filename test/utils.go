package test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pratyushtiwary/toki/config"
	"github.com/stretchr/testify/assert"
)

func getTestFilePath(relativePath ...string) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	pathParts := append([]string{dir}, relativePath...)
	return filepath.Join(pathParts...)
}

func ParseTestConfig(configFile string) []*config.PipelineStepConfig {
	pipelineConfig, err := config.NewPipelineConfig(getTestFilePath("./resources/", configFile), nil)

	if err != nil {
		panic(err)
	}

	return pipelineConfig
}

func ParseTestConfigWithProjectConfig(configFile string, projectConfigFile string) ([]*config.PipelineStepConfig, *config.ProjectConfig) {
	projectConfig, err := config.NewProjectConfig(getTestFilePath("./resources/", projectConfigFile))

	if err != nil {
		panic(err)
	}

	pipelineConfig, err := config.NewPipelineConfig(getTestFilePath("./resources/", configFile), projectConfig)

	if err != nil {
		panic(err)
	}

	return pipelineConfig, projectConfig
}

func GetFirstStep(t *testing.T, pipelineConfigPath string, projectConfigPath *string) (*config.PipelineStepConfig, *config.ProjectConfig) {
	var pipelineConfig []*config.PipelineStepConfig
	var projectConfig *config.ProjectConfig
	if projectConfigPath != nil {
		pipelineConfig, projectConfig = ParseTestConfigWithProjectConfig(pipelineConfigPath, *projectConfigPath)
		assert.NotNil(t, projectConfig)
		assert.IsType(t, &config.ProjectConfig{}, projectConfig)
	} else {
		pipelineConfig = ParseTestConfig(pipelineConfigPath)
		assert.NotNil(t, pipelineConfig)
	}

	pipelineStep := pipelineConfig[0]

	return pipelineStep, projectConfig
}
