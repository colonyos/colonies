package database

import "github.com/colonyos/colonies/pkg/core"

type FileDatabase interface {
	AddFile(file *core.File) error
	GetFileByID(colonyName string, fileID string) (*core.File, error)
	GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error)
	GetFileByName(colonyName string, label string, name string) ([]*core.File, error)
	GetFilenamesByLabel(colonyName string, label string) ([]string, error)
	GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error)
	RemoveFileByID(colonyName string, fileID string) error
	RemoveFileByName(colonyName string, label string, name string) error
	GetFileLabels(colonyName string) ([]*core.Label, error)
	GetFileLabelsByName(colonyName string, name string, exact bool) ([]*core.Label, error)
	CountFilesWithLabel(colonyName string, label string) (int, error)
	CountFiles(colonyName string) (int, error)
}