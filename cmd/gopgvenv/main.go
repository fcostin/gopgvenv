package main

import "errors"
import "fmt"
import "io/ioutil"
import "log"
import "net"
import "os"
import "os/exec"
import "path/filepath"
import "runtime"
import "syscall"
import "strconv"
import "strings"
import "time"

func getPGBinDir() string {
	cmd := exec.Command("pg_config", "--bindir")
	outb, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(outb))
}

func subprocess(binPath string, args []string) (string, int) {
	cmd := exec.Command(binPath, args...)
	outb, err := cmd.CombinedOutput()
	var exitcode int = 0
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// Program exited with nonzero exit code. See if we can extract the exit code value.
			status, ok := exiterr.Sys().(syscall.WaitStatus)
			if ok {
				// ref: https://groups.google.com/forum/#!topic/golang-nuts/dKbL1oOiCIY
				// ref: https://groups.google.com/forum/#!msg/golang-nuts/8XIlxWgpdJw/Z8s2N-SoWHsJ
				// This appears to work in both Linux and Windows:
				exitcode = int(status.ExitStatus())
			} else {
				log.Println("Failed to extract exit code, using bogus value of -1")
				exitcode = -1
			}
		} else {
			// Some other failure. Panic.
			log.Panicln("exec failed:", binPath, args, string(outb))
		}
	}
	return string(outb), exitcode
}

func subprocess_(binPath string, args []string) int {
	cmd := exec.Command(binPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Panicln("Command failed to start:", err)
	}
	err = cmd.Wait()
	var exitcode int = 0
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// Program exited with nonzero exit code. See if we can extract the exit code value.
			status, ok := exiterr.Sys().(syscall.WaitStatus)
			if ok {
				// ref: https://groups.google.com/forum/#!topic/golang-nuts/dKbL1oOiCIY
				// ref: https://groups.google.com/forum/#!msg/golang-nuts/8XIlxWgpdJw/Z8s2N-SoWHsJ
				// This appears to work in both Linux and Windows:
				exitcode = int(status.ExitStatus())
			} else {
				log.Println("Failed to extract exit code, using bogus value of -1")
				exitcode = -1
			}
		} else {
			// Some other failure. Panic.
			log.Panicln("exec failed:", binPath, args)
		}
	}
	return exitcode
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// PostgresServer interface

type PostgresServer interface {
	Start()
	Stop()
	ConnectionUri() string
}

// PGCtlPostresServer -- implementation of PostgresServer using pg_ctl

type PGCtlPostgresServer struct {
	binDir   string
	logPath  string
	dataDir  string
	sockDir  string
	host     string
	database string
	port     int
	options  []string
}

func (h *PGCtlPostgresServer) Start() {
	all_options := []string{}
	all_options = append(all_options,
		"-i", "-h", h.host, "-p", strconv.Itoa(h.port),
		"-k", h.sockDir,
	)
	all_options = append(all_options, h.options...)
	postgresOptions := fmtOptionString(all_options...)
	pgctlPath := filepath.Join(h.binDir, "pg_ctl")
	// TODO log
	status := subprocess_(pgctlPath, []string{"start", "-w", "-o", postgresOptions, "-D", h.dataDir})
	if status != 0 {
		log.Panicln("pg_ctl failed with exit code %d", status)
	}
}

func (h *PGCtlPostgresServer) Stop() {
	pgctlPath := filepath.Join(h.binDir, "pg_ctl")
	errout, status := subprocess(pgctlPath, []string{"stop", "-D", h.dataDir})
	if status != 0 {
		log.Panicln("pg_ctl failed with exit code %d; details: %s", status, errout)
	}
}

func (h *PGCtlPostgresServer) ConnectionUri() string {
	return fmt.Sprintf("postgresql://%s:%d/%s", h.host, h.port, h.database)
}

// RawPostgresServer -- implementation of PostgresServer directly using postgres

type RawPostgresServer struct {
	binDir   string
	logPath  string
	dataDir  string
	sockDir  string
	host     string
	database string
	port     int
	options  []string
	cmd      *exec.Cmd
}

