package main

import (
	"C"
)
import (
	"fmt"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/fs"
	log "github.com/sirupsen/logrus"
)

//export sync
func sync(chost *C.char, cport C.int, cinsecure C.int, cskiptlsverify C.int, cdir *C.char, clabel *C.char, ckeeplocal C.int, ccolonyname *C.char, cprvkey *C.char) C.int {
	host := C.GoString(chost)
	port := int(cport)
	insecure := int(cinsecure) != 0
	skipTLSVerify := int(cskiptlsverify) != 0
	dir := C.GoString(cdir)
	label := C.GoString(clabel)
	keepLocal := int(ckeeplocal) != 0
	colonyName := C.GoString(ccolonyname)
	prvKey := C.GoString(cprvkey)

	log.WithFields(log.Fields{"host": host, "port": port, "insecure": insecure, "skipTLSVerify": skipTLSVerify, "dir": dir, "label": label, "keepLocal": keepLocal, "colonyName": colonyName, "prvKey": prvKey}).Info("Syncing...")

	client := client.CreateColoniesClient(host, port, insecure, skipTLSVerify)

	fsClient, err := fs.CreateFSClient(client, colonyName, prvKey)
	if err != nil {
		fmt.Println("failed to create fs client:", err)
		return C.int(1)
	}

	syncPlans, err := fsClient.CalcSyncPlans(dir, label, keepLocal)
	if err != nil {
		fmt.Println("failed to calculate sync plans:", err)
		return C.int(1)
	}

	for _, syncPlan := range syncPlans {
		err = fsClient.ApplySyncPlan(colonyName, syncPlan)
		if err != nil {
			fmt.Println("failed to apply sync plan:", err)
			return C.int(1)
		}
	}

	return C.int(0)
}

func main() {}
