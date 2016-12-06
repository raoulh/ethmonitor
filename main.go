// +build go1.7

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"github.com/jawher/mow.cli"
	"github.com/k-takata/go-iscygpty"
	"github.com/mattn/go-isatty"
)

const (
	CharStar     = "\u2737"
	CharAbort    = "\u2718"
	CharCheck    = "\u2714"
	CharWarning  = "\u26A0"
	CharArrow    = "\u2012\u25b6"
	CharVertLine = "\u2502"
)

var (
	blue       = color.New(color.FgBlue).SprintFunc()
	errorRed   = color.New(color.FgRed).SprintFunc()
	errorBgRed = color.New(color.BgRed, color.FgBlack).SprintFunc()
	green      = color.New(color.FgGreen).SprintFunc()
	cyan       = color.New(color.FgCyan).SprintFunc()
	bgCyan     = color.New(color.FgWhite).SprintFunc()
)

var (
	optConfig  *string
	isTerminal bool
)

func exit(err error, exit int) {
	fmt.Fprintln(os.Stderr, errorRed(CharAbort), err)
	cli.Exit(exit)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// fix for cygwin terminal
	if iscygpty.IsCygwinPty(os.Stdout.Fd()) || isatty.IsTerminal(os.Stdout.Fd()) {
		isTerminal = true
	}

	app := cli.App("ethmonitor", "Command line tool to monitor ethminer")

	optConfig = app.StringArg("CONFIG", "ethmonitor.conf", "Config file to use")
	app.Spec = "[CONFIG]"
	app.Action = func() {
		err := readConfig(*optConfig)
		if err != nil {
			exit(err, 1)
		}

		err = monitorEthminer()
		if err != nil {

			sendEmail("[MINING] - Reboot!", fmt.Sprintf("Mining RIG is unstable, reboot is planned"))

			//Force a reboot
			if err := exec.Command("cmd", "/C", "shutdown", "/r", "/f").Run(); err != nil {
				fmt.Println("Failed to initiate shutdown:", err)
			}

			exit(err, 1)
		}
	}

	if err := app.Run(os.Args); err != nil {
		exit(err, 1)
	}
}
