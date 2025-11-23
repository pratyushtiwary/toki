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

func (cS *CustomStrategy) IsExpired() (bool, error) {
	time_to_refresh := cS.lastSyncTime.Add(time.Duration(cS.refreshAfter) * time.Minute)
	now := time.Now()

	return now.After(time_to_refresh) || now.Equal(time_to_refresh), nil
}

func (cS *CustomStrategy) Refresh(force bool) error {
	expired, err := cS.IsExpired()

	if err != nil {
		return err
	}

	if !expired && !force {
		return nil
	}

	err = cS.authCommand.Run(cS.processGroupId)
	if err != nil {
		return err
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

func NewCustomStrategy(config *config.StepAuthConfig) (*CustomStrategy, *config.StepAuthConfig, error) {
	authCommand, err := process.NewCommand(config.Params.Command, nil)

	if err != nil {
		return nil, nil, err
	}

	return &CustomStrategy{
		expiry:         *config.Params.Expiry,
		refreshAfter:   *config.Params.RefreshAfter,
		lastSyncTime:   time.Now(),
		processGroupId: config.Params.GetParentProcessGroup(),
		authCommand:    authCommand,
	}, config, nil
}
