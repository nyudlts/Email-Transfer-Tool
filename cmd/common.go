package cmd

var (
	email   string
	domains = make(map[string]string)
)

func init() {
	domains["gmail.com"] = "imap.gmail.com:993"
}
