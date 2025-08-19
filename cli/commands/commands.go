package commands

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"

	"github.com/steve/pman/cli/client"
	"github.com/steve/pman/cli/config"
	"github.com/steve/pman/cli/tree"
	"golang.org/x/term"
)

// expandCombinedFlags expands combined single-character flags like -rf into -r -f
// and reorders arguments to put flags before positional arguments
func expandCombinedFlags(args []string) []string {
	// First pass: expand combined flags and collect all elements
	var expanded []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && len(arg) > 2 && arg[1] != '-' {
			// This is a combined flag like -rf
			for i := 1; i < len(arg); i++ {
				expanded = append(expanded, "-"+string(arg[i]))
			}
		} else {
			expanded = append(expanded, arg)
		}
	}

	// Second pass: separate flags (with their values) from positional arguments
	var result []string
	var positional []string
	
	for i := 0; i < len(expanded); i++ {
		arg := expanded[i]
		if strings.HasPrefix(arg, "-") {
			// This is a flag
			result = append(result, arg)
			// Check if this flag expects a value
			if arg == "-g" || arg == "--group" || arg == "-s" || arg == "-u" || arg == "-p" || arg == "--expire" {
				// Get the next argument as the value if it exists and isn't a flag
				if i+1 < len(expanded) && !strings.HasPrefix(expanded[i+1], "-") {
					i++
					result = append(result, expanded[i])
				}
			}
		} else {
			// This is a positional argument
			positional = append(positional, arg)
		}
	}

	// Combine flags and positional arguments with flags first
	return append(result, positional...)
}

func ShowHelp() {
	fmt.Println("pman - Password Manager")
	fmt.Println("Usage: pman <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  login       Login to pman server")
	fmt.Println("  logout      Logout from pman server")
	fmt.Println("  setgroup    Set default group")
	fmt.Println("  add/put     Add password")
	fmt.Println("  get         Get password")
	fmt.Println("  ls/list     List passwords")
	fmt.Println("  edit        Edit password")
	fmt.Println("  rm/del      Delete password")
	fmt.Println("  info        Show password info")
	fmt.Println("  version     Show version")
	fmt.Println("  status      Show server status")
	fmt.Println("  passwd      Change password")
	fmt.Println("  whoami      Show current user, server and default group")
	fmt.Println("")
	fmt.Println("Admin commands:")
	fmt.Println("  useradd     Add user")
	fmt.Println("  userdel     Delete user")
	fmt.Println("  userupdate  Update user")
	fmt.Println("  userlist    List users")
	fmt.Println("  userdisable Disable user")
	fmt.Println("  userenable  Enable user")
}

func getAuthenticatedClient() (*client.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	if cfg.Server == "" || cfg.Token == "" {
		return nil, fmt.Errorf("not logged in. Please run 'pman login' first")
	}

	return client.NewClient(cfg.Server, cfg.Token), nil
}

func resolveGroup(groupFlag string) (string, error) {
	if groupFlag != "" {
		return groupFlag, nil
	}

	group := config.GetGroup()
	if group == "" {
		return "", fmt.Errorf("no group specified. Use -g flag, set PMAN_GROUP environment variable, or run 'pman setgroup'")
	}

	return group, nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(passwordBytes), nil
}

func readPasswordTwice(prompt string) (string, error) {
	password1, err := readPassword(prompt + ": ")
	if err != nil {
		return "", err
	}

	password2, err := readPassword("Confirm password: ")
	if err != nil {
		return "", err
	}

	if password1 != password2 {
		return "", fmt.Errorf("passwords do not match")
	}

	return password1, nil
}

func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	editors := []string{"nano", "vim", "vi", "code", "notepad"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return "vi"
}

func editWithEditor(content string) (string, error) {
	tmpfile, err := os.CreateTemp("", "pman-edit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		tmpfile.Close()
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	editor := getEditor()
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %v", err)
	}

	editedContent, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %v", err)
	}

	return strings.TrimSpace(string(editedContent)), nil
}

