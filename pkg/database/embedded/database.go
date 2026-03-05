package embedded

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/embedded/diskstore"
	"github.com/colonyos/colonies/pkg/database/embedded/flusher"
	"github.com/colonyos/colonies/pkg/database/embedded/index"
	"github.com/colonyos/colonies/pkg/database/embedded/store"
	"github.com/colonyos/colonies/pkg/database/embedded/wal"
)

type EmbeddedDatabase struct {
	dataDir string
	wal     wal.WAL
	flusher *flusher.Flusher

	// Colony store (keyed by colony name)
	colonies    *store.Store[string, core.Colony]
	coloniesIdx struct {
		byID *index.MapIndex[string, string] // colonyID -> colonyName
	}

	// User store (keyed by "colonyName:userName")
	users    *store.Store[string, core.User]
	usersIdx struct {
		byColony *index.MapIndex[string, string] // colonyName -> set of compound keys
		byID     *index.MapIndex[string, string] // userID -> compound key
	}

	// Executor store (keyed by executorID)
	executors    *store.Store[string, core.Executor]
	executorsIdx struct {
		byColony    *index.MapIndex[string, string] // colonyName -> set of executorIDs
		byName      *index.MapIndex[string, string] // "colonyName:executorName" -> executorID
		byBlueprint *index.MapIndex[string, string] // blueprintID -> set of executorIDs
	}

	// Function store (keyed by functionID)
	functions    *store.Store[string, core.Function]
	functionsIdx struct {
		byColony   *index.MapIndex[string, string] // colonyName -> set of functionIDs
		byExecutor *index.MapIndex[string, string] // "colonyName:executorName" -> set of functionIDs
		byFullName *index.MapIndex[string, string] // "colonyName:executorName:funcName" -> functionID
	}

	// Generator store (keyed by generatorID)
	generators    *store.Store[string, core.Generator]
	generatorsIdx struct {
		byColony *index.MapIndex[string, string] // colonyName -> set of generatorIDs
		byName   *index.MapIndex[string, string] // "colonyName:name" -> generatorID
	}

	// GeneratorArg store (keyed by argID)
	generatorArgs    *store.Store[string, core.GeneratorArg]
	generatorArgsIdx struct {
		byGenerator *index.MapIndex[string, string] // generatorID -> set of argIDs
		byColony    *index.MapIndex[string, string] // colonyName -> set of argIDs
	}

	// Cron store (keyed by cronID)
	crons    *store.Store[string, core.Cron]
	cronsIdx struct {
		byColony *index.MapIndex[string, string] // colonyName -> set of cronIDs
		byName   *index.MapIndex[string, string] // "colonyName:cronName" -> cronID
	}

	// Snapshot store (keyed by snapshotID)
	snapshots    *store.Store[string, core.Snapshot]
	snapshotsIdx struct {
		byColony *index.MapIndex[string, string] // colonyName -> set of snapshotIDs
		byName   *index.MapIndex[string, string] // "colonyName:snapshotName" -> snapshotID
	}

	// Location store (keyed by locationID)
	locations    *store.Store[string, core.Location]
	locationsIdx struct {
		byColony *index.MapIndex[string, string] // colonyName -> set of locationIDs
		byName   *index.MapIndex[string, string] // "colonyName:locationName" -> locationID
	}

	// Server store (keyed by "serverid")
	server *store.Store[string, string]

	// Blueprint stores
	blueprintDefs    *store.Store[string, core.BlueprintDefinition]
	blueprintDefsIdx struct {
		byNamespace *index.MapIndex[string, string] // namespace -> set of IDs
		byName      *index.MapIndex[string, string] // "namespace:name" -> ID
		byKind      *index.MapIndex[string, string] // kind -> set of IDs
	}

	blueprints    *store.Store[string, core.Blueprint]
	blueprintsIdx struct {
		byNamespace *index.MapIndex[string, string] // namespace -> set of IDs
		byName      *index.MapIndex[string, string] // "namespace:name" -> ID
		byKind      *index.MapIndex[string, string] // kind -> set of IDs
	}

	blueprintHistory    *store.Store[string, core.BlueprintHistory]
	blueprintHistoryIdx struct {
		byBlueprint *index.MapIndex[string, string] // blueprintID -> set of historyIDs
	}

	// Process store (keyed by processID)
	processes    *store.Store[string, core.Process]
	processesIdx struct {
		byColony   *index.CompoundIndex                // colony -> state -> OrderedIndex[by priorityTime]
		byExecutor *index.MapIndex[string, string]     // executorID -> set of processIDs
		byGraph    *index.MapIndex[string, string]     // processGraphID -> set of processIDs
	}

	// Attribute store (keyed by attributeID)
	attributes    *store.Store[string, core.Attribute]
	attributesIdx struct {
		byTarget *index.MapIndex[string, string] // targetID -> set of attributeIDs
		byColony *index.MapIndex[string, string] // colonyName -> set of attributeIDs
		byGraph  *index.MapIndex[string, string] // processGraphID -> set of attributeIDs
	}

	// ProcessGraph store (keyed by processGraphID)
	processGraphs    *store.Store[string, core.ProcessGraph]
	processGraphsIdx struct {
		byColony *index.CompoundIndex // colony -> state -> OrderedIndex[by submissionTime]
	}

	// Log store (keyed by generated ID)
	logs    *store.Store[string, core.Log]
	logsIdx struct {
		byProcess  *index.MapIndex[string, string] // processID -> set of logIDs
		byExecutor *index.MapIndex[string, string] // executorName -> set of logIDs
		byColony   *index.MapIndex[string, string] // colonyName -> set of logIDs
	}

	// File store (keyed by fileID)
	files    *store.Store[string, core.File]
	filesIdx struct {
		byColony *index.MapIndex[string, string] // colonyName -> set of fileIDs
		byLabel  *index.MapIndex[string, string] // "colonyName:label" -> set of fileIDs
		byName   *index.MapIndex[string, string] // "colonyName:label:name" -> set of fileIDs
	}

	// File sequence counter
	fileSeqCounter int64
}

