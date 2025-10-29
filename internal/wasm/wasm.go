//go:build js && wasm
// +build js,wasm

package wasm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall/js"

	"github.com/spf13/cobra"
)

var (
	// Use a single buffer pair with mutex protection
	stdoutBuffer bytes.Buffer
	stderrBuffer bytes.Buffer
	bufferMutex  sync.Mutex

	// Original stdout/stderr
	originalStdout *os.File
	originalStderr *os.File

	// Writers for capturing output
	stdoutWriter io.Writer
	stderrWriter io.Writer

	// Debug flag
	debugMode = false
)

// Add this near the top of the file after your imports

// WasmOutputBuffer is used to directly capture output from Cobra commands
var WasmOutputBuffer = &DirectOutputBuffer{
	buffer: &bytes.Buffer{},
}

// DirectOutputBuffer provides a buffer that can be used directly by Cobra commands
type DirectOutputBuffer struct {
	buffer *bytes.Buffer
	mutex  sync.Mutex
}

func (d *DirectOutputBuffer) Write(p []byte) (n int, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// More detailed console logging
	content := string(p)
	js.Global().Get("console").Call("log", fmt.Sprintf("📝 BUFFER WRITE [%d bytes]:", len(p)))

	// Split multi-line output for better readability
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line != "" {
			js.Global().Get("console").Call("log", fmt.Sprintf("  │ %s", line))
		}
	}

	// Also write to console directly for visibility during debugging
	js.Global().Get("console").Call("debug", content)

	return d.buffer.Write(p)
}

// Add this function to help debug command traversal
func TraceCommandExecution(cmd *cobra.Command, args []string) {
	js.Global().Get("console").Call("group", "⚡ COMMAND TRAVERSAL")

	// Log the command hierarchy
	currentCmd := cmd
	cmdPath := []string{}
	for currentCmd != nil {
		cmdPath = append([]string{currentCmd.Name()}, cmdPath...)
		currentCmd = currentCmd.Parent()
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("Command path: %s", strings.Join(cmdPath, " → ")))
	js.Global().Get("console").Call("log", fmt.Sprintf("Args: %v", args))

	// Show available subcommands at this level
	if len(cmd.Commands()) > 0 {
		subNames := []string{}
		for _, sub := range cmd.Commands() {
			subNames = append(subNames, sub.Name())
		}
		js.Global().Get("console").Call("log", fmt.Sprintf("Available subcommands: %v", subNames))
	}

	js.Global().Get("console").Call("groupEnd")
}

func TraceCommand(command string, args []string) {
	js.Global().Get("console").Call("group", fmt.Sprintf("🔍 COMMAND: %s", command))
	js.Global().Get("console").Call("log", fmt.Sprintf("Full command: %s", command))
	js.Global().Get("console").Call("log", fmt.Sprintf("Parsed args: %v", args))

	// Log environment variables that might affect command execution
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")
	env := os.Getenv("MEGAPORT_ENVIRONMENT")

	js.Global().Get("console").Call("log", fmt.Sprintf("Environment: %s", env))
	js.Global().Get("console").Call("log", fmt.Sprintf("Auth configured: %v", accessKey != "" && secretKey != ""))
	js.Global().Get("console").Call("groupEnd")
}

func (d *DirectOutputBuffer) String() string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	result := d.buffer.String()
	// Debug read operation
	fmt.Printf("WASM Debug: Reading buffer, length: %d\n", len(result))
	return result
}

func (d *DirectOutputBuffer) Reset() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.buffer.Reset()
}

// EnableDebugMode turns on additional debugging
func EnableDebugMode() {
	debugMode = true
	js.Global().Set("wasmDebug", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return debugMode
	}))
}

