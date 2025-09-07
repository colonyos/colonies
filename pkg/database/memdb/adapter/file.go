package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// FileDatabase interface implementation

func (a *ColonyOSAdapter) AddFile(file *core.File) error {
	doc := &memdb.VelocityDocument{
		ID:     file.ID,
		Fields: a.fileToFields(file),
	}
	
	return a.db.Insert(context.Background(), FilesCollection, doc)
}

func (a *ColonyOSAdapter) GetFileByID(colonyName string, fileID string) (*core.File, error) {
	doc, err := a.db.Get(context.Background(), FilesCollection, fileID)
	if err != nil {
		return nil, err
	}
	
	file, err := a.fieldsToFile(doc.Fields)
	if err != nil {
		return nil, err
	}
	
	if file.ColonyName != colonyName {
		return nil, fmt.Errorf("file not found in colony")
	}
	
	return file, nil
}

func (a *ColonyOSAdapter) GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	files, err := a.GetFileByName(colonyName, label, name)
	if err != nil {
		return nil, err
	}
	
	// Return the most recent file (highest sequence number)
	if len(files) == 0 {
		return files, nil
	}
	
	var latest *core.File
	for _, file := range files {
		if latest == nil || file.SequenceNumber > latest.SequenceNumber {
			latest = file
		}
	}
	
	return []*core.File{latest}, nil
}

func (a *ColonyOSAdapter) GetFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	result, err := a.db.List(context.Background(), FilesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var files []*core.File
	for _, doc := range result {
		file, err := a.fieldsToFile(doc.Fields)
		if err == nil && 
		   file.ColonyName == colonyName && 
		   file.Label == label && 
		   file.Name == name {
			files = append(files, file)
		}
	}
	
	return files, nil
}

func (a *ColonyOSAdapter) GetFiles(colonyName string) ([]*core.File, error) {
	return a.GetFilesByColonyName(colonyName)
}

