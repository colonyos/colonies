package cli

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
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

var mutex sync.Mutex

func init() {
	workerCmd.AddCommand(workerStartCmd)
	workerCmd.AddCommand(workerRegisterCmd)
	workerCmd.AddCommand(workerUnregisterCmd)
	rootCmd.AddCommand(workerCmd)

	workerStartCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	workerStartCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	workerStartCmd.Flags().StringVarP(&RuntimeName, "name", "", "", "Runtime name")
	workerStartCmd.Flags().StringVarP(&RuntimeType, "runtimetype", "", "", "Runtime type")
	workerStartCmd.Flags().StringVarP(&CPU, "cpu", "", "", "CPU info")
	workerStartCmd.Flags().IntVarP(&Cores, "cores", "", -1, "Cores")
	workerStartCmd.Flags().IntVarP(&Mem, "mem", "", -1, "Memory [MiB]")
	workerStartCmd.Flags().StringVarP(&GPU, "gpu", "", "", "GPU info")
	workerStartCmd.Flags().IntVarP(&GPUs, "gpus", "", -1, "Number of GPUs")
	workerStartCmd.Flags().StringVarP(&LogDir, "logdir", "", "", "Log directory")
	workerStartCmd.Flags().IntVarP(&Timeout, "timeout", "", 100, "Max time to wait for a process assignment")

	workerRegisterCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	workerRegisterCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	workerRegisterCmd.Flags().StringVarP(&RuntimeName, "name", "", "", "Runtime name")
	workerRegisterCmd.Flags().StringVarP(&RuntimeType, "type", "", "", "Runtime type")
	workerRegisterCmd.Flags().StringVarP(&CPU, "cpu", "", "", "CPU info")
	workerRegisterCmd.Flags().IntVarP(&Cores, "cores", "", -1, "Cores")
	workerRegisterCmd.Flags().IntVarP(&Mem, "mem", "", -1, "Memory [MiB]")
	workerRegisterCmd.Flags().StringVarP(&GPU, "gpu", "", "", "GPU info")
	workerRegisterCmd.Flags().IntVarP(&GPUs, "gpus", "", -1, "Number of GPUs")
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage workers",
	Long:  "Manage workers",
}

var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Register and start a worker",
	Long:  "Register and start a worker",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Starting a worker")
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

		if LogDir == "" {
			LogDir = os.Getenv("COLONIES_LOGDIR")
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		crypto := crypto.CreateCrypto()
		runtimePrvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)
		runtimeID, err := crypto.GenerateID(runtimePrvKey)
		CheckError(err)

		err = os.WriteFile("/tmp/runtimeid", []byte(runtimeID), 0644)
		CheckError(err)

		err = os.WriteFile("/tmp/runtimeprvkey", []byte(runtimePrvKey), 0644)
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

		log.Info("Saving runtimePrvKey to /tmp/runtimeprvkey")
		err = os.WriteFile("/tmp/runtimeprvkey", []byte(runtimePrvKey), 0644)
		CheckError(err)

		log.WithFields(log.Fields{"RuntimeID": runtimeID, "RuntimeName": RuntimeName, "RuntimeType": RuntimeType, "ColonyID": ColonyID, "CPU": CPU, "Cores": Cores, "Mem": Mem, "GPU": GPU, "GPUs": GPUs}).Info("Register a new Runtime")
		runtime := core.CreateRuntime(runtimeID, RuntimeType, RuntimeName, ColonyID, CPU, Cores, Mem, GPU, GPUs, time.Now(), time.Now())
		_, err = client.AddRuntime(runtime, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"RuntimeID": runtimeID}).Info("Approving Runtime")
		err = client.ApproveRuntime(runtimeID, ColonyPrvKey)
		CheckError(err)

		var assignedProcess *core.Process
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			if assignedProcess != nil {
				log.WithFields(log.Fields{"ProcessID": assignedProcess.ID}).Info("Closing process as failed")
				client.Fail(assignedProcess.ID, []string{"SIGTERM"}, runtimePrvKey)
			}
			unregisterRuntime(client)
			os.Exit(0)
		}()

		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime, "ServerHost": ServerHost, "ServerPort": ServerPort}).Info("Worker now waiting for processes to be execute")

		for {
			assignedProcess, err = client.AssignProcess(ColonyID, Timeout, runtimePrvKey)
			if err != nil {
				switch err.(type) {
				case *url.Error:
					fmt.Println("Connection error, trying to reconnect ...")
					time.Sleep(2 * time.Second)
					continue
				default:
					if strings.HasPrefix(err.Error(), "No processes can be selected for runtime with Id") {
						continue
					} else {
						CheckError(err)
					}
				}
			}

			log.WithFields(log.Fields{"ProcessID": assignedProcess.ID}).Info("Worker was assigned a process")
			log.WithFields(log.Fields{"Func": assignedProcess.ProcessSpec.Func, "Args": assignedProcess.ProcessSpec.Args}).Info("Lauching process")
			execCmd := assignedProcess.ProcessSpec.Args
			execCmd = append([]string{assignedProcess.ProcessSpec.Func}, execCmd...)
			execCmdStr := strings.Join(execCmd[:], " ")

			cmd := exec.Command("sh", "-c", execCmdStr)
			cmd.Env = os.Environ()
			for _, attribute := range assignedProcess.Attributes {
				cmd.Env = append(cmd.Env, attribute.Key+"="+attribute.Value)
			}

			cmd.Env = append(cmd.Env, "COLONIES_COLONYID="+ColonyID)
			cmd.Env = append(cmd.Env, "COLONIES_PROCESSID="+assignedProcess.ID)
			cmd.Env = append(cmd.Env, "COLONIES_SERVERHOST="+ServerHost)
			cmd.Env = append(cmd.Env, "COLONIES_SERVERPORT="+strconv.Itoa(ServerPort))
			cmd.Env = append(cmd.Env, "COLONIES_RUNTIMEID="+runtimeID)
			cmd.Env = append(cmd.Env, "COLONIES_RUNTIMEPRVKEY="+runtimePrvKey)

			// Get output
			stdout, err := cmd.StdoutPipe()
			CheckError(err)
			cmd.Stderr = cmd.Stdout

			failure := false
			if err = cmd.Start(); err != nil {
				log.Error(err)
				failure = true
			}

			output := ""
			for {
				tmp := make([]byte, 1)
				_, err := stdout.Read(tmp)
				if err != nil {
					break
				}
				fmt.Print(string(tmp))
				if LogDir != "" {
					f, err := os.OpenFile(LogDir+"/"+assignedProcess.ID+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					CheckError(err)
					defer f.Close()
					_, err = f.WriteString(string(tmp))
					CheckError(err)
				}
				output += string(bytes.Trim(tmp, "\x00"))
			}

			attribute := core.CreateAttribute(assignedProcess.ID, ColonyID, "", core.OUT, "output", output)
			client.AddAttribute(attribute, runtimePrvKey)

			failure = false
			if err = cmd.Wait(); err != nil {
				log.Error(err)
				failure = true
			}

			if failure {
				log.WithFields(log.Fields{"processID": assignedProcess.ID}).Info("Closing process as failed")
				client.Fail(assignedProcess.ID, []string{"Process failed"}, runtimePrvKey)
			} else {
				log.WithFields(log.Fields{"processID": assignedProcess.ID}).Info("Closing process as successful")
				client.Close(assignedProcess.ID, runtimePrvKey)
			}
		}
	},
}

var workerRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a worker",
	Long:  "Register a worker",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Registering a worker")
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

		crypto := crypto.CreateCrypto()

		runtimePrvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)
		runtimeID, err := crypto.GenerateID(runtimePrvKey)
		CheckError(err)

		log.Info("Saving runtimeID to /tmp/runtimeid")
		err = os.WriteFile("/tmp/runtimeid", []byte(runtimeID), 0644)
		CheckError(err)

		err = os.WriteFile("/tmp/runtimeprvkey", []byte(runtimePrvKey), 0644)
		CheckError(err)
		log.Info("Saving runtimePrvKey to /tmp/runtimeprvkey")

		if RuntimeName == "" {
			RuntimeName = os.Getenv("COLONIES_RUNTIMENAME")
			if os.Getenv("HOSTNAME") != "" {
				RuntimeName += "."
				RuntimeName += os.Getenv("HOSTNAME")
			}
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

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		log.WithFields(log.Fields{"runtimeID": runtimeID, "runtimeName": RuntimeName, "runtimeType:": RuntimeType, "colonyID": ColonyID, "CPU": CPU, "Cores": Cores, "Mem": Mem, "GPU": GPU, "GPUs": GPUs}).Info("Register a new Runtime")
		runtime := core.CreateRuntime(runtimeID, RuntimeType, RuntimeName, ColonyID, CPU, Cores, Mem, GPU, GPUs, time.Now(), time.Now())
		_, err = client.AddRuntime(runtime, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"runtimeID": runtimeID}).Info("Approving Runtime")
		err = client.ApproveRuntime(runtimeID, ColonyPrvKey)
		CheckError(err)
	},
}

var workerUnregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister an already started worker",
	Long:  "Unregister an already started worker",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Starting a worker")
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

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		unregisterRuntime(client)
		os.Exit(0)
	},
}

func unregisterRuntime(client *client.ColoniesClient) {
	mutex.Lock()
	defer mutex.Unlock()

	runtimeIDBytes, err := os.ReadFile("/tmp/runtimeid")
	CheckError(err)

	runtimeID := string(runtimeIDBytes)

	err = client.DeleteRuntime(runtimeID, ColonyPrvKey)
	CheckError(err)

	log.WithFields(log.Fields{"RuntimeID": runtimeID}).Info("Runtime unregistered")
}
