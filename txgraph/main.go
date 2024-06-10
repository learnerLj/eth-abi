package txgraph

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"tx-analyze/staticts"
	"tx-analyze/types"
	"tx-analyze/utils"

	cmap "github.com/orcaman/concurrent-map/v2"
)

var funcSigMap map[string][]string
var addrABIMap map[string]string
var eventSigMap map[string]string

func MakeGraph(fromDir, toDir, signatureDbPath, eventDbPath, abiDbPath string,
	maxConcurrentFiles, maxConcurrentTx int, dryRun bool) {
	utils.Init()

	log.SetFlags(log.Lshortfile)

	funcSigMap = utils.LoadFunctionSignatures(signatureDbPath)
	eventSigMap = utils.LoadEventSignatures(eventDbPath)
	addrABIMap = utils.LoadABIs(abiDbPath)
	log.Println("loaded function signature, event signature and ABI databases")

	semaphore := make(chan struct{}, maxConcurrentFiles)
	var wg sync.WaitGroup

	err := filepath.Walk(fromDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		semaphore <- struct{}{}
		toPath := filepath.Join(toDir, info.Name())
		go func(sp, writeTo string) {
			defer wg.Done()
			//`txHash` -> `Graph`
			txsGraph := cmap.New[types.Graph]()

			scanFile(sp, &txsGraph, maxConcurrentTx)

			if !dryRun {
				if err = utils.GraphToFile(&txsGraph, writeTo); err != nil {
					log.Fatal(err)
				}
			}

			<-semaphore
		}(path, toPath)

		return nil
	})
	wg.Wait()

	if err != nil {
		log.Printf("遍历目录时发生错误：%v\n", err)
	}

	log.Println("All works are done!")
	staticts.Statistic()
}
