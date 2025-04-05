package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mguptahub/nanoHttp/internal/server"
)

const (
	appName    = "nanoHttp"
	appVersion = "1.0.0"
)

func main() {
	// Create config directory if it doesn't exist
	configDir := filepath.Join(os.Getenv("HOME"), ".nanoHttp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize server manager
	manager := server.NewManager(configDir)

	// Parse command line arguments
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "add":
		handleAdd(manager)
	case "start":
		handleStart(manager)
	case "stop":
		handleStop(manager)
	case "delete":
		handleDelete(manager)
	case "list":
		handleList(manager)
	case "update":
		handleUpdate()
	case "version":
		fmt.Printf("%s version %s\n", appName, appVersion)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("Usage: %s <command> [options]\n\n", appName)
	fmt.Println("Commands:")
	fmt.Println("  add     Add a new server instance")
	fmt.Println("  start   Start a server instance")
	fmt.Println("  stop    Stop a server instance")
	fmt.Println("  delete  Delete a server instance")
	fmt.Println("  list    List all server instances")
	fmt.Println("  update  Check for and install updates")
	fmt.Println("  version Show version information")
	fmt.Println("\nUse --help with any command for detailed usage information")
}

func handleAdd(manager *server.Manager) {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addCmd.Usage = func() {
		fmt.Println("Usage: nanoHttp add [options]")
		fmt.Println("\nOptions:")
		fmt.Printf("  -d | -allow-dir-listing     Allow directory listing\n")
		fmt.Printf("  -n | -name                  Instance name (required)\n")
		fmt.Printf("  -p | -port                  Port number (default 8080)\n")
		fmt.Printf("  -w | -web-folder            Web root folder (required)\n")
	}

	var (
		name            string
		port            int
		webFolder       string
		allowDirListing bool
	)

	// Define flags with aliases
	addCmd.StringVar(&name, "name", "", "")
	addCmd.StringVar(&name, "n", "", "")
	addCmd.IntVar(&port, "port", 8080, "")
	addCmd.IntVar(&port, "p", 8080, "")
	addCmd.StringVar(&webFolder, "web-folder", "", "")
	addCmd.StringVar(&webFolder, "w", "", "")
	addCmd.BoolVar(&allowDirListing, "allow-dir-listing", false, "")
	addCmd.BoolVar(&allowDirListing, "d", false, "")

	// Handle --help explicitly
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		addCmd.Usage()
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		addCmd.Usage()
		os.Exit(1)
	}

	addCmd.Parse(os.Args[2:])

	if name == "" || webFolder == "" {
		fmt.Println("Error: name and web-folder are required")
		addCmd.Usage()
		os.Exit(1)
	}

	instance := &server.Instance{
		Name:            name,
		Port:            port,
		WebFolder:       webFolder,
		AllowDirListing: allowDirListing,
	}

	if err := manager.AddInstance(instance); err != nil {
		fmt.Printf("Error adding instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Instance '%s' added successfully\n", name)
}

func handleStart(manager *server.Manager) {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startCmd.Usage = func() {
		fmt.Println("Usage: nanoHttp start <instance-name> [options]")
		fmt.Println("\nOptions:")
		fmt.Printf("  -f | -foreground            Run server in foreground (current process)\n")
	}

	var foreground bool
	startCmd.BoolVar(&foreground, "foreground", false, "")
	startCmd.BoolVar(&foreground, "f", false, "")

	if len(os.Args) < 3 {
		fmt.Println("Error: instance name is required")
		startCmd.Usage()
		os.Exit(1)
	}

	// Check if --help is requested
	if os.Args[2] == "--help" || os.Args[2] == "-h" {
		startCmd.Usage()
		os.Exit(0)
	}

	name := os.Args[2]
	if err := startCmd.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if foreground {
		runServerInForeground(manager, name)
	} else {
		if err := manager.StartInstance(name); err != nil {
			fmt.Printf("Error starting instance: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Instance '%s' started successfully\n", name)
	}
}

func runServerInForeground(manager *server.Manager, name string) {
	server, err := manager.GetServer(name)
	if err != nil {
		fmt.Printf("Error getting server instance: %v\n", err)
		os.Exit(1)
	}

	// Create a new context for the server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create the HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", server.GetConfig().Port),
		Handler: server.CreateHandler(),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	// Start the server
	go func() {
		fmt.Printf("Starting server '%s' in foreground mode on port %d\n", name, server.GetConfig().Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server %s: %v\n", name, err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Printf("\nShutting down server '%s'...\n", name)

	// Shutdown the server
	if err := httpServer.Shutdown(context.Background()); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server stopped successfully")
}

func handleStop(manager *server.Manager) {
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	stopCmd.Usage = func() {
		fmt.Println("Usage: nanoHttp stop <instance-name>")
		fmt.Println("\nDescription:")
		fmt.Println("  Stop a running server instance. If the instance is not running,")
		fmt.Println("  an error will be returned. This command will gracefully shutdown")
		fmt.Println("  the server, allowing in-flight requests to complete.")
	}

	// Handle --help explicitly
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		stopCmd.Usage()
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		fmt.Println("Error: instance name is required")
		stopCmd.Usage()
		os.Exit(1)
	}

	name := os.Args[2]
	if err := manager.StopInstance(name); err != nil {
		fmt.Printf("Error stopping instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Instance '%s' stopped successfully\n", name)
}

func handleDelete(manager *server.Manager) {
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteCmd.Usage = func() {
		fmt.Println("Usage: nanoHttp delete <instance-name>")
		fmt.Println("\nDescription:")
		fmt.Println("  Delete a server instance configuration. If the instance is running,")
		fmt.Println("  it will be stopped first. This command will remove all configuration")
		fmt.Println("  data for the specified instance. This action cannot be undone.")
	}

	// Handle --help explicitly
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		deleteCmd.Usage()
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		fmt.Println("Error: instance name is required")
		deleteCmd.Usage()
		os.Exit(1)
	}

	name := os.Args[2]
	if err := manager.DeleteInstance(name); err != nil {
		fmt.Printf("Error deleting instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Instance '%s' deleted successfully\n", name)
}

func handleList(manager *server.Manager) {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.Usage = func() {
		fmt.Println("Usage: nanoHttp list [options]")
		fmt.Println("\nOptions:")
		fmt.Printf("  -s | -simple                Use simple list format instead of table\n")
	}

	var simpleFormat bool
	listCmd.BoolVar(&simpleFormat, "simple", false, "")
	listCmd.BoolVar(&simpleFormat, "s", false, "")

	// Handle --help explicitly
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		listCmd.Usage()
		os.Exit(0)
	}

	listCmd.Parse(os.Args[2:])

	instances := manager.ListInstances()
	if len(instances) == 0 {
		fmt.Println("No instances found")
		return
	}

	if simpleFormat {
		// Simple format display
		fmt.Println("Server Instances:")
		for _, instance := range instances {
			status := "stopped"
			statusColor := "\033[31m" // Red
			if instance.IsRunning {
				status = "running"
				statusColor = "\033[32m" // Green
			}

			dirListing := "no"
			if instance.AllowDirListing {
				dirListing = "yes"
			}

			pid := "-"
			if instance.IsRunning && instance.PID > 0 {
				pid = fmt.Sprintf("%d", instance.PID)
			}

			fmt.Printf("Name: %s\n", instance.Name)
			fmt.Printf("  Port: %d\n", instance.Port)
			fmt.Printf("  Web Folder: %s\n", instance.WebFolder)
			fmt.Printf("  Dir Listing: %s\n", dirListing)
			fmt.Printf("  Status: %s%s\033[0m\n", statusColor, status)
			fmt.Printf("  PID: %s\n\n", pid)
		}
		return
	}

	const (
		nameWidth   = 12
		portWidth   = 6
		folderWidth = 12
		dirWidth    = 12
		statusWidth = 8
		pidWidth    = 8
	)

	// Print table header
	fmt.Println("Server Instances:")
	fmt.Printf("┌%s┬%s┬%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", portWidth+2),
		strings.Repeat("─", folderWidth+2),
		strings.Repeat("─", dirWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", pidWidth+2))

	fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		nameWidth, "Name",
		portWidth, "Port",
		folderWidth, "Web Folder",
		dirWidth, "Dir Listing",
		statusWidth, "Status",
		pidWidth, "PID")

	fmt.Printf("├%s┼%s┼%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", portWidth+2),
		strings.Repeat("─", folderWidth+2),
		strings.Repeat("─", dirWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", pidWidth+2))

	// Print each instance as a table row
	for _, instance := range instances {
		status := "stopped"
		statusColor := "\033[31m" // Red
		if instance.IsRunning {
			status = "running"
			statusColor = "\033[32m" // Green
		}

		dirListing := "no"
		if instance.AllowDirListing {
			dirListing = "yes"
		}

		webFolder := truncateString(instance.WebFolder, folderWidth)
		name := truncateString(instance.Name, nameWidth)

		pid := "-"
		if instance.IsRunning && instance.PID > 0 {
			pid = fmt.Sprintf("%d", instance.PID)
		}

		fmt.Printf("│ %-*s │ %*d │ %-*s │ %-*s │ %s%-*s\033[0m │ %-*s │\n",
			nameWidth, name,
			portWidth, instance.Port,
			folderWidth, webFolder,
			dirWidth, dirListing,
			statusColor, statusWidth, status,
			pidWidth, pid)
	}

	fmt.Printf("└%s┴%s┴%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", portWidth+2),
		strings.Repeat("─", folderWidth+2),
		strings.Repeat("─", dirWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", pidWidth+2))
}

// truncateString truncates a string if it's longer than maxLen and adds "..."
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func handleUpdate() {
	// TODO: Implement update functionality
	fmt.Println("Update functionality not implemented yet")
}
