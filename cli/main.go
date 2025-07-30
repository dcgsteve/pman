package main

import (
	"fmt"
	"os"

	"github.com/steve/pman/cli/commands"
)

var (
	Version = "1.0.3"
)

func main() {
	if len(os.Args) < 2 {
		commands.ShowHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "login":
		commands.Login(args)
	case "logout":
		commands.Logout(args)
	case "setgroup":
		commands.SetGroup(args)
	case "add", "put":
		commands.Add(args)
	case "get":
		commands.Get(args)
	case "ls", "list":
		commands.List(args)
	case "edit":
		commands.Edit(args)
	case "rm", "del", "delete":
		commands.Delete(args)
	case "info":
		commands.Info(args)
	case "version":
		fmt.Printf("pman: v%s\n", Version)
	case "status":
		commands.Status(args)
	case "useradd":
		commands.UserAdd(args)
	case "userdel":
		commands.UserDel(args)
	case "userupdate":
		commands.UserUpdate(args)
	case "userlist":
		commands.UserList(args)
	case "userdisable":
		commands.UserDisable(args)
	case "userenable":
		commands.UserEnable(args)
	case "passwd":
		commands.Passwd(args)
	case "whoami":
		commands.Whoami(args)
	case "help", "--help", "-h":
		commands.ShowHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		commands.ShowHelp()
		os.Exit(1)
	}
}
