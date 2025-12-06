package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(n int) *int {
	return &n
}

func uintPtr(n uintptr) *uintptr {
	return &n
}

func Setup() StepAuthConfig {
	customStrategyConfig := CustomStrategyConfig{
		Expiry:             intPtr(2),
		RefreshAfter:       intPtr(1),
		parentProcessGroup: nil,
	}

	stepAuthConfig := StepAuthConfig{
		StrategyName: "custom",
		Params: AuthParamsConfig{
			StrategyConfig: StrategyConfig{
				Command: "echo 1",
			},
			CustomStrategyConfig: customStrategyConfig,
		},
		Name: "custom",
	}
	return stepAuthConfig
}

func TestCustomStrategyConfigHappyPath(t *testing.T) {
	stepAuthConfig := Setup()

	err := ValidateCustomStrategyConfig(&stepAuthConfig, nil)

	assert.Nil(t, err)
}

func TestCustomStrategyConfigCommandNotSet(t *testing.T) {
	stepAuthConfig := Setup()

	stepAuthConfig.Params.Command = ""

	err := ValidateCustomStrategyConfig(&stepAuthConfig, nil)

	assert.NotNil(t, err)
	assert.NotNil(t, err.Error(), "custom strategy: command is missing")
}

func TestCustomStrategyConfigExpiryNotSet(t *testing.T) {
	stepAuthConfig := Setup()

	stepAuthConfig.Params.Expiry = nil

	err := ValidateCustomStrategyConfig(&stepAuthConfig, nil)

	assert.NotNil(t, err)
	assert.NotNil(t, err.Error(), "custom strategy: expiry is missing")
}
func TestCustomStrategyConfigRefreshAfterNotSet(t *testing.T) {
	stepAuthConfig := Setup()

	stepAuthConfig.Params.RefreshAfter = nil

	err := ValidateCustomStrategyConfig(&stepAuthConfig, nil)

	assert.NotNil(t, err)
	assert.NotNil(t, err.Error(), "custom strategy: refresh_after is missing")
}

func TestCustomStrategyConfigExpiryLessThan0(t *testing.T) {
	stepAuthConfig := Setup()

	stepAuthConfig.Params.Expiry = intPtr(-1)

	err := ValidateCustomStrategyConfig(&stepAuthConfig, nil)

	assert.NotNil(t, err)
	assert.NotNil(t, err.Error(), "custom strategy: invalid value provided for expire, make sure value is positive")
}
func TestCustomStrategyConfigRefreshAfterLessThan0(t *testing.T) {
	stepAuthConfig := Setup()

	stepAuthConfig.Params.RefreshAfter = intPtr(-1)

	err := ValidateCustomStrategyConfig(&stepAuthConfig, nil)

	assert.NotNil(t, err)
	assert.NotNil(t, err.Error(), "custom strategy: invalid value provided for refresh_after, make sure value is positive")
}

func TestCustomStrategyConfigRefreshAfterGreaterThanExpiry0(t *testing.T) {
	stepAuthConfig := Setup()

	stepAuthConfig.Params.RefreshAfter = intPtr(2)
	stepAuthConfig.Params.Expiry = intPtr(1)

	err := ValidateCustomStrategyConfig(&stepAuthConfig, nil)

	assert.NotNil(t, err)
	assert.NotNil(t, err.Error(), "custom strategy: refresh_after should be less than expiry")
}

func TestCustomStrategyConfigGetParentProcessGroup(t *testing.T) {
	stepAuthConfig := Setup()

	customStrategy := stepAuthConfig.Params.CustomStrategyConfig

	assert.Nil(t, customStrategy.GetParentProcessGroup())

	customStrategy.parentProcessGroup = uintPtr(2)

	assert.Equal(t, customStrategy.GetParentProcessGroup(), uintPtr(2))
}