func confirmAction(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func Login(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("login", flag.ExitOnError)
	server := fs.String("s", "", "Server URL")
	email := fs.String("u", "", "Email address")
	password := fs.String("p", "", "Password")
	expireDays := fs.Int("expire", 0, "Token expiry in days")

	fs.Parse(args)

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	var serverURL, userEmail, userPassword string
	var expire int

	if *server != "" {
		serverURL = *server
	} else if cfg.Server != "" {
		serverURL = cfg.Server
	} else {
		fmt.Print("Server URL: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		serverURL = strings.TrimSpace(input)
	}

	if *email != "" {
		userEmail = *email
	} else if cfg.Email != "" {
		fmt.Printf("Email [%s]: ", cfg.Email)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			userEmail = cfg.Email
		} else {
			userEmail = input
		}
	} else {
		fmt.Print("Email: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		userEmail = strings.TrimSpace(input)
	}

	if *password != "" {
		userPassword = *password
	} else {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()
		userPassword = string(passwordBytes)
	}

	if *expireDays > 0 {
		expire = *expireDays
	}

	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "https://" + serverURL
	}

	c := client.NewClient(serverURL, "")
	loginResp, err := c.Login(userEmail, userPassword, expire)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}

	cfg.Server = serverURL
	cfg.Email = userEmail
	cfg.Token = loginResp.Token

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Login successful")
}

func Logout(args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Token == "" {
		fmt.Println("Not currently logged in")
		return
	}

	if err := cfg.ClearToken(); err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Logout successful")
}

func SetGroup(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman setgroup <group>\n")
		os.Exit(1)
	}

	groupName := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	cfg.DefaultGroup = groupName

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Default group set to: %s\n", groupName)
}

func Add(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("add", flag.ExitOnError)
	groupFlag := fs.String("g", "", "Group name")
	groupLongFlag := fs.String("group", "", "Group name")

	fs.Parse(args)
	remainingArgs := fs.Args()

	if len(remainingArgs) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: pman add <path> [password]\n")
		os.Exit(1)
	}

	path := remainingArgs[0]

	group := *groupFlag
	if group == "" {
		group = *groupLongFlag
	}

	resolvedGroup, err := resolveGroup(group)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var password string

	if len(remainingArgs) > 1 {
		password = remainingArgs[1]
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			passwordBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
			password = string(passwordBytes)
		} else {
			password, err = readPasswordTwice("Password")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
				os.Exit(1)
			}
		}
	}

	if err := client.CreatePassword(path, password, resolvedGroup); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating password: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Password added successfully: %s\n", path)
}

