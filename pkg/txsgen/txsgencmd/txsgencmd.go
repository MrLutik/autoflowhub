package txsgencmd

import (
	"fmt"
	"log"
	"sync"

	"github.com/docker/docker/client"
	"github.com/mrlutik/autoflowhub/pkg/keygen/usecase"
	"github.com/mrlutik/autoflowhub/pkg/txsgen/docker"
	"github.com/spf13/cobra"
)

const (
	use              = "accgen"
	shortDescription = "command to generate accounts for Kira Network testnet"
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
			countOfTxs, err := cmd.Flags().GetInt("count")
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

			client, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				return err
			}

			list, err := readKeys(dirOfKeys)
			if err != nil && list != nil {
				return err
			}
			processTransactions(client, list, blockToListen, countOfTxs)
			return nil
		},
	}

	accgencmd.PersistentFlags().StringP("keys-dir", "d", "", "Keys directory (relative or absolute path)")
	accgencmd.PersistentFlags().IntP("count", "c", 0, "Count of keys which will be added")
	accgencmd.PersistentFlags().IntP("block", "b", 0, "which block to listen")
	accgencmd.PersistentFlags().IntP("txAmount", "t", 0, "how much transactions from generated users you want")

	return accgencmd
}

func processTransactions(client *client.Client, list []string, BlockToListen, TxAmount int) {
	waitGroup := &sync.WaitGroup{}
	c := make(chan int)

	go docker.BlockListener(client, "validator", BlockToListen, waitGroup, c)

	arr := make([]*docker.User, len(list))
	for i := range list {
		arr[i] = &docker.User{Key: list[i], Balance: 0}
	}

	disruptSum := TxAmount * 100
	docker.DisruptTokensBetweenAllAccounts(client, waitGroup, disruptSum, arr[:])

	<-c
	waitGroup.Wait()

	for _, u := range arr {
		fmt.Println(u)
	}

	txcount := docker.TransactionSpam(client, waitGroup, TxAmount, arr)
	waitGroup.Wait()
	fmt.Println(txcount)
}

func readKeys(path string) ([]string, error) {
	reader := usecase.NewKeysReader(path)
	list, err := reader.GetAllAddresses()
	if err != nil {
		return nil, err
	}
	return list, nil
}