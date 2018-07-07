package main

import "fmt"
import "io/ioutil"
import "log"
import "os"
import "os/exec"
import "path/filepath"
import "strings"

func getPGBinDir() string {
	cmd := exec.Command("pg_config", "--bindir")
	outb, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(outb))
}

func pgInitDb(pgBinDir string, pgDataDir string) {
	binPath := filepath.Join(pgBinDir, "pg_ctl")
	args := []string{"initdb", "--pgdata", pgDataDir}
	cmd := exec.Command(binPath, args...)
	outb, err := cmd.CombinedOutput()
	if err != nil {
		log.Panicln("exec failed:", binPath, args, string(outb))
	}
}

func main() {
	pgBinDir := getPGBinDir()

	fmt.Println("pg_bindir=", pgBinDir)

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

	pgInitDb(pgBinDir, pgDataDir)
}
