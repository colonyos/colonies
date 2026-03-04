package embedded

import (
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func copyFile(f *core.File) *core.File {
	c := *f
	return &c
}

func (db *EmbeddedDatabase) AddFile(file *core.File) error {
	seqNr := atomic.AddInt64(&db.fileSeqCounter, 1)
	file.SequenceNumber = seqNr
	file.Added = time.Now()

	fileCopy := copyFile(file)
	if err := db.files.Put(fileCopy.ID, fileCopy); err != nil {
		return err
	}

	db.filesIdx.byColony.Add(fileCopy.ID, fileCopy.ColonyName)
	db.filesIdx.byLabel.Add(fileCopy.ID, fileCopy.ColonyName+":"+fileCopy.Label)
	db.filesIdx.byName.Add(fileCopy.ID, fileCopy.ColonyName+":"+fileCopy.Label+":"+fileCopy.Name)

	return nil
}

func (db *EmbeddedDatabase) GetFileByID(colonyName string, fileID string) (*core.File, error) {
	f, ok := db.files.Get(fileID)
	if !ok {
		return nil, nil
	}
	if f.ColonyName != colonyName {
		return nil, nil
	}
	return copyFile(f), nil
}

func (db *EmbeddedDatabase) GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	key := colonyName + ":" + label + ":" + name
	ids := db.filesIdx.byName.Lookup(key)

	var best *core.File
	for _, id := range ids {
		if f, ok := db.files.Get(id); ok {
			if f.ColonyName == colonyName {
				if best == nil || f.SequenceNumber > best.SequenceNumber {
					best = f
				}
			}
		}
	}

	if best == nil {
		return nil, nil
	}
	return []*core.File{copyFile(best)}, nil
}

func (db *EmbeddedDatabase) GetFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	key := colonyName + ":" + label + ":" + name
	ids := db.filesIdx.byName.Lookup(key)

	var result []*core.File
	for _, id := range ids {
		if f, ok := db.files.Get(id); ok {
			if f.ColonyName == colonyName {
				result = append(result, copyFile(f))
			}
		}
	}

	// Sort by sequence number descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].SequenceNumber > result[j].SequenceNumber
	})

	return result, nil
}

func (db *EmbeddedDatabase) GetFilenamesByLabel(colonyName string, label string) ([]string, error) {
	key := colonyName + ":" + label
	ids := db.filesIdx.byLabel.Lookup(key)

	nameSet := make(map[string]struct{})
	for _, id := range ids {
		if f, ok := db.files.Get(id); ok {
			nameSet[f.Name] = struct{}{}
		}
	}

	var filenames []string
	for name := range nameSet {
		filenames = append(filenames, name)
	}

	return filenames, nil
}

func (db *EmbeddedDatabase) GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error) {
	key := colonyName + ":" + label
	ids := db.filesIdx.byLabel.Lookup(key)

	// Keep only the latest version per filename
	latest := make(map[string]*core.File)
	for _, id := range ids {
		if f, ok := db.files.Get(id); ok {
			if existing, found := latest[f.Name]; !found || f.SequenceNumber > existing.SequenceNumber {
				latest[f.Name] = f
			}
		}
	}

	var result []*core.FileData
	for _, f := range latest {
		result = append(result, &core.FileData{
			Name:       f.Name,
			Checksum:   f.Checksum,
			Size:       f.Size,
			S3Filename: f.Reference.S3Object.Object,
		})
	}

	return result, nil
}

func (db *EmbeddedDatabase) RemoveFileByID(colonyName string, fileID string) error {
	f, ok := db.files.Get(fileID)
	if !ok {
		return nil
	}
	if f.ColonyName != colonyName {
		return nil
	}
	db.removeFileFromIndexes(fileID, f)
	return db.files.Delete(fileID)
}

func (db *EmbeddedDatabase) RemoveFileByName(colonyName string, label string, name string) error {
	key := colonyName + ":" + label + ":" + name
	ids := db.filesIdx.byName.Lookup(key)
	for _, id := range ids {
		if f, ok := db.files.Get(id); ok {
			if f.ColonyName == colonyName {
				db.removeFileFromIndexes(id, f)
				db.files.Delete(id)
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) GetFileLabels(colonyName string) ([]*core.Label, error) {
	ids := db.filesIdx.byColony.Lookup(colonyName)

	labelCount := make(map[string]int)
	for _, id := range ids {
		if f, ok := db.files.Get(id); ok {
			labelCount[f.Label]++
		}
	}

	var labels []*core.Label
	for name, count := range labelCount {
		labels = append(labels, &core.Label{Name: name, Files: count})
	}

	return labels, nil
}

func (db *EmbeddedDatabase) GetFileLabelsByName(colonyName string, name string, exact bool) ([]*core.Label, error) {
	ids := db.filesIdx.byColony.Lookup(colonyName)

	labelSet := make(map[string]struct{})
	for _, id := range ids {
		if f, ok := db.files.Get(id); ok {
			if exact {
				// Match label == name or label starts with name+"/"
				if f.Label == name || strings.HasPrefix(f.Label, name+"/") {
					labelSet[f.Label] = struct{}{}
				}
			} else {
				// Prefix match
				if strings.HasPrefix(f.Label, name) {
					labelSet[f.Label] = struct{}{}
				}
			}
		}
	}

	if len(labelSet) == 0 {
		if exact {
			return nil, nil
		}
		return nil, nil
	}

	var labels []*core.Label
	for labelName := range labelSet {
		count, _ := db.CountFilesWithLabel(colonyName, labelName)
		labels = append(labels, &core.Label{Name: labelName, Files: count})
	}

	return labels, nil
}

func (db *EmbeddedDatabase) CountFilesWithLabel(colonyName string, label string) (int, error) {
	key := colonyName + ":" + label
	return db.filesIdx.byLabel.Count(key), nil
}

func (db *EmbeddedDatabase) CountFiles(colonyName string) (int, error) {
	return db.filesIdx.byColony.Count(colonyName), nil
}

func (db *EmbeddedDatabase) removeFileFromIndexes(id string, f *core.File) {
	db.filesIdx.byColony.Remove(id, f.ColonyName)
	db.filesIdx.byLabel.Remove(id, f.ColonyName+":"+f.Label)
	db.filesIdx.byName.Remove(id, f.ColonyName+":"+f.Label+":"+f.Name)
}
