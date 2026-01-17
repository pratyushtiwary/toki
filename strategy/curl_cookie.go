package strategy

import (
	"bufio"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/log"
	"github.com/pratyushtiwary/toki/process"
)

type CurlCookieStrategy struct {
	Strategy
	curlCookieFilepath string
	authCommand        process.ProcessInterface
	expiryTime         *time.Time
	processGroupId     *uintptr
}

func (cCS *CurlCookieStrategy) calcMinExpiry() error {
	file, err := os.Open(cCS.curlCookieFilepath)

	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var minExpiry int64 = math.MaxInt64

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#HttpOnly_") {
			continue
		}

		parts := strings.Split(line, "\t")

		if len(parts) < 5 {
			continue
		}

		expiryStr := parts[4]

		expiry, err := strconv.ParseInt(expiryStr, 10, 64)

		if err != nil {
			return nil
		}

		minExpiry = min(minExpiry, expiry)
	}

	if minExpiry == math.MaxInt {
		now := time.Now()
		cCS.expiryTime = &now
		return nil
	}

	expiryTime := time.Unix(minExpiry, 0)

	cCS.expiryTime = &expiryTime

	return nil
}

func (cCS *CurlCookieStrategy) IsExpired() (bool, error) {
	if cCS.expiryTime == nil {
		err := cCS.calcMinExpiry()

		if err != nil {
			return false, err
		}
	}

	now := time.Now()

	return cCS.expiryTime.Add(-1*time.Minute).Before(now) || cCS.expiryTime.Add(-1*time.Minute).Equal(now), nil
}

func (cCS *CurlCookieStrategy) Refresh(force bool, verbose bool) error {
	expired, err := cCS.IsExpired()

	if err != nil {
		return err
	}

	if !expired && !force {
		return nil
	}

	err = cCS.authCommand.Run(cCS.processGroupId, []string{})
	if err != nil {
		return err
	}

	err = cCS.authCommand.WaitTillFinished()
	if err != nil {
		return errors.New(cCS.authCommand.GetStderrBuffer().String())
	}

	if verbose {
		log.Info("Stdout: %s", cCS.authCommand.GetStdoutBuffer().String())
	}

	// calc expiry time again after refresh
	cCS.calcMinExpiry()

	return nil
}

func (cCS *CurlCookieStrategy) Cleanup() error {
	return cCS.authCommand.Cleanup()
}

func NewCurlCookieStrategy(config *config.StepAuthConfig) (*CurlCookieStrategy, *config.StepAuthConfig, error) {
	authCommand, err := process.NewCommand(config.Params.Command, nil)

	if err != nil {
		return nil, nil, err
	}

	return &CurlCookieStrategy{
		curlCookieFilepath: config.Params.CurlCookieStrategyConfig.CurlCookieFile,
		authCommand:        authCommand,
		processGroupId:     config.Params.CurlCookieStrategyConfig.GetParentProcessGroup(),
	}, config, nil
}
