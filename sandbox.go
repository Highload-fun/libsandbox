// Package sandbox provides a Go API for executing programs inside a sandbox environment.
//
// It is a thin, declarative wrapper over the sandbox tool, responsible only for configuration
// and process invocation. The sandbox executable defines all execution semantics and guarantees.
//
// Sandbox tool documentation and usage:
//
//	https://github.com/Highload-fun/sandbox
package sandbox

import (
	"context"
	"os/exec"
	"strconv"
)

// Path points to the sandbox executable.
// It is used as the program invoked by exec.Command.
var Path = "/usr/bin/sandbox"

// Sandbox is a mutable builder that describes how a program should be executed inside a sandbox.
//
// It accumulates filesystem mappings, environment configuration, resource limits, and execution
// parameters, which are later translated into sandbox tool arguments.
type Sandbox struct {
	path          string
	files         []file
	mountDirs     []mountDir
	env           []string
	noNewNet      bool
	cgroup        string
	cpuSet        string
	memLimit      uint64
	saveUsageStat string
	execDir       string
}

type file struct {
	src      string
	dst      string
	withLibs bool
}

type mountDir struct {
	src string
	dst string
}

// New creates a new sandbox configuration for the given sandbox root path.
func New(path string) *Sandbox {
	return &Sandbox{path: path}
}

// AddFile declares that a file from the host must be available inside the sandbox at the given location.
func (s *Sandbox) AddFile(src, dst string, withLibs bool) *Sandbox {
	s.files = append(s.files, file{
		src:      src,
		dst:      dst,
		withLibs: withLibs,
	})

	return s
}

// MountDir declares that a directory from the host filesystem must be accessible inside the sandbox.
func (s *Sandbox) MountDir(src, dst string) *Sandbox {
	s.mountDirs = append(s.mountDirs, mountDir{
		src: src,
		dst: dst,
	})

	return s
}

// AddEnv adds an environment variable that will be visible to the sandboxed process.
func (s *Sandbox) AddEnv(value string) *Sandbox {
	s.env = append(s.env, value)

	return s
}

// SetNoNewNet configures whether the sandboxed process is isolated from the network.
func (s *Sandbox) SetNoNewNet(v bool) *Sandbox {
	s.noNewNet = v

	return s
}

// SetCGroup assigns the sandboxed process to a control group.
func (s *Sandbox) SetCGroup(name string) *Sandbox {
	s.cgroup = name

	return s
}

// SetCpuSet restricts which CPUs the sandboxed process may use.
func (s *Sandbox) SetCpuSet(set string) *Sandbox {
	s.cpuSet = set

	return s
}

// SetMemLimit limits memory usage of the sandboxed process.
func (s *Sandbox) SetMemLimit(limit uint64) *Sandbox {
	s.memLimit = limit

	return s
}

// SaveUsageStat enables persisting execution statistics after the process exits.
func (s *Sandbox) SaveUsageStat(filename string) *Sandbox {
	s.saveUsageStat = filename

	return s
}

// ExecDir sets the working directory inside the sandbox where the command will be executed.
func (s *Sandbox) ExecDir(dir string) *Sandbox {
	s.execDir = dir

	return s
}

// Command constructs an exec.Cmd that runs a command inside the configured sandbox.
func (s *Sandbox) Command(path string, args ...string) *exec.Cmd {
	return s.CommandContext(nil, path, args...)
}

// CommandContext is identical to Command, but allows the execution to be bound to a context.
func (s *Sandbox) CommandContext(ctx context.Context, path string, args ...string) *exec.Cmd {
	execArgs := s.BuildExecArgs(path, args)

	if ctx == nil {
		return exec.Command(Path, execArgs...)
	}

	return exec.CommandContext(ctx, Path, execArgs...)
}

// BuildExecArgs converts the sandbox configuration into a complete argument list for the sandbox executable.
func (s *Sandbox) BuildExecArgs(path string, args []string) []string {
	execArgs := []string{s.path}

	for _, f := range s.files {
		if f.withLibs {
			execArgs = append(execArgs, "--add_elf_file")
		} else {
			execArgs = append(execArgs, "--add_file")
		}

		execArgs = append(execArgs, f.src, f.dst)
	}

	for _, d := range s.mountDirs {
		execArgs = append(execArgs, "--mount_dir", d.src, d.dst)
	}

	for _, e := range s.env {
		execArgs = append(execArgs, "--env", e)
	}

	if s.noNewNet {
		execArgs = append(execArgs, "--no_new_net")
	}

	if s.cgroup != "" {
		execArgs = append(execArgs, "--cgroup", s.cgroup)
	}

	if s.cpuSet != "" {
		execArgs = append(execArgs, "--cpuset", s.cpuSet)
	}

	if s.memLimit != 0 {
		execArgs = append(execArgs, "--mem_limit", strconv.FormatUint(s.memLimit, 10))
	}

	if s.saveUsageStat != "" {
		execArgs = append(execArgs, "--save_usage_stat", s.saveUsageStat)
	}

	if s.execDir != "" {
		execArgs = append(execArgs, "--exec_dir", s.execDir)
	}

	execArgs = append(execArgs, "--", path)
	execArgs = append(execArgs, args...)
	return execArgs
}
