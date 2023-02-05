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
	workerStartCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	workerStartCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	workerStartCmd.Flags().StringVarP(&ExecutorName, "name", "", "", "Executor name")
	workerStartCmd.Flags().StringVarP(&ExecutorType, "executortype", "", "", "Executor type")
	workerStartCmd.Flags().StringVarP(&CPU, "cpu", "", "", "CPU info")
	workerStartCmd.Flags().IntVarP(&Cores, "cores", "", -1, "Cores")
	workerStartCmd.Flags().IntVarP(&Mem, "mem", "", -1, "Memory [MiB]")
	workerStartCmd.Flags().StringVarP(&GPU, "gpu", "", "", "GPU info")
	workerStartCmd.Flags().IntVarP(&GPUs, "gpus", "", -1, "Number of GPUs")
	workerStartCmd.Flags().StringVarP(&LogDir, "logdir", "", "", "Log directory")
	workerStartCmd.Flags().IntVarP(&Timeout, "timeout", "", 100, "Max time to wait for a process assignment")
	workerStartCmd.Flags().Float64VarP(&Long, "long", "", 0, "Longitude")
	workerStartCmd.Flags().Float64VarP(&Lat, "lat", "", 0, "Latitude")

	workerRegisterCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	workerRegisterCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	workerRegisterCmd.Flags().StringVarP(&ExecutorName, "name", "", "", "Executor name")
	workerRegisterCmd.Flags().StringVarP(&ExecutorType, "type", "", "", "Executor type")
	workerRegisterCmd.Flags().StringVarP(&CPU, "cpu", "", "", "CPU info")
	workerRegisterCmd.Flags().IntVarP(&Cores, "cores", "", -1, "Cores")
	workerRegisterCmd.Flags().IntVarP(&Mem, "mem", "", -1, "Memory [MiB]")
	workerRegisterCmd.Flags().StringVarP(&GPU, "gpu", "", "", "GPU info")
	workerRegisterCmd.Flags().IntVarP(&GPUs, "gpus", "", -1, "Number of GPUs")
}

