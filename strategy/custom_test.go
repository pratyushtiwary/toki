package strategy

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/process"
	"github.com/pratyushtiwary/toki/process/mocks"
	"github.com/pratyushtiwary/toki/test"
	"github.com/stretchr/testify/assert"
)

func assertCustomStrategyHappyPath(t *testing.T, err error, customStrategy *CustomStrategy, _config *config.StepAuthConfig, authConfig *config.StepAuthConfig) {
	assert.Nil(t, err)
	assert.IsType(t, &CustomStrategy{}, customStrategy)
	assert.Equal(t, authConfig, _config)
}

func mockNewCommand(authCommand process.ProcessInterface) func() {
	var originalNewCommand = process.NewCommand
	process.NewCommand = func(execCommand string, parentProcessGroup *uintptr) (process.ProcessInterface, error) {
		if authCommand == nil {
			return nil, errors.New("test error")
		}
		return authCommand, nil
	}
	var resetNewCommand = func() {
		process.NewCommand = originalNewCommand
	}

	return resetNewCommand
}

func TestCustomStrategyHappyPath(t *testing.T) {
	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _config, err := NewCustomStrategy(authConfig)

	assertCustomStrategyHappyPath(t, err, customStrategy, _config, authConfig)
}

func TestCustomStrategyNewCommandFailed(t *testing.T) {
	defer mockNewCommand(nil)()
	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	originalNewCommand := process.NewCommand

	customStrategy, _config, err := NewCustomStrategy(authConfig)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "test error")
	assert.Nil(t, customStrategy)
	assert.Nil(t, _config)

	process.NewCommand = originalNewCommand
}

func TestCustomStrategyIsExpired(t *testing.T) {
	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	// happy path
	expired, err := customStrategy.IsExpired()

	assert.Nil(t, err)
	assert.False(t, expired)

	// expired case - now is equal to expired time
	customStrategy.lastSyncTime = customStrategy.lastSyncTime.Add(-(time.Duration(customStrategy.refreshAfter) * time.Minute))
	expired, err = customStrategy.IsExpired()

	assert.Nil(t, err)
	assert.True(t, expired, "now == lastSyncTime, IsExpired should return true")

	// expired case - now is before expired time
	customStrategy.lastSyncTime = customStrategy.lastSyncTime.Add(-(time.Duration(customStrategy.refreshAfter+5) * time.Minute))
	expired, err = customStrategy.IsExpired()

	assert.Nil(t, err)
	assert.True(t, expired, "now < lastSyncTime, IsExpired should return true")
}

func TestCustomStrategyRefreshHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProc := mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Run(gomock.Nil(), gomock.Eq([]string{})).Return(nil).Times(1),
		mockProc.EXPECT().WaitTillFinished().Return(nil).Times(1),
	)

	defer mockNewCommand(mockProc)()

	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	customStrategy.lastSyncTime = customStrategy.lastSyncTime.Add(-(time.Duration(customStrategy.refreshAfter) * time.Minute))

	err := customStrategy.Refresh(false, false)

	assert.Nil(t, err)
}

func TestCustomStrategyRefreshNotExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProc := mocks.NewMockProcessInterface(ctrl)

	defer mockNewCommand(mockProc)()

	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	err := customStrategy.Refresh(false, false)

	assert.Nil(t, err)
}
func TestCustomStrategyRefreshForce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProc := mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Run(gomock.Nil(), gomock.Eq([]string{})).Return(nil).Times(1),
		mockProc.EXPECT().WaitTillFinished().Return(nil).Times(1),
	)

	defer mockNewCommand(mockProc)()

	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	err := customStrategy.Refresh(true, false)

	assert.Nil(t, err)
}

func TestCustomStrategyRefreshRunReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedError := errors.New("test error")

	mockProc := mocks.NewMockProcessInterface(ctrl)
	mockProc.EXPECT().Run(gomock.Nil(), gomock.Any()).Return(expectedError).Times(1)

	defer mockNewCommand(mockProc)()

	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	customStrategy.lastSyncTime = customStrategy.lastSyncTime.Add(-(time.Duration(customStrategy.refreshAfter) * time.Minute))

	err := customStrategy.Refresh(false, false)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedError.Error())
}

func TestCustomStrategyRefreshWaitTillFinishedReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedError := errors.New("test error")
	mockedStdErrBuffer := bytes.NewBufferString("test stderr")

	mockProc := mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Run(gomock.Nil(), gomock.Eq([]string{})).Return(nil).Times(1),
		mockProc.EXPECT().WaitTillFinished().Return(expectedError).Times(1),
		mockProc.EXPECT().GetStderrBuffer().Return(mockedStdErrBuffer).Times(1),
	)

	defer mockNewCommand(mockProc)()

	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	customStrategy.lastSyncTime = customStrategy.lastSyncTime.Add(-(time.Duration(customStrategy.refreshAfter) * time.Minute))

	err := customStrategy.Refresh(false, false)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), mockedStdErrBuffer.String())
}

func TestCustomStrategyCleanupHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProc := mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Cleanup().Return(nil).Times(1),
	)

	defer mockNewCommand(mockProc)()

	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	err := customStrategy.Cleanup()

	assert.Nil(t, err)
}

func TestCustomStrategyCleanupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedError := errors.New("test error")

	mockProc := mocks.NewMockProcessInterface(ctrl)
	gomock.InOrder(
		mockProc.EXPECT().Cleanup().Return(expectedError).Times(1),
	)

	defer mockNewCommand(mockProc)()

	customStrategyConfig, _ := test.GetFirstStep(t, "customPipelineConfig.json", nil)
	authConfig := customStrategyConfig.AuthCommands[0]

	customStrategy, _, _ := NewCustomStrategy(authConfig)

	err := customStrategy.Cleanup()

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedError.Error())
}
