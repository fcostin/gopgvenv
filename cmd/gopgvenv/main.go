package main

import "fmt"
import "os/exec"
import "strings"

func getPGBinDir() string {
	cmd := exec.Command("pg_config", "--bindir")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(out))
}

func main() {
	pgBinDir := getPGBinDir()
	fmt.Println("pg_bindir=", pgBinDir)
}