func CreateEmbeddedDatabase(dataDir string) *EmbeddedDatabase {
	return &EmbeddedDatabase{dataDir: dataDir}
}

func identity(s string) string { return s }

func (db *EmbeddedDatabase) Initialize() error {
	if err := os.MkdirAll(db.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	walPath := filepath.Join(db.dataDir, "wal")
	w, err := wal.NewFileWAL(walPath, wal.SyncNone)
	if err != nil {
		return fmt.Errorf("failed to create WAL: %w", err)
	}

	// Create all stores without WAL first (WAL is set after replay)
	if err := db.createStores(nil); err != nil {
		return err
	}

	// Replay WAL entries into stores
	if err := db.replayWAL(w); err != nil {
		return fmt.Errorf("failed to replay WAL: %w", err)
	}

	// Load from disk (won't overwrite records already populated by WAL replay)
	if err := db.loadAll(); err != nil {
		return fmt.Errorf("failed to load from disk: %w", err)
	}

	// Now set WAL on all stores so future writes are logged
	db.setWALOnStores(w)
	db.wal = w

	// Rebuild all indexes from loaded data
	db.rebuildIndexes()

	// Start background flusher
	db.flusher = flusher.NewFlusher(5 * time.Second)
	db.addStoresToFlusher()
	db.flusher.Start()

	return nil
}

func (db *EmbeddedDatabase) createStores(w wal.WAL) error {
	baseDir := filepath.Join(db.dataDir, "data")

	mkDisk := func(entity string) error {
		dir := filepath.Join(baseDir, entity)
		return os.MkdirAll(dir, 0755)
	}

	entities := []string{
		"colonies", "users", "executors", "functions",
		"generators", "generatorargs", "crons", "snapshots",
		"locations", "server", "blueprintdefs", "blueprints",
		"blueprinthistory", "processes", "attributes",
		"processgraphs", "logs", "files",
	}

	for _, e := range entities {
		if err := mkDisk(e); err != nil {
			return fmt.Errorf("failed to create data dir for %s: %w", e, err)
		}
	}

	// Helper to create typed DiskStore
	createDiskStore := func(entity string) string {
		return filepath.Join(baseDir, entity)
	}

	// Colonies
	coloniesDisk, err := diskstore.NewDiskStore[core.Colony](createDiskStore("colonies"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for colonies: %w", err)
	}
	db.colonies = store.NewStore(store.Config[string, core.Colony]{
		Disk: coloniesDisk, WAL: w, EntityName: "colonies",
		KeyToStr: identity, StrToKey: identity,
	})

	// Users
	usersDisk, err := diskstore.NewDiskStore[core.User](createDiskStore("users"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for users: %w", err)
	}
	db.users = store.NewStore(store.Config[string, core.User]{
		Disk: usersDisk, WAL: w, EntityName: "users",
		KeyToStr: identity, StrToKey: identity,
	})

	// Executors
	executorsDisk, err := diskstore.NewDiskStore[core.Executor](createDiskStore("executors"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for executors: %w", err)
	}
	db.executors = store.NewStore(store.Config[string, core.Executor]{
		Disk: executorsDisk, WAL: w, EntityName: "executors",
		KeyToStr: identity, StrToKey: identity,
	})

	// Functions
	functionsDisk, err := diskstore.NewDiskStore[core.Function](createDiskStore("functions"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for functions: %w", err)
	}
	db.functions = store.NewStore(store.Config[string, core.Function]{
		Disk: functionsDisk, WAL: w, EntityName: "functions",
		KeyToStr: identity, StrToKey: identity,
	})

	// Generators
	generatorsDisk, err := diskstore.NewDiskStore[core.Generator](createDiskStore("generators"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for generators: %w", err)
	}
	db.generators = store.NewStore(store.Config[string, core.Generator]{
		Disk: generatorsDisk, WAL: w, EntityName: "generators",
		KeyToStr: identity, StrToKey: identity,
	})

	// GeneratorArgs
	generatorArgsDisk, err := diskstore.NewDiskStore[core.GeneratorArg](createDiskStore("generatorargs"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for generatorargs: %w", err)
	}
	db.generatorArgs = store.NewStore(store.Config[string, core.GeneratorArg]{
		Disk: generatorArgsDisk, WAL: w, EntityName: "generatorargs",
		KeyToStr: identity, StrToKey: identity,
	})

	// Crons
	cronsDisk, err := diskstore.NewDiskStore[core.Cron](createDiskStore("crons"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for crons: %w", err)
	}
	db.crons = store.NewStore(store.Config[string, core.Cron]{
		Disk: cronsDisk, WAL: w, EntityName: "crons",
		KeyToStr: identity, StrToKey: identity,
	})

	// Snapshots
	snapshotsDisk, err := diskstore.NewDiskStore[core.Snapshot](createDiskStore("snapshots"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for snapshots: %w", err)
	}
	db.snapshots = store.NewStore(store.Config[string, core.Snapshot]{
		Disk: snapshotsDisk, WAL: w, EntityName: "snapshots",
		KeyToStr: identity, StrToKey: identity,
	})

	// Locations
	locationsDisk, err := diskstore.NewDiskStore[core.Location](createDiskStore("locations"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for locations: %w", err)
	}
	db.locations = store.NewStore(store.Config[string, core.Location]{
		Disk: locationsDisk, WAL: w, EntityName: "locations",
		KeyToStr: identity, StrToKey: identity,
	})

	// Server
	serverDisk, err := diskstore.NewDiskStore[string](createDiskStore("server"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for server: %w", err)
	}
	db.server = store.NewStore(store.Config[string, string]{
		Disk: serverDisk, WAL: w, EntityName: "server",
		KeyToStr: identity, StrToKey: identity,
	})

	// Blueprint Definitions
	blueprintDefsDisk, err := diskstore.NewDiskStore[core.BlueprintDefinition](createDiskStore("blueprintdefs"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for blueprintdefs: %w", err)
	}
	db.blueprintDefs = store.NewStore(store.Config[string, core.BlueprintDefinition]{
		Disk: blueprintDefsDisk, WAL: w, EntityName: "blueprintdefs",
		KeyToStr: identity, StrToKey: identity,
	})

	// Blueprints
	blueprintsDisk, err := diskstore.NewDiskStore[core.Blueprint](createDiskStore("blueprints"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for blueprints: %w", err)
	}
	db.blueprints = store.NewStore(store.Config[string, core.Blueprint]{
		Disk: blueprintsDisk, WAL: w, EntityName: "blueprints",
		KeyToStr: identity, StrToKey: identity,
	})

	// Blueprint History
	blueprintHistoryDisk, err := diskstore.NewDiskStore[core.BlueprintHistory](createDiskStore("blueprinthistory"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for blueprinthistory: %w", err)
	}
	db.blueprintHistory = store.NewStore(store.Config[string, core.BlueprintHistory]{
		Disk: blueprintHistoryDisk, WAL: w, EntityName: "blueprinthistory",
		KeyToStr: identity, StrToKey: identity,
	})

	processesDisk, err := diskstore.NewDiskStore[core.Process](createDiskStore("processes"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for processes: %w", err)
	}
	db.processes = store.NewStore(store.Config[string, core.Process]{
		Disk: processesDisk, WAL: w, EntityName: "processes",
		KeyToStr: identity, StrToKey: identity,
	})

	attributesDisk, err := diskstore.NewDiskStore[core.Attribute](createDiskStore("attributes"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for attributes: %w", err)
	}
	db.attributes = store.NewStore(store.Config[string, core.Attribute]{
		Disk: attributesDisk, WAL: w, EntityName: "attributes",
		KeyToStr: identity, StrToKey: identity,
	})

	processGraphsDisk, err := diskstore.NewDiskStore[core.ProcessGraph](createDiskStore("processgraphs"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for processgraphs: %w", err)
	}
	db.processGraphs = store.NewStore(store.Config[string, core.ProcessGraph]{
		Disk: processGraphsDisk, WAL: w, EntityName: "processgraphs",
		KeyToStr: identity, StrToKey: identity,
	})

	logsDisk, err := diskstore.NewDiskStore[core.Log](createDiskStore("logs"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for logs: %w", err)
	}
	db.logs = store.NewStore(store.Config[string, core.Log]{
		Disk: logsDisk, WAL: w, EntityName: "logs",
		KeyToStr: identity, StrToKey: identity,
	})

	filesDisk, err := diskstore.NewDiskStore[core.File](createDiskStore("files"), "")
	if err != nil {
		return fmt.Errorf("failed to create disk store for files: %w", err)
	}
	db.files = store.NewStore(store.Config[string, core.File]{
		Disk: filesDisk, WAL: w, EntityName: "files",
		KeyToStr: identity, StrToKey: identity,
	})

	// Create all indexes
	db.createIndexes()

	return nil
}

func (db *EmbeddedDatabase) createIndexes() {
	db.coloniesIdx.byID = index.NewMapIndex[string, string]()

	db.usersIdx.byColony = index.NewMapIndex[string, string]()
	db.usersIdx.byID = index.NewMapIndex[string, string]()

	db.executorsIdx.byColony = index.NewMapIndex[string, string]()
	db.executorsIdx.byName = index.NewMapIndex[string, string]()
	db.executorsIdx.byBlueprint = index.NewMapIndex[string, string]()

	db.functionsIdx.byColony = index.NewMapIndex[string, string]()
	db.functionsIdx.byExecutor = index.NewMapIndex[string, string]()
	db.functionsIdx.byFullName = index.NewMapIndex[string, string]()

	db.generatorsIdx.byColony = index.NewMapIndex[string, string]()
	db.generatorsIdx.byName = index.NewMapIndex[string, string]()

	db.generatorArgsIdx.byGenerator = index.NewMapIndex[string, string]()
	db.generatorArgsIdx.byColony = index.NewMapIndex[string, string]()

	db.cronsIdx.byColony = index.NewMapIndex[string, string]()
	db.cronsIdx.byName = index.NewMapIndex[string, string]()

	db.snapshotsIdx.byColony = index.NewMapIndex[string, string]()
	db.snapshotsIdx.byName = index.NewMapIndex[string, string]()

	db.locationsIdx.byColony = index.NewMapIndex[string, string]()
	db.locationsIdx.byName = index.NewMapIndex[string, string]()

	db.blueprintDefsIdx.byNamespace = index.NewMapIndex[string, string]()
	db.blueprintDefsIdx.byName = index.NewMapIndex[string, string]()
	db.blueprintDefsIdx.byKind = index.NewMapIndex[string, string]()

	db.blueprintsIdx.byNamespace = index.NewMapIndex[string, string]()
	db.blueprintsIdx.byName = index.NewMapIndex[string, string]()
	db.blueprintsIdx.byKind = index.NewMapIndex[string, string]()

	db.blueprintHistoryIdx.byBlueprint = index.NewMapIndex[string, string]()

	// Process indexes
	db.processesIdx.byColony = index.NewCompoundIndex()
	db.processesIdx.byExecutor = index.NewMapIndex[string, string]()
	db.processesIdx.byGraph = index.NewMapIndex[string, string]()

	// Attribute indexes
	db.attributesIdx.byTarget = index.NewMapIndex[string, string]()
	db.attributesIdx.byColony = index.NewMapIndex[string, string]()
	db.attributesIdx.byGraph = index.NewMapIndex[string, string]()

	// ProcessGraph indexes
	db.processGraphsIdx.byColony = index.NewCompoundIndex()

	// Log indexes
	db.logsIdx.byProcess = index.NewMapIndex[string, string]()
	db.logsIdx.byExecutor = index.NewMapIndex[string, string]()
	db.logsIdx.byColony = index.NewMapIndex[string, string]()

	// File indexes
	db.filesIdx.byColony = index.NewMapIndex[string, string]()
	db.filesIdx.byLabel = index.NewMapIndex[string, string]()
	db.filesIdx.byName = index.NewMapIndex[string, string]()
}

func (db *EmbeddedDatabase) replayWAL(w wal.WAL) error {
	// Lock all stores for replay
	db.colonies.Lock()
	db.users.Lock()
	db.executors.Lock()
	db.functions.Lock()
	db.generators.Lock()
	db.generatorArgs.Lock()
	db.crons.Lock()
	db.snapshots.Lock()
	db.locations.Lock()
	db.server.Lock()
	db.blueprintDefs.Lock()
	db.blueprints.Lock()
	db.blueprintHistory.Lock()
	db.processes.Lock()
	db.attributes.Lock()
	db.processGraphs.Lock()
	db.logs.Lock()
	db.files.Lock()

	defer func() {
		db.colonies.Unlock()
		db.users.Unlock()
		db.executors.Unlock()
		db.functions.Unlock()
		db.generators.Unlock()
		db.generatorArgs.Unlock()
		db.crons.Unlock()
		db.snapshots.Unlock()
		db.locations.Unlock()
		db.server.Unlock()
		db.blueprintDefs.Unlock()
		db.blueprints.Unlock()
		db.blueprintHistory.Unlock()
		db.processes.Unlock()
		db.attributes.Unlock()
		db.processGraphs.Unlock()
		db.logs.Unlock()
		db.files.Unlock()
	}()

	return w.Replay(func(entry wal.Entry) error {
		switch entry.Entity {
		case "colonies":
			return replayEntry(entry, db.colonies)
		case "users":
			return replayEntry(entry, db.users)
		case "executors":
			return replayEntry(entry, db.executors)
		case "functions":
			return replayEntry(entry, db.functions)
		case "generators":
			return replayEntry(entry, db.generators)
		case "generatorargs":
			return replayEntry(entry, db.generatorArgs)
		case "crons":
			return replayEntry(entry, db.crons)
		case "snapshots":
			return replayEntry(entry, db.snapshots)
		case "locations":
			return replayEntry(entry, db.locations)
		case "server":
			return replayEntry(entry, db.server)
		case "blueprintdefs":
			return replayEntry(entry, db.blueprintDefs)
		case "blueprints":
			return replayEntry(entry, db.blueprints)
		case "blueprinthistory":
			return replayEntry(entry, db.blueprintHistory)
		case "processes":
			return replayEntry(entry, db.processes)
		case "attributes":
			return replayEntry(entry, db.attributes)
		case "processgraphs":
			return replayEntry(entry, db.processGraphs)
		case "logs":
			return replayEntry(entry, db.logs)
		case "files":
			return replayEntry(entry, db.files)
		}
		return nil
	})
}

func replayEntry[K comparable, V any](entry wal.Entry, s *store.Store[K, V]) error {
	// The store must be locked by the caller
	switch entry.Op {
	case wal.OpPut:
		var v V
		if err := json.Unmarshal(entry.Data, &v); err != nil {
			return fmt.Errorf("failed to unmarshal WAL entry for %s: %w", entry.Entity, err)
		}
		// strToKey is captured in store, but we need the key as K
		// Since all our stores use string keys, entry.Key works
		s.ReplayPut(any(entry.Key).(K), &v)
	case wal.OpDelete:
		s.ReplayDelete(any(entry.Key).(K))
	}
	return nil
}

func (db *EmbeddedDatabase) loadAll() error {
	stores := []interface{ LoadAll() error }{
		db.colonies, db.users, db.executors, db.functions,
		db.generators, db.generatorArgs, db.crons, db.snapshots,
		db.locations, db.server, db.blueprintDefs, db.blueprints,
		db.blueprintHistory, db.processes, db.attributes,
		db.processGraphs, db.logs, db.files,
	}
	for _, s := range stores {
		if err := s.LoadAll(); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) setWALOnStores(w wal.WAL) {
	db.colonies.SetWAL(w)
	db.users.SetWAL(w)
	db.executors.SetWAL(w)
	db.functions.SetWAL(w)
	db.generators.SetWAL(w)
	db.generatorArgs.SetWAL(w)
	db.crons.SetWAL(w)
	db.snapshots.SetWAL(w)
	db.locations.SetWAL(w)
	db.server.SetWAL(w)
	db.blueprintDefs.SetWAL(w)
	db.blueprints.SetWAL(w)
	db.blueprintHistory.SetWAL(w)
	db.processes.SetWAL(w)
	db.attributes.SetWAL(w)
	db.processGraphs.SetWAL(w)
	db.logs.SetWAL(w)
	db.files.SetWAL(w)
}

func (db *EmbeddedDatabase) rebuildIndexes() {
	// Colonies
	for _, c := range db.colonies.All() {
		db.coloniesIdx.byID.Add(c.Name, c.ID)
	}

	// Users
	for _, u := range db.users.All() {
		key := u.ColonyName + ":" + u.Name
		db.usersIdx.byColony.Add(key, u.ColonyName)
		db.usersIdx.byID.Add(key, u.ID)
	}

	// Executors
	for _, e := range db.executors.All() {
		db.executorsIdx.byColony.Add(e.ID, e.ColonyName)
		db.executorsIdx.byName.Add(e.ID, e.ColonyName+":"+e.Name)
		if e.BlueprintID != "" {
			db.executorsIdx.byBlueprint.Add(e.ID, e.BlueprintID)
		}
	}

	// Functions
	for _, f := range db.functions.All() {
		db.functionsIdx.byColony.Add(f.FunctionID, f.ColonyName)
		db.functionsIdx.byExecutor.Add(f.FunctionID, f.ColonyName+":"+f.ExecutorName)
		db.functionsIdx.byFullName.Add(f.FunctionID, f.ColonyName+":"+f.ExecutorName+":"+f.FuncName)
	}

	// Generators
	for _, g := range db.generators.All() {
		db.generatorsIdx.byColony.Add(g.ID, g.ColonyName)
		db.generatorsIdx.byName.Add(g.ID, g.ColonyName+":"+g.Name)
	}

	// GeneratorArgs
	for _, ga := range db.generatorArgs.All() {
		db.generatorArgsIdx.byGenerator.Add(ga.ID, ga.GeneratorID)
		db.generatorArgsIdx.byColony.Add(ga.ID, ga.ColonyName)
	}

	// Crons
	for _, c := range db.crons.All() {
		db.cronsIdx.byColony.Add(c.ID, c.ColonyName)
		db.cronsIdx.byName.Add(c.ID, c.ColonyName+":"+c.Name)
	}

	// Snapshots
	for _, s := range db.snapshots.All() {
		db.snapshotsIdx.byColony.Add(s.ID, s.ColonyName)
		db.snapshotsIdx.byName.Add(s.ID, s.ColonyName+":"+s.Name)
	}

	// Locations
	for _, l := range db.locations.All() {
		db.locationsIdx.byColony.Add(l.ID, l.ColonyName)
		db.locationsIdx.byName.Add(l.ID, l.ColonyName+":"+l.Name)
	}

	// Blueprint Definitions
	for _, sd := range db.blueprintDefs.All() {
		db.blueprintDefsIdx.byNamespace.Add(sd.ID, sd.Metadata.ColonyName)
		db.blueprintDefsIdx.byName.Add(sd.ID, sd.Metadata.ColonyName+":"+sd.Metadata.Name)
		db.blueprintDefsIdx.byKind.Add(sd.ID, sd.Spec.Names.Kind)
	}

	// Blueprints
	for _, b := range db.blueprints.All() {
		db.blueprintsIdx.byNamespace.Add(b.ID, b.Metadata.ColonyName)
		db.blueprintsIdx.byName.Add(b.ID, b.Metadata.ColonyName+":"+b.Metadata.Name)
		db.blueprintsIdx.byKind.Add(b.ID, b.Kind)
	}

	// Blueprint History
	for _, h := range db.blueprintHistory.All() {
		db.blueprintHistoryIdx.byBlueprint.Add(h.ID, h.BlueprintID)
	}

	// Processes
	for _, p := range db.processes.All() {
		db.processesIdx.byColony.Add(p.FunctionSpec.Conditions.ColonyName, p.State, index.IndexEntry[string]{SortKey: p.PriorityTime, PrimaryKey: p.ID})
		if p.AssignedExecutorID != "" {
			db.processesIdx.byExecutor.Add(p.ID, p.AssignedExecutorID)
		}
		if p.ProcessGraphID != "" {
			db.processesIdx.byGraph.Add(p.ID, p.ProcessGraphID)
		}
	}

	// Attributes
	for _, a := range db.attributes.All() {
		db.attributesIdx.byTarget.Add(a.ID, a.TargetID)
		db.attributesIdx.byColony.Add(a.ID, a.TargetColonyName)
		if a.TargetProcessGraphID != "" {
			db.attributesIdx.byGraph.Add(a.ID, a.TargetProcessGraphID)
		}
	}

	// ProcessGraphs
	for _, g := range db.processGraphs.All() {
		db.processGraphsIdx.byColony.Add(g.ColonyName, g.State, index.IndexEntry[string]{SortKey: g.SubmissionTime.UnixNano(), PrimaryKey: g.ID})
	}

	// Logs (use ForEach since core.Log has no ID field)
	db.logs.ForEach(func(logID string, l *core.Log) {
		db.logsIdx.byProcess.Add(logID, l.ProcessID)
		db.logsIdx.byExecutor.Add(logID, l.ExecutorName)
		db.logsIdx.byColony.Add(logID, l.ColonyName)
	})

	// Files
	db.fileSeqCounter = 0
	for _, f := range db.files.All() {
		db.filesIdx.byColony.Add(f.ID, f.ColonyName)
		db.filesIdx.byLabel.Add(f.ID, f.ColonyName+":"+f.Label)
		db.filesIdx.byName.Add(f.ID, f.ColonyName+":"+f.Label+":"+f.Name)
		if f.SequenceNumber > db.fileSeqCounter {
			db.fileSeqCounter = f.SequenceNumber
		}
	}
}

func (db *EmbeddedDatabase) addStoresToFlusher() {
	db.flusher.AddStore(db.colonies)
	db.flusher.AddStore(db.users)
	db.flusher.AddStore(db.executors)
	db.flusher.AddStore(db.functions)
	db.flusher.AddStore(db.generators)
	db.flusher.AddStore(db.generatorArgs)
	db.flusher.AddStore(db.crons)
	db.flusher.AddStore(db.snapshots)
	db.flusher.AddStore(db.locations)
	db.flusher.AddStore(db.server)
	db.flusher.AddStore(db.blueprintDefs)
	db.flusher.AddStore(db.blueprints)
	db.flusher.AddStore(db.blueprintHistory)
	db.flusher.AddStore(db.processes)
	db.flusher.AddStore(db.attributes)
	db.flusher.AddStore(db.processGraphs)
	db.flusher.AddStore(db.logs)
	db.flusher.AddStore(db.files)
}

func (db *EmbeddedDatabase) Close() {
	if db.flusher != nil {
		db.flusher.Stop()
	}
	if db.wal != nil {
		db.wal.Truncate()
		db.wal.Close()
	}
}

func (db *EmbeddedDatabase) Drop() error {
	if db.flusher != nil {
		db.flusher.Stop()
		db.flusher = nil
	}
	if db.wal != nil {
		db.wal.Close()
		db.wal = nil
	}
	return os.RemoveAll(db.dataDir)
}

func (db *EmbeddedDatabase) ApplyRetentionPolicy(retentionPeriod int64) error {
	cutoff := time.Now().Add(-time.Duration(retentionPeriod) * time.Second)

	// Delete SUCCESS attributes whose parent process has SubmissionTime < cutoff
	for _, a := range db.attributes.All() {
		if a.State != core.SUCCESS {
			continue
		}
		if p, ok := db.processes.Get(a.TargetID); ok {
			if p.SubmissionTime.Before(cutoff) {
				db.removeAttributeFromIndexes(a.ID, a)
				db.attributes.Delete(a.ID)
			}
		}
	}

	// Delete logs where Timestamp < cutoff
	// Collect IDs first to avoid deadlock (ForEach holds RLock, Delete needs write Lock)
	type logToDelete struct {
		id           string
		processID    string
		executorName string
		colonyName   string
	}
	var logsToDelete []logToDelete
	db.logs.ForEach(func(logID string, l *core.Log) {
		logTime := time.Unix(0, l.Timestamp)
		if logTime.Before(cutoff) {
			logsToDelete = append(logsToDelete, logToDelete{
				id:           logID,
				processID:    l.ProcessID,
				executorName: l.ExecutorName,
				colonyName:   l.ColonyName,
			})
		}
	})
	for _, ld := range logsToDelete {
		db.logsIdx.byProcess.Remove(ld.id, ld.processID)
		db.logsIdx.byExecutor.Remove(ld.id, ld.executorName)
		db.logsIdx.byColony.Remove(ld.id, ld.colonyName)
		db.logs.Delete(ld.id)
	}

	// Delete SUCCESS processes where SubmissionTime < cutoff
	for _, p := range db.processes.All() {
		if p.State == core.SUCCESS && p.SubmissionTime.Before(cutoff) {
			db.removeProcessFromIndexes(p)
			db.RemoveAllAttributesByTargetID(p.ID)
			db.processes.Delete(p.ID)
		}
	}

	// Delete SUCCESS process graphs where SubmissionTime < cutoff
	for _, g := range db.processGraphs.All() {
		if g.State == core.SUCCESS && g.SubmissionTime.Before(cutoff) {
			db.processGraphsIdx.byColony.Remove(g.ColonyName, g.State, index.IndexEntry[string]{
				SortKey:    g.SubmissionTime.UnixNano(),
				PrimaryKey: g.ID,
			})
			db.processGraphs.Delete(g.ID)
		}
	}

	return nil
}
