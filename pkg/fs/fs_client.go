package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

type FSClient struct {
	coloniesClient *client.ColoniesClient
	colonyID       string
	executorPrvKey string
	s3Client       *S3Client
}

type FileInfo struct {
	Name       string
	Checksum   string
	Size       int64
	S3Filename string
}

type SyncPlan struct {
	Dir           string
	LocalMissing  []*FileInfo
	RemoteMissing []*FileInfo
	Conflicts     []*FileInfo
	KeepLocal     bool
	Label         string
}

func CreateFSClient(coloniesClient *client.ColoniesClient, colonyID string, executorPrvKey string) (*FSClient, error) {
	fsClient := &FSClient{}
	fsClient.coloniesClient = coloniesClient
	fsClient.colonyID = colonyID
	fsClient.executorPrvKey = executorPrvKey

	s3Client, err := CreateS3Client()
	fsClient.s3Client = s3Client
	if err != nil {
		return nil, err
	}

	return fsClient, nil
}

func checksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buffer := make([]byte, 10000)
	hasher := sha256.New()
	if _, err := io.CopyBuffer(hasher, f, buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (fsClient *FSClient) uploadFile(syncPlan *SyncPlan, fileInfo *FileInfo) error {
	fileStat, err := os.Stat(syncPlan.Dir + "/" + fileInfo.Name)
	if err != nil {
		return err
	}
	s3Object := core.S3Object{
		Server:        fsClient.s3Client.Endpoint,
		Port:          -1,
		TLS:           fsClient.s3Client.TLS,
		AccessKey:     fsClient.s3Client.AccessKey,
		SecretKey:     fsClient.s3Client.SecretKey,
		Region:        fsClient.s3Client.Region,
		EncryptionKey: "",
		EncryptionAlg: "",
		Object:        core.GenerateRandomID(),
		Bucket:        fsClient.s3Client.BucketName,
	}
	ref := core.Reference{Protocol: "s3", S3Object: s3Object}
	coloniesFile := &core.File{
		ColonyID:    fsClient.colonyID,
		Label:       syncPlan.Label,
		Name:        fileInfo.Name,
		Size:        fileStat.Size(),
		Checksum:    fileInfo.Checksum,
		ChecksumAlg: "SHA256",
		Reference:   ref}

	err = fsClient.s3Client.Upload(syncPlan.Dir, coloniesFile.Name, coloniesFile.Reference.S3Object.Object, coloniesFile.Size)
	if err != nil {
		return err
	}
	_, err = fsClient.coloniesClient.AddFile(coloniesFile, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	return nil
}

func (fsClient *FSClient) ApplySyncPlan(colonyID string, syncPlan *SyncPlan) error {
	// 1. Upload all remote missing files
	for _, fileInfo := range syncPlan.RemoteMissing {
		err := fsClient.uploadFile(syncPlan, fileInfo)
		if err != nil {
			return err
		}
	}

	// 2. Download all local missing files
	for _, fileInfo := range syncPlan.LocalMissing {
		err := fsClient.s3Client.Download(fileInfo.Name, fileInfo.S3Filename, syncPlan.Dir)
		if err != nil {
			return err
		}
	}

	// 3. Handle conflicts
	// If keepLocalFiles then upload conflicting files to server else download conflicting files to local filesystem
	if syncPlan.KeepLocal {
		for _, fileInfo := range syncPlan.Conflicts {
			err := fsClient.uploadFile(syncPlan, fileInfo)
			if err != nil {
				return err
			}
		}
	} else {
		for _, fileInfo := range syncPlan.Conflicts {
			err := fsClient.s3Client.Download(fileInfo.Name, fileInfo.S3Filename, syncPlan.Dir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fsClient *FSClient) CalcSyncPlan(dir string, label string, keepLocal bool) (*SyncPlan, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	remoteFilenames, err := fsClient.coloniesClient.GetFilenames(fsClient.colonyID, label, fsClient.executorPrvKey)
	if err != nil {
		return nil, err
	}

	var remoteFileMap = make(map[string]string)
	var remoteS3FilenameMap = make(map[string]string)
	var remoteFileSizeMap = make(map[string]int64)
	for _, remoteFilename := range remoteFilenames {
		remoteColoniesFile, err := fsClient.coloniesClient.GetLatestFileByName(fsClient.colonyID, label, remoteFilename, fsClient.executorPrvKey)
		if err != nil {
			return nil, err
		}
		for _, revision := range remoteColoniesFile {
			remoteFileMap[revision.Name] = revision.Checksum
			remoteFileSizeMap[revision.Name] = revision.Size
			remoteS3FilenameMap[revision.Name] = revision.Reference.S3Object.Object
		}
	}

	var localFileMap = make(map[string]string)
	var localFileSizeMap = make(map[string]int64)
	for _, file := range files {
		if !file.IsDir() {
			checksum, err := checksum(dir + "/" + file.Name())
			if err != nil {
				return nil, err
			}
			localFileMap[file.Name()] = checksum

			fi, err := os.Stat(dir + "/" + file.Name())
			if err != nil {
				return nil, err
			}
			localFileSizeMap[file.Name()] = fi.Size()
		}
	}

	// Find out which files are missing at the server
	var remoteMissing []*FileInfo
	for filename, checksum := range localFileMap {
		_, ok := remoteFileMap[filename]
		if !ok {
			// File missing on server
			size := localFileSizeMap[filename]
			remoteMissing = append(remoteMissing, &FileInfo{Name: filename, Checksum: checksum, Size: size, S3Filename: ""})
		}
	}

	// Find out which files are missing locally
	var localMissing []*FileInfo
	for filename, checksum := range remoteFileMap {
		_, ok := localFileMap[filename]
		if !ok {
			// File missing locally
			size := remoteFileSizeMap[filename]
			s3Filename := remoteS3FilenameMap[filename]
			localMissing = append(localMissing, &FileInfo{Name: filename, Checksum: checksum, Size: size, S3Filename: s3Filename})
		}
	}

	// Calculate conflicts
	var conflicts []*FileInfo
	for filename, checksum := range remoteFileMap {
		// File exists locally, but does not match file on server
		_, ok := localFileMap[filename]
		if ok {
			if localFileMap[filename] != checksum {
				if keepLocal {
					localChecksum := localFileMap[filename]
					size := localFileSizeMap[filename]
					conflicts = append(conflicts, &FileInfo{Name: filename, Checksum: localChecksum, Size: size, S3Filename: ""})
				} else {
					size := remoteFileSizeMap[filename]
					s3Filename := remoteS3FilenameMap[filename]
					conflicts = append(conflicts, &FileInfo{Name: filename, Checksum: checksum, Size: size, S3Filename: s3Filename})
				}
			}
		}
	}

	return &SyncPlan{
		LocalMissing:  localMissing,
		RemoteMissing: remoteMissing,
		Conflicts:     conflicts,
		Dir:           dir,
		Label:         label,
		KeepLocal:     keepLocal}, nil
}

func (fsClient *FSClient) Download(colonyID string, fileID string, downloadDir string) error {
	file, err := fsClient.coloniesClient.GetFileByID(colonyID, fileID, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	if len(file) != 1 {
		return errors.New("Failed to get file info")
	}

	return fsClient.s3Client.Download(file[0].Name, file[0].Reference.S3Object.Object, downloadDir)
}

func (fsClient *FSClient) RemoveFileByID(colonyID string, fileID string) error {
	file, err := fsClient.coloniesClient.GetFileByID(colonyID, fileID, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	if len(file) != 1 {
		return errors.New("Failed to get file info")
	}

	err = fsClient.s3Client.Remove(file[0].Reference.S3Object.Object)
	if err != nil {
		return err
	}

	return fsClient.coloniesClient.RemoveFileByID(colonyID, fileID, fsClient.executorPrvKey)
}

func (fsClient *FSClient) RemoveFileByName(colonyID string, label string, name string) error {
	file, err := fsClient.coloniesClient.GetFileByName(colonyID, label, name, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	for _, revision := range file {
		err = fsClient.s3Client.Remove(revision.Reference.S3Object.Object)
		if err != nil {
			return err
		}
		err = fsClient.coloniesClient.RemoveFileByID(colonyID, revision.ID, fsClient.executorPrvKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fsClient *FSClient) RemoveAllFilesWithLabel(label string) error {
	filenames, err := fsClient.coloniesClient.GetFilenames(fsClient.colonyID, label, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		file, err := fsClient.coloniesClient.GetFileByName(fsClient.colonyID, label, filename, fsClient.executorPrvKey)
		if err != nil {
			return err
		}
		for _, f := range file {
			log.WithFields(log.Fields{"Filename": f.Reference.S3Object.Object, "BucketName": fsClient.s3Client.BucketName}).Debug("Removing file from S3")
			err = fsClient.s3Client.Remove(f.Reference.S3Object.Object)
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{"ColonyID": fsClient.colonyID, "FileID": f.ID}).Debug("Remove file from Colonies FS")
			err = fsClient.coloniesClient.RemoveFileByID(fsClient.colonyID, f.ID, fsClient.executorPrvKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fsClient *FSClient) DownloadSnapshot(snapshotID string, downloadDir string) error {
	snpashot, err := fsClient.coloniesClient.GetSnapshotByID(fsClient.colonyID, snapshotID, fsClient.executorPrvKey)
	if err != nil {
		return err
	}
	for _, fileID := range snpashot.FileIDs {
		file, err := fsClient.coloniesClient.GetFileByID(fsClient.colonyID, fileID, fsClient.executorPrvKey)
		if len(file) != 1 {
			return errors.New("Failed to download file")
		}
		err = fsClient.s3Client.Download(file[0].Name, file[0].Reference.S3Object.Object, downloadDir)
		if err != nil {
			return err
		}
	}

	return nil
}
