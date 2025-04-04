package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mguptahub/justHttp/internal/server"
)

const (
	appName    = "justHttp"
	appVersion = "1.0.0"
)

func main() {
	// Create config directory if it doesn't exist
	configDir := filepath.Join(os.Getenv("HOME"), ".justHttp")
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
		fmt.Println("Usage: justHttp start <instance-name>")
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
		fmt.Println("Usage: justHttp stop <instance-name>")
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
		fmt.Println("Usage: justHttp delete <instance-name>")
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
	instances := manager.ListInstances()
	if len(instances) == 0 {
		fmt.Println("No instances found")
		return
	}

	fmt.Println("Server instances:")
	fmt.Println("----------------")
	for _, instance := range instances {
		status := "stopped"
		if instance.IsRunning {
			status = "running"
		}
		fmt.Printf("Name: %s\n", instance.Name)
		fmt.Printf("Port: %d\n", instance.Port)
		fmt.Printf("Web Folder: %s\n", instance.WebFolder)
		fmt.Printf("Directory Listing: %v\n", instance.AllowDirListing)
		if instance.SSLCertFolder != "" {
			fmt.Printf("SSL Certificates: %s\n", instance.SSLCertFolder)
		}
		fmt.Printf("Status: %s\n", status)
		fmt.Println("----------------")
	}
}

func handleUpdate() {
	// TODO: Implement update functionality
	fmt.Println("Update functionality not implemented yet")
}
