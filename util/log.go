package util

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"errors"
)

var timer *time.Timer
var interval time.Duration
var infoFile, warningFile string
var infoWriter, warningWriter bufio.Writer

func SetupLog() {
	dir := "logs/"
	infoFile = dir + "info.txt"
	warningFile = dir + "warning.txt"
	interval = 2 * time.Second

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Could not create directory for logs: %s\n", err))
	}

	if _, err := os.Stat(infoFile); errors.Is(err, os.ErrNotExist) {
		os.Create(infoFile)
	} else if err != nil {
		panic(fmt.Sprintf("Could not create info file for logs: %s\n", err))
	}

	if _, err := os.Stat(warningFile); errors.Is(err, os.ErrNotExist) {
		os.Create(warningFile)
	} else if err != nil {
		panic(fmt.Sprintf("Could not create warning file for logs: %s\n", err))
	}

	file, err := os.Open(infoFile)
	if err != nil {
		panic(fmt.Sprintf("Could not open info file for logs: %s\n", err))
	}

	infoWriter = *bufio.NewWriter(file)

	file, err = os.Open(warningFile)
	if err != nil {
		panic(fmt.Sprintf("Could not open warning file for logs: %s\n", err))
	}

	warningWriter = *bufio.NewWriter(file)

}

func startLogging() {
	if timer == nil {
		timer = time.AfterFunc(interval, logBuffs)
	}
}

func LogInfoDef(str string) {
	fmt.Println("log info def")
	if timer == nil {
		startLogging()
	}

	fmt.Fprintln(&infoWriter, str)
}

func LogWarningDef(str string) {
	fmt.Println("log info def")
	if timer == nil {
		startLogging()
	}

	fmt.Fprintln(&warningWriter, str)
}

func logBuffs() {
	fmt.Println("log immediate")
	if infoWriter.Buffered() > 0 {
		infoWriter.Flush()
	}

	if warningWriter.Buffered() > 0 {
		warningWriter.Flush()
	}

	if timer != nil {
		timer.Reset(interval)
	}
}

func StopLogging() {
	if timer != nil {
		timer.Stop()
		timer = nil
	}

	logBuffs()
}