func (h *RawPostgresServer) Start() {
	all_options := []string{}
	all_options = append(all_options,
		"-i", "-h", h.host, "-p", strconv.Itoa(h.port),
		"-k", h.sockDir, "-D", h.dataDir,
	)
	all_options = append(all_options, h.options...)

	postgresPath := filepath.Join(h.binDir, "postgres")
	h.cmd = exec.Command(postgresPath, all_options...)
	// TODO log stdout and stderr to logPath
	h.cmd.Stdout = os.Stdout
	h.cmd.Stderr = os.Stderr

	error_chan := make(chan error, 1)
	ready_chan := make(chan int, 1)
	term_chan := make(chan error, 1)

	cleanLaunch := false

	err := h.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// ensure postgres server is cleaned up if it fails to launch
	// cleanly or we are terminated while we are waiting for it to
	// be ready.
	defer func() {
		if !cleanLaunch {
			// attempt to kill postgres if we're exiting without a clean launch
			h.Stop()
		}
	}()

	// monitor for unexpected server termination during launch
	go func() {
		err := h.cmd.Wait()
		term_chan <- err
	}()

	// monitor for server using pg_isready
	go func(pgIsReadyPath string, dbname string) {
		// Wait for server to boot
		t_timeout := time.Now().Add(time.Duration(60) * time.Second)
		for time.Now().Before(t_timeout) {
			// ref: https://www.postgresql.org/docs/9.6/static/app-pg-isready.html
			errout, status := subprocess(pgIsReadyPath, []string{"--dbname=" + dbname})
			if status == 0 {
				// server accepting connections. signal success
				ready_chan <- 1
				return
			} else if status == 1 {
				// server responded but rejected connection. wait.
			} else if status == 2 {
				// server did not respond. wait.
			} else {
				// some unexpected error
				error_chan <- errors.New("unexpected pg_isready error waiting for postgres server --dbname=" + dbname + "; detail: " + errout)
				return
			}
			time.Sleep(time.Second)
		}
		error_chan <- errors.New("pg_isready timed out waiting for postgres server --dbname=" + dbname)
	}(filepath.Join(h.binDir, "pg_isready"), h.ConnectionUri())

	// wait to see what happens
	select {
	case err := <-error_chan:
		log.Panicln("error during boot:", err)
	case err := <-term_chan:
		log.Panicln("postgres server terminated during boot:", err)
	case <-ready_chan:
		log.Println("postgres server ready")
		cleanLaunch = true
	}

}

func (h *RawPostgresServer) Stop() {
	if h.cmd != nil {
		err := h.cmd.Process.Kill()
		if err != nil {
			log.Fatal(err)
		}
		h.cmd = nil
	}
}

func (h *RawPostgresServer) ConnectionUri() string {
	return fmt.Sprintf("postgresql://%s:%d/%s", h.host, h.port, h.database)
}

func fmtOptionString(options ...string) string {
	return strings.Join(options, " ")
}

type ShellExecutor interface {
	Exec(userCommand []string) error
}

func LinuxShellExec(args []string) int {
	// Unsure if this is correct.
	// Alternatively, we could join args with " ", but
	// that obviously seems wrong as it doesnt preserve
	// the number of args passed to the user's command
	// in cases such as:
	// $ foo " " "  " "   "
	// vs
	// $ gopgvenv foo " " "  " "   "
	shArgs := []string{"-c"}
	shArgs = append(shArgs, args...)
	return subprocess_("/bin/sh", shArgs)
}

