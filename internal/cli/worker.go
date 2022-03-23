package cli

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"errors"
	"os"

	"os/exec"
	"os/signal"
	"syscall"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	workerCmd.AddCommand(workerStartCmd)
	workerCmd.AddCommand(workerUnregisterCmd)
	rootCmd.AddCommand(workerCmd)

	workerStartCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	workerStartCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	workerStartCmd.Flags().StringVarP(&RuntimeName, "name", "", "", "Runtime name")
	workerStartCmd.Flags().StringVarP(&RuntimeType, "type", "", "", "Runtime type")
	workerStartCmd.Flags().StringVarP(&CPU, "cpu", "", "", "CPU info")
	workerStartCmd.Flags().IntVarP(&Cores, "cores", "", -1, "Cores")
	workerStartCmd.Flags().IntVarP(&Mem, "mem", "", -1, "Memory [MiB]")
	workerStartCmd.Flags().StringVarP(&GPU, "gpu", "", "", "GPU info")
	workerStartCmd.Flags().IntVarP(&GPUs, "gpus", "", -1, "Number of GPUs")
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage a Colonies Worker",
	Long:  "Manage a Colonies Worker",
}

var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Colonies Worker",
	Long:  "Start a Colonies Worker",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Starting a Colonies Worker")
		parseServerEnv()

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey = os.Getenv("COLONIES_COLONYPRVKEY")
		}
		if ColonyPrvKey == "" {
			keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
			CheckError(err)

			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			unregisterRuntime(client)
			os.Exit(0)
		}()

		crypto := crypto.CreateCrypto()
		runtimePrvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)
		runtimeID, err := crypto.GenerateID(runtimePrvKey)
		CheckError(err)

		err = os.WriteFile("/tmp/runtimeid", []byte(runtimeID), 0644)
		CheckError(err)

		if RuntimeName == "" {
			RuntimeName = os.Getenv("COLONIES_RUNTIMENAME")
		}

		if RuntimeName == "" {
			CheckError(errors.New("Runtime name not specified"))
		}

		if RuntimeType == "" {
			RuntimeType = os.Getenv("COLONIES_RUNTIMETYPE")
		}

		if RuntimeType == "" {
			CheckError(errors.New("Runtime type not specified"))
		}

		log.Info("Saving runtimeID to /tmp/runtimeid")
		err = os.WriteFile("/tmp/runtimeid", []byte(runtimeID), 0644)
		CheckError(err)

		log.WithFields(log.Fields{"runtimeID": runtimeID, "runtimeName": RuntimeName, "runtimeType:": RuntimeType, "colonyID": ColonyID, "CPU": CPU, "Cores": Cores, "Mem": Mem, "GPU": GPU, "GPUs": GPUs}).Info("Register a new Runtime")
		runtime := core.CreateRuntime(runtimeID, RuntimeType, RuntimeName, ColonyID, CPU, Cores, Mem, GPU, GPUs, time.Now(), time.Now())
		_, err = client.AddRuntime(runtime, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"runtimeID": runtimeID}).Info("Approving Runtime")
		err = client.ApproveRuntime(runtimeID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Colonies Worker ready to serve")
		for {
			assignedProcess, err := client.AssignProcess(ColonyID, runtimePrvKey)
			if err != nil {
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			log.WithFields(log.Fields{"processID": assignedProcess.ID}).Info("Colonies Worker was assigned a process")
			log.WithFields(log.Fields{"Cmd": assignedProcess.ProcessSpec.Cmd, "Args": assignedProcess.ProcessSpec.Args}).Info("Executing")
			cmd := exec.Command(assignedProcess.ProcessSpec.Cmd, assignedProcess.ProcessSpec.Args...)
			cmd.Env = os.Environ()
			for _, attribute := range assignedProcess.Attributes {
				cmd.Env = append(cmd.Env, attribute.Key+"="+attribute.Value)
			}
			cmd.Env = append(cmd.Env, "COLONIES_COLONYID="+ColonyID)
			cmd.Env = append(cmd.Env, "COLONIES_SERVER_HOST="+ServerHost)
			cmd.Env = append(cmd.Env, "COLONIES_SERVER_PORT="+strconv.Itoa(ServerPort))
			cmd.Env = append(cmd.Env, "COLONIES_RUNTIMEID="+runtimeID)
			cmd.Env = append(cmd.Env, "COLONIES_RUNTIMEPRVKEY="+runtimePrvKey)

			// Get output
			stdout, err := cmd.StdoutPipe()
			cmd.Stderr = cmd.Stdout
			CheckError(err)

			failure := false
			if err = cmd.Start(); err != nil {
				log.Error(err)
				failure = true
			}

			output := ""
			for {
				tmp := make([]byte, 400)
				_, err := stdout.Read(tmp)
				fmt.Print(tmp)
				if err != nil {
					break
				}
				output += string(bytes.Trim(tmp, "\x00"))
			}

			attribute := core.CreateAttribute(assignedProcess.ID, ColonyID, core.OUT, "output", output)
			client.AddAttribute(attribute, runtimePrvKey)

			failure = false
			if err = cmd.Wait(); err != nil {
				log.Error(err)
				failure = true
			}

			if failure {
				client.CloseFailed(assignedProcess.ID, runtimePrvKey)
				log.WithFields(log.Fields{"processID": assignedProcess.ID}).Info("Closing process as Failed")
			} else {
				client.CloseSuccessful(assignedProcess.ID, runtimePrvKey)
				log.WithFields(log.Fields{"processID": assignedProcess.ID}).Info("Closing process as Successful")
			}
		}
	},
}

var workerUnregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister an already started Colonies Worker",
	Long:  "Unregister an already started Colonies Worker",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Starting a Colonies Worker")
		parseServerEnv()

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey = os.Getenv("COLONIES_COLONYPRVKEY")
		}
		if ColonyPrvKey == "" {
			keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
			CheckError(err)

			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure

		unregisterRuntime(client)
		os.Exit(0)
	},
}

func unregisterRuntime(client *client.ColoniesClient) {
	runtimeIDBytes, err := os.ReadFile("/tmp/runtimeid")
	CheckError(err)
	runtimeID := string(runtimeIDBytes)

	err = client.DeleteRuntime(runtimeID, ColonyPrvKey)
	CheckError(err)

	log.WithFields(log.Fields{"runtimeID": runtimeID}).Info("Runtime unregistered")
}
