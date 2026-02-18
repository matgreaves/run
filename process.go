package run

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/matgreaves/run/onexit"
)

var _ Runner = Process{}

// Process runs an extermal
type Process struct {
	Name string
	// path to executable
	Path string
	// working directory
	Dir string
	// command-line args
	Args []string
	// environment variables
	Env map[string]string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// InheritOSEnv inherits environment variables from os.Environ.
	// Env takes priority over inherited variables.
	InheritOSEnv bool
}

// Run implements [Runner] starting the external process.
//
// The process can be shut down by cancelling ctx. In this case the process and all child processes
//
//	will receive a SIGINT.
//
// If this program or p do not terminate gracefully then a SIGKILL will be sent to the process group.
func (p Process) Run(ctx context.Context) error {
	var err error
	p.Path, err = exec.LookPath(p.Path)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, p.Path, p.Args...)
	cmd.Dir = p.Dir
	cmd.Stdin = p.Stdin
	cmd.Stdout = p.Stdout
	cmd.Stderr = p.Stderr
	cmd.Env = p.environ()

	// Give the external process its own group to more easily clean up it and all of its children.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
		return nil
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	cancel, err := onexit.Kill(p.Name, -cmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		cmd.Cancel()
		return fmt.Errorf("run: failed to register killer: %w", err)
	}
	defer cancel()

	err = cmd.Wait()
	return err
}

func Command(cmd string, args ...string) Process {
	return Process{
		Name: filepath.Base(cmd),
		Path: cmd,
		Args: args,
	}
}

// environ builds the environment variable slice for the subprocess.
//
// If neither Env nor InheritOSEnv is set, returns nil (inherit parent env).
// If InheritOSEnv is set, starts with os.Environ().
// Env entries are always appended last so they take priority.
func (p Process) environ() []string {
	if len(p.Env) == 0 && !p.InheritOSEnv {
		return nil
	}

	var env []string
	if p.InheritOSEnv {
		env = append(env, os.Environ()...)
	}

	for k, v := range p.Env {
		env = append(env, k+"="+v)
	}
	return env
}
