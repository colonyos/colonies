package embedded

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func copyGenerator(g *core.Generator) *core.Generator {
	cp := *g
	return &cp
}

func copyGeneratorArg(a *core.GeneratorArg) *core.GeneratorArg {
	cp := *a
	return &cp
}

func (db *EmbeddedDatabase) AddGenerator(generator *core.Generator) error {
	existingGenerator, err := db.GetGeneratorByName(generator.ColonyName, generator.Name)
	if err != nil {
		return err
	}

	if existingGenerator != nil {
		return errors.New("Generator with name <" + generator.Name + "> in Colony <" + generator.ColonyName + "> already exists")
	}

	cp := copyGenerator(generator)
	if err := db.generators.Put(cp.ID, cp); err != nil {
		return err
	}

	db.generatorsIdx.byColony.Add(cp.ID, cp.ColonyName)
	db.generatorsIdx.byName.Add(cp.ID, cp.ColonyName+":"+cp.Name)

	return nil
}

func (db *EmbeddedDatabase) SetGeneratorLastRun(generatorID string) error {
	g, ok := db.generators.Get(generatorID)
	if !ok {
		return nil
	}

	cp := *g
	cp.LastRun = time.Now()
	return db.generators.Put(cp.ID, &cp)
}

func (db *EmbeddedDatabase) SetGeneratorFirstPack(generatorID string) error {
	g, ok := db.generators.Get(generatorID)
	if !ok {
		return nil
	}

	cp := *g
	cp.FirstPack = time.Now()
	return db.generators.Put(cp.ID, &cp)
}

func (db *EmbeddedDatabase) GetGeneratorByID(generatorID string) (*core.Generator, error) {
	g, ok := db.generators.Get(generatorID)
	if !ok {
		return nil, nil
	}
	return copyGenerator(g), nil
}

func (db *EmbeddedDatabase) GetGeneratorByName(colonyName string, name string) (*core.Generator, error) {
	ids := db.generatorsIdx.byName.Lookup(colonyName + ":" + name)
	if len(ids) == 0 {
		return nil, nil
	}

	g, ok := db.generators.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return copyGenerator(g), nil
}

func (db *EmbeddedDatabase) FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error) {
	ids := db.generatorsIdx.byColony.Lookup(colonyName)
	var generators []*core.Generator
	for _, id := range ids {
		if len(generators) >= count {
			break
		}
		if g, ok := db.generators.Get(id); ok {
			generators = append(generators, copyGenerator(g))
		}
	}
	return generators, nil
}

func (db *EmbeddedDatabase) FindAllGenerators() ([]*core.Generator, error) {
	all := db.generators.All()
	result := make([]*core.Generator, len(all))
	for i, g := range all {
		result[i] = copyGenerator(g)
	}
	return result, nil
}

func (db *EmbeddedDatabase) RemoveGeneratorByID(generatorID string) error {
	g, ok := db.generators.Get(generatorID)
	if !ok {
		return nil
	}

	db.generatorsIdx.byColony.Remove(generatorID, g.ColonyName)
	db.generatorsIdx.byName.Remove(generatorID, g.ColonyName+":"+g.Name)

	if err := db.generators.Delete(generatorID); err != nil {
		return err
	}

	return db.RemoveAllGeneratorArgsByGeneratorID(generatorID)
}

func (db *EmbeddedDatabase) RemoveAllGeneratorsByColonyName(colonyName string) error {
	ids := db.generatorsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if g, ok := db.generators.Get(id); ok {
			db.generatorsIdx.byName.Remove(id, g.ColonyName+":"+g.Name)
		}
		db.generatorsIdx.byColony.Remove(id, colonyName)
		if err := db.generators.Delete(id); err != nil {
			return err
		}
	}

	return db.RemoveAllGeneratorArgsByColonyName(colonyName)
}

func (db *EmbeddedDatabase) AddGeneratorArg(generatorArg *core.GeneratorArg) error {
	cp := copyGeneratorArg(generatorArg)
	if err := db.generatorArgs.Put(cp.ID, cp); err != nil {
		return err
	}

	db.generatorArgsIdx.byGenerator.Add(cp.ID, cp.GeneratorID)
	db.generatorArgsIdx.byColony.Add(cp.ID, cp.ColonyName)

	return nil
}

func (db *EmbeddedDatabase) GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error) {
	ids := db.generatorArgsIdx.byGenerator.Lookup(generatorID)
	var args []*core.GeneratorArg
	for _, id := range ids {
		if len(args) >= count {
			break
		}
		if a, ok := db.generatorArgs.Get(id); ok {
			args = append(args, copyGeneratorArg(a))
		}
	}
	return args, nil
}

func (db *EmbeddedDatabase) CountGeneratorArgs(generatorID string) (int, error) {
	return db.generatorArgsIdx.byGenerator.Count(generatorID), nil
}

func (db *EmbeddedDatabase) RemoveGeneratorArgByID(generatorArgsID string) error {
	a, ok := db.generatorArgs.Get(generatorArgsID)
	if !ok {
		return nil
	}

	db.generatorArgsIdx.byGenerator.Remove(generatorArgsID, a.GeneratorID)
	db.generatorArgsIdx.byColony.Remove(generatorArgsID, a.ColonyName)

	return db.generatorArgs.Delete(generatorArgsID)
}

func (db *EmbeddedDatabase) RemoveAllGeneratorArgsByGeneratorID(generatorID string) error {
	ids := db.generatorArgsIdx.byGenerator.Lookup(generatorID)
	for _, id := range ids {
		if a, ok := db.generatorArgs.Get(id); ok {
			db.generatorArgsIdx.byColony.Remove(id, a.ColonyName)
		}
		db.generatorArgsIdx.byGenerator.Remove(id, generatorID)
		if err := db.generatorArgs.Delete(id); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllGeneratorArgsByColonyName(colonyName string) error {
	ids := db.generatorArgsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if a, ok := db.generatorArgs.Get(id); ok {
			db.generatorArgsIdx.byGenerator.Remove(id, a.GeneratorID)
		}
		db.generatorArgsIdx.byColony.Remove(id, colonyName)
		if err := db.generatorArgs.Delete(id); err != nil {
			return err
		}
	}
	return nil
}
