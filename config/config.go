package config

type AuthParamsConfig struct {
	StrategyConfig           `yaml:",inline"`
	CustomStrategyConfig     `yaml:",inline"`
	CurlCookieStrategyConfig `yaml:",inline"`
}

type StepAuthConfig struct {
	StrategyName string           `json:"strategy" yaml:"strategy"`
	Params       AuthParamsConfig `json:"params" yaml:"params"`
	Name         string           `json:"name,omitempty" yaml:"name,omitempty"`
}

type StepConfig struct {
	Command      string            `json:"command,omitempty" yaml:"command,omitempty"`
	AuthCommands []*StepAuthConfig `json:"auth,omitempty" yaml:"auth,omitempty"`
}
