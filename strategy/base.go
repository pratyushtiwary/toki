package strategy

type Strategy interface {
	IsExpired() (bool, error)
	Refresh(bool) error
	Cleanup() error
}

type StrategyConfig interface {
}
