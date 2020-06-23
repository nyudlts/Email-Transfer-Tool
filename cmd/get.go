package cmd

import (
	"bufio"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/mcnijman/go-emailaddress"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().StringVarP(&email, "email", "e", "mail@example.com", "email address")
	getCmd.PersistentFlags().BoolVarP(&all, "all", "a", false, "get all mailboxes")
	getCmd.PersistentFlags().StringVarP(&mailbox, "mailbox", "m", "inbox", "mailbox to capture")
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
		imapClient, err := getClient(domain)
		if err != nil {
			fmt.Println(err)
			fmt.Println("exiting")
			os.Exit(1)
		}
		fmt.Printf("  * Client connected to %s\n", domain)
		defer imapClient.Close()

		if err := imapClient.Login(email, getPassword()); err != nil {
			fmt.Println(err)
			fmt.Println("exiting")
			os.Exit(1)
		}

		mailboxes := getMailboxes(imapClient)
		if mailboxContains(mailboxes, mailbox) {
			backupMailbox(imapClient)
		} else {
			fmt.Printf("  ! account does not contain mailbox %s\n", mailbox)
			fmt.Println("exiting")
			os.Exit(0)
		}

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

func getClient(domain string) (*client.Client, error) {
	// Connect to server
	c, err := client.DialTLS(domain, nil)
	if err != nil {
		return c, err
	}
	return c, nil
}

func getPassword() string {
	fmt.Print("Enter your password: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func getMailboxes(imapClient *client.Client) []string {
	mailboxes := []string{}
	mailboxChannel := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- imapClient.List("", "*", mailboxChannel)
	}()
	for m := range mailboxChannel {
		mailboxes = append(mailboxes, m.Name)
	}
	return mailboxes
}

func mailboxContains(mbs []string, mb string) bool {
	for _, a := range mbs {
		if a == mb {
			return true
		}
	}
	return false
}

func backupMailbox(imapClient *client.Client) {
	mbox, err := imapClient.Select(mailbox, false)
	if err != nil {
		fmt.Println(err)
		fmt.Println("exiting")
		os.Exit(1)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(uint32(1), mbox.Messages)

	messages := make(chan *imap.Message, mbox.Messages)
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	go func() {
		if err := imapClient.Fetch(seqSet, items, messages); err != nil {
			panic(err)
		}
	}()

	for msg := range messages {
		msgBody := msg.GetBody(&section)
		reader, err := mail.CreateReader(msgBody)
		if err != nil {
			panic(err)
		}
		header := reader.Header
		fields := header.Fields()
		fmt.Printf("\nFrom %v\n", header.Get("From"))
		for fields.Next() {
			fmt.Printf("%v: %v\n", fields.Key(), fields.Value())
		}
	}

}
