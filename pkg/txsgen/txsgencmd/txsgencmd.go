package txsgencmd

import (
	"fmt"
	"log"
	"sync"

	"github.com/docker/docker/client"
	"github.com/mrlutik/autoflowhub/internal/docker"
	"github.com/mrlutik/autoflowhub/pkg/keygen/usecase"
	"github.com/spf13/cobra"
)

const (
	use              = "txgen"
	shortDescription = "command to generate transactions from generated accounts"
	longDescription  = "dummy field for some very useful longdescription"
)

func New() *cobra.Command {
	accgencmd := &cobra.Command{
		Use:   use,
		Short: shortDescription,
		Long:  longDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			dirOfKeys, err := cmd.Flags().GetString("keys-dir")
			if err != nil {
				log.Fatalf("Error reading keys-dir flag: %v", err)
				return err
			}
			countOfTxs, err := cmd.Flags().GetInt("countOfTxs")
			if err != nil {
				log.Fatalf("Error reading txs count flag: %v", err)
				return err
			}
			blockToListen, err := cmd.Flags().GetInt("blockToListen")
			if err != nil {
				log.Fatalf("Error reading block flag: %v", err)
			}

			log.Println("Keys directory path: ", dirOfKeys)
			log.Println("Amount of txs to propogate: ", countOfTxs)
			log.Println("Waiting for block: ", blockToListen)
			log.Println("Starting to propogate txs:")

			client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return err
			}

			list, err := readKeys(dirOfKeys)
			if err != nil && list != nil {
				return err
			}
			if dirOfKeys == "" || blockToListen == 0 {
				log.Fatal("Please provide all required parameters: blockToListen, keys-dir. -h for help")
			}
			processTransactions(client, list, blockToListen, countOfTxs)
			return nil
		},
	}

	accgencmd.PersistentFlags().StringP("keys-dir", "d", "", "Keys directory (relative or absolute path)")
	accgencmd.PersistentFlags().IntP("blockToListen", "b", 0, "which block to listen")
	accgencmd.PersistentFlags().IntP("countOfTxs", "c", 0, "how much transactions from generated users you want, if = 0 then default value = 7000000/4")

	return accgencmd
}

func processTransactions(client *client.Client, list []string, BlockToListen, TxAmount int) {
	if TxAmount == 0 {
		TxAmount = 7000000 / 4
	}
	blockTolisten := BlockToListen
	if len(list) < 1 {
		panic("keys list empty")
	}
	var arr []*docker.User = make([]*docker.User, len(list))
	for i := range list {
		arr[i] = &docker.User{Key: list[i], Balance: 0}
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	c := make(chan int)
	go docker.BlockListener(client, "validator", blockTolisten, waitGroup, c)
	<-c
	txcount := docker.TransactionSpam(client, waitGroup, TxAmount, arr)
	waitGroup.Wait()
	fmt.Println(*txcount, "txs was completed")
}

func readKeys(path string) ([]string, error) {
	reader := usecase.NewKeysReader(path)
	list, err := reader.GetAllAddresses()
	if err != nil {
		return nil, err
	}
	return list, nil
}
