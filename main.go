package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/shlex"
)

func main() {
	dirName := filepath.Dir(os.Args[0])
	if dirName == "." {
		if exeFile, err := os.Executable(); err != nil {
			log.Fatal(fmt.Errorf("get executable failed: %w", err))
		} else {
			dirName = filepath.Dir(exeFile)
		}
	} else if !filepath.IsAbs(dirName) {
		if dirAbs, err := filepath.Abs(dirName); err != nil {
			log.Fatal(fmt.Errorf("parse dirname failed: %w", err))
		} else {
			dirName = dirAbs
		}
	}

	exeName := filepath.Base(os.Args[0])
	if exeReal, err := filepath.EvalSymlinks(filepath.Join(dirName, exeName)); err != nil {
		log.Fatal(fmt.Errorf("parse symlinks failed: %w", err))
	} else if exeName == exeReal {
		log.Println("Usage:")
		log.Println("ln -s exex docker")
		log.Println("echo \"-H ssh://test\" > docker.args")
		log.Println("./docker info")
		os.Exit(1)
	}

	exeArgs := os.Args[1:]

	if bs, err := ioutil.ReadFile(filepath.Join(dirName, fmt.Sprintf("%s.args", exeName))); err == nil {
		if args, err := shlex.Split(string(bs)); err != nil {
			log.Fatal(fmt.Errorf("parse extend args failed: %w", err))
		} else {
			exeArgs = append(args, exeArgs...)
		}
	}

	if bs, err := ioutil.ReadFile(filepath.Join(dirName, fmt.Sprintf("%s.alias", exeName))); err == nil {
		exeName = strings.TrimSpace(string(bs))
	}

	cmd := exec.Command(exeName, exeArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(fmt.Errorf("get stdin pipe failed: %w", err))
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(fmt.Errorf("command init failed: %w", err))
	}

	doClose := false

	go func() {
		if _, err := io.Copy(stdin, os.Stdin); err != nil {
			log.Fatal(fmt.Errorf("stdin copy failed: %w", err))
		}
		_ = stdin.Close()
		// try close process by send signal if stdin close not working
		<-time.After(3 * time.Second)
		doClose = true
		_ = cmd.Process.Signal(syscall.SIGHUP)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)
	go func() {
		sigRecv := <-sigChan
		doClose = true
		if err := cmd.Process.Signal(sigRecv); err != nil {
			log.Println(fmt.Errorf("send %s to process failed: %w", sigRecv, err))
		}
	}()

	if err := cmd.Wait(); err != nil && !doClose {
		log.Fatal(fmt.Errorf("command run failed: %w", err))
	}
}
