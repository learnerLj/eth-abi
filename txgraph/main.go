package txgraph

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	abiDecoder "tx-analyze/abidec"
	"tx-analyze/staticts"
	"tx-analyze/types"
	"tx-analyze/utils"

	cmap "github.com/orcaman/concurrent-map/v2"
)

func MakeGraph(fromDir, toDir string, abidb *abiDecoder.ABIDB,
	maxConcurrentFiles, maxConcurrentTx int, dryRun bool) {

	log.SetFlags(log.Lshortfile)

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

			scanFile(sp, &txsGraph, maxConcurrentTx, abidb)

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
		log.Println(err)
	}

	log.Println("All works are done!")
	staticts.Statistic()
}
