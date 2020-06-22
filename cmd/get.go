package cmd

import (
	"fmt"
	"github.com/mcnijman/go-emailaddress"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().StringVarP(&email, "email", "e", "mail@example.com", "email address")
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get",
	Run: func(cmd *cobra.Command, args []string) {
		domain, err := getDomain()
		if err != nil {
			fmt.Println(err)
			fmt.Println("exiting")
			os.Exit(1)
		}
		fmt.Println(domain)
	},
}

func getDomain() (string, error) {
	domain := ""
	emailAddreess, err := emailaddress.Parse(email)
	if err != nil {
		return domain, err
	}

	if domain, ok := domains[emailAddreess.Domain]; ok {
		return domain, nil
	} else {
		return domain, fmt.Errorf("%s is not a supported domain", emailAddreess.Domain)
	}
}
