package config

import (
	"encoding/json"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type PipelineAuthParamsConfig struct {
	Expiry       *int   `json:"expiry,omitempty"`
	RefreshAfter *int   `json:"refresh_after,omitempty"`
	Command      string `json:"command"`
}

type StepAuthConfig struct {
	StrategyName string                   `json:"strategy"`
	Params       PipelineAuthParamsConfig `json:"params"`
}

type PipelineStepConfig struct {
	Command      string           `json:"command"`
	AuthCommands []StepAuthConfig `json:"auth"`
}

func NewPipelineConfig(filepath string) ([]PipelineStepConfig, error) {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	var steps []PipelineStepConfig

	if strings.HasSuffix(filepath, "yml") || strings.HasSuffix(filepath, "yaml") {
		err = yaml.Unmarshal(data, &steps)
	} else {
		err = json.Unmarshal(data, &steps)
	}

	if err != nil {
		return nil, err
	}

	return steps, nil
}
