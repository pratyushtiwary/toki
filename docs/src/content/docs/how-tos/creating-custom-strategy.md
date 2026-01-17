---
title: Creating custom strategy
description: This guide shows how you can create and register a custom strategy in Toki
---

In this tutorial we'll try to re-invent the `curl_cookie` strategy in Toki.

## Prerequisites

Before we start, you should have:
- Cloned [Toki repo](https://github.com/pratyushtiwary/toki),
- A shell environment

## Step 1: Understanding the architecture

A strategy is made up of 2 components:
- Config: Config defines the input for the stategy implementation, &
- Strategy implementation: Strategy implementation implements `IsExpired`, `Refresh` & `Cleanup` methods

## Step 2: Defining config

Config is very important for any strategy within Toki. A config controls how expiry and refresh will happen. For `curl_cookie` strategy we need curl cookie path to determine when the cookie is going to expire along with a refresh command.

Configs are defined under config module. Each stategy config extends `StrategyConfig` base struct. The new config for our strategy would look something like the following:
```go
type CurlCookieStrategyConfig struct {
	CurlCookieFile     string `json:"curl_cookie_file,omitempty" yaml:"curl_cookie_file,omitempty"`
}
```

Save the above code inside `config` dir with `curl_cookie_strategy_config.go` as file name

:::note
Notice how above we haven't added `Command` as one of the config parameters, this is because StrategyConfig(base struct for all StrategyConfigs) already has `Command` defined
:::

## Step 3: Creating strategy implementation

Now that we have config defined we need a implementation to consume this config. Let's start by creating a dummy implementation with noop for now.

Start by creating a new file under `strategy` dir with `curl_cookie.go` as file name.

Copy paste the following content in the file:
```go
package strategy

import (
	"errors"

	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/log"
	"github.com/pratyushtiwary/toki/process"
)

type CurlCookieStrategy struct {
	Strategy
	curlCookieFilepath string
	authCommand        process.ProcessInterface
	processGroupId     *uintptr
}

func (cCS *CurlCookieStrategy) IsExpired() (bool, error) {
	return false, nil
}

func (cS *CurlCookieStrategy) Refresh(force bool, verbose bool) error {
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
```

In the code above we have created a function to instantiate CurlCookieStrategy struct and we've defined the implementation struct along with required `IsExpired`, `Refresh` & `Cleanup` methods

## Step 4: Registering strategy

Toki maintains an internal registry of strategies which gets called automatically based on current step's configuration. To register our new strategy to Toki we can edit `strategy\registry.go` file and add a new case for our strategy.

Let's add the following code snippet to registry file:
```go
	case "curl_cookie":
		err := config.ValidateCurlCookieStrategyConfig(stepConfig)

		if err != nil {
			return nil, nil, err
		}

		return NewCurlCookieStrategy(stepConfig)
```

The code snippet above is first validating the config and then returning curl cookie strategy. Strategies in Toki are supposed to perform validation and creation in the registry function

## Step 5: Testing

Now that we have a noop strategy lets try to test it.

Run the following command in your terminal to generate a test-cookie file using `curl`:
```sh
curl -X POST https://dummyjson.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "emilys", "password": "emilyspass", "expiresInMins": 60}' \
  -c test-cookies
```

Copy paste the following pipeline config in `test.pipeline.json` file:
```json
[
  {
    "command": "sleep 20m",
    "auth": [
      {
        "strategy": "curl_cookie",
        "params": {
          "expiry": 60,
          "curl_cookie_file": "<ABS/REL PATH to test-cookies>",
          "command": "echo 1"
        }
      }
    ]
  }
]
```

To run the above config run `go run . test.pipeline.config` in your terminal

## Step 6: Implementing Refresh method

Refresh logic would be simple for curl cookies, we check if cookies are expired and then run the auth command

Replace Refresh function with the code snippet below:
```go
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

	return nil
}
```

In the code snippet above we are checking whether the token has expired, if either the token has expired or it was a force refresh we run the auth command else we exit function early

## Step 7: Implementing IsExpired method

Before we start with the implementation lets start by understanding how to read curl cookie files. Curl has a good write up on the format of file which can be found [here](https://curl.se/docs/http-cookies.html). Here's a gist of it:
```text
Field number, what type and example data and the meaning of it:

- string example.com - the domain name
- boolean FALSE - include subdomains
- string /foobar/ - path
- boolean TRUE - send/receive over HTTPS only
- number 1462299217 - expires at - seconds since Jan 1st 1970, or 0
- string person - name of the cookie
- string daniel - value of the cookie
```

We are interested in the 5th field which is for expiry. The plan is to fetch all the expiry values, get the minimum from there and convert it to datetime and then check if that value >= now.

Replace IsExpired function with the below code snippet:
```go
func (cCS *CurlCookieStrategy) IsExpired() (bool, error) {
	file, err := os.Open(cCS.curlCookieFilepath)

	if err != nil {
		return false, err
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
			return false, nil
		}

		minExpiry = min(minExpiry, expiry)
	}

	now := time.Now()

	if minExpiry == math.MaxInt {
		expiryTime := now
	} else {
        expiryTime := time.Unix(minExpiry, 0)
    }

	return expiryTime.Add(-1*time.Minute).Before(now) || expiryTime.Add(-1*time.Minute).Equal(now), nil
}
```

In the above code we are parsing the curl cookie file everytime we check for expiry and returning true/false based on whether the min expiry set in the cookie file is >= now with a -1 minute offset

## Step 8: Optimizing IsExpired method

Toki calls IsExpired quite often, so we don't wanna make it so that IsExpired reads curl cookie file everytime. We will use a cached value for minExpiry and will invalidate the cache on every refresh.

We will start by adding a new data member to our struct called `expiryTime` of `*time.Time` type, and we'll also extract the logic for calculating min expiry time into a separate method called `calcMinExpiry`. In `IsExpired` method we'll add a guard to check if `expiryTime` is set or else we'll call `calcMinExpiry`. In `Refresh` method we'll call `calcMinExpiry` once refresh is successful.

Here's the updated code for curl_cookie strategy implementation:
```go
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
```


Congratulations on making it this far and implementing your first strategy in Toki! 🚀