func debugAuthInfo(this js.Value, args []js.Value) interface{} {
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")
	env := os.Getenv("MEGAPORT_ENVIRONMENT")

	// Mask secret key for security
	maskedSecret := ""
	if secretKey != "" {
		if len(secretKey) > 4 {
			maskedSecret = secretKey[:4] + "..." + secretKey[len(secretKey)-4:]
		} else {
			maskedSecret = "****"
		}
	}

	return map[string]interface{}{
		"accessKeySet":     accessKey != "",
		"accessKeyPreview": accessKey[:4] + "..." + accessKey[len(accessKey)-4:],
		"secretKeySet":     secretKey != "",
		"secretKeyPreview": maskedSecret,
		"environment":      env,
	}
}

// RegisterJSFunctions registers Go functions with JavaScript
func RegisterJSFunctions() {
	// Export file system operations
	js.Global().Set("readConfigFile", js.FuncOf(readConfigFile))
	js.Global().Set("writeConfigFile", js.FuncOf(writeConfigFile))
	js.Global().Set("debugAuthInfo", js.FuncOf(debugAuthInfo))

	// Export storage operations
	js.Global().Set("saveToLocalStorage", js.FuncOf(saveToLocalStorage))
	js.Global().Set("loadFromLocalStorage", js.FuncOf(loadFromLocalStorage))

	// Debug functions
	js.Global().Set("resetWasmOutput", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ResetOutputBuffers()
		return true
	}))

	js.Global().Set("getWasmOutput", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return GetCapturedOutput()
	}))

	InstallCommandHooks()

	// Add a debug mode toggle function
	js.Global().Set("toggleWasmDebug", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		debugMode = !debugMode
		return debugMode
	}))

	// Add a function to dump the full buffer contents for debugging
	js.Global().Set("dumpBuffers", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		stdoutContent := stdoutBuffer.String()
		stderrContent := stderrBuffer.String()
		directContent := WasmOutputBuffer.String()

		return map[string]interface{}{
			"stdout": stdoutContent,
			"stderr": stderrContent,
			"direct": directContent,
		}
	}))
}

// ResetOutputBuffers clears the output buffers
func ResetOutputBuffers() {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	stdoutBuffer.Reset()
	stderrBuffer.Reset()
	WasmOutputBuffer.Reset()

	// CRITICAL: Also clear all the global output variables
	js.Global().Delete("wasmJSONOutput")
	js.Global().Delete("wasmCSVOutput")
	js.Global().Delete("wasmTableOutput")

	if debugMode {
		js.Global().Get("console").Call("log", "Output buffers reset (including all structured output globals)")
	}
}

// GetCapturedOutput returns all captured output
func GetCapturedOutput() string {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	out := stdoutBuffer.String()
	err := stderrBuffer.String()
	direct := WasmOutputBuffer.String()

	// IMPORTANT: Also check for structured output from output package
	// Check for JSON output
	jsonOutput := ""
	if wasmJSONGlobal := js.Global().Get("wasmJSONOutput"); !wasmJSONGlobal.IsUndefined() && !wasmJSONGlobal.IsNull() {
		jsonOutput = wasmJSONGlobal.String()
		js.Global().Get("console").Call("log", fmt.Sprintf("📝 Found JSON output: %d bytes", len(jsonOutput)))
	}

	// Check for CSV output
	csvOutput := ""
	if wasmCSVGlobal := js.Global().Get("wasmCSVOutput"); !wasmCSVGlobal.IsUndefined() && !wasmCSVGlobal.IsNull() {
		csvOutput = wasmCSVGlobal.String()
		js.Global().Get("console").Call("log", fmt.Sprintf("📊 Found CSV output: %d bytes", len(csvOutput)))
	}

	// Check for table output
	tableOutput := ""
	if wasmTableWriterGlobal := js.Global().Get("wasmTableOutput"); !wasmTableWriterGlobal.IsUndefined() && !wasmTableWriterGlobal.IsNull() {
		tableOutput = wasmTableWriterGlobal.String()
		js.Global().Get("console").Call("log", fmt.Sprintf("📊 Found table output: %d bytes", len(tableOutput)))
	}

	// Log what was captured in each buffer
	js.Global().Get("console").Call("group", "📤 OUTPUT CAPTURE RESULTS")
	js.Global().Get("console").Call("log", fmt.Sprintf("stdout buffer: [%d bytes]", len(out)))
	js.Global().Get("console").Call("log", fmt.Sprintf("stderr buffer: [%d bytes]", len(err)))
	js.Global().Get("console").Call("log", fmt.Sprintf("direct buffer: [%d bytes]", len(direct)))
	js.Global().Get("console").Call("log", fmt.Sprintf("JSON buffer: [%d bytes]", len(jsonOutput)))
	js.Global().Get("console").Call("log", fmt.Sprintf("CSV buffer: [%d bytes]", len(csvOutput)))
	js.Global().Get("console").Call("log", fmt.Sprintf("table buffer: [%d bytes]", len(tableOutput)))

	// Priority order: JSON > CSV > table > direct > stdout/stderr combined
	// This ensures structured output formats take precedence
	finalOutput := ""
	outputSource := ""

	if jsonOutput != "" {
		finalOutput = jsonOutput
		outputSource = "JSON buffer"
	} else if csvOutput != "" {
		finalOutput = csvOutput
		outputSource = "CSV buffer"
	} else if tableOutput != "" {
		finalOutput = tableOutput
		outputSource = "table buffer"
	} else if direct != "" {
		finalOutput = direct
		outputSource = "direct buffer"
	} else {
		finalOutput = out + err
		outputSource = "combined stdout/stderr"
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("Using %s for output", outputSource))

	js.Global().Get("console").Call("log", fmt.Sprintf("Final output length: %d bytes", len(finalOutput)))
	js.Global().Get("console").Call("groupEnd")

	return finalOutput
}

