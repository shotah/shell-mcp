// Package tools runs one shell command. The host publishes it as shell__command_run.
package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

const (
	defaultTimeout = 30 * time.Second
	maxTimeout     = 120 * time.Second
	maxCommand     = 100_000
	maxBody        = 4000
)

// Run executes command with cwd set to root.
// The result leads with the exit code, then the tail of the output.
func Run(ctx context.Context, root, command string, timeout time.Duration) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", errors.New("command is required")
	}
	if len(command) > maxCommand {
		return "", errors.New("command is too long")
	}
	if strings.ContainsRune(command, 0) {
		return "", errors.New("invalid command")
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if timeout > maxTimeout {
		timeout = maxTimeout
	}
	shell, err := shellPath()
	if err != nil {
		return "", err
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := startShell(runCtx, shell, command)
	cmd.Dir = root
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	timedOut := errors.Is(runCtx.Err(), context.DeadlineExceeded)
	exit := 0
	if timedOut {
		exit = 124
	} else if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exit = ee.ExitCode()
		} else {
			return "", err
		}
	}
	return formatRun(exit, stdout.String(), stderr.String(), timedOut), nil
}

func shellPath() (string, error) {
	if _, err := os.Stat("/bin/sh"); err == nil {
		return "/bin/sh", nil
	}
	path, err := exec.LookPath("sh")
	if err != nil {
		return "", errors.New("sh not found")
	}
	return path, nil
}

// startShell starts a shell in its own process group so a timeout kills the
// command and the children it spawned. Killing the shell alone leaves the
// child holding the output pipes until it exits.
func startShell(ctx context.Context, shell, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, shell, "-c", command) //nolint:gosec // G204: command_run is this tool; cwd is the workspace jail
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 200 * time.Millisecond
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == nil || errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	return cmd
}

func formatRun(exit int, stdout, stderr string, timedOut bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "exit: %d\n", exit)
	if timedOut {
		b.WriteString("timed out\n")
	}
	body := joinStreams(stdout, stderr)
	b.WriteString(tail(body, maxBody))
	if !strings.HasSuffix(b.String(), "\n") {
		b.WriteByte('\n')
	}
	return b.String()
}

func joinStreams(stdout, stderr string) string {
	stdout = strings.TrimRight(stdout, "\n")
	stderr = strings.TrimRight(stderr, "\n")
	switch {
	case stdout == "" && stderr == "":
		return ""
	case stderr == "":
		return stdout
	case stdout == "":
		return stderr
	default:
		return "--- stdout ---\n" + stdout + "\n--- stderr ---\n" + stderr
	}
}

func tail(s string, limit int) string {
	if s == "" || utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return "…[truncated head]\n" + string(runes[len(runes)-limit:])
}
