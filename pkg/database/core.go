package database

type DatabaseCore interface {
	Close()
	Initialize() error
	Drop() error
	ApplyRetentionPolicy(retentionPeriod int64) error
}