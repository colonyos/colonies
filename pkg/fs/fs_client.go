package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
)

type FSClient struct {
	coloniesClient *client.ColoniesClient
	colonyID       string
	executorPrvKey string
	s3Client       *S3Client
}

type FileInfo struct {
	Name     string
	Checksum string
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

	err = fsClient.s3Client.Upload(syncPlan.Dir, coloniesFile.Name, coloniesFile.Size)
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
		fsClient.uploadFile(syncPlan, fileInfo)
	}

	// 2. Download all local missing files
	for _, fileInfo := range syncPlan.LocalMissing {
		err := fsClient.s3Client.Download(fileInfo.Name, syncPlan.Dir)
		if err != nil {
			return err
		}
	}

	// 3. Handle conflicts
	// If keepLocalFiles then upload conflicting files to server else download conflicting files to local filesystem
	if syncPlan.KeepLocal {
		for _, fileInfo := range syncPlan.Conflicts {
			fsClient.uploadFile(syncPlan, fileInfo)
		}
	} else {
		for _, fileInfo := range syncPlan.Conflicts {
			err := fsClient.s3Client.Download(fileInfo.Name, syncPlan.Dir)
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
	for _, remoteFilename := range remoteFilenames {
		remoteColoniesFile, err := fsClient.coloniesClient.GetFileByName(fsClient.colonyID, label, remoteFilename, fsClient.executorPrvKey)
		if err != nil {
			return nil, err
		}
		for _, revision := range remoteColoniesFile {
			remoteFileMap[revision.Name] = revision.Checksum
		}
	}

	var localFileMap = make(map[string]string)
	for _, file := range files {
		if !file.IsDir() {
			checksum, err := checksum(dir + "/" + file.Name())
			if err != nil {
				return nil, err
			}
			localFileMap[file.Name()] = checksum
		}
	}

	// Find out which files are missing at the server
	var remoteMissing []*FileInfo
	for filename, checksum := range localFileMap {
		_, ok := remoteFileMap[filename]
		if !ok {
			// File missing on server
			remoteMissing = append(remoteMissing, &FileInfo{Name: filename, Checksum: checksum})
		}
	}

	// Find out which files are missing locally
	var localMissing []*FileInfo
	for filename, checksum := range remoteFileMap {
		_, ok := localFileMap[filename]
		if !ok {
			// File missing locally
			localMissing = append(localMissing, &FileInfo{Name: filename, Checksum: checksum})
		}
	}

	// Calculate conflicts
	var conflicts []*FileInfo
	for filename, checksum := range remoteFileMap {
		// File exists locally, but does not match file on server
		if localFileMap[filename] != checksum {
			if keepLocal {
				localChecksum := localFileMap[filename]
				conflicts = append(conflicts, &FileInfo{Name: filename, Checksum: localChecksum})
			} else {
				conflicts = append(conflicts, &FileInfo{Name: filename, Checksum: checksum})
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
