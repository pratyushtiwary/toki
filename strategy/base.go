package strategy

type Strategy interface {
	IsExpired() (bool, error)
	Refresh(bool, bool) error
	Cleanup() error
}
