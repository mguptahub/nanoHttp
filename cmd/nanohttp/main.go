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
	case "serve":
		handleServe(manager)
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
	fmt.Println("  serve   Serve a server instance")
	fmt.Println("  version Show version information")
	fmt.Println("\nUse --help with any command for detailed usage information")
}

func handleAdd(manager *server.Manager) {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	name := addCmd.String("name", "", "Instance name (required)")
	port := addCmd.Int("port", 8080, "Port number")
	webFolder := addCmd.String("web-folder", "", "Web root folder (required)")
	allowDirListing := addCmd.Bool("allow-dir-listing", false, "Allow directory listing")
	sslCertFolder := addCmd.String("ssl-cert-folder", "", "SSL certificates folder")

	addCmd.Parse(os.Args[2:])

	if *name == "" || *webFolder == "" {
		fmt.Println("Error: name and web-folder are required")
		addCmd.PrintDefaults()
		os.Exit(1)
	}

	instance := &server.Instance{
		Name:            *name,
		Port:            *port,
		WebFolder:       *webFolder,
		AllowDirListing: *allowDirListing,
		SSLCertFolder:   *sslCertFolder,
	}

	if err := manager.AddInstance(instance); err != nil {
		fmt.Printf("Error adding instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Instance '%s' added successfully\n", *name)
}

func handleStart(manager *server.Manager) {
	if len(os.Args) < 3 {
		fmt.Println("Error: instance name is required")
		fmt.Println("Usage: nanoHttp start <instance-name>")
		os.Exit(1)
	}

	name := os.Args[2]
	if err := manager.StartInstance(name); err != nil {
		fmt.Printf("Error starting instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Instance '%s' started successfully\n", name)
}

func handleStop(manager *server.Manager) {
	if len(os.Args) < 3 {
		fmt.Println("Error: instance name is required")
		fmt.Println("Usage: nanoHttp stop <instance-name>")
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
	if len(os.Args) < 3 {
		fmt.Println("Error: instance name is required")
		fmt.Println("Usage: nanoHttp delete <instance-name>")
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
	// Add flag parsing for list command
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	simpleFormat := listCmd.Bool("simple", false, "Use simple list format instead of table")
	listCmd.Parse(os.Args[2:])

	instances := manager.ListInstances()
	if len(instances) == 0 {
		fmt.Println("No instances found")
		return
	}

	if *simpleFormat {
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

			sslCert := "none"
			if instance.SSLCertFolder != "" {
				sslCert = instance.SSLCertFolder
			}

			pid := "-"
			if instance.IsRunning && instance.PID > 0 {
				pid = fmt.Sprintf("%d", instance.PID)
			}

			fmt.Printf("Name: %s\n", instance.Name)
			fmt.Printf("  Port: %d\n", instance.Port)
			fmt.Printf("  Web Folder: %s\n", instance.WebFolder)
			fmt.Printf("  Dir Listing: %s\n", dirListing)
			fmt.Printf("  SSL Cert: %s\n", sslCert)
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
		sslWidth    = 12
		statusWidth = 8
		pidWidth    = 8
	)

	// Print table header
	fmt.Println("Server Instances:")
	fmt.Printf("┌%s┬%s┬%s┬%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", portWidth+2),
		strings.Repeat("─", folderWidth+2),
		strings.Repeat("─", dirWidth+2),
		strings.Repeat("─", sslWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", pidWidth+2))

	fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		nameWidth, "Name",
		portWidth, "Port",
		folderWidth, "Web Folder",
		dirWidth, "Dir Listing",
		sslWidth, "SSL Cert",
		statusWidth, "Status",
		pidWidth, "PID")

	fmt.Printf("├%s┼%s┼%s┼%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", portWidth+2),
		strings.Repeat("─", folderWidth+2),
		strings.Repeat("─", dirWidth+2),
		strings.Repeat("─", sslWidth+2),
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

		sslCert := "none"
		if instance.SSLCertFolder != "" {
			sslCert = truncateString(instance.SSLCertFolder, sslWidth)
		}

		webFolder := truncateString(instance.WebFolder, folderWidth)
		name := truncateString(instance.Name, nameWidth)

		pid := "-"
		if instance.IsRunning && instance.PID > 0 {
			pid = fmt.Sprintf("%d", instance.PID)
		}

		fmt.Printf("│ %-*s │ %*d │ %-*s │ %-*s │ %-*s │ %s%-*s\033[0m │ %-*s │\n",
			nameWidth, name,
			portWidth, instance.Port,
			folderWidth, webFolder,
			dirWidth, dirListing,
			sslWidth, sslCert,
			statusColor, statusWidth, status,
			pidWidth, pid)
	}

	fmt.Printf("└%s┴%s┴%s┴%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", portWidth+2),
		strings.Repeat("─", folderWidth+2),
		strings.Repeat("─", dirWidth+2),
		strings.Repeat("─", sslWidth+2),
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

func handleServe(manager *server.Manager) {
	if len(os.Args) < 3 {
		fmt.Println("Error: instance name is required")
		fmt.Println("Usage: nanoHttp serve <instance-name>")
		os.Exit(1)
	}

	name := os.Args[2]
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
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server %s: %v\n", name, err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-sigChan

	// Shutdown the server
	if err := httpServer.Shutdown(context.Background()); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
		os.Exit(1)
	}
}
