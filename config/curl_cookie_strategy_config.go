package config

import (
	"os"
	"path/filepath"
)

type CurlCookieStrategyConfig struct {
	CurlCookieFile     string `json:"curl_cookie_file,omitempty" yaml:"curl_cookie_file,omitempty"`
	parentProcessGroup *uintptr
}

func (cCSC *CurlCookieStrategyConfig) GetParentProcessGroup() *uintptr {
	return cCSC.parentProcessGroup
}

func ValidateCurlCookieStrategyConfig(config *StepAuthConfig) error {
	curlCookieFilepath, err := filepath.Abs(config.Params.CurlCookieFile)

	if err != nil {
		return err
	}

	_, err = os.Stat(curlCookieFilepath)

	if err != nil {
		return err
	}

	config.Params.CurlCookieFile = curlCookieFilepath
	return nil
}
