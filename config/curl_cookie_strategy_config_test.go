package config

import (
	"errors"
	"os"
	"testing"

	test_utils "github.com/pratyushtiwary/toki/testutils"
	test_path_utils "github.com/pratyushtiwary/toki/testutils/path"
	"github.com/stretchr/testify/assert"
)

func SetupCurlCookie() CurlCookieStrategyConfig {
	curl_cookie_file := test_path_utils.GetTestFilePath("curl_configs", "test-cookies-valid")

	curl_cookie_strategy_config := CurlCookieStrategyConfig{
		CurlCookieFile:     curl_cookie_file,
		parentProcessGroup: nil,
	}

	return curl_cookie_strategy_config
}

func TestCurlCookieStrategyConfigHappyPath(t *testing.T) {
	curl_cookie_strategy_config := SetupCurlCookie()

	stepAuthConfig := StepAuthConfig{
		StrategyName: "curl_cookie",
		Params: AuthParamsConfig{
			StrategyConfig: StrategyConfig{
				Command: "echo 1",
			},
			CurlCookieStrategyConfig: curl_cookie_strategy_config,
		},
		Name: "custom",
	}

	err := ValidateCurlCookieStrategyConfig(&stepAuthConfig)

	assert.Nil(t, err)
}

func TestCurlCookieStrategyConfigInvalidFile(t *testing.T) {
	curl_cookie_strategy_config := SetupCurlCookie()

	curl_cookie_strategy_config.CurlCookieFile = "../test"

	stepAuthConfig := StepAuthConfig{
		StrategyName: "curl_cookie",
		Params: AuthParamsConfig{
			StrategyConfig: StrategyConfig{
				Command: "echo 1",
			},
			CurlCookieStrategyConfig: curl_cookie_strategy_config,
		},
		Name: "custom",
	}

	err := ValidateCurlCookieStrategyConfig(&stepAuthConfig)

	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestCurlCookieStrategyConfigGetParentProcessGroup(t *testing.T) {
	curl_cookie_strategy_config := SetupCurlCookie()

	assert.Nil(t, curl_cookie_strategy_config.GetParentProcessGroup())

	curl_cookie_strategy_config.parentProcessGroup = test_utils.UintPtr(2)
	assert.Equal(t, curl_cookie_strategy_config.GetParentProcessGroup(), test_utils.UintPtr(2))
}
