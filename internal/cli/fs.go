package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/fs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	snapshotCmd.AddCommand(createSnapshotCmd)
	snapshotCmd.AddCommand(downloadSnapshotCmd)
	snapshotCmd.AddCommand(listSnapshotsCmd)
	snapshotCmd.AddCommand(infoSnapshotCmd)
	snapshotCmd.AddCommand(removeSnapshotCmd)
	snapshotCmd.AddCommand(removeAllSnapshotsCmd)

	labelsCmd.AddCommand(listLabelsCmd)
	labelsCmd.AddCommand(removeLabelCmd)

	fsCmd.AddCommand(syncCmd)
	fsCmd.AddCommand(labelsCmd)
	fsCmd.AddCommand(listFilesCmd)
	fsCmd.AddCommand(getFileInfoCmd)
	fsCmd.AddCommand(getFileCmd)
	fsCmd.AddCommand(removeFileCmd)
	fsCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(fsCmd)

	syncCmd.Flags().StringVarP(&SyncDir, "dir", "d", "", "Local directory to sync")
	syncCmd.MarkFlagRequired("dir")
	syncCmd.Flags().StringVarP(&StorageDriver, "driver", "", "s3", "Storage driver")
	syncCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	syncCmd.MarkFlagRequired("label")
	syncCmd.Flags().BoolVarP(&Dry, "dry", "", false, "Dry run")
	syncCmd.Flags().BoolVarP(&Yes, "yes", "", false, "Anser yes to all questions")
	syncCmd.Flags().BoolVarP(&KeepLocal, "keeplocal", "", true, "Keep local files in case of conflicts")
	syncCmd.Flags().BoolVarP(&SyncPlans, "syncplans", "", false, "Print sync plans details")
	syncCmd.Flags().BoolVarP(&Quite, "quite", "", false, "No outputs")

	listFilesCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	syncCmd.MarkFlagRequired("label")

	getFileInfoCmd.Flags().StringVarP(&FileID, "fileid", "i", "", "File Id")
	getFileInfoCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	getFileInfoCmd.Flags().StringVarP(&Filename, "name", "n", "", "Filename")

	getFileCmd.Flags().StringVarP(&FileID, "fileid", "i", "", "File Id")
	getFileCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	getFileCmd.Flags().StringVarP(&Filename, "name", "n", "", "Filename")
	getFileCmd.Flags().StringVarP(&DownloadDir, "dir", "d", "", "Local directory to download file to")

	removeFileCmd.Flags().StringVarP(&FileID, "fileid", "i", "", "File Id")
	removeFileCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	removeFileCmd.Flags().StringVarP(&Filename, "name", "n", "", "Filename")
	removeFileCmd.Flags().StringVarP(&DownloadDir, "dir", "d", "", "Local directory to download file to")

	createSnapshotCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	createSnapshotCmd.MarkFlagRequired("label")
	createSnapshotCmd.Flags().StringVarP(&SnapshotName, "snapshotname", "n", "", "Snapshot name")
	createSnapshotCmd.MarkFlagRequired("label")

	downloadSnapshotCmd.Flags().StringVarP(&SnapshotID, "snapshotid", "i", "", "Snapshot Id")
	downloadSnapshotCmd.Flags().StringVarP(&SnapshotName, "snapshotname", "n", "", "Snapshot name")
	downloadSnapshotCmd.Flags().StringVarP(&DownloadDir, "dir", "d", "", "Local directory to download files to")

	infoSnapshotCmd.Flags().StringVarP(&SnapshotID, "snapshotid", "i", "", "Snapshot Id")
	infoSnapshotCmd.Flags().StringVarP(&SnapshotName, "snapshotname", "n", "", "Snapshot name")

	removeSnapshotCmd.Flags().StringVarP(&SnapshotID, "snapshotid", "i", "", "Snapshot Id")
	removeSnapshotCmd.Flags().StringVarP(&SnapshotName, "snapshotname", "n", "", "Snapshot name")

	removeLabelCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	removeLabelCmd.MarkFlagRequired("label")
	removeLabelCmd.Flags().BoolVarP(&Yes, "yes", "", false, "Anser yes to all questions")
}

