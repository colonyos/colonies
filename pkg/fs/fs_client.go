package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	log "github.com/sirupsen/logrus"
)

type FSClient struct {
	coloniesClient *client.ColoniesClient
	colonyName     string
	executorPrvKey string
	s3Client       *S3Client
	Quiet          bool
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

func CreateFSClient(coloniesClient *client.ColoniesClient, colonyName string, executorPrvKey string) (*FSClient, error) {
	fsClient := &FSClient{}
	fsClient.coloniesClient = coloniesClient
	fsClient.colonyName = colonyName
	fsClient.executorPrvKey = executorPrvKey

	s3Client, err := CreateS3Client()
	if err != nil {
		return nil, err
	}
	fsClient.s3Client = s3Client

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

func (fsClient *FSClient) uploadFile(syncPlan *SyncPlan, fileInfo *FileInfo, tracker *progress.Tracker, quite bool) error {
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
		ColonyName:  fsClient.colonyName,
		Label:       syncPlan.Label,
		Name:        fileInfo.Name,
		Size:        fileStat.Size(),
		Checksum:    fileInfo.Checksum,
		ChecksumAlg: "SHA256",
		Reference:   ref}

	if coloniesFile.Size > 0 {
		err = fsClient.s3Client.Upload(syncPlan.Dir, coloniesFile.Name, coloniesFile.Reference.S3Object.Object, coloniesFile.Size, tracker, quite)
		if err != nil {
			return err
		}
	}

	_, err = fsClient.coloniesClient.AddFile(coloniesFile, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	return nil
}

func (fsClient *FSClient) ApplySyncPlan(colonyName string, syncPlan *SyncPlan) error {
	totalCalls := len(syncPlan.RemoteMissing) + len(syncPlan.LocalMissing) + len(syncPlan.Conflicts)
	if totalCalls == 0 {
		return nil
	}

	aggErrChan := make(chan error, totalCalls)

	if _, err := os.Stat(syncPlan.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(syncPlan.Dir, 0755)
		if err != nil {
			return err
		}
	}

	totalUploadSize := int64(0)
	totalDownloadSize := int64(0)
	totalConflictSize := int64(0)

	var pw progress.Writer
	if !fsClient.Quiet {
		pw = utils.ProgressBar(totalCalls)
		go pw.Render()

		for _, fileInfo := range syncPlan.RemoteMissing {
			totalUploadSize += fileInfo.Size
		}

		for _, fileInfo := range syncPlan.LocalMissing {
			totalDownloadSize += fileInfo.Size
		}

		for _, fileInfo := range syncPlan.Conflicts {
			totalConflictSize += fileInfo.Size
		}
	}

	pool := utils.NewWorkerPool(50).Start()

	var uploadTracker progress.Tracker
	var downloadTracker progress.Tracker
	var conflictTracker progress.Tracker

	startTracker := false
	if !fsClient.Quiet {
		messageUploadTracker := fmt.Sprintf("Uploading %s", syncPlan.Label)
		uploadTracker = progress.Tracker{Message: messageUploadTracker, Total: totalUploadSize, Units: progress.UnitsBytes}
		if len(syncPlan.RemoteMissing) > 0 && totalUploadSize > 0 {
			pw.AppendTracker(&uploadTracker)
			startTracker = true
			uploadTracker.Start()
		}

		messageDownloadTracker := fmt.Sprintf("Downloading %s", syncPlan.Dir)
		downloadTracker = progress.Tracker{Message: messageDownloadTracker, Total: totalDownloadSize, Units: progress.UnitsBytes}
		if len(syncPlan.LocalMissing) > 0 && totalDownloadSize > 0 {
			pw.AppendTracker(&downloadTracker)
			startTracker = true
			downloadTracker.Start()
		}

		var messageConflictTracker string
		if syncPlan.KeepLocal {
			messageConflictTracker = fmt.Sprintf("Conflict (keeplocal) %s", syncPlan.Label)
		} else {
			messageConflictTracker = fmt.Sprintf("Conflict (keepremote) %s", syncPlan.Label)
		}
		conflictTracker = progress.Tracker{Message: messageConflictTracker, Total: totalConflictSize, Units: progress.UnitsBytes}
		if len(syncPlan.Conflicts) > 0 && totalConflictSize > 0 {
			pw.AppendTracker(&conflictTracker)
			startTracker = true
			conflictTracker.Start()
		}
	}

	// 1. Upload all remote missing files
	for _, fileInfo := range syncPlan.RemoteMissing {
		errChan := pool.Call(func(arg interface{}) error {
			f := arg.(*FileInfo)
			return fsClient.uploadFile(syncPlan, f, &uploadTracker, fsClient.Quiet)
		}, fileInfo)
		go func() {
			err := <-errChan
			aggErrChan <- err
		}()
	}

	// 2. Download all local missing files
	for _, fileInfo := range syncPlan.LocalMissing {
		errChan := pool.Call(func(arg interface{}) error {
			f := arg.(*FileInfo)
			if f.Size > 0 {
				return fsClient.s3Client.Download(f.Name, f.S3Filename, syncPlan.Dir, &downloadTracker, fsClient.Quiet)
			} else {
				file, err := os.Create(syncPlan.Dir + "/" + f.Name)
				if err != nil {
					return err
				}
				defer file.Close()
				return nil
			}
		}, fileInfo)
		go func() {
			err := <-errChan
			aggErrChan <- err
		}()
	}

	// 3. Handle conflicts
	// If keepLocalFiles then upload conflicting files to server else download conflicting files to local filesystem
	if syncPlan.KeepLocal {
		for _, fileInfo := range syncPlan.Conflicts {
			errChan := pool.Call(func(arg interface{}) error {
				f := arg.(*FileInfo)
				return fsClient.uploadFile(syncPlan, f, &conflictTracker, fsClient.Quiet)
			}, fileInfo)
			go func() {
				err := <-errChan
				aggErrChan <- err
			}()
		}
	} else {
		for _, fileInfo := range syncPlan.Conflicts {
			errChan := pool.Call(func(arg interface{}) error {
				f := arg.(*FileInfo)
				return fsClient.s3Client.Download(f.Name, f.S3Filename, syncPlan.Dir, &conflictTracker, fsClient.Quiet)
			}, fileInfo)
			go func() {
				err := <-errChan
				aggErrChan <- err
			}()
		}
	}

	expectedErrs := totalCalls
	counter := 0
O:
	for {
		select {
		case err := <-aggErrChan:
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Debug("Error in worker")
				return err
			}
			counter++
			expectedErrs--
			if expectedErrs == 0 {
				break O
			}
		}
	}

	if !fsClient.Quiet && startTracker {
		for {
			if !pw.IsRenderInProgress() {
				break
			}
		}

		uploadTracker.MarkAsDone()
		downloadTracker.MarkAsDone()
		conflictTracker.MarkAsDone()
		pw.Stop()
	}

	return nil
}

func (fsClient *FSClient) CalcSyncPlans(dir string, label string, keepLocal bool) ([]*SyncPlan, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		return nil, errors.New(dir + " is not a directory")
	}

	if !strings.HasPrefix(label, "/") {
		label = "/" + label
	}

	syncPlans := make(map[string]*SyncPlan)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			l := ""

			path = strings.Replace(path, `\`, `/`, -1)
			dir = strings.Replace(dir, `\`, `/`, -1)
			if len(strings.TrimPrefix(path, dir)) > 0 { // XXX this line does not work on windows
				l = label + strings.TrimPrefix(path, dir)
			} else {
				l = label
			}

			log.WithFields(log.Fields{"Label": l, "Dir:": dir, "Path": path}).Debug("Calculating sync plan")
			syncPlan, err := fsClient.CalcSyncPlan(path, l, keepLocal)
			if err != nil {
				return err
			}
			syncPlans[l] = syncPlan
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	allLabels, err := fsClient.coloniesClient.GetFileLabelsByName(fsClient.colonyName, label, fsClient.executorPrvKey)
	if err != nil {
		return nil, err
	}
	for _, l := range allLabels {
		if _, ok := syncPlans[l.Name]; !ok {
			subdir := strings.TrimPrefix(l.Name, label)
			syncPlan, err := fsClient.CalcSyncPlan(dir+subdir, l.Name, keepLocal)
			if err != nil {
				return nil, err
			}
			syncPlans[l.Name] = syncPlan
		}
	}

	var a []*SyncPlan
	for _, v := range syncPlans {
		a = append(a, v)
	}

	return a, nil
}

func (fsClient *FSClient) CalcSyncPlan(dir string, label string, keepLocal bool) (*SyncPlan, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.WithFields(log.Fields{"Dir": dir}).Debug("Directory does not exists")
		files = nil
	}

	if fsClient.coloniesClient == nil {
		return nil, errors.New("coloniesClient is nil")
	}

	var pw progress.Writer
	var localTracker progress.Tracker
	if !fsClient.Quiet {
		pw = utils.ProgressBar(2)
		pw.Style().Visibility.Value = false
		go pw.Render()

		if len(files) > 0 {
			localTrackerMessage := fmt.Sprintf("Analyzing %s", dir)
			localTracker = progress.Tracker{Message: localTrackerMessage, Total: int64(len(files)), Units: progress.UnitsDefault}
			pw.AppendTracker(&localTracker)
			localTracker.Start()
		}
	}

	log.WithFields(log.Fields{"Label": label, "Dir:": dir}).Debug("Getting remoteFilenames")
	remoteFileDataArr, err := fsClient.coloniesClient.GetFileData(fsClient.colonyName, label, fsClient.executorPrvKey)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"Label": label, "Dir:": dir, "RemoteFileData": len(remoteFileDataArr)}).Debug("Done getting remoteFileData")

	var remoteFileMap = make(map[string]string)
	var remoteS3FilenameMap = make(map[string]string)
	var remoteFileSizeMap = make(map[string]int64)

	for _, remoteFileData := range remoteFileDataArr {
		remoteFileMap[remoteFileData.Name] = remoteFileData.Checksum
		remoteFileSizeMap[remoteFileData.Name] = remoteFileData.Size
		remoteS3FilenameMap[remoteFileData.Name] = remoteFileData.S3Filename
	}

	var localFileMap = make(map[string]string)
	var localFileSizeMap = make(map[string]int64)
	for _, file := range files {
		// Strange, file.IsDir() says that a file is a not a directory when it is
		// The workaround seems to obtain a new fileinfo struct
		fileInfo, err := os.Stat(dir + "/" + file.Name())
		if err != nil {
			return nil, err
		}
		if !fileInfo.IsDir() {
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

		if !fsClient.Quiet {
			localTracker.Increment(int64(1))
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

	if !fsClient.Quiet {
		if len(files) > 0 {
			for {
				if !pw.IsRenderInProgress() {
					break
				}
			}
			localTracker.MarkAsDone()
			pw.Stop()
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

func (fsClient *FSClient) Download(colonyName string, fileID string, downloadDir string) error {
	file, err := fsClient.coloniesClient.GetFileByID(colonyName, fileID, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	if len(file) != 1 {
		return errors.New("Failed to get file info")
	}

	pw := utils.ProgressBar(1)
	go pw.Render()

	var downloadTracker progress.Tracker
	if !fsClient.Quiet {
		messageDownloadTracker := fmt.Sprintf("Downloading %s", file[0].Name)
		downloadTracker = progress.Tracker{Message: messageDownloadTracker, Total: file[0].Size, Units: progress.UnitsBytes}
		pw.AppendTracker(&downloadTracker)
		downloadTracker.Start()
	}

	err = fsClient.s3Client.Download(file[0].Name, file[0].Reference.S3Object.Object, downloadDir, &downloadTracker, fsClient.Quiet)

	if !fsClient.Quiet {
		for {
			if !pw.IsRenderInProgress() {
				break
			}
		}

		downloadTracker.MarkAsDone()
	}
	return err
}

func (fsClient *FSClient) RemoveFileByID(colonyName string, fileID string) error {
	file, err := fsClient.coloniesClient.GetFileByID(colonyName, fileID, fsClient.executorPrvKey)
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

	return fsClient.coloniesClient.RemoveFileByID(colonyName, fileID, fsClient.executorPrvKey)
}

func (fsClient *FSClient) RemoveFileByName(colonyName string, label string, name string) error {
	file, err := fsClient.coloniesClient.GetFileByName(colonyName, label, name, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	for _, revision := range file {
		err = fsClient.s3Client.Remove(revision.Reference.S3Object.Object)
		if err != nil {
			return err
		}
		err = fsClient.coloniesClient.RemoveFileByID(colonyName, revision.ID, fsClient.executorPrvKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fsClient *FSClient) RemoveAllFilesWithLabel(label string) error {
	allLabels, err := fsClient.coloniesClient.GetFileLabelsByName(fsClient.colonyName, label, fsClient.executorPrvKey)
	if err != nil {
		return err
	}

	var pw progress.Writer
	var labelTracker progress.Tracker

	if !fsClient.Quiet {
		pw = utils.ProgressBar(2)
		pw.Style().Visibility.Value = false
		go pw.Render()

		labelTrackerMessage := fmt.Sprintf("Analyzing labels")
		labelTracker = progress.Tracker{Message: labelTrackerMessage, Total: int64(len(allLabels)), Units: progress.UnitsDefault}
		pw.AppendTracker(&labelTracker)
		labelTracker.Start()
	}

	allFileDataArr := make(map[*core.Label][]*core.FileData)
	for _, l := range allLabels {
		fileDataArr, err := fsClient.coloniesClient.GetFileData(fsClient.colonyName, l.Name, fsClient.executorPrvKey)
		if err != nil {
			return err
		}
		allFileDataArr[l] = fileDataArr
		if !fsClient.Quiet {
			labelTracker.Increment(int64(1))
		}
	}

	if !fsClient.Quiet {
		labelTracker.MarkAsDone()
	}

	totalRevisions := 0
	for _, innerSlice := range allFileDataArr {
		totalRevisions += len(innerSlice)
	}

	var removeTracker progress.Tracker
	if !fsClient.Quiet {
		removeTrackerMessage := fmt.Sprintf("Removing files")
		removeTracker = progress.Tracker{Message: removeTrackerMessage, Total: int64(totalRevisions), Units: progress.UnitsDefault}
		pw.AppendTracker(&removeTracker)
		removeTracker.Start()
	}

	aggErrChan := make(chan error, totalRevisions)
	pool := utils.NewWorkerPool(5).Start()

	type w struct {
		l          string
		filename   string
		s3Filename string
	}

	for l, fileDataArr := range allFileDataArr {
		for _, fileData := range fileDataArr {
			errChan := pool.Call(func(arg interface{}) error {
				w := arg.(w)
				log.WithFields(log.Fields{"S3Filename": w.filename, "BucketName": fsClient.s3Client.BucketName}).Debug("Removing file from S3")
				err := fsClient.s3Client.Remove(w.s3Filename)
				if err != nil {
					return err
				}
				log.WithFields(log.Fields{"ColonyName": fsClient.colonyName, "Filename": w.filename}).Debug("Remove file from Colonies FS")
				err = fsClient.coloniesClient.RemoveFileByName(fsClient.colonyName, w.l, w.filename, fsClient.executorPrvKey)
				if err != nil {
					return err
				}
				if !fsClient.Quiet {
					removeTracker.Increment(int64(1))
				}
				return nil
			}, w{l: l.Name, filename: fileData.Name, s3Filename: fileData.S3Filename})
			go func() {
				err := <-errChan
				aggErrChan <- err
			}()

		}
	}

	expectedErrs := totalRevisions
	counter := 0
O:
	for {
		select {
		case err := <-aggErrChan:
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Debug("Error in worker")
				return err
			}
			counter++
			expectedErrs--
			if expectedErrs == 0 {
				break O
			}
		}
	}

	if !fsClient.Quiet {
		labelTracker.MarkAsDone()
	}

	if !fsClient.Quiet {
		for {
			if !pw.IsRenderInProgress() {
				break
			}
		}

		pw.Stop()
	}

	return nil
}

func (fsClient *FSClient) DownloadSnapshot(snapshotID string, downloadDir string) error {
	snapshot, err := fsClient.coloniesClient.GetSnapshotByID(fsClient.colonyName, snapshotID, fsClient.executorPrvKey)
	if err != nil {
		return err
	}
	for _, fileID := range snapshot.FileIDs {
		file, err := fsClient.coloniesClient.GetFileByID(fsClient.colonyName, fileID, fsClient.executorPrvKey)
		if len(file) != 1 {
			return errors.New("Failed to download file, no revision found")
		}
		downloadFile := false
		// Check if we already have the file
		checksum, err := checksum(downloadDir + "/" + file[0].Name)
		if err != nil {
			downloadFile = true
		} else {
			if checksum != file[0].Checksum {
				downloadFile = true
			}
		}
		if downloadFile {
			dir := strings.TrimPrefix(file[0].Label, snapshot.Label)
			if len(dir) == 0 {
				dir = "/"
			}
			dir = downloadDir + dir
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
			}

			var pw progress.Writer
			var downloadTracker progress.Tracker
			if !fsClient.Quiet {
				pw = utils.ProgressBar(1)
				messageDownloadTracker := fmt.Sprintf("Downloading %s", file[0].Name)
				downloadTracker = progress.Tracker{Message: messageDownloadTracker, Total: file[0].Size, Units: progress.UnitsBytes}
				pw.AppendTracker(&downloadTracker)
				downloadTracker.Start()
			}

			err = fsClient.s3Client.Download(file[0].Name, file[0].Reference.S3Object.Object, dir, &downloadTracker, fsClient.Quiet)
			if err != nil {
				return err
			}

			if !fsClient.Quiet {
				for {
					if !pw.IsRenderInProgress() {
						break
					}
				}

				downloadTracker.MarkAsDone()
			}

			log.WithFields(log.Fields{"Filename": file[0].Name, "DownloadDir": downloadDir}).Debug("Downloading file")
		} else {
			log.WithFields(log.Fields{"Filename": file[0].Name, "DownloadDir": downloadDir}).Debug("Skipping file, already downloaded")
		}
	}

	return nil
}
