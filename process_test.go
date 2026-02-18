package run_test

import (
	"bytes"
	"testing"

	"github.com/matgreaves/run"
	"github.com/matryer/is"
)

func TestCommand(t *testing.T) {
	is := is.New(t)
	p := run.Command("echo", "Hello, World!")
	buf := &bytes.Buffer{}
	p.Stdout = buf
	err := p.Run(t.Context())
	is.NoErr(err)
	is.Equal(buf.String(), "Hello, World!\n")
}

func TestProcess_Env(t *testing.T) {
	is := is.New(t)
	buf := &bytes.Buffer{}
	p := run.Process{
		Name:   "env-test",
		Path:   "sh",
		Args:   []string{"-c", `echo "$MY_VAR"`},
		Env:    map[string]string{"MY_VAR": "hello-from-env"},
		Stdout: buf,
	}
	err := p.Run(t.Context())
	is.NoErr(err)
	is.Equal(buf.String(), "hello-from-env\n")
}

func TestProcess_InheritOSEnv(t *testing.T) {
	is := is.New(t)
	t.Setenv("RUN_TEST_INHERIT", "inherited-value")

	buf := &bytes.Buffer{}
	p := run.Process{
		Name:         "inherit-test",
		Path:         "sh",
		Args:         []string{"-c", `echo "$RUN_TEST_INHERIT"`},
		InheritOSEnv: true,
		Stdout:       buf,
	}
	err := p.Run(t.Context())
	is.NoErr(err)
	is.Equal(buf.String(), "inherited-value\n")
}

func TestProcess_EnvOverridesInherited(t *testing.T) {
	is := is.New(t)
	t.Setenv("RUN_TEST_OVERRIDE", "original")

	buf := &bytes.Buffer{}
	p := run.Process{
		Name:         "override-test",
		Path:         "sh",
		Args:         []string{"-c", `echo "$RUN_TEST_OVERRIDE"`},
		InheritOSEnv: true,
		Env:          map[string]string{"RUN_TEST_OVERRIDE": "replaced"},
		Stdout:       buf,
	}
	err := p.Run(t.Context())
	is.NoErr(err)
	is.Equal(buf.String(), "replaced\n")
}

