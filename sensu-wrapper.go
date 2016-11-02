package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"os/exec"
	//"strings"
	"syscall"
)

func run_command(cmdName string, cmdArgs []string) (int, string) {

	// the command we're going to run
	cmd := exec.Command(cmdName, cmdArgs...)

	// assign vars for output and stderr
	var output bytes.Buffer
	var stderr bytes.Buffer

	// get the stdout and stderr and assign to pointers
	cmd.Stderr = &stderr
	cmd.Stdout = &output

	// The command never started successfully
	if err := cmd.Start(); err != nil {
		log.Fatalf("Command not found: %s", cmdName)
	}

	// Here's the good stuff
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// Command ! exit 0, capture it
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				// Check it's nagios compliant
				if status.ExitStatus() == 1 || status.ExitStatus() == 2 {
					return status.ExitStatus(), stderr.String()
				} else {
					// If not, force an exit code 2
					return 2, stderr.String()
				}
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}
	// We didn't get captured, continue!
	return 0, output.String()
}

func main() {

	type Output struct {
		Name     string   `json:"name"`
		Status   int      `json:"status"`
		Output   string   `json:"output"`
		Ttl      int      `json:"ttl,omitempty"`
		Source   string   `json:"source,omitempty"`
		Handlers []string `json:"handlers,omitempty"`
	}

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "dry-run, D", Usage: "Output to stdout or not"},
		cli.StringFlag{Name: "name, N", Usage: "The name of the check"},
		cli.IntFlag{Name: "ttl, T", Usage: "The TTL for the check"},
		cli.StringFlag{Name: "source, S", Usage: "The source of the check"},
		cli.StringSliceFlag{Name: "handlers, H", Usage: "The handlers to use for the check"},
	}

	app.Name = "Sensu Wrapper"
	app.Version = "0.1"
	app.Usage = "Execute a command and send the result to a sensu socket"
	app.Authors = []cli.Author{
		cli.Author{
			Name: "Lee Briggs",
		},
	}
	app.Action = func(c *cli.Context) error {

		if !c.IsSet("name") {
			return cli.NewExitError("No Check Name Specified", 128)
		}

		// runs the command args
		status, output := run_command(c.Args().First(), c.Args().Tail())

		sensu_values := &Output{
			Name:     c.String("name"),
			Status:   status,
			Output:   output,
			Ttl:      c.Int("ttl"),
			Source:   c.String("source"),
			Handlers: c.StringSlice("handlers"),
		}

		json, _ := json.Marshal(sensu_values)

		if c.Bool("dry-run") {
			fmt.Println(string(json))
			return nil
		} else {
			fmt.Println("no stdout")
			return nil
		}

	}

	app.Run(os.Args)
}
