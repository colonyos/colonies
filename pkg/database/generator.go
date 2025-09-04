package database

import "github.com/colonyos/colonies/pkg/core"

type GeneratorDatabase interface {
	AddGenerator(generator *core.Generator) error
	SetGeneratorLastRun(generatorID string) error
	SetGeneratorFirstPack(generatorID string) error
	GetGeneratorByID(generatorID string) (*core.Generator, error)
	GetGeneratorByName(colonyName string, name string) (*core.Generator, error)
	FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error)
	FindAllGenerators() ([]*core.Generator, error)
	RemoveGeneratorByID(generatorID string) error
	RemoveAllGeneratorsByColonyName(colonyName string) error
	AddGeneratorArg(generatorArg *core.GeneratorArg) error
	GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error)
	CountGeneratorArgs(generatorID string) (int, error)
	RemoveGeneratorArgByID(generatorArgsID string) error
	RemoveAllGeneratorArgsByGeneratorID(generatorID string) error
	RemoveAllGeneratorArgsByColonyName(generatorID string) error
}