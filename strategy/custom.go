package strategy

import (
	"errors"
	"time"

	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/process"
)

type CustomStrategy struct {
	Strategy
	expiry         int
	refreshAfter   int
	lastSyncTime   time.Time
	authCommand    *process.Process
	processGroupId *uintptr
}

type CustomStrategyConfig struct {
	StrategyConfig
	expiry             int
	refreshAfter       int
	parentProcessGroup *uintptr
	command            string
}

func (cS *CustomStrategy) IsExpired() (bool, error) {
	time_to_refresh := cS.lastSyncTime.Add(time.Duration(cS.refreshAfter) * time.Minute)
	now := time.Now()

	return now.After(time_to_refresh) || now.Equal(time_to_refresh), nil
}

func (cS *CustomStrategy) Refresh(force bool) error {
	expired, err := cS.IsExpired()

	if err != nil {
		return nil
	}

	if expired || force {
		err := cS.authCommand.Run(cS.processGroupId)

		if err != nil {
			return err
		}
	}

	err = cS.authCommand.WaitTillFinished()

	if err != nil {
		return errors.New(cS.authCommand.GetStderrBuffer().String())
	}

	cS.lastSyncTime = time.Now()

	return nil
}

func (cS *CustomStrategy) Cleanup() error {
	return cS.authCommand.Cleanup()
}

func NewCustomStrategy(config *CustomStrategyConfig) (*CustomStrategy, error) {
	authCommand, err := process.NewCommand(config.command, nil)

	if err != nil {
		return nil, err
	}

	return &CustomStrategy{
		expiry:         config.expiry,
		refreshAfter:   config.refreshAfter,
		lastSyncTime:   time.Now(),
		processGroupId: config.parentProcessGroup,
		authCommand:    authCommand,
	}, nil
}

func NewCustomStrategyConfig(param config.PipelineAuthParamsConfig, processGroupId *uintptr) (*CustomStrategyConfig, error) {
	expiry := param.Expiry
	refreshAfter := param.RefreshAfter

	if expiry == nil {
		return nil, errors.New("custom strategy: expiry is missing")
	}

	if refreshAfter == nil {
		return nil, errors.New("custom strategy: refresh_after is missing")
	}

	if *expiry < 0 {
		return nil, errors.New("custom strategy: invalid value provided for expire, make sure value is positive")
	}

	if *refreshAfter < 0 {
		return nil, errors.New("custom strategy: invalid value provided for refresh_after, make sure value is positive")
	}

	if *refreshAfter >= *expiry {
		return nil, errors.New("custom strategy: refresh_after should be less than expiry")
	}

	return &CustomStrategyConfig{
		expiry:             *expiry,
		refreshAfter:       *refreshAfter,
		parentProcessGroup: processGroupId,
		command:            param.Command,
	}, nil
}
