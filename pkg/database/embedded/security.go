package embedded

func (db *EmbeddedDatabase) SetServerID(oldServerID, newServerID string) error {
	if oldServerID == "" {
		return db.server.Put("serverid", &newServerID)
	}

	return db.server.Put("serverid", &newServerID)
}

func (db *EmbeddedDatabase) GetServerID() (string, error) {
	v, ok := db.server.Get("serverid")
	if !ok {
		return "", nil
	}
	return *v, nil
}

func (db *EmbeddedDatabase) ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error {
	colony, ok := db.colonies.Get(colonyName)
	if !ok {
		return nil
	}

	// Update index
	db.coloniesIdx.byID.Remove(colonyName, oldColonyID)

	cp := *colony
	cp.ID = newColonyID
	if err := db.colonies.Put(colonyName, &cp); err != nil {
		return err
	}

	db.coloniesIdx.byID.Add(colonyName, newColonyID)

	return nil
}

func (db *EmbeddedDatabase) ChangeUserID(colonyName string, oldUserID, newUserID string) error {
	keys := db.usersIdx.byID.Lookup(oldUserID)
	for _, key := range keys {
		u, ok := db.users.Get(key)
		if !ok {
			continue
		}
		if u.ColonyName == colonyName {
			db.usersIdx.byID.Remove(key, oldUserID)
			cp := *u
			cp.ID = newUserID
			if err := db.users.Put(key, &cp); err != nil {
				return err
			}
			db.usersIdx.byID.Add(key, newUserID)
			return nil
		}
	}
	return nil
}

func (db *EmbeddedDatabase) ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error {
	executor, ok := db.executors.Get(oldExecutorID)
	if !ok {
		return nil
	}
	if executor.ColonyName != colonyName {
		return nil
	}

	// Remove old entry and indexes
	db.executorsIdx.byColony.Remove(oldExecutorID, executor.ColonyName)
	db.executorsIdx.byName.Remove(oldExecutorID, executor.ColonyName+":"+executor.Name)
	if executor.BlueprintID != "" {
		db.executorsIdx.byBlueprint.Remove(oldExecutorID, executor.BlueprintID)
	}
	if err := db.executors.Delete(oldExecutorID); err != nil {
		return err
	}

	// Add with new ID (copy to avoid modifying stored pointer)
	cp := copyExecutor(executor)
	cp.ID = newExecutorID
	if err := db.executors.Put(newExecutorID, cp); err != nil {
		return err
	}
	db.executorsIdx.byColony.Add(newExecutorID, cp.ColonyName)
	db.executorsIdx.byName.Add(newExecutorID, cp.ColonyName+":"+cp.Name)
	if cp.BlueprintID != "" {
		db.executorsIdx.byBlueprint.Add(newExecutorID, cp.BlueprintID)
	}

	// Update any processes assigned to this executor
	ids := db.processesIdx.byExecutor.Lookup(oldExecutorID)
	for _, id := range ids {
		if p, ok := db.processes.Get(id); ok {
			if p.AssignedExecutorID == oldExecutorID {
				db.processesIdx.byExecutor.Remove(p.ID, oldExecutorID)
				pcp := *p
				pcp.AssignedExecutorID = newExecutorID
				db.processes.Put(pcp.ID, &pcp)
				db.processesIdx.byExecutor.Add(pcp.ID, newExecutorID)
			}
		}
	}

	return nil
}
