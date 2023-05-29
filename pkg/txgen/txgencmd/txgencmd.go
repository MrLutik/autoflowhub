package txgencmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mrlutik/autoflowhub/internal/models"
	"github.com/mrlutik/autoflowhub/pkg/keygen/usecase"
	"github.com/spf13/cobra"
)

const (
	use              = "txgen"
	shortDescription = "Command to generate transaction from generated accounts"
	longDescription  = "Dummy field for some very useful longdescription"
)

func New() *cobra.Command {
	txgen := &cobra.Command{
		Use:   use,
		Short: shortDescription,
		Long:  longDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			dirOfKeys, err := cmd.Flags().GetString("keys-dir")
			if err != nil {
				log.Fatalf("Error reading keys-dir flag: %v", err)
				return err
			}
			log.Println("Keys directory path: ", dirOfKeys)
			sender, err := cmd.Flags().GetString("from")
			if err != nil {
				log.Fatalf("Error reading sender flag: %v", err)
				return err
			}
			recv, err := cmd.Flags().GetString("to")
			if err != nil {
				log.Fatalf("Error reading receiver address/es: %v", err)
			}
			sum, err := cmd.Flags().GetString("amount")
			if err != nil {
				log.Fatalf("Error reading amount to send: %v", err)
			}
			if dirOfKeys == "" || sender == "" || recv == "" || sum == "" {
				return errors.New("Flags are mandatory. can't be empty")
			}
			if ok, err := checkPath(recv); err == nil {
				if !ok {
					if err := processString(); err != nil {
						log.Fatalln("Something went wrong")
						return err
					}
					return nil
				}
				if err := processPath(recv, sender); err != nil {
					log.Fatalln("Something went wrong")
					return err
				}
				log.Println("Path processed")
				return nil
			}

			return nil
		},
	}
	txgen.PersistentFlags().StringP("keys-dir", "d", "", "Keys directory (relative or absolute path)")
	txgen.PersistentFlags().StringP("from", "f", "", "The address to send txs from")
	txgen.PersistentFlags().StringP("to", "t", "", "The address/addresses to send tx to or path to file")
	txgen.PersistentFlags().StringP("amount", "a", "", "Amount to send. Example: 100ukex")

	return txgen
}

func checkPath(path string) (bool, error) {
	if strings.HasPrefix(path, "~/") {
		log.Println("Tild expansion spoted")
		usr, err := user.Current()
		if err != nil {
			return false, err
		}
		log.Println("Correcting path")
		path = filepath.Join(usr.HomeDir, path[2:])
	}

	_, err := os.Stat(path)
	if err == nil {
		log.Printf("path %v exist. processing...", path)
		return true, nil
	}
	if os.IsNotExist(err) {
		log.Printf("path %v doesn't exist.", path)
		return false, nil
	}
	return false, err
}

func processPath(path, sender string) error {
	log.Println("Processing path...")
	reader := usecase.NewKeysReader(path)
	adrs, err := reader.GetAllAddresses()
	if err != nil {
		log.Fatalf("Failed to get addresses from path %v", path)
	}
	var msgs []models.Message
	for _, adr := range adrs {
		msg := models.Message{
			Type:        "/cosmos.bank.v1beta1.MsgSend",
			FromAddress: sender,
			ToAddress:   adr,
			Amount: []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			}{
				{
					Denom:  "ukex",
					Amount: "100",
				},
			},
		}
		msgs = append(msgs, msg)
	}

	tx := models.Tx{
		Body:     models.Body{TimeoutHeight: "0", Messages: msgs},
		AuthInfo: models.AuthInfo{Fee: models.Fee{GasLimit: "200000"}},
	}
	jsonData, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func processString() error {
	log.Println("Processing string")
	return nil
}