// SetupIO redirects stdout/stderr to our buffers using WasmOutputBuffer
// In WASM, os.Pipe() is not implemented, so we don't need complex IO redirection
// The output is already captured through WasmOutputBuffer which Cobra commands use
func SetupIO() {
	// Save original stdout/stderr references
	originalStdout = os.Stdout
	originalStderr = os.Stderr

	// Note: In WASM, we don't need to redirect os.Stdout/Stderr because:
	// 1. Cobra commands are configured to write to WasmOutputBuffer directly
	// 2. Table output writes to WasmTableWriter and sets wasmTableOutput global
	// 3. Any fmt.Print() calls go to the console via wasm_exec.js

	// The output capture happens through:
	// - Direct writes to WasmOutputBuffer from Cobra commands
	// - Table output via WasmTableWriter -> wasmTableOutput global
	// - These are collected in GetCapturedOutput()

	if debugMode {
		js.Global().Get("console").Call("log", "✅ WASM IO setup complete (using WasmOutputBuffer)")
	}
}

// customWriter provides a custom io.Writer implementation that handles WASM quirks
type customWriter struct {
	writer io.Writer
}

func (cw *customWriter) Write(p []byte) (n int, err error) {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	return cw.writer.Write(p)
}

// CaptureOutput runs a function and captures its stdout/stderr output
func CaptureOutput(fn func()) string {
	// Reset buffers before capture
	ResetOutputBuffers()
	WasmOutputBuffer.Reset()

	// Create pipes for output capture
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Save original stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Replace with pipes
	os.Stdout = wOut
	os.Stderr = wErr

	// Execute function in a controlled environment
	done := make(chan bool)
	go func() {
		fn()
		// Ensure all output is flushed
		wOut.Close()
		wErr.Close()
		done <- true
	}()

	// Read from pipes in parallel
	var stdoutStr, stderrStr string
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		stdoutStr = buf.String()
		wg.Done()
	}()

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		stderrStr = buf.String()
		wg.Done()
	}()

	// Wait for the command to finish
	<-done

	// Wait for all reads to complete
	wg.Wait()

	// Restore original stdout/stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Combine captured outputs
	combinedOutput := stdoutStr + stderrStr

	// Also check our direct buffer which may have captured output
	directOutput := WasmOutputBuffer.String()

	// Store in our main buffers for later access
	bufferMutex.Lock()
	stdoutBuffer.WriteString(stdoutStr)
	stderrBuffer.WriteString(stderrStr)
	bufferMutex.Unlock()

	// Return all output - prefer direct capture if available
	if directOutput != "" {
		return directOutput
	}
	return combinedOutput
}

