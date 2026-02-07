package strategy

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/process"
	process_mocks "github.com/pratyushtiwary/toki/process/mocks"
	test_config_utils "github.com/pratyushtiwary/toki/testutils/config"
	test_mock_utils "github.com/pratyushtiwary/toki/testutils/mocks"
	test_path_utils "github.com/pratyushtiwary/toki/testutils/path"
	"github.com/stretchr/testify/assert"
)

func assertCurlCookieStrategyHappyPath(t *testing.T, err error, curlCookieStrategy *CurlCookieStrategy, _config *config.StepAuthConfig, authConfig *config.StepAuthConfig) {
	assert.Nil(t, err)
	assert.IsType(t, &CurlCookieStrategy{}, curlCookieStrategy)
	assert.Equal(t, authConfig, _config)
}

func TestCurlCookieStrategyRefreshIsExpired(t *testing.T) {
	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-expired")

	curlCookieStrategy, _, err := NewCurlCookieStrategy(authConfig)

	assert.Nil(t, err)

	expired, err := curlCookieStrategy.IsExpired()

	assert.Nil(t, err)
	assert.True(t, expired)
}

func TestCurlCookieStrategyRefreshIsExpiredReturnsError(t *testing.T) {
	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "invalid-cookie-file")

	curlCookieStrategy, _, err := NewCurlCookieStrategy(authConfig)

	assert.Nil(t, err)

	// non-existent file path
	_, err = curlCookieStrategy.IsExpired()
	_, isPathError := err.(*os.PathError)

	assert.NotNil(t, err)
	assert.True(t, isPathError)

	// invalid cookie file
	curlCookieStrategy.curlCookieFilepath = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-invalid")
	_, err = curlCookieStrategy.IsExpired()
	_, isNumError := err.(*strconv.NumError)

	assert.NotNil(t, err)
	assert.True(t, isNumError)
}

func TestCurlCookieStrategyNewCommandFailed(t *testing.T) {
	resetNewCommand := test_mock_utils.MockNewCommand(nil)
	defer resetNewCommand()
	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	originalNewCommand := process.NewCommand

	curlCookieStrategy, _config, err := NewCurlCookieStrategy(authConfig)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "test error")
	assert.Nil(t, curlCookieStrategy)
	assert.Nil(t, _config)

	process.NewCommand = originalNewCommand
}

func TestCurlCookieStrategyHappyPath(t *testing.T) {
	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-valid")

	curlCookieStrategy, _config, err := NewCurlCookieStrategy(authConfig)

	assertCurlCookieStrategyHappyPath(t, err, curlCookieStrategy, _config, authConfig)

	err = curlCookieStrategy.Refresh(false, false)

	assert.Nil(t, err)
}

func TestCurlCookieStrategyRefreshIsExpiredThrowsError(t *testing.T) {
	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-invalid")

	curlCookieStrategy, _, _ := NewCurlCookieStrategy(authConfig)

	err := curlCookieStrategy.Refresh(false, false)
	_, isNumError := err.(*strconv.NumError)

	assert.NotNil(t, err)
	assert.True(t, isNumError)
}

func TestCurlCookieStrategyRefreshForce(t *testing.T) {
	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-valid")

	curlCookieStrategy, _, _ := NewCurlCookieStrategy(authConfig)

	err := curlCookieStrategy.Refresh(true, false)

	assert.Nil(t, err)
}

func TestCurlCookieStrategyRefreshRunReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedError := errors.New("test error")

	mockProc := process_mocks.NewMockProcessInterface(ctrl)
	mockProc.EXPECT().Run(gomock.Nil(), gomock.Any()).Return(expectedError).Times(1)

	resetNewCommand := test_mock_utils.MockNewCommand(mockProc)
	defer resetNewCommand()

	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-expired")

	curlCookieStrategy, _, _ := NewCurlCookieStrategy(authConfig)

	err := curlCookieStrategy.Refresh(false, false)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedError.Error())
}

func TestCurlCookieStrategyRefreshWaitTillFinishedReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedError := errors.New("test error")
	mockedStdErrBuffer := bytes.NewBufferString("test stderr")

	mockProc := process_mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Run(gomock.Nil(), gomock.Eq([]string{})).Return(nil).Times(1),
		mockProc.EXPECT().WaitTillFinished().Return(expectedError).Times(1),
		mockProc.EXPECT().GetStderrBuffer().Return(mockedStdErrBuffer).Times(1),
	)

	resetNewCommand := test_mock_utils.MockNewCommand(mockProc)
	defer resetNewCommand()

	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-expired")

	curlCookieStrategy, _, _ := NewCurlCookieStrategy(authConfig)

	err := curlCookieStrategy.Refresh(false, false)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), mockedStdErrBuffer.String())
}

func TestCurlCookieStrategyCleanupHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProc := process_mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Cleanup().Return(nil).Times(1),
	)

	resetNewCommand := test_mock_utils.MockNewCommand(mockProc)
	defer resetNewCommand()

	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-valid")

	curlCookieStrategy, _, _ := NewCurlCookieStrategy(authConfig)

	err := curlCookieStrategy.Cleanup()

	assert.Nil(t, err)
}

func TestCurlCookieStrategyCleanupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedError := errors.New("test error")

	mockProc := process_mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Cleanup().Return(expectedError).Times(1),
	)

	resetNewCommand := test_mock_utils.MockNewCommand(mockProc)
	defer resetNewCommand()

	curlCookieStrategyConfig, _ := test_config_utils.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := curlCookieStrategyConfig.AuthCommands[1]

	authConfig.Params.CurlCookieFile = test_path_utils.GetTestFilePath("curl_configs", "test-cookies-valid")

	curlCookieStrategy, _, _ := NewCurlCookieStrategy(authConfig)

	err := curlCookieStrategy.Cleanup()

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedError.Error())
}
