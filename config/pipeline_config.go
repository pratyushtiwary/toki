package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pratyushtiwary/toki/log"
	"gopkg.in/yaml.v3"
)

type PipelineStepConfig struct {
	Use               string `json:"use,omitempty" yaml:"use,omitempty"`
	StepConfig        `yaml:",inline"`
	projectStepConfig *ProjectStepConfig
}

func (pSC *PipelineStepConfig) GetProjectStepConfig() *ProjectStepConfig {
	return pSC.projectStepConfig
}

func (sAC *StepAuthConfig) IsEqual(stepAuthConfig StepAuthConfig) bool {
	if sAC.StrategyName != stepAuthConfig.StrategyName {
		return false
	}

	if sAC.Params.Command != stepAuthConfig.Params.Command {
		return false
	}

	return true
}

func (pSC *PipelineStepConfig) ValidatePipelineStepConfig(projectConfig *ProjectConfig) error {

	if len(pSC.Use) == 0 && len(pSC.Command) == 0 {
		return errors.New("either of `use` or `command` needs to be present in the pipeline step")
	}

	if len(pSC.Use) != 0 {
		if projectConfig == nil {
			if len(pSC.Command) > 0 {
				log.Warn("`use` would be skipped as no project config is supplied")
			} else {
				return errors.New("`command` needs to be present since project config is not supplied and this step uses `use`")
			}
		}

		if projectConfig != nil {
			_, err := projectConfig.GetStep(pSC.Use)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (pSC *PipelineStepConfig) MergeWithProjectConfig(projectConfig *ProjectConfig) error {
	if projectConfig == nil || len(pSC.Use) == 0 {
		return nil
	}

	// validate self
	pSC.ValidatePipelineStepConfig(projectConfig)

	projectConfigStepName := pSC.Use
	projectConfigStep, err := projectConfig.GetStep(projectConfigStepName)

	if err != nil {
		return err
	}

	// validate project config step
	projectConfigStep.ValidateProjectStepConfig()

	// pipeline config attributes get more preference

	if len(pSC.Command) == 0 {
		pSC.Command = projectConfigStep.Command
	}

	if len(pSC.AuthCommands) == 0 {
		pSC.AuthCommands = projectConfigStep.AuthCommands
	}

	pSC.projectStepConfig = projectConfigStep // this is done so that strategies can also perform lookup on project step config if needed

	return nil
}

func NewPipelineConfig(filepath string, projectConfig *ProjectConfig) ([]*PipelineStepConfig, error) {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	var steps []*PipelineStepConfig

	if strings.HasSuffix(filepath, "yml") || strings.HasSuffix(filepath, "yaml") {
		err = yaml.Unmarshal(data, &steps)
	} else {
		err = json.Unmarshal(data, &steps)
	}

	if err != nil {
		return nil, err
	}

	for stepIdx, step := range steps {
		err := step.ValidatePipelineStepConfig(projectConfig)

		if err != nil {
			return nil, fmt.Errorf("step %d failed validation, error: %s", stepIdx+1, err.Error())
		}

		err = step.MergeWithProjectConfig(projectConfig)
		if err != nil {
			return nil, fmt.Errorf("step %d failed validation, error: %s", stepIdx+1, err.Error())
		}
	}

	return steps, nil
}
