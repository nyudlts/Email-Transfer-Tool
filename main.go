package main

import (
	"fmt"
	"github.com/nyudlts/go-mail/cmd"
)

var vers = "0.0.0"

func main() {
	fmt.Println("** email-transfer-tool v.", vers, " **")
	cmd.Execute()
}
