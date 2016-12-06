package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	errEthminer       = errors.New("ethminer: error occured")
	errTooMuchFailure = errors.New("ethminer: too much failure")
	config            Config
)

type Config struct {
	EthminerBin   string            `json:"ethminer_bin"`
	Env           map[string]string `json:"env"`
	EthminerArgs  string            `json:"ethminer_args"`
	MaxErrorCount int               `json:"max_error_count"`
	StartDelay    int               `json:"start_delay_sec"`
	EmailNotif    string            `json:"email_notif"`
	EmailPass     string            `json:"email_pass"`
}

func readConfig(c string) (err error) {
	log.Println(CharArrow+"Reading config from", c)

	cfile, err := ioutil.ReadFile(c)
	if err != nil {
		log.Println("Reading config file failed")
		return
	}

	if err = json.Unmarshal(cfile, &config); err != nil {
		log.Println("Unmarshal config file failed")
		return
	}

	return
}

func monitorEthminer() (err error) {
	globalErrorCount := 0

	for {
		log.Printf("%s Waiting %ds before starting miner...", CharStar, config.StartDelay)
		for i := 0; i < config.StartDelay; i++ {
			log.Printf("Starting in %ds", config.StartDelay-i)
			time.Sleep(time.Second * 1)
		}

		log.Println(CharArrow + "Start miner")
		sendEmail("[MINING] - Start miner", fmt.Sprintf("Miner is going to be started attempt: #%d", globalErrorCount))

		err = runMiner()
		if err == errEthminer {
			globalErrorCount++
		}

		if globalErrorCount > config.MaxErrorCount {
			return errTooMuchFailure
		}
	}
}

func runMiner() (err error) {
	errorCount = 0

	cmd := exec.Command(config.EthminerBin, strings.Split(config.EthminerArgs, " ")...)
	env := os.Environ()
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	cmd.Env = env

	stdout, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	//	stderr, err := cmd.StderrPipe()
	//	if err != nil {
	//		return err
	//	}

	//	multi := io.MultiReader(stdout, stderr)

	// start the command after having set up the pipe
	if err = cmd.Start(); err != nil {
		return err
	}

	// read command's stdout line by line
	in := bufio.NewScanner(stdout)

	for in.Scan() {
		ln := in.Text()

		//write line to stdout
		fmt.Println(ln)

		//and process the line for info
		err = processLine(ln)
		if err != nil {
			//something failed in parsing line
			if err == errEthminer {
				cmd.Process.Kill()
				return
			}
		}
	}
	if err = in.Err(); err != nil {
		fmt.Printf("error: %s", err)
	}

	//if here miner has exited. this is not good
	cmd.Wait()

	return errEthminer
}

const (
	ReCatchLog = `\s*(\S+)\s+([0-9:]+).(\w+)\s+(.*)`
)

var (
	RegLine = regexp.MustCompile(ReCatchLog)

	errorCount = 0
)

type LogLine struct {
	Type      string
	Time      string
	Namespace string
	Message   string
}

func parseLine(s string) (l *LogLine) {
	s = strings.TrimSpace(s)
	t := RegLine.FindStringSubmatch(s)
	l = new(LogLine)

	if len(t) < 4 {
		l.Type = "-"
		l.Message = "Parse failed: " + s
	} else {
		l.Type = t[1]
		l.Time = t[2]
		l.Namespace = t[3]
		l.Message = t[4]
	}

	return
}

func processLine(ln string) (err error) {
	l := parseLine(ln)

	//Error occured
	if l.Type == "X" {
		errorCount++
		fmt.Printf("type is %s errorCount is %d\n", l.Type, errorCount)
	}

	//When too much error, stop miner
	if errorCount > config.MaxErrorCount {
		sendEmail("[MINING] - Miner failure!", fmt.Sprintf("Miner has failed with:\n%s", l.Message))
		err = errEthminer
	}

	return
}

/*
done := make(chan error, 1)
go func() {
    done <- cmd.Wait()
}()
select {
case <-time.After(3 * time.Second):
    if err := cmd.Process.Kill(); err != nil {
        log.Fatal("failed to kill: ", err)
    }
    log.Println("process killed as timeout reached")
case err := <-done:
    if err != nil {
        log.Printf("process done with error = %v", err)
    } else {
        log.Print("process done gracefully without error")
    }
}


*/
