package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	for {
		cmd := exec.Command("git", "pull")
		cmd.Dir = "../"
		_, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		cmd = exec.Command("go", "build", "-o", "gobot", ".")
		cmd.Dir = "../"
		_, err = cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		cmd = exec.Command("./gobot")
		cmd.Dir = "../"
		stdout, err := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		if err != nil {
			log.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			log.Fatalf("cmd.Start: %v", err)
		}

		go func() {
			for {
				tmp := make([]byte, 1024)
				_, err := stdout.Read(tmp)
				fmt.Print(string(tmp))
				if err != nil {
					break
				}
			}
		}()

		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		sigQuit := make(chan os.Signal, 1)
		signal.Notify(sigQuit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		select {
		case <-sigQuit:
			cmd.Process.Signal(syscall.SIGTERM)
			log.Print("Shutting down damon...")
			os.Exit(0)

		case err := <-done:
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if exiterr.ExitCode() == 0 {
						log.Print("Exiting...\n")
						os.Exit(0)
					}
					log.Println("Restarting the bot...")
				}
			}
		}
	}
}