var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Register and start a local Unix process executor",
	Long:  "Register and start a local Unix process executor",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Starting a local executor running Unix processes")
		parseServerEnv()

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")
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
		executorPrvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)
		executorID, err := crypto.GenerateID(executorPrvKey)
		CheckError(err)

		err = os.WriteFile("/tmp/executorid", []byte(executorID), 0644)
		CheckError(err)

		err = os.WriteFile("/tmp/executorprvkey", []byte(executorPrvKey), 0644)
		CheckError(err)

		if ExecutorName == "" {
			ExecutorName = os.Getenv("COLONIES_EXECUTOR_NAME")
		}

		if ExecutorName == "" {
			CheckError(errors.New("Executor name not specified"))
		}

		if ExecutorType == "" {
			ExecutorType = os.Getenv("COLONIES_EXECUTOR_TYPE")
		}

		if ExecutorType == "" {
			CheckError(errors.New("Executor type not specified"))
		}

		log.Info("Saving executorID to /tmp/executorid")
		err = os.WriteFile("/tmp/executorid", []byte(executorID), 0644)
		CheckError(err)

		log.Info("Saving executorPrvKey to /tmp/executorprvkey")
		err = os.WriteFile("/tmp/executorprvkey", []byte(executorPrvKey), 0644)
		CheckError(err)

		log.WithFields(log.Fields{"ExecutorID": executorID, "ExecutorName": ExecutorName, "ExecutorType": ExecutorType, "ColonyID": ColonyID, "Long": Long, "Lat": Lat}).Info("Register a new Executor")
		executor := core.CreateExecutor(executorID, ExecutorType, ExecutorName, ColonyID, time.Now(), time.Now())
		executor.Location.Long = Long
		executor.Location.Lat = Lat
		_, err = client.AddExecutor(executor, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ExecutorID": executorID}).Info("Approving Executor")
		err = client.ApproveExecutor(executorID, ColonyPrvKey)
		CheckError(err)

		var assignedProcess *core.Process
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			if assignedProcess != nil {
				log.WithFields(log.Fields{"ProcessID": assignedProcess.ID}).Info("Closing process as failed")
				client.Fail(assignedProcess.ID, []string{"SIGTERM"}, executorPrvKey)
			}
			unregisterExecutor(client)
			os.Exit(0)
		}()

		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime, "ServerHost": ServerHost, "ServerPort": ServerPort}).Info("Worker now waiting for processes to be execute")

		for {
			assignedProcess, err = client.AssignProcess(ColonyID, Timeout, executorPrvKey)
			if err != nil {
				switch err.(type) {
				case *url.Error:
					fmt.Println("Connection error, trying to reconnect ...")
					time.Sleep(2 * time.Second)
					continue
				default:
					if strings.HasPrefix(err.Error(), "No processes can be selected for executor with Id") {
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

			cmd.Env = append(cmd.Env, "COLONIES_COLONY_ID="+ColonyID)
			cmd.Env = append(cmd.Env, "COLONIES_PROCESS_ID="+assignedProcess.ID)
			cmd.Env = append(cmd.Env, "COLONIES_SERVER_HOST="+ServerHost)
			cmd.Env = append(cmd.Env, "COLONIES_SERVER_PORT="+strconv.Itoa(ServerPort))
			cmd.Env = append(cmd.Env, "COLONIES_EXECUTOR_ID="+executorID)
			cmd.Env = append(cmd.Env, "COLONIES_EXECUTOR_PRVKEY="+executorPrvKey)

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

			failure = false
			if err = cmd.Wait(); err != nil {
				log.Error(err)
				failure = true
			}

			if failure {
				log.WithFields(log.Fields{"processID": assignedProcess.ID}).Info("Closing process as failed")
				client.Fail(assignedProcess.ID, []string{output}, executorPrvKey)
			} else {
				log.WithFields(log.Fields{"processID": assignedProcess.ID}).Info("Closing process as successful")
				client.CloseWithOutput(assignedProcess.ID, []string{output}, executorPrvKey)
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
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")
		}
		if ColonyPrvKey == "" {
			keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
			CheckError(err)

			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		crypto := crypto.CreateCrypto()

		executorPrvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)
		executorID, err := crypto.GenerateID(executorPrvKey)
		CheckError(err)

		log.Info("Saving executorID to /tmp/executorid")
		err = os.WriteFile("/tmp/executorid", []byte(executorID), 0644)
		CheckError(err)

		err = os.WriteFile("/tmp/executorprvkey", []byte(executorPrvKey), 0644)
		CheckError(err)
		log.Info("Saving executorPrvKey to /tmp/executorprvkey")

		if ExecutorName == "" {
			ExecutorName = os.Getenv("COLONIES_EXECUTOR_NAME")
			if os.Getenv("HOSTNAME") != "" {
				ExecutorName += "."
				ExecutorName += os.Getenv("HOSTNAME")
			}
		}

		if ExecutorName == "" {
			CheckError(errors.New("Executor name not specified"))
		}

		if ExecutorType == "" {
			ExecutorType = os.Getenv("COLONIES_EXECUTOR_TYPE")
		}

		if ExecutorType == "" {
			CheckError(errors.New("Executor type not specified"))
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		log.WithFields(log.Fields{"executorID": executorID, "executorName": ExecutorName, "executorType:": ExecutorType, "colonyID": ColonyID}).Info("Register a new Executor")
		executor := core.CreateExecutor(executorID, ExecutorType, ExecutorName, ColonyID, time.Now(), time.Now())
		_, err = client.AddExecutor(executor, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"executorID": executorID}).Info("Approving Executor")
		err = client.ApproveExecutor(executorID, ColonyPrvKey)
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
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")
		}
		if ColonyPrvKey == "" {
			keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
			CheckError(err)

			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		unregisterExecutor(client)
		os.Exit(0)
	},
}

func unregisterExecutor(client *client.ColoniesClient) {
	mutex.Lock()
	defer mutex.Unlock()

	executorIDBytes, err := os.ReadFile("/tmp/executorid")
	CheckError(err)

	executorID := string(executorIDBytes)

	err = client.DeleteExecutor(executorID, ColonyPrvKey)
	CheckError(err)

	log.WithFields(log.Fields{"ExecutorID": executorID}).Info("Executor unregistered")
}