func SplitArgs(cmd string) []string {
	// Trim any leading/trailing whitespace
	cmd = strings.TrimSpace(cmd)

	// Debug the raw command
	fmt.Printf("WASM Debug: Parsing command: '%s'\n", cmd)

	var args []string
	inQuote := false
	var currentArg strings.Builder
	quoteChar := rune(0) // Track which quote char opened the current quote

	for _, r := range cmd {
		switch {
		case (r == '"' || r == '\'') && (inQuote && r == quoteChar):
			// Closing quote that matches opening quote
			inQuote = false
			quoteChar = rune(0)
		case (r == '"' || r == '\'') && !inQuote:
			// Opening quote
			inQuote = true
			quoteChar = r
		case r == ' ' && !inQuote:
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			currentArg.WriteRune(r)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	// Remove program name if user included it
	// The main_wasm.go will add it back, so we don't want duplicates
	cleanedArgs := []string{}
	for _, arg := range args {
		// Skip any occurrence of the program name
		if arg == "megaport-cli" || arg == "./megaport-cli" || arg == "megaport" {
			continue
		}
		cleanedArgs = append(cleanedArgs, arg)
	}

	// Debug the cleaned arguments
	fmt.Printf("WASM Debug: Original args: %v\n", args)
	fmt.Printf("WASM Debug: Cleaned args (program name removed): %v\n", cleanedArgs)

	return cleanedArgs
}

func readConfigFile(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error": "Filename required",
		}
	}

	filename := args[0].String()

	// Get from localStorage
	content := js.Global().Get("localStorage").Call("getItem", "megaport_fs_"+filename)

	if content.IsNull() {
		// If config file doesn't exist, create an empty one with default structure
		if filename == "config.json" {
			defaultConfig := map[string]interface{}{
				"profiles":      map[string]interface{}{},
				"activeProfile": "",
				"defaults":      map[string]interface{}{},
			}
			defaultConfigBytes, _ := json.Marshal(defaultConfig)
			return map[string]interface{}{
				"content": string(defaultConfigBytes),
			}
		}
		return map[string]interface{}{
			"error": "File not found: " + filename,
		}
	}

	return map[string]interface{}{
		"content": content.String(),
	}
}

func writeConfigFile(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return map[string]interface{}{
			"error": "Filename and content required",
		}
	}

	filename := args[0].String()
	content := args[1].String()

	// Save to localStorage
	js.Global().Get("localStorage").Call("setItem", "megaport_fs_"+filename, content)

	return map[string]interface{}{
		"success": true,
	}
}

// localStorage wrappers
func saveToLocalStorage(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return false
	}

	key := args[0].String()
	value := args[1].String()

	js.Global().Get("localStorage").Call("setItem", key, value)
	return true
}

func loadFromLocalStorage(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return ""
	}

	key := args[0].String()
	return js.Global().Get("localStorage").Call("getItem", key)
}

// Add this function to install hook for specific commands
func InstallCommandHooks() {
	// Create a global JavaScript function to log command-specific details
	js.Global().Set("logLocationCommand", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		message := "No message provided"
		if len(args) > 0 {
			message = args[0].String()
		}

		js.Global().Get("console").Call("group", "🌎 LOCATIONS COMMAND DEBUG")
		js.Global().Get("console").Call("log", message)

		// Log environment status
		accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
		secretKey := os.Getenv("MEGAPORT_SECRET_KEY")
		env := os.Getenv("MEGAPORT_ENVIRONMENT")

		js.Global().Get("console").Call("log", fmt.Sprintf("Environment: %s", env))
		js.Global().Get("console").Call("log", fmt.Sprintf("Access Key: %v", accessKey != ""))
		js.Global().Get("console").Call("log", fmt.Sprintf("Secret Key: %v", secretKey != ""))
		js.Global().Get("console").Call("groupEnd")

		return nil
	}))
}
