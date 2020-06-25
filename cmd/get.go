package cmd

import (
	"bufio"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-mbox"
	"github.com/emersion/go-message/mail"
	"github.com/mcnijman/go-emailaddress"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().StringVarP(&email, "email", "e", "mail@example.com", "email address")
	getCmd.PersistentFlags().BoolVarP(&all, "all", "a", false, "get all mailboxes")
	getCmd.PersistentFlags().StringVarP(&mailbox, "mailbox", "m", "inbox", "mailbox to capture")
	getCmd.PersistentFlags().StringVarP(&location, "location", "l", "/tmp", "location to write mbox file")
}

var emailAddress *emailaddress.EmailAddress
var err error

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get",
	Run: func(cmd *cobra.Command, args []string) {

		emailAddress, err = emailaddress.Parse(email)
		if err != nil {
			fmt.Println(err)
			fmt.Println("exiting")
			os.Exit(1)
		}

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
			mboxName := fmt.Sprintf("%s_AT_%s_%s.mbox", strings.ReplaceAll(emailAddress.LocalPart, ".", "_"), strings.ReplaceAll(emailAddress.Domain, ".", "_"), mailbox)
			f, err := os.Create(filepath.Join(location, mboxName))
			if err != nil {
				fmt.Println(err)
				fmt.Println("exiting")
				os.Exit(1)
			}

			fileWriter := bufio.NewWriter(f)
			mboxWriter := mbox.NewWriter(fileWriter)
			defer mboxWriter.Close()

			backupMailbox(imapClient, mboxWriter)

		} else {
			fmt.Printf("  ! account does not contain mailbox %s\n", mailbox)
			fmt.Println("exiting")
			os.Exit(0)
		}

	},
}

func getDomain() (string, error) {

	if domain, ok := domains[emailAddress.Domain]; ok {
		return domain, nil
	} else {
		return domain, fmt.Errorf("%s is not a supported domain", emailAddress.Domain)
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
	fmt.Print("  * Enter your password: ")
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

func backupMailbox(imapClient *client.Client, writer *mbox.Writer) {
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

	count := 1
	for msg := range messages {
		count = count + 1
		fmt.Println("  * Writing email ", count, " of ", mbox.Messages)
		msgBody := msg.GetBody(&section)
		mr, err := mail.CreateReader(msgBody)
		if err != nil {
			fmt.Println(err)
			fmt.Println("exiting")
			os.Exit(1)
		}
		header := mr.Header
		fields := header.Fields()
		date, _ := header.Date()
		mw, _ := writer.CreateMessage(header.Get("From"), date)

		for fields.Next() {
			headerLine := strings.NewReader(fmt.Sprintf("%v: %v\n", fields.Key(), fields.Value()))
			io.Copy(mw, headerLine)
		}

		io.Copy(mw, strings.NewReader("\n"))

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			io.Copy(mw, p.Body)

		}

	}
}