var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "Manage file storage",
	Long:  "Manage file storage",
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage file snapshots",
	Long:  "Manage file snapshots",
}

func printSyncPlans(syncPlans []*fs.SyncPlan) {
	filesToDownload := 0
	filesToUpload := 0
	conflicts := 0

	for _, syncPlan := range syncPlans {
		filesToDownload += len(syncPlan.LocalMissing)
		filesToUpload += len(syncPlan.RemoteMissing)
		conflicts += len(syncPlan.Conflicts)
	}

	var conflictResolution string
	if KeepLocal {
		conflictResolution = "replace-remote"
	} else {
		conflictResolution = "replace-local"
	}

	log.WithFields(log.Fields{"Conflict resolution": conflictResolution, "Download": filesToDownload, "Upload": filesToUpload, "Conflicts": conflicts}).Info("Sync plans completed")

	if SyncPlans {
		printSyncPlansDetails(syncPlans)
	} else {
		log.Info("Add --syncplan flag to view the sync plan in more detail")
	}
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize a directory with a file storage",
	Long:  "Synchronize a directory with a file storage",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if Quite && !Yes {
			CheckError(errors.New("--quite and --yes flags must be used together, please add --yes"))
		}

		fileInfo, err := os.Stat(SyncDir)
		if err == nil {
			if !fileInfo.IsDir() {
				CheckError(errors.New(SyncDir + " is not a directory"))
			}
		}

		Label = strings.TrimRight(Label, "/")
		Label = strings.TrimLeft(Label, "/")
		Label = "/" + Label

		err = os.MkdirAll(SyncDir, 0755)
		CheckError(err)

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyName, PrvKey)
		CheckError(err)

		if Quite {
			fsClient.Quiet = true
		}

		if !Quite {
			log.Info("Calculating sync plans")
		}

		syncPlans, err := fsClient.CalcSyncPlans(SyncDir, Label, KeepLocal)
		CheckError(err)

		counter := 0
		for _, syncPlan := range syncPlans {
			if len(syncPlan.LocalMissing) == 0 && len(syncPlan.RemoteMissing) == 0 && len(syncPlan.Conflicts) == 0 {
				counter++
			}
		}

		if counter == len(syncPlans) {
			if !Quite {
				log.WithFields(log.Fields{"Label": Label, "SyncDir": SyncDir}).Info("Synchronizing, nothing to do, already synchronized")
			}
			os.Exit(0)
		}

		if Dry {
			printSyncPlans(syncPlans)
		} else {
			if Yes {
				for _, syncPlan := range syncPlans {
					err = fsClient.ApplySyncPlan(ColonyName, syncPlan)
					CheckError(err)
				}
			} else {
				printSyncPlans(syncPlans)
				fmt.Print("\nAre you sure you want to continue? (yes,no): ")
				reader := bufio.NewReader(os.Stdin)
				reply, _ := reader.ReadString('\n')
				if reply == "yes\n" || reply == "y\n" {
					for _, syncPlan := range syncPlans {
						err = fsClient.ApplySyncPlan(ColonyName, syncPlan)
						CheckError(err)
					}
				}
			}
		}
	},
}

var labelsCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage file labels",
	Long:  "Manage file labels",
}

var listLabelsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all labels",
	Long:  "List all labels",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		labels, err := client.GetFileLabels(ColonyName, PrvKey)
		CheckError(err)

		if len(labels) > 0 {
			printLabelsTable(labels)
		} else {
			log.Info("No labels found")
		}
	},
}

var removeLabelCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a label",
	Long:  "Remove a label",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyName, PrvKey)
		CheckError(err)

		if Yes {
			err = fsClient.RemoveAllFilesWithLabel(Label)
			CheckError(err)
			log.WithFields(log.Fields{"Label": Label}).Debug("Label removed")
		} else {
			fmt.Print("All files with label <" + Label + "/*> will be removed. Local files are not removed.\n\nAre you sure you want to continue?  (yes,no): ")
			reader := bufio.NewReader(os.Stdin)
			reply, _ := reader.ReadString('\n')
			if reply == "yes\n" || reply == "y\n" {
				err = fsClient.RemoveAllFilesWithLabel(Label)
				CheckError(err)
				log.WithFields(log.Fields{"Label": Label}).Debug("Label removed")
			}
		}
	},
}

type fileInfo struct {
	filename  string
	fileID    string
	size      string
	added     time.Time
	revisions string
}

var listFilesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all files with a label",
	Long:  "List all files with a label",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		fileDataArr, err := client.GetFileData(ColonyName, Label, PrvKey)
		CheckError(err)

		if len(fileDataArr) == 0 {
			log.Info("No files found")
			os.Exit(0)
		}

		var files []fileInfo
		for _, fileData := range fileDataArr {
			coloniesFile, err := client.GetLatestFileByName(ColonyName, Label, fileData.Name, PrvKey)
			CheckError(err)

			allRevisions, err := client.GetFileByName(ColonyName, Label, fileData.Name, PrvKey)
			CheckError(err)

			if len(coloniesFile) != 1 {
				CheckError(errors.New("Failed to get file info from Colonies server"))
			}
			files = append(files, fileInfo{filename: fileData.Name, size: strconv.FormatInt(coloniesFile[0].Size/1024, 10) + " KiB", fileID: coloniesFile[0].ID, added: coloniesFile[0].Added, revisions: strconv.Itoa(len(allRevisions))})
		}

		printFilesTable(files)
	},
}

var getFileInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info about a file",
	Long:  "Get info about a file",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		var err error
		var files []*core.File
		if FileID != "" {
			files, err = client.GetFileByID(ColonyName, FileID, PrvKey)
			CheckError(err)
		} else if Filename != "" && Label != "" {
			files, err = client.GetFileByName(ColonyName, Label, Filename, PrvKey)
			CheckError(err)
		} else {
			CheckError(errors.New("FileId nor filename + label were specified"))
		}

		counter := 0
		for _, file := range files {
			printFileInfoTable(file)
			if counter != len(files)-1 {
				fmt.Println()
			}
			counter++
		}
	},
}

var getFileCmd = &cobra.Command{
	Use:   "get",
	Short: "Download a file from file storage",
	Long:  "Download a file from file storage",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if DownloadDir == "" {
			CheckError(errors.New("Download directory must be specified"))
		}

		var err error
		var coloniesFiles []*core.File
		if FileID != "" {
			coloniesFiles, err = client.GetFileByID(ColonyName, FileID, PrvKey)
			CheckError(err)
		} else if Filename != "" && Label != "" {
			fmt.Println(Label)
			coloniesFiles, err = client.GetLatestFileByName(ColonyName, Label, Filename, PrvKey)
			CheckError(err)
		} else {
			CheckError(errors.New("FileId nor filename + label were specified"))
		}

		if len(coloniesFiles) != 1 {
			CheckError(errors.New("Failed to get file info"))
		}

		err = os.MkdirAll(DownloadDir, 0755)
		if err == nil {
			CheckError(err)
		}

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyName, PrvKey)
		CheckError(err)

		err = fsClient.Download(ColonyName, coloniesFiles[0].ID, DownloadDir)
		CheckError(err)

		log.WithFields(log.Fields{"DownloadDir": DownloadDir, FileID: Insecure}).Debug("Downloaded file")
	},
}

var removeFileCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a file from file storage",
	Long:  "Remove a file from file storage",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyName, PrvKey)
		CheckError(err)

		if FileID != "" {
			err = fsClient.RemoveFileByID(ColonyName, FileID)
			CheckError(err)
		} else if Filename != "" && Label != "" {
			err = fsClient.RemoveFileByName(ColonyName, Label, Filename)
			CheckError(err)
		} else {
			CheckError(errors.New("FileId nor filename + label were specified"))
		}

		log.WithFields(log.Fields{"FileID": FileID, "Label": Label, "Name": Filename}).Info("Removed file (local file is not removed)")
	},
}

var createSnapshotCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a snapshot",
	Long:  "Create a snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if !strings.HasSuffix(Label, "/") {
			Label += "/"
		}
		if !strings.HasPrefix(Label, "/") {
			Label = "/" + Label
		}

		snapshot, err := client.CreateSnapshot(ColonyName, Label, SnapshotName, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"Label": Label, "SnapshotName": SnapshotName}).Info("Snapshot created")

		printSnapshotTable(snapshot, client)
	},
}

var downloadSnapshotCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a snapshot",
	Long:  "Download a snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyName, PrvKey)
		CheckError(err)

		if DownloadDir == "" {
			CheckError(errors.New("Download dir must be specified"))
		}

		err = os.MkdirAll(DownloadDir, 0755)
		if err == nil {
			CheckError(err)
		}
		if !strings.HasSuffix(DownloadDir, "/") {
			DownloadDir += "/"
		}

		if SnapshotID != "" {
			err = fsClient.DownloadSnapshot(SnapshotID, DownloadDir)
			CheckError(err)
			log.WithFields(log.Fields{"SnapshotId": SnapshotID, "DownloadDir": DownloadDir}).Debug("Download snapshot")
		} else if SnapshotName != "" {
			snapshot, err := client.GetSnapshotByName(ColonyName, SnapshotName, PrvKey)
			CheckError(err)
			err = fsClient.DownloadSnapshot(snapshot.ID, DownloadDir)
			CheckError(err)
			log.WithFields(log.Fields{"SnapshotName": SnapshotName, "SnapshotId": snapshot.ID, "DownloadDir": DownloadDir}).Debug("Download snapshot")
		} else {
			CheckError(errors.New("Snapshot Id nor name was provided"))
		}
	},
}

var listSnapshotsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all snapshots",
	Long:  "List all snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		snapshots, err := client.GetSnapshotsByColonyName(ColonyName, PrvKey)
		CheckError(err)

		if len(snapshots) > 0 {
			printSnapshotsTable(snapshots)
		} else {
			log.Info("No snapshots found")
		}
	},
}

var infoSnapshotCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info about a snapshot",
	Long:  "Get info about a snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if SnapshotID != "" {
			snapshot, err := client.GetSnapshotByID(ColonyName, SnapshotID, PrvKey)
			CheckError(err)
			printSnapshotTable(snapshot, client)
		} else if SnapshotName != "" {
			snapshot, err := client.GetSnapshotByName(ColonyName, SnapshotName, PrvKey)
			CheckError(err)
			printSnapshotTable(snapshot, client)
		} else {
			CheckError(errors.New("Snapshot Id nor name was provided"))
		}
	},
}

var removeSnapshotCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a snapshot",
	Long:  "Remove a snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		var err error
		if SnapshotID != "" {
			err = client.RemoveSnapshotByID(ColonyName, SnapshotID, PrvKey)
			CheckError(err)
			log.WithFields(log.Fields{"SnapshotId": SnapshotID}).Info("Snapshot removed")
		} else if SnapshotName != "" {
			err = client.RemoveSnapshotByName(ColonyName, SnapshotName, PrvKey)
			CheckError(err)
			log.WithFields(log.Fields{"SnapshotName": SnapshotName}).Info("Snapshot removed")
		} else {
			CheckError(errors.New("Snapshot Id nor name was provided"))
		}
	},
}

var removeAllSnapshotsCmd = &cobra.Command{
	Use:   "removeall",
	Short: "Remove all snapshots",
	Long:  "Remove all snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		fmt.Print("WARNING!!! Are you sure you want to remove all snapshots in colony <" + ColonyName + ">. This operation cannot be undone! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			err := client.RemoveAllSnapshots(ColonyName, PrvKey)
			CheckError(err)
		} else {
			log.Info("Aborting ...")
			os.Exit(0)
		}

		log.Info("All snapshots removed")
	},
}
