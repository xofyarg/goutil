// Experimental daemonize package for go, implemented using a dirty
// environment variable and ForkExec.
package daemon

import (
	"errors"
	"os"
	"syscall"
)

// environment variable used to distinguish parent/child. Can be set
// before calling Daemon.
var DaemonEnv = "_GODAEMON"

// Simulate daemon(3). Chdir to "/" if nochdir is false, close
// stand{in,out,err} if noclose is false.
func Start(nochdir, noclose bool) error {
	switch os.Getenv(DaemonEnv) {
	case "":
		if err := os.Setenv(DaemonEnv, "1"); err != nil {
			return err
		}
		if err := parent(); err != nil {
			return err
		}
		os.Exit(0)
	case "1":
		if err := os.Setenv(DaemonEnv, "2"); err != nil {
			return err
		}
		if err := child(noclose); err != nil {
			return err
		}
		os.Exit(0)
		// return nil
	case "2":
		// TODO: add this call after go1.4
		// os.Unsetenv(DaemonEnv)

		// keep chroot in the last step to keep working directory information
		if !nochdir {
			os.Chdir("/")
		}
		return nil
	default:
		return errors.New("environment variable exists")
	}
	return nil
}

func parent() error {
	attr := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2},
		// Create new session, being session leader, drop control
		// terminal. This cannot be done in parent since parent is already
		// process/group leader.
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	}

	_, _, err := syscall.StartProcess(os.Args[0], os.Args, attr)
	if err != nil {
		return err
	}
	return nil
}

func child(noclose bool) error {
	var files []uintptr
	if noclose {
		files = []uintptr{0, 1, 2}
	}

	attr := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: files,
	}

	_, _, err := syscall.StartProcess(os.Args[0], os.Args, attr)
	if err != nil {
		return err
	}

	return nil
}
