package database

type Database interface {
	DatabaseCore
	UserDatabase
	ColonyDatabase
	ExecutorDatabase
	FunctionDatabase
	ProcessDatabase
	AttributeDatabase
	ProcessGraphDatabase
	GeneratorDatabase
	CronDatabase
	LogDatabase
	FileDatabase
	SnapshotDatabase
	BlueprintDatabase
	SecurityDatabase
	LocationDatabase
}