func (a *ColonyOSAdapter) GetFilesByColonyName(colonyName string) ([]*core.File, error) {
	result, err := a.db.List(context.Background(), FilesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var files []*core.File
	for _, doc := range result {
		file, err := a.fieldsToFile(doc.Fields)
		if err == nil && file.ColonyName == colonyName {
			files = append(files, file)
		}
	}
	
	return files, nil
}

func (a *ColonyOSAdapter) GetFilesByLabel(colonyName, label string) ([]*core.File, error) {
	result, err := a.db.List(context.Background(), FilesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var files []*core.File
	for _, doc := range result {
		file, err := a.fieldsToFile(doc.Fields)
		if err == nil && file.ColonyName == colonyName && file.Label == label {
			files = append(files, file)
		}
	}
	
	return files, nil
}

func (a *ColonyOSAdapter) GetFilesByName(colonyName, name string) ([]*core.File, error) {
	result, err := a.db.List(context.Background(), FilesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var files []*core.File
	for _, doc := range result {
		file, err := a.fieldsToFile(doc.Fields)
		if err == nil && file.ColonyName == colonyName && file.Name == name {
			files = append(files, file)
		}
	}
	
	return files, nil
}

func (a *ColonyOSAdapter) GetFilenamesByLabel(colonyName string, label string) ([]string, error) {
	files, err := a.GetFilesByLabel(colonyName, label)
	if err != nil {
		return nil, err
	}
	
	nameSet := make(map[string]bool)
	var names []string
	for _, file := range files {
		if !nameSet[file.Name] {
			nameSet[file.Name] = true
			names = append(names, file.Name)
		}
	}
	
	return names, nil
}

func (a *ColonyOSAdapter) GetFileDataByID(fileID string) ([]byte, error) {
	// This is a simplified implementation - in a real system,
	// this would fetch file data from S3 or another storage system
	// based on the file's Reference field
	return nil, fmt.Errorf("file data retrieval not implemented")
}

func (a *ColonyOSAdapter) GetFileDataByName(colonyName, fileName string) ([]byte, error) {
	// This is a simplified implementation - in a real system,
	// this would fetch file data from S3 or another storage system
	return nil, fmt.Errorf("file data retrieval not implemented")
}

func (a *ColonyOSAdapter) GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error) {
	files, err := a.GetFilesByLabel(colonyName, label)
	if err != nil {
		return nil, err
	}
	
	var fileDataList []*core.FileData
	for _, file := range files {
		fileData := &core.FileData{
			Name:       file.Name,
			Checksum:   file.Checksum,
			Size:       file.Size,
			S3Filename: file.Reference.S3Object.Object,
		}
		fileDataList = append(fileDataList, fileData)
	}
	
	return fileDataList, nil
}

func (a *ColonyOSAdapter) GetFileLabels(colonyName string) ([]*core.Label, error) {
	files, err := a.GetFilesByColonyName(colonyName)
	if err != nil {
		return nil, err
	}
	
	labelCounts := make(map[string]int)
	for _, file := range files {
		labelCounts[file.Label]++
	}
	
	var labels []*core.Label
	for labelName, count := range labelCounts {
		label := &core.Label{
			Name:  labelName,
			Files: count,
		}
		labels = append(labels, label)
	}
	
	return labels, nil
}

func (a *ColonyOSAdapter) GetFileLabelsByName(colonyName string, name string, exact bool) ([]*core.Label, error) {
	files, err := a.GetFilesByColonyName(colonyName)
	if err != nil {
		return nil, err
	}
	
	labelCounts := make(map[string]int)
	for _, file := range files {
		nameMatches := false
		if exact {
			nameMatches = file.Name == name
		} else {
			nameMatches = strings.Contains(strings.ToLower(file.Name), strings.ToLower(name))
		}
		
		if nameMatches {
			labelCounts[file.Label]++
		}
	}
	
	var labels []*core.Label
	for labelName, count := range labelCounts {
		label := &core.Label{
			Name:  labelName,
			Files: count,
		}
		labels = append(labels, label)
	}
	
	return labels, nil
}

func (a *ColonyOSAdapter) RemoveFileByID(colonyName string, fileID string) error {
	// Verify file exists in the specified colony
	_, err := a.GetFileByID(colonyName, fileID)
	if err != nil {
		return err
	}
	
	return a.db.Delete(context.Background(), FilesCollection, fileID)
}

func (a *ColonyOSAdapter) RemoveFileByName(colonyName string, label string, name string) error {
	files, err := a.GetFileByName(colonyName, label, name)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if err := a.db.Delete(context.Background(), FilesCollection, file.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveFilesByColonyName(colonyName string) error {
	files, err := a.GetFilesByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if err := a.db.Delete(context.Background(), FilesCollection, file.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveFilesByLabel(colonyName, label string) error {
	files, err := a.GetFilesByLabel(colonyName, label)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if err := a.db.Delete(context.Background(), FilesCollection, file.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveFilesByName(colonyName, name string) error {
	files, err := a.GetFilesByName(colonyName, name)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if err := a.db.Delete(context.Background(), FilesCollection, file.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) CountFilesWithLabel(colonyName string, label string) (int, error) {
	files, err := a.GetFilesByLabel(colonyName, label)
	if err != nil {
		return 0, err
	}
	
	return len(files), nil
}

func (a *ColonyOSAdapter) CountFiles(colonyName string) (int, error) {
	return a.CountFilesByColonyName(colonyName)
}

func (a *ColonyOSAdapter) CountFilesByColonyName(colonyName string) (int, error) {
	files, err := a.GetFilesByColonyName(colonyName)
	if err != nil {
		return 0, err
	}
	
	return len(files), nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) fileToFields(file *core.File) map[string]interface{} {
	fields := map[string]interface{}{
		"id":              file.ID,
		"colony_name":     file.ColonyName,
		"label":           file.Label,
		"name":            file.Name,
		"size":            file.Size,
		"sequence_number": file.SequenceNumber,
		"checksum":        file.Checksum,
		"checksum_alg":    file.ChecksumAlg,
		"added":           file.Added,
	}
	
	// Serialize Reference
	if refData, err := json.Marshal(file.Reference); err == nil {
		fields["reference"] = string(refData)
	}
	
	return fields
}

func (a *ColonyOSAdapter) fieldsToFile(fields map[string]interface{}) (*core.File, error) {
	file := &core.File{}
	
	if id, ok := fields["id"].(string); ok {
		file.ID = id
	}
	if colonyName, ok := fields["colony_name"].(string); ok {
		file.ColonyName = colonyName
	}
	if label, ok := fields["label"].(string); ok {
		file.Label = label
	}
	if name, ok := fields["name"].(string); ok {
		file.Name = name
	}
	if size, ok := fields["size"].(int64); ok {
		file.Size = size
	} else if size, ok := fields["size"].(float64); ok {
		file.Size = int64(size)
	}
	if seqNum, ok := fields["sequence_number"].(int64); ok {
		file.SequenceNumber = seqNum
	} else if seqNum, ok := fields["sequence_number"].(float64); ok {
		file.SequenceNumber = int64(seqNum)
	}
	if checksum, ok := fields["checksum"].(string); ok {
		file.Checksum = checksum
	}
	if checksumAlg, ok := fields["checksum_alg"].(string); ok {
		file.ChecksumAlg = checksumAlg
	}
	
	// Deserialize Reference
	if refStr, ok := fields["reference"].(string); ok {
		var reference core.Reference
		if err := json.Unmarshal([]byte(refStr), &reference); err == nil {
			file.Reference = reference
		}
	}
	
	return file, nil
}