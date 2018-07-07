package main

import "fmt"
import "io/ioutil"
import "log"
import "net"
import "os"
import "os/exec"
import "path/filepath"
import "strconv"
import "strings"

func getPGBinDir() string {
	cmd := exec.Command("pg_config", "--bindir")
	outb, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(outb))
}

func subprocess(binPath string, args []string) string {
	cmd := exec.Command(binPath, args...)
	outb, err := cmd.CombinedOutput()
	if err != nil {
		log.Panicln("exec failed:", binPath, args, string(outb))
	}
	return string(outb)
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

func fmtOptionString(options ...string) string {
	return strings.Join(options, " ")
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

	pgctl := filepath.Join(pgBinDir, "pg_ctl")

	subprocess(pgctl, []string{"initdb", "-o", initdbOptions, "--pgdata", pgDataDir})

	port, err := getFreePort()
	if err != nil {
		panic(err)
	}

	fmt.Println("port=", port)

	pghost := "postgres"
	pgdatabase := "postgres"

	pgLogPath := filepath.Join(workDir, "postgres.log")

	// refer: https://www.postgresql.org/docs/current/static/libpq-envars.html
	os.Setenv("PGHOST", pghost)
	os.Setenv("PGDATABASE", pgdatabase)
	os.Setenv("PGPORT", strconv.Itoa(port))

	postgresOptions := fmtOptionString(
		"-i", "-h", "localhost", "-p", strconv.Itoa(port),
		"-k", pgSockDir,
		"-F",
	)

	fmt.Println("postgresOptions=", postgresOptions)
	subprocess(pgctl, []string{"start", "-w", "--log", pgLogPath, "-o", postgresOptions, "-D", pgDataDir})
	defer func() {
		subprocess(pgctl, []string{"stop", "-D", pgDataDir})
	}()

	pgurl := fmt.Sprintf("postgresql://localhost:%d/%s", port, pgdatabase)
	os.Setenv("PGURL", pgurl)

	fmt.Println(subprocess("sh", []string{"-c", fmtOptionString("psql", "--dbname", "$PGURL", "-c", "\"select now();\"")}))

}
