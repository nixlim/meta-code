package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// STDIOTransport implements the Transport interface for STDIO-based communication
// with subprocess MCP servers.
type STDIOTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	reader *bufio.Reader
	writer *bufio.Writer

	codec     jsonrpc.Codec
	connected bool
	mu        sync.RWMutex
	writeMu   sync.Mutex // Protects writer for concurrent sends

	// Error channel for stderr output
	errChan chan error
	// Done channel to signal shutdown
	done chan struct{}
	// Process wait result
	processErr chan error
	waitOnce   sync.Once
}

// NewSTDIOTransport creates a new STDIO transport for the given command.
// The command should be the path to the MCP server executable with any required arguments.
func NewSTDIOTransport(cmd *exec.Cmd) (*STDIOTransport, error) {
	if cmd == nil {
		return nil, fmt.Errorf("command cannot be nil")
	}

	// Get pipes for stdin, stdout, and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	transport := &STDIOTransport{
		cmd:        cmd,
		stdin:      stdin,
		stdout:     stdout,
		stderr:     stderr,
		reader:     bufio.NewReader(stdout),
		writer:     bufio.NewWriter(stdin),
		codec:      &JSONCodec{},
		connected:  true,
		errChan:    make(chan error, 1),
		done:       make(chan struct{}),
		processErr: make(chan error, 1),
	}

	// Start monitoring stderr in a goroutine
	go transport.monitorStderr()

	// Start monitoring process exit
	go transport.monitorProcess()

	return transport, nil
}

// Send sends a message over the STDIO transport
func (t *STDIOTransport) Send(ctx context.Context, message jsonrpc.Message) error {
	t.mu.RLock()
	if !t.connected {
		t.mu.RUnlock()
		return fmt.Errorf("transport is not connected")
	}
	t.mu.RUnlock()

	// Protect concurrent writes
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	// Encode the message
	if err := t.codec.Encode(t.writer, message); err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	// Flush the writer to ensure the message is sent
	if err := t.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// Receive receives a message from the STDIO transport
func (t *STDIOTransport) Receive(ctx context.Context) (jsonrpc.Message, error) {
	t.mu.RLock()
	if !t.connected {
		t.mu.RUnlock()
		return nil, fmt.Errorf("transport is not connected")
	}
	t.mu.RUnlock()

	// Create a channel for the result
	type result struct {
		msg jsonrpc.Message
		err error
	}
	resultChan := make(chan result, 1)

	// Decode message in a goroutine to support context cancellation
	go func() {
		msg, err := t.codec.Decode(t.reader)
		resultChan <- result{msg: msg, err: err}
	}()

	// Wait for either the result or context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultChan:
		if res.err != nil {
			return nil, fmt.Errorf("failed to decode message: %w", res.err)
		}
		return res.msg, nil
	case <-t.done:
		return nil, fmt.Errorf("transport closed")
	}
}

// SendBatch sends multiple messages as a batch
func (t *STDIOTransport) SendBatch(ctx context.Context, messages []jsonrpc.Message) error {
	t.mu.RLock()
	if !t.connected {
		t.mu.RUnlock()
		return fmt.Errorf("transport is not connected")
	}
	t.mu.RUnlock()

	// Protect concurrent writes
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	// Encode the batch
	if err := t.codec.EncodeBatch(t.writer, messages); err != nil {
		return fmt.Errorf("failed to encode batch: %w", err)
	}

	// Flush the writer
	if err := t.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// ReceiveBatch receives multiple messages as a batch
func (t *STDIOTransport) ReceiveBatch(ctx context.Context) ([]jsonrpc.Message, error) {
	t.mu.RLock()
	if !t.connected {
		t.mu.RUnlock()
		return nil, fmt.Errorf("transport is not connected")
	}
	t.mu.RUnlock()

	// Create a channel for the result
	type result struct {
		msgs []jsonrpc.Message
		err  error
	}
	resultChan := make(chan result, 1)

	// Decode batch in a goroutine to support context cancellation
	go func() {
		msgs, err := t.codec.DecodeBatch(t.reader)
		resultChan <- result{msgs: msgs, err: err}
	}()

	// Wait for either the result or context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultChan:
		if res.err != nil {
			return nil, fmt.Errorf("failed to decode batch: %w", res.err)
		}
		return res.msgs, nil
	case <-t.done:
		return nil, fmt.Errorf("transport closed")
	}
}

