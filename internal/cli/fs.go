package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/fs"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	snapshotCmd.AddCommand(createSnapshotCmd)
	snapshotCmd.AddCommand(downloadSnapshotCmd)
	snapshotCmd.AddCommand(listSnapshotsCmd)
	snapshotCmd.AddCommand(infoSnapshotCmd)
	snapshotCmd.AddCommand(removeSnapshotCmd)

	fsCmd.AddCommand(syncCmd)
	fsCmd.AddCommand(getLabelsCmd)
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

func printSyncPlan(syncPlan *fs.SyncPlan) {
	if len(syncPlan.RemoteMissing) > 0 {
		fmt.Println("The files will be uploaded:")
		var uploaded [][]string
		for _, file := range syncPlan.RemoteMissing {
			uploaded = append(uploaded, []string{file.Name, strconv.FormatInt(file.Size/1024, 10) + " KiB", Label})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"File", "Size", "Label"})
		for _, v := range uploaded {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	} else {
		fmt.Println("No files will be uploaded")
	}

	if len(syncPlan.LocalMissing) > 0 {
		fmt.Println("\nThese files will be downloaded to directory <" + SyncDir + ">:")
		var downloaded [][]string
		for _, file := range syncPlan.LocalMissing {
			downloaded = append(downloaded, []string{file.Name, strconv.FormatInt(file.Size/1024, 10) + " KiB", Label})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"File", "Size", "Label"})
		for _, v := range downloaded {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	} else {
		fmt.Println("No files will be downloaded")
	}
	if len(syncPlan.Conflicts) > 0 {
		if syncPlan.KeepLocal {
			fmt.Println("These files will be replaced at the server:")
		} else {
			fmt.Println("These files will be replaced at directory <" + SyncDir + ">:")
		}
		var conflicts [][]string
		for _, file := range syncPlan.Conflicts {
			conflicts = append(conflicts, []string{file.Name, strconv.FormatInt(file.Size/1024, 10) + " KiB", Label})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"File", "Size", "Label"})
		for _, v := range conflicts {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	}
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize a directory with a file storage",
	Long:  "Synchronize a directory with a file storage",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyID, ExecutorPrvKey)
		CheckError(err)

		err = os.Mkdir(SyncDir, 0755)
		if err == nil {
			CheckError(err)
		}

		syncPlan, err := fsClient.CalcSyncPlan(SyncDir, Label, KeepLocal)
		CheckError(err)

		if len(syncPlan.LocalMissing) == 0 && len(syncPlan.RemoteMissing) == 0 && len(syncPlan.Conflicts) == 0 {
			fmt.Println("Nothing to do " + SyncDir + " is already synchronized with label " + Label)
			os.Exit(0)
		}

		if Dry {
			printSyncPlan(syncPlan)
		} else {
			if Yes {
				err = fsClient.ApplySyncPlan(ColonyID, syncPlan)
				CheckError(err)
			} else {
				printSyncPlan(syncPlan)
				fmt.Print("\nAre you sure you want to continue (yes,no): ")
				reader := bufio.NewReader(os.Stdin)
				reply, _ := reader.ReadString('\n')
				if reply == "yes\n" || reply == "y\n" {
					err = fsClient.ApplySyncPlan(ColonyID, syncPlan)
					CheckError(err)
				}
			}
		}
	},
}

