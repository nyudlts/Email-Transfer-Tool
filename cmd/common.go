package cmd

var (
	all     bool
	email   string
	domains = make(map[string]string)
	mailbox string
)

func init() {
	domains["gmail.com"] = "imap.gmail.com:993"
	domains["nyu.edu"] = "imap.gmail.com:993"
}