// Close closes the transport connection
func (t *STDIOTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil
	}

	t.connected = false
	close(t.done)

	// Close pipes
	var errs []error
	if err := t.stdin.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close stdin: %w", err))
	}
	if err := t.stdout.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close stdout: %w", err))
	}
	if err := t.stderr.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close stderr: %w", err))
	}

	// Trigger process wait if not already done
	go t.monitorProcess()

	// Wait for process to exit or timeout
	select {
	case err := <-t.processErr:
		if err != nil && err.Error() != "signal: killed" && !strings.Contains(err.Error(), "no child processes") {
			errs = append(errs, fmt.Errorf("process exit error: %w", err))
		}
	case <-time.After(5 * time.Second):
		// Force kill if it doesn't exit gracefully
		if t.cmd.Process != nil {
			if err := t.cmd.Process.Kill(); err != nil && !strings.Contains(err.Error(), "process already finished") {
				errs = append(errs, fmt.Errorf("failed to kill process: %w", err))
			}
		}
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// IsConnected returns true if the transport is connected
func (t *STDIOTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// GetProcessInfo returns information about the subprocess
func (t *STDIOTransport) GetProcessInfo() (pid int, running bool) {
	if t.cmd != nil && t.cmd.Process != nil {
		pid = t.cmd.Process.Pid
		running = t.IsConnected()
	}
	return
}

// monitorStderr monitors the stderr output from the subprocess
func (t *STDIOTransport) monitorStderr() {
	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		// Log stderr output for debugging
		// In production, this could be sent to a logger
		select {
		case t.errChan <- fmt.Errorf("stderr: %s", line):
		default:
			// Don't block if nobody is reading
		}
	}
}

// monitorProcess monitors the subprocess for unexpected exits
func (t *STDIOTransport) monitorProcess() {
	// Wait for the process only once
	t.waitOnce.Do(func() {
		err := t.cmd.Wait()
		t.processErr <- err
		close(t.processErr)
		
		// Mark as disconnected
		t.mu.Lock()
		t.connected = false
		t.mu.Unlock()
		
		// Report error if unexpected
		if err != nil && err.Error() != "signal: killed" {
			select {
			case t.errChan <- fmt.Errorf("process exited: %w", err):
			default:
			}
		}
	})
}

// GetLastError returns the last error from stderr or process monitoring
func (t *STDIOTransport) GetLastError() error {
	select {
	case err := <-t.errChan:
		return err
	default:
		return nil
	}
}

// JSONCodec implements the Codec interface for JSON encoding/decoding
type JSONCodec struct{}

// Encode encodes a message to JSON with newline delimiter
func (c *JSONCodec) Encode(w io.Writer, message jsonrpc.Message) error {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(message); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// Decode decodes a message from JSON
func (c *JSONCodec) Decode(r io.Reader) (jsonrpc.Message, error) {
	decoder := json.NewDecoder(r)
	
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Parse the raw message to determine its type
	return jsonrpc.ParseMessage([]byte(raw))
}

// EncodeBatch encodes multiple messages as a JSON array
func (c *JSONCodec) EncodeBatch(w io.Writer, messages []jsonrpc.Message) error {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(messages); err != nil {
		return fmt.Errorf("failed to encode batch: %w", err)
	}
	return nil
}

// DecodeBatch decodes multiple messages from a JSON array
func (c *JSONCodec) DecodeBatch(r io.Reader) ([]jsonrpc.Message, error) {
	decoder := json.NewDecoder(r)
	
	var raw []json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode batch: %w", err)
	}

	messages := make([]jsonrpc.Message, 0, len(raw))
	for i, rawMsg := range raw {
		msg, err := jsonrpc.ParseMessage([]byte(rawMsg))
		if err != nil {
			return nil, fmt.Errorf("failed to parse message %d: %w", i, err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}