func Get(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("get", flag.ExitOnError)
	groupFlag := fs.String("g", "", "Group name")
	groupLongFlag := fs.String("group", "", "Group name")

	fs.Parse(args)
	remainingArgs := fs.Args()

	if len(remainingArgs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman get <path>\n")
		os.Exit(1)
	}

	path := remainingArgs[0]

	group := *groupFlag
	if group == "" {
		group = *groupLongFlag
	}

	resolvedGroup, err := resolveGroup(group)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	password, err := client.GetPassword(path, resolvedGroup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting password: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(password)
}

func List(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("list", flag.ExitOnError)
	groupFlag := fs.String("g", "", "Group name")
	groupLongFlag := fs.String("group", "", "Group name")

	fs.Parse(args)
	remainingArgs := fs.Args()

	var pathPrefix string
	if len(remainingArgs) > 0 {
		pathPrefix = remainingArgs[0]
	}

	group := *groupFlag
	if group == "" {
		group = *groupLongFlag
	}

	resolvedGroup, err := resolveGroup(group)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	paths, err := client.ListPasswords(resolvedGroup, pathPrefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing passwords: %v\n", err)
		os.Exit(1)
	}

	rootName := resolvedGroup
	if pathPrefix != "" {
		rootName = pathPrefix
	}

	tree.PrintTree(paths, rootName)
}

func Edit(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	groupFlag := fs.String("g", "", "Group name")
	groupLongFlag := fs.String("group", "", "Group name")

	fs.Parse(args)
	remainingArgs := fs.Args()

	if len(remainingArgs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman edit <path>\n")
		os.Exit(1)
	}

	path := remainingArgs[0]

	group := *groupFlag
	if group == "" {
		group = *groupLongFlag
	}

	resolvedGroup, err := resolveGroup(group)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var currentPassword string
	var isNewPassword bool

	currentPassword, err = client.GetPassword(path, resolvedGroup)
	if err != nil {
		if strings.Contains(err.Error(), "password not found") {
			isNewPassword = true
			currentPassword = ""
		} else {
			fmt.Fprintf(os.Stderr, "Error getting current password: %v\n", err)
			os.Exit(1)
		}
	}

	newPassword, err := editWithEditor(currentPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error editing password: %v\n", err)
		os.Exit(1)
	}

	if newPassword == "" {
		fmt.Fprintf(os.Stderr, "Error: Password cannot be empty\n")
		os.Exit(1)
	}

	if newPassword == currentPassword && !isNewPassword {
		fmt.Println("Password unchanged")
		return
	}

	if isNewPassword {
		if err := client.CreatePassword(path, newPassword, resolvedGroup); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating password: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Password created successfully: %s\n", path)
	} else {
		if err := client.UpdatePassword(path, newPassword, resolvedGroup); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating password: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Password updated successfully: %s\n", path)
	}
}

func Delete(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	groupFlag := fs.String("g", "", "Group name")
	groupLongFlag := fs.String("group", "", "Group name")
	forceFlag := fs.Bool("f", false, "Force deletion without confirmation")
	recursiveFlag := fs.Bool("r", false, "Recursive deletion")

	fs.Parse(args)
	remainingArgs := fs.Args()

	if len(remainingArgs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman rm/del/delete <path> [-f] [-r]\n")
		os.Exit(1)
	}

	path := remainingArgs[0]

	group := *groupFlag
	if group == "" {
		group = *groupLongFlag
	}

	resolvedGroup, err := resolveGroup(group)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Confirmation unless force flag is set
	if !*forceFlag {
		var message string
		if *recursiveFlag {
			message = fmt.Sprintf("Are you sure you want to recursively delete '%s' and all contents", path)
		} else {
			message = fmt.Sprintf("Are you sure you want to delete '%s'", path)
		}

		if !confirmAction(message) {
			fmt.Println("Deletion cancelled")
			return
		}
	}

	if *recursiveFlag {
		count, err := client.DeletePasswordRecursive(path, resolvedGroup)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting passwords: %v\n", err)
			os.Exit(1)
		}

		if count == 1 {
			fmt.Printf("Successfully deleted 1 password\n")
		} else {
			fmt.Printf("Successfully deleted %d passwords\n", count)
		}
	} else {
		if err := client.DeletePassword(path, resolvedGroup); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting password: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully deleted: %s\n", path)
	}
}

