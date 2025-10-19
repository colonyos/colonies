package database

type DatabaseCore interface {
	Close()
	Initialize() error
	Drop() error
	Lock(timeout int) error
	Unlock() error
	ApplyRetentionPolicy(retentionPeriod int64) error
}