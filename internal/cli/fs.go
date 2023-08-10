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
	fsCmd.AddCommand(syncCmd)
	fsCmd.AddCommand(getLabelsCmd)
	fsCmd.AddCommand(listFilesCmd)
	fsCmd.AddCommand(getFileInfoCmd)
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

	getFileInfoCmd.Flags().StringVarP(&FileID, "fileid", "", "", "File Id")
	getFileInfoCmd.Flags().StringVarP(&Label, "label", "l", "", "Label")
	getFileInfoCmd.Flags().StringVarP(&Filename, "name", "n", "", "Filename")

}

var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "Manage file storage",
	Long:  "Manage file storage",
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
		fmt.Println("\nThese files will be downloaded " + SyncDir + ":")
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
			fmt.Println("These files will be replaced at " + SyncDir + ":")
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

		log.WithFields(log.Fields{"SyncDir": SyncDir, "Label": Label, "Dry": Dry}).Debug("Starting a file storage client")
		fsClient, err := fs.CreateFSClient(client, ColonyID, ExecutorPrvKey)
		CheckError(err)

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