func WindowsShellExec(args []string) int {
	// Here be dragons.

	// In Linux, the shell is responsible for passing argv to programs.
	// In Windows, each program receives a single command string from the
	// operating system and is responsible for turning that into argv, if
	// it so chooses. Similarly, a windows program launching a subprogram
	// with command line arguments needs to encode all those arguments
	// as a single giant string.

	// Go aspires to provide a cross-platform API for Go programs that
	// wish to receive command line arguments (via os.Args) or execute
	// processes with command line arguments (via e.g. exec.Command ).

	// Therefore, like all Windows programs, the Windows implementation
	// of a Go program therefore needs to implement:
	// (i)  parsing a big windows command string into an ANSI C style
	//      argv []string; and
	// (ii) formatting []string back into a big parseable windows
	//      command string when launching command line processes
	//
	// Both of these are invoked automatically so I don't believe we
	// need to do anything special here!

	// Go implements big-windows-string -> argv []string like so:
	// https://github.com/golang/go/commit/39c8d2b7faed06b0e91a1ad7906231f53aab45d1
	// and the inverse transform as `makeCmdLine` like so:
	// https://golang.org/src/syscall/exec_windows.go

	// For more background on this kind of thing, please refer to this
	// microsoft dev blog:

	// ref: https://blogs.msdn.microsoft.com/twistylittlepassagesallalike/2011/04/23/everyone-quotes-command-line-arguments-the-wrong-way/

	// Throw all our user args directly into exec.Command where
	// they'll be encoded with `makeCmdLine`
	cmdArgs := []string{"/c"}
	cmdArgs = append(cmdArgs, args...)
	return subprocess_("cmd", cmdArgs)
}

func main() {
	pgBinDir := getPGBinDir()

	fmt.Println("pgBinDir=", pgBinDir)

	workDir, err := ioutil.TempDir("", "gopgvenv_")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = os.RemoveAll(workDir)
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("workDir=", workDir)

	pgDataDir := filepath.Join(workDir, "pgdata")
	err = os.MkdirAll(pgDataDir, 0755)
	if err != nil {
		panic(err)
	}

	fmt.Println("pgDataDir=", pgDataDir)

	pgSockDir := filepath.Join(workDir, "pgsock")
	err = os.MkdirAll(pgSockDir, 0755)
	if err != nil {
		panic(err)
	}

	fmt.Println("pgSockDir=", pgSockDir)

	initdbOptions := fmtOptionString(
		"-A", "trust",
	)
	pgctlPath := filepath.Join(pgBinDir, "pg_ctl")

	errout, status := subprocess(pgctlPath, []string{"initdb", "-o", initdbOptions, "--pgdata", pgDataDir})
	if status != 0 {
		log.Panicln("initdb failed with exit code %d; details: %s", status, errout)
	}

	port, err := getFreePort()
	if err != nil {
		panic(err)
	}

	fmt.Println("port=", port)

	pgLogPath := filepath.Join(workDir, "postgres.log")

	var server PostgresServer

	if true {
		server = &PGCtlPostgresServer{
			binDir:   pgBinDir,
			logPath:  pgLogPath,
			dataDir:  pgDataDir,
			sockDir:  pgSockDir,
			host:     "localhost",
			database: "postgres",
			port:     port,
			options:  []string{"-F"}, // disposable db: data persistence not required, disable fsync for speed.
		}
	} else {
		server = &RawPostgresServer{
			binDir:   pgBinDir,
			logPath:  pgLogPath,
			dataDir:  pgDataDir,
			sockDir:  pgSockDir,
			host:     "localhost",
			database: "postgres",
			port:     port,
			options:  []string{"-F"}, // disposable db: data persistence not required, disable fsync for speed.
			cmd:      nil,
		}
	}

	server.Start()
	defer server.Stop()

	os.Setenv("PGURL", server.ConnectionUri())

	userCommand := os.Args[1:]
	if len(userCommand) == 0 {
		panic("No user command was given: nothing to do after starting postgres server.")
	}

	userStatus := 0
	if runtime.GOOS == "windows" {
		userStatus = WindowsShellExec(userCommand)
	} else {
		userStatus = LinuxShellExec(userCommand)
	}
	if userStatus != 0 {
		log.Panicln("user command failed with exit code %d")
	}

	// TODO: define convention for gopgvenv exit code. for example:
	// 0 everything in order
	// 1 internal error
	// 2 user command error
	// 3 postgres-related environmental error
}