var getLabelsCmd = &cobra.Command{
	Use:   "labels",
	Short: "List all registered labels",
	Long:  "List all registered labels",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		coloniesLabels, err := client.GetFileLabels(ColonyID, ExecutorPrvKey)
		CheckError(err)

		var labels [][]string
		for _, coloniesLabel := range coloniesLabels {
			labels = append(labels, []string{coloniesLabel.Name, strconv.Itoa(coloniesLabel.Files)})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Label", "Number of files"})
		for _, v := range labels {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()

	},
}

type fileInfo struct {
	filename string
	fileID   string
	size     string
	added    time.Time
}

var listFilesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all files with a label",
	Long:  "List all files with a label",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		filenames, err := client.GetFilenames(ColonyID, Label, ExecutorPrvKey)
		CheckError(err)

		if len(filenames) == 0 {
			fmt.Println("No files found")
			os.Exit(0)
		}

		var fi []fileInfo
		for _, filename := range filenames {
			coloniesFile, err := client.GetLatestFileByName(ColonyID, Label, filename, ExecutorPrvKey)
			CheckError(err)

			if len(coloniesFile) != 1 {
				CheckError(errors.New("Failed to get file info from Colonies server"))
			}
			fi = append(fi, fileInfo{filename: filename, size: strconv.FormatInt(coloniesFile[0].Size/1024, 10) + " KiB", fileID: coloniesFile[0].ID, added: coloniesFile[0].Added})
		}

		var fileData [][]string
		for _, f := range fi {
			fileData = append(fileData, []string{f.filename, f.size, f.fileID, f.added.Format(TimeLayout)})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Filename", "Size", "Latest ID", "Added"})
		for _, v := range fileData {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	},
}

func printFileInfo(coloniesFile *core.File) {
	fileData := [][]string{
		[]string{"Filename", coloniesFile.Name},
		[]string{"Id", coloniesFile.ID},
		[]string{"ColonyId", coloniesFile.ColonyID},
		[]string{"Added", coloniesFile.Added.Format(TimeLayout)},
		[]string{"Sequence Number", strconv.FormatInt(coloniesFile.SequenceNumber, 10)},
		[]string{"Label", coloniesFile.Label},
		[]string{"Size", strconv.FormatInt(coloniesFile.Size/1024, 10) + " KiB"},
		[]string{"Checksum", coloniesFile.Checksum},
		[]string{"Checksum Alg", coloniesFile.ChecksumAlg},
		[]string{"Protocol", coloniesFile.Reference.Protocol},
		[]string{"S3 Endpoint", coloniesFile.Reference.S3Object.Server},
		[]string{"S3 TLS", strconv.FormatBool(coloniesFile.Reference.S3Object.TLS)},
		[]string{"S3 Region", coloniesFile.Reference.S3Object.Region},
		[]string{"S3 Bucket", coloniesFile.Reference.S3Object.Bucket},
		[]string{"S3 Object", coloniesFile.Reference.S3Object.Object},
		[]string{"S3 Accesskey", "*********************************"},
		[]string{"S3 Secretkey", "*********************************"},
		[]string{"Encryption Key", "*********************************"},
		[]string{"Encryption Alg", ""},
	}
	fileTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range fileData {
		fileTable.Append(v)
	}
	fileTable.SetAlignment(tablewriter.ALIGN_LEFT)
	fileTable.Render()
}

var getFileInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info about a file",
	Long:  "Get info about a file",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		var coloniesFiles []*core.File
		if FileID != "" {
			coloniesFiles, err = client.GetFileByID(ColonyID, FileID, ExecutorPrvKey)
			CheckError(err)
		} else if Filename != "" && Label != "" {
			coloniesFiles, err = client.GetFileByName(ColonyID, Label, Filename, ExecutorPrvKey)
			CheckError(err)
		} else {
			CheckError(errors.New("FileId nor filename + label were specified"))
		}

		counter := 0
		for _, coloniesFile := range coloniesFiles {
			printFileInfo(coloniesFile)
			if counter != len(coloniesFiles)-1 {
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		var coloniesFiles []*core.File
		if FileID != "" {
			coloniesFiles, err = client.GetFileByID(ColonyID, FileID, ExecutorPrvKey)
			CheckError(err)
		} else if Filename != "" && Label != "" {
			fmt.Println(Label)
			coloniesFiles, err = client.GetLatestFileByName(ColonyID, Label, Filename, ExecutorPrvKey)
			CheckError(err)
		} else {
			CheckError(errors.New("FileId nor filename + label were specified"))
		}

		if len(coloniesFiles) != 1 {
			CheckError(errors.New("Failed to get file info"))
		}

		err = os.Mkdir(DownloadDir, 0755)
		if err == nil {
			CheckError(err)
		}

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyID, ExecutorPrvKey)
		CheckError(err)

		err = fsClient.Download(ColonyID, coloniesFiles[0].ID, DownloadDir)
		CheckError(err)

		log.WithFields(log.Fields{"DownloadDir": DownloadDir, FileID: Insecure}).Debug("Downloaded file")
	},
}

var removeFileCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a file from file storage",
	Long:  "Remove a file from file storage",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyID, ExecutorPrvKey)
		CheckError(err)

		if FileID != "" {
			err = fsClient.RemoveFileByID(ColonyID, FileID)
			CheckError(err)
		} else if Filename != "" && Label != "" {
			fmt.Println(Label)
			err = fsClient.RemoveFileByName(ColonyID, Label, Filename)
			CheckError(err)
		} else {
			CheckError(errors.New("FileId nor filename + label were specified"))
		}

		log.WithFields(log.Fields{"FileID": FileID, "Label": Label, "Name": Filename}).Debug("Removed file, local file is not deleted")
	},
}

func printSnapshot(snapshot *core.Snapshot) {
	snapshotData := [][]string{
		[]string{"Id", snapshot.ID},
		[]string{"ColonyId", snapshot.ColonyID},
		[]string{"Label", snapshot.Label},
		[]string{"Name", snapshot.Name},
		[]string{"Added", snapshot.Added.Format(TimeLayout)},
	}
	snapshotTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range snapshotData {
		snapshotTable.Append(v)
	}
	snapshotTable.SetAlignment(tablewriter.ALIGN_LEFT)
	snapshotTable.Render()

	if len(snapshot.FileIDs) > 0 {
		fmt.Println()
		var fileIDData [][]string
		for _, fileID := range snapshot.FileIDs {
			fileIDData = append(fileIDData, []string{fileID})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"File IDs"})
		for _, v := range fileIDData {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	}
}

var createSnapshotCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a snapshot",
	Long:  "Create a snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		snapshot, err := client.CreateSnapshot(ColonyID, Label, SnapshotName, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"Label": Label, "SnapshotName": SnapshotName}).Debug("Creating snapshot")

		fmt.Println("Snapshot:")
		printSnapshot(snapshot)
	},
}

var downloadSnapshotCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a snapshot",
	Long:  "Download a snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyID, ExecutorPrvKey)
		CheckError(err)

		err = os.Mkdir(DownloadDir, 0755)
		if err == nil {
			CheckError(err)
		}

		if SnapshotID != "" {
			err = fsClient.DownloadSnapshot(SnapshotID, DownloadDir)
			CheckError(err)
			log.WithFields(log.Fields{"SnapshotId": SnapshotID, "DownloadDir": DownloadDir}).Debug("Download snapshot")
		} else if SnapshotName != "" {
			err = fsClient.DownloadSnapshot(SnapshotID, DownloadDir)
			snapshot, err := client.GetSnapshotByName(ColonyID, SnapshotName, ExecutorPrvKey)
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		snapshots, err := client.GetSnapshotsByColonyID(ColonyID, ExecutorPrvKey)
		CheckError(err)

		var snapshotData [][]string
		for _, s := range snapshots {
			snapshotData = append(snapshotData, []string{s.Name, s.ID, s.Label, strconv.Itoa(len(s.FileIDs)), s.Added.Format(TimeLayout)})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "ID", "Label", "Files", "Added"})
		for _, v := range snapshotData {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	},
}

var infoSnapshotCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info about a snapshot",
	Long:  "Get info about a snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		if SnapshotID != "" {
			snapshot, err := client.GetSnapshotByID(ColonyID, SnapshotID, ExecutorPrvKey)
			CheckError(err)
			printSnapshot(snapshot)
		} else if SnapshotName != "" {
			snapshot, err := client.GetSnapshotByName(ColonyID, SnapshotName, ExecutorPrvKey)
			CheckError(err)
			printSnapshot(snapshot)
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		log.Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyID, ExecutorPrvKey)
		CheckError(err)

		fmt.Println(fsClient)

		if SnapshotID != "" {
			// TODO
		} else if SnapshotName != "" && Label != "" {
			// TODO
		} else {
			CheckError(errors.New("Snapshot Id nor name was provided"))
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Removing snb´")
	},
}