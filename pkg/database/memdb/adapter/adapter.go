package adapter

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// ColonyOSAdapter adapts VelocityDB for ColonyOS Database interface
type ColonyOSAdapter struct {
	db *memdb.VelocityDB
}

// Collections used by ColonyOS
const (
	ColoniesCollection     = "colonies"
	ProcessesCollection    = "processes"
	ExecutorsCollection    = "executors"
	FunctionsCollection    = "functions"
	UsersCollection        = "users"
	AttributesCollection   = "attributes"
	ProcessGraphsCollection = "processgraphs"
	LogsCollection         = "logs"
	FilesCollection        = "files"
	CronsCollection        = "crons"
	GeneratorsCollection   = "generators"
	SnapshotsCollection    = "snapshots"
	SecurityCollection     = "security"
)

// NewColonyOSAdapter creates a new ColonyOS adapter
func NewColonyOSAdapter(config *memdb.VelocityConfig) (*ColonyOSAdapter, error) {
	db, err := memdb.NewVelocityDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create VelocityDB: %w", err)
	}

	return &ColonyOSAdapter{db: db}, nil
}

// DatabaseCore interface implementation
func (a *ColonyOSAdapter) Close() {
	a.db.Close()
}

func (a *ColonyOSAdapter) IsEnabled() bool {
	return true
}

func (a *ColonyOSAdapter) Initialize() error {
	return nil
}

func (a *ColonyOSAdapter) Drop() error {
	return nil
}

func (a *ColonyOSAdapter) Lock(timeout int) error {
	return nil
}

func (a *ColonyOSAdapter) Unlock() error {
	return nil
}

func (a *ColonyOSAdapter) ApplyRetentionPolicy(retentionPeriod int64) error {
	return nil
}

// Verify interface compliance
var _ database.Database = (*ColonyOSAdapter)(nil)