func Info(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("info", flag.ExitOnError)
	groupFlag := fs.String("g", "", "Group name")
	groupLongFlag := fs.String("group", "", "Group name")
	jsonFlag := fs.Bool("json", false, "Output in JSON format")

	fs.Parse(args)
	remainingArgs := fs.Args()

	if len(remainingArgs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman info <path> [--json]\n")
		os.Exit(1)
	}

	path := remainingArgs[0]

	group := *groupFlag
	if group == "" {
		group = *groupLongFlag
	}

	resolvedGroup, err := resolveGroup(group)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	passwordInfo, err := client.GetPasswordInfo(path, resolvedGroup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting password info: %v\n", err)
		os.Exit(1)
	}

	if *jsonFlag {
		jsonOutput, err := json.MarshalIndent(passwordInfo, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Printf("Path: %s\n", passwordInfo.Path)
		fmt.Printf("Created by: %s\n", passwordInfo.CreatedBy)
		fmt.Printf("Created at: %s\n", passwordInfo.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Last Updated by: %s\n", passwordInfo.UpdatedBy)
		fmt.Printf("Last Updated at: %s\n", passwordInfo.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
}

func Status(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("status", flag.ExitOnError)
	serverFlag := fs.String("s", "", "Server URL to check (defaults to configured server)")

	fs.Parse(args)

	var serverURL string

	if *serverFlag != "" {
		serverURL = *serverFlag
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			serverURL = "https://" + serverURL
		}
	} else {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if cfg.Server == "" {
			fmt.Fprintf(os.Stderr, "No server configured. Please login first or specify server with -s flag\n")
			os.Exit(1)
		}

		serverURL = cfg.Server
	}

	// Create client without token for health check
	client := client.NewClient(serverURL, "")

	if err := client.CheckHealth(); err != nil {
		fmt.Printf("Server not OK (%s)\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("Server OK")
}

func UserAdd(args []string) {
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: pman useradd <email> <role> <groups>\n")
		fmt.Fprintf(os.Stderr, "Example: pman useradd \"user@email.com\" \"admin\" \"team1:rw,team2:ro\"\n")
		os.Exit(1)
	}

	email := args[0]
	role := args[1]
	groups := args[2]

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	password, err := client.CreateUser(email, role, groups)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User created successfully: %s\n", email)
	fmt.Printf("Generated password: %s\n", password)
}

func UserDel(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("userdel", flag.ExitOnError)
	forceFlag := fs.Bool("f", false, "Force deletion without confirmation")

	fs.Parse(args)
	remainingArgs := fs.Args()

	if len(remainingArgs) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman userdel <email> [-f]\n")
		os.Exit(1)
	}

	email := remainingArgs[0]

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Confirmation unless force flag is set
	if !*forceFlag {
		message := fmt.Sprintf("Are you sure you want to delete user '%s'", email)
		if !confirmAction(message) {
			fmt.Println("User deletion cancelled")
			return
		}
	}

	if err := client.DeleteUser(email); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User deleted successfully: %s\n", email)
}

func UserUpdate(args []string) {
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: pman userupdate <email> <role> <groups>\n")
		fmt.Fprintf(os.Stderr, "Example: pman userupdate \"user@email.com\" \"admin\" \"team1:ro,team2:ro,team3:rw\"\n")
		os.Exit(1)
	}

	email := args[0]
	role := args[1]
	groups := args[2]

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := client.UpdateUser(email, role, groups); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User updated successfully: %s\n", email)
}

func UserList(args []string) {
	args = expandCombinedFlags(args)

	fs := flag.NewFlagSet("userlist", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output in JSON format")

	fs.Parse(args)

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	users, err := client.ListUsers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing users: %v\n", err)
		os.Exit(1)
	}

	if *jsonFlag {
		jsonOutput, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	} else {
		// Sort users by email (domain then user)
		sort.Slice(users, func(i, j int) bool {
			return users[i].Email < users[j].Email
		})

		fmt.Printf("%-30s %-10s %-15s %s\n", "EMAIL", "ROLE", "STATUS", "GROUPS")
		fmt.Printf("%-30s %-10s %-15s %s\n", strings.Repeat("-", 30), strings.Repeat("-", 10), strings.Repeat("-", 15), strings.Repeat("-", 20))

		for _, user := range users {
			status := "enabled"
			if !user.Enabled {
				status = "disabled"
			}
			fmt.Printf("%-30s %-10s %-15s %s\n", user.Email, user.Role, status, user.Groups)
		}
	}
}

func UserDisable(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman userdisable <email>\n")
		os.Exit(1)
	}

	email := args[0]

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := client.DisableUser(email); err != nil {
		fmt.Fprintf(os.Stderr, "Error disabling user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User disabled successfully: %s\n", email)
}

func UserEnable(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: pman userenable <email>\n")
		os.Exit(1)
	}

	email := args[0]

	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := client.EnableUser(email); err != nil {
		fmt.Fprintf(os.Stderr, "Error enabling user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User enabled successfully: %s\n", email)
}

func Whoami(args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "Not logged in. Please run 'pman login' first\n")
		os.Exit(1)
	}

	fmt.Printf("User: %s\n", cfg.Email)
	fmt.Printf("Server: %s\n", cfg.Server)

	// Show the effective default group
	effectiveGroup := config.GetGroup()
	if effectiveGroup == "" {
		fmt.Printf("Default group: (not set)\n")
	} else {
		fmt.Printf("Active group: %s\n", effectiveGroup)
	}
}

func Passwd(args []string) {
	client, err := getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		// User mode: change own password
		currentPassword, err := readPassword("Current password: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading current password: %v\n", err)
			os.Exit(1)
		}

		newPassword, err := readPasswordTwice("New password")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading new password: %v\n", err)
			os.Exit(1)
		}

		if err := client.ChangePassword(currentPassword, newPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error changing password: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Password changed successfully")
	} else if len(args) == 1 {
		// Admin mode: change another user's password
		email := args[0]

		newPassword, err := readPasswordTwice("New password")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading new password: %v\n", err)
			os.Exit(1)
		}

		if err := client.AdminChangePassword(email, newPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error changing password for %s: %v\n", email, err)
			os.Exit(1)
		}

		fmt.Printf("Password changed successfully for: %s\n", email)
	} else {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  pman passwd                    - Change your own password\n")
		fmt.Fprintf(os.Stderr, "  pman passwd <email>           - Change another user's password (admin only)\n")
		os.Exit(1)
	}
}
