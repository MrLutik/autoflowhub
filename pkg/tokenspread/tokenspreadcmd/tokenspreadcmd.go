package tokenspreadcmd

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
	use              = "tkspread"
	shortDescription = "command to spread tokens on generated accounts"
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
			countOfTxs, err := cmd.Flags().GetInt("txAmount")
			if err != nil {
				log.Fatalf("Error reading txs count flag: %v", err)
				return err
			}

			log.Println("Keys directory path: ", dirOfKeys)
			log.Println("Amount of txs to propogate: ", countOfTxs)
			log.Println("Starting to spread tokens:")

			client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return err
			}

			list, err := readKeys(dirOfKeys)
			if err != nil && list != nil {
				return err
			}
			if dirOfKeys == "" {
				log.Fatal("Please provide all required parameters: keys-dir. -h for help")
			}
			processTransactions(client, list, countOfTxs)
			return nil
		},
	}

	accgencmd.PersistentFlags().StringP("keys-dir", "d", "", "Keys directory (relative or absolute path)")
	accgencmd.PersistentFlags().IntP("txAmount", "t", 0, "how much transactions from generated users you want, if = 0 then default value = 7000000/4")

	return accgencmd
}
func processTransactions(client *client.Client, list []string, TxAmount int) {
	txamount := TxAmount

	disruptSum := txamount * 100
	if disruptSum == 0 {
		disruptSum = 7000000 / 4 * 100
	}

	if len(list) < 1 {
		panic("keys list empty")
	}
	disruptSum += len(list)
	var arr []*docker.User = make([]*docker.User, len(list))
	for i := range list {
		arr[i] = &docker.User{Key: list[i], Balance: 0}
	}
	// var arr []*docker.User = make([]*docker.User, 2500)
	// for i := 0; i < 2500; i++ {
	// 	arr[i] = &docker.User{Key: "pepelaugh", Balance: 0}
	// }
	waitGroup := &sync.WaitGroup{}
	fmt.Println(arr[0])
	docker.DisruptTokensBetweenAllAccounts(client, waitGroup, disruptSum, arr)
	waitGroup.Wait()
	for u := range arr {
		fmt.Println(arr[u])
	}
}

func readKeys(path string) ([]string, error) {
	reader := usecase.NewKeysReader(path)
	list, err := reader.GetAllAddresses()
	if err != nil {
		return nil, err
	}
	return list, nil
}
