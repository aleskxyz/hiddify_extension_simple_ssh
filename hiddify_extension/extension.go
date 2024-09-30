package hiddify_extension

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	ex "github.com/hiddify/hiddify-core/extension"
	ui "github.com/hiddify/hiddify-core/extension/ui"
	"golang.org/x/crypto/ssh"
)

// Console color settings
var (
	red    = color.New(color.FgRed).Add(color.Bold)
	green  = color.New(color.FgGreen).Add(color.Underline)
	yellow = color.New(color.FgYellow)
)

// Extension-specific data struct
type HiddifyExtensionSimpleSshData struct {
	IP       string `json:"ip"`       // SSH server IP
	Port     string `json:"port"`     // SSH port
	Username string `json:"username"` // SSH username
	Password string `json:"password"` // SSH password
	Command  string `json:"command"`  // Command to execute on SSH server
}

// Form field keys
const (
	IPKey       = "ip"
	PortKey     = "port"
	UsernameKey = "username"
	PasswordKey = "password"
	CommandKey  = "command"
)

// HiddifyExtensionSimpleSsh represents the extension's core functionality
type HiddifyExtensionSimpleSsh struct {
	ex.Base[HiddifyExtensionSimpleSshData]
	console string             // Stores console output
	cancel  context.CancelFunc // Function to cancel background tasks
}

// GetUI provides the form for user input
func (e *HiddifyExtensionSimpleSsh) GetUI() ui.Form {
	// UI form creation
	return ui.Form{
		Title:       "Simple SSH Command Executor",
		Description: "Execute a command on a remote SSH server",
		Buttons:     []string{ui.Button_Cancel, ui.Button_Submit},
		Fields: []ui.FormField{
			{
				Type:        ui.FieldInput,
				Key:         IPKey,
				Label:       "IP Address",
				Placeholder: "Enter the SSH server IP address",
				Required:    true,
				Value:       e.Base.Data.IP,
			},
			{
				Type:        ui.FieldInput,
				Key:         PortKey,
				Label:       "Port",
				Placeholder: "Enter the SSH server port",
				Required:    true,
				Value:       e.Base.Data.Port,
				Validator:   ui.ValidatorDigitsOnly, // Only allow digits
			},
			{
				Type:        ui.FieldInput,
				Key:         UsernameKey,
				Label:       "Username",
				Placeholder: "Enter SSH username",
				Required:    true,
				Value:       e.Base.Data.Username,
			},
			{
				Type:        ui.FieldPassword, // Hide password input
				Key:         PasswordKey,
				Label:       "Password",
				Placeholder: "Enter SSH password",
				Required:    true,
				Value:       e.Base.Data.Password,
			},
			{
				Type:        ui.FieldInput,
				Key:         CommandKey,
				Label:       "Command",
				Placeholder: "Enter command to execute",
				Required:    true,
				Value:       e.Base.Data.Command,
			},
			{
				Type:  ui.FieldConsole,
				Key:   "console",
				Label: "Console Output",
				Value: e.console, // Display console output
				Lines: 20,
			},
		},
	}
}

// setFormData validates and sets form data
func (e *HiddifyExtensionSimpleSsh) setFormData(data map[string]string) error {
	// Validate and store form inputs
	if val, ok := data[IPKey]; ok {
		e.Base.Data.IP = val
	}
	if val, ok := data[PortKey]; ok {
		e.Base.Data.Port = val
	}
	if val, ok := data[UsernameKey]; ok {
		e.Base.Data.Username = val
	}
	if val, ok := data[PasswordKey]; ok {
		e.Base.Data.Password = val
	}
	if val, ok := data[CommandKey]; ok {
		e.Base.Data.Command = val
	}
	return nil
}

// backgroundTask connects to the SSH server and executes the command
func (e *HiddifyExtensionSimpleSsh) backgroundTask(ctx context.Context) {
	// Prepare SSH connection configuration
	config := &ssh.ClientConfig{
		User: e.Base.Data.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(e.Base.Data.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Skip host key verification (for simplicity)
		Timeout:         5 * time.Second,
	}

	// Connect to the SSH server
	address := fmt.Sprintf("%s:%s", e.Base.Data.IP, e.Base.Data.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		e.addAndUpdateConsole(red.Sprint("Failed to connect: "), err.Error())
		return
	}
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		e.addAndUpdateConsole(red.Sprint("Failed to create SSH session: "), err.Error())
		return
	}
	defer session.Close()

	// Execute the command and get output
	output, err := session.CombinedOutput(e.Base.Data.Command)
	if err != nil {
		e.addAndUpdateConsole(red.Sprint("Command execution failed: "), err.Error())
		return
	}

	// Print the output
	e.addAndUpdateConsole(green.Sprint("Command executed successfully:\n"), string(output))
}

// addAndUpdateConsole adds messages to the console and updates the UI
func (e *HiddifyExtensionSimpleSsh) addAndUpdateConsole(message ...any) {
	e.console = fmt.Sprintln(message...) + e.console
	e.UpdateUI(e.GetUI()) // Refresh the UI with new console content
}

// SubmitData processes form submission and starts the background task
func (e *HiddifyExtensionSimpleSsh) SubmitData(data map[string]string) error {
	// Validate and set the form data
	err := e.setFormData(data)
	if err != nil {
		e.ShowMessage("Invalid data", err.Error())
		return err
	}

	// Cancel any ongoing background task
	if e.cancel != nil {
		e.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	// Start SSH command execution in the background
	go e.backgroundTask(ctx)

	return nil
}

// Cancel stops the background task
func (e *HiddifyExtensionSimpleSsh) Cancel() error {
	if e.cancel != nil {
		e.cancel()     // Cancel background task
		e.cancel = nil // Clear cancel function
	}
	return nil
}

// Stop is called when the extension is closed
func (e *HiddifyExtensionSimpleSsh) Stop() error {
	return e.Cancel()
}

// NewHiddifyExtensionSimpleSsh initializes a new instance of HiddifyExtensionSimpleSsh
func NewHiddifyExtensionSimpleSsh() ex.Extension {
	return &HiddifyExtensionSimpleSsh{
		Base: ex.Base[HiddifyExtensionSimpleSshData]{ // Set default values
			Data: HiddifyExtensionSimpleSshData{
				IP:       "127.0.0.1",
				Port:     "22",
				Username: "",
				Password: "",
				Command:  "echo 'Hello, World!'",
			},
		},
		console: yellow.Sprint("Ready to execute commands over SSH\n"),
	}
}

// init registers the extension with metadata
func init() {
	ex.RegisterExtension(
		ex.ExtensionFactory{
			Id:          "github.com/aleskxyz/hiddify_extension_simple_ssh/hiddify_extension",
			Title:       "Simple SSH",
			Description: "An extension to execute commands on a remote SSH server",
			Builder:     NewHiddifyExtensionSimpleSsh,
		},
	)
}
