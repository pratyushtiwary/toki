package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type ProjectConfigInterface interface {
	AddStep(*ProjectStepConfig) error
	GetStep(string) (*ProjectStepConfig, error)
}

type ProjectConfigStepInterface interface {
	ensureInitialized()
	AddAuthCommand(*StepAuthConfig) error
	GetAuthCommand(string) (*StepAuthConfig, error)
	ValidateProjectStepConfig() error
}

type ProjectStepConfig struct {
	ProjectConfigStepInterface
	StepConfig `yaml:",inline"`
	Name       string `json:"name" yaml:"name"`
	strategies map[string]*StepAuthConfig
	rwLock     sync.RWMutex
}

type ProjectConfig struct {
	ProjectConfigInterface
	steps  map[string]*ProjectStepConfig
	rwLock sync.RWMutex
}

func (pC *ProjectConfig) AddStep(step *ProjectStepConfig) error {
	pC.rwLock.Lock()
	defer pC.rwLock.Unlock()

	pC.steps[step.Name] = step

	for _, stepAuthConfig := range step.AuthCommands {
		err := step.AddAuthCommand(stepAuthConfig)

		if err != nil {
			return err
		}
	}
	return nil
}

func (pC *ProjectConfig) GetStep(stepName string) (*ProjectStepConfig, error) {
	pC.rwLock.RLock()
	defer pC.rwLock.RUnlock()
	value, exists := pC.steps[stepName]

	if !exists {
		return nil, fmt.Errorf("`%s` step doesn't exist in project config", stepName)
	}

	return value, nil
}

func (pSC *ProjectStepConfig) ensureInitialized() {
	pSC.strategies = make(map[string]*StepAuthConfig)
}

func (pSC *ProjectStepConfig) AddAuthCommand(stepAuthConfig *StepAuthConfig) error {
	pSC.rwLock.Lock()
	defer pSC.rwLock.Unlock()

	if len(stepAuthConfig.Name) == 0 {
		return errors.New("`name` missing from auth config")
	}

	pSC.strategies[stepAuthConfig.Name] = stepAuthConfig

	return nil
}

func (pSC *ProjectStepConfig) GetAuthCommand(commandName string) (*StepAuthConfig, error) {
	pSC.rwLock.RLock()
	defer pSC.rwLock.RUnlock()

	value, exists := pSC.strategies[commandName]

	if !exists {
		return nil, fmt.Errorf("command with name `%s` doesn't exists", commandName)
	}

	return value, nil
}

func (pC *ProjectStepConfig) ValidateProjectStepConfig() error {
	if len(pC.Command) == 0 {
		return errors.New("`command` needs to be present in project step")
	}

	return nil
}

func NewProjectConfig(projectConfigFilepath string) (ProjectConfigInterface, error) {
	if len(projectConfigFilepath) == 0 {
		return nil, errors.New("no project config file path provided")
	}

	projectConfigFilepath, err := filepath.Abs(projectConfigFilepath)

	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(projectConfigFilepath)

	if err != nil {
		return nil, err
	}

	var steps []ProjectStepConfig

	if strings.HasSuffix(projectConfigFilepath, "yml") || strings.HasSuffix(projectConfigFilepath, "yaml") {
		err = yaml.Unmarshal(data, &steps)
	} else {
		err = json.Unmarshal(data, &steps)
	}

	if err != nil {
		return nil, err
	}

	projectConfig := ProjectConfig{
		steps: make(map[string]*ProjectStepConfig),
	}

	for stepIdx := range steps {
		step := &steps[stepIdx]
		step.ensureInitialized()

		err := step.ValidateProjectStepConfig()

		if err != nil {
			return nil, fmt.Errorf("step %d failed validation, error: %s", stepIdx+1, err.Error())
		}

		err = projectConfig.AddStep(step)

		if err != nil {
			return nil, err
		}
	}

	return &projectConfig, nil
}
