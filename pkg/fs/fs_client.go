package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/client"
)

type FSClient struct {
	coloniesClient *client.ColoniesClient
	colonyID       string
	executorPrvKey string
}

type FileInfo struct {
	Name     string
	Checksum string
}

func CreateFSClient(coloniesClient *client.ColoniesClient, colonyID string, executorPrvKey string) *FSClient {
	fsClient := &FSClient{}
	fsClient.coloniesClient = coloniesClient
	fsClient.colonyID = colonyID
	fsClient.executorPrvKey = executorPrvKey

	return fsClient

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

func (fsClient *FSClient) CalcSyncPlan(dir string, prefix string) ([]*FileInfo, []*FileInfo, []*FileInfo, []*FileInfo, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	remoteFilenames, err := fsClient.coloniesClient.GetFilenames(fsClient.colonyID, prefix, fsClient.executorPrvKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var remoteFileMap = make(map[string]string)
	for _, remoteFilename := range remoteFilenames {
		remoteColoniesFile, err := fsClient.coloniesClient.GetFileByName(fsClient.colonyID, prefix, remoteFilename, fsClient.executorPrvKey)
		if err != nil {
			return nil, nil, nil, nil, err
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
				return nil, nil, nil, nil, err
			}
			localFileMap[file.Name()] = checksum
		}
	}

	// Find out which files are missing at the server
	var remoteMissingFiles []*FileInfo
	var remoteOverWrite []*FileInfo
	for filename, checksum := range localFileMap {
		_, ok := remoteFileMap[filename]
		if !ok {
			// File missing on server
			remoteMissingFiles = append(remoteMissingFiles, &FileInfo{Name: filename, Checksum: checksum})
		} else {
			// File is on server, but does not match local file
			if remoteFileMap[filename] != checksum {
				remoteOverWrite = append(remoteOverWrite, &FileInfo{Name: filename, Checksum: checksum})
			}
		}
	}

	// Find out which files are missing locally
	var localMissingFiles []*FileInfo
	var localOverWrite []*FileInfo
	for filename, checksum := range remoteFileMap {
		_, ok := localFileMap[filename]
		if !ok {
			// File missing locally
			localMissingFiles = append(localMissingFiles, &FileInfo{Name: filename, Checksum: checksum})
		} else {
			// File exists locally, but does not match file on server
			if localFileMap[filename] != checksum {
				localOverWrite = append(localOverWrite, &FileInfo{Name: filename, Checksum: checksum})
			}
		}
	}

	return localMissingFiles, remoteMissingFiles, localOverWrite, remoteOverWrite, err
}
