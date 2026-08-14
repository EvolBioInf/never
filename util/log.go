package util

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"

	"strconv"
)

type Writable interface {
	Keys() []string
	Values() []string
}

type WarningEntry struct {
	timestamp,
	Warning string
}

type InfoEntry struct {
	timestamp,
	RequestIp,
	RequestUrl,
	RequestMethod string
	ResponseCode int
	ResponseSize,
	ResponseTime,
	Description string
}

var infoBuffer, warningBuffer *csv.Writer
var infoFile, warningFile *os.File
var infoMutex, warningMutex sync.Mutex
var timer *time.Timer
var interval time.Duration

func SetupLog() {
	dir := "logs/"
	interval = 2 * time.Second
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(fmt.Sprintf("Could not create directory for logs: %s\n", err))
	}

	infoFile, infoBuffer = handleFileOpen(dir+"info.csv", InfoEntry{})
	warningFile, warningBuffer = handleFileOpen(
		dir+"warning.csv",
		WarningEntry{},
	)

	LogInfoDef(InfoEntry{Description: "Setup log"})
}

func handleFileOpen(path string, entry Writable) (*os.File, *csv.Writer) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(fmt.Sprintf("Could not open %s for logs: %s\n", path, err))
	}

	buffer := csv.NewWriter(f)
	fi, err := os.Stat(path)
	if err != nil {
		panic(fmt.Sprintf("Could not read stats of %s: %s\n", path, err))
	}

	if fi.Size() == 0 {
		buffer.Write(entry.Keys())
	}

	return f, buffer
}

func (e WarningEntry) Keys() []string {
	return []string{"timestamp", "warning"}
}

func (e WarningEntry) Values() []string {
	return []string{e.timestamp, e.Warning}
}

func (e InfoEntry) Keys() []string {
	return []string{
		"timestamp",
		"request_ip",
		"request_url",
		"request_method",
		"response_code",
		"response_size",
		"response_time",
		"description",
	}
}

func (e InfoEntry) Values() []string {
	code := ""
	if e.ResponseCode != 0 {
		code = strconv.Itoa(e.ResponseCode)
	}

	return []string{
		e.timestamp,
		e.RequestIp,
		e.RequestUrl,
		e.RequestMethod,
		code,
		e.ResponseSize,
		e.ResponseTime,
		e.Description,
	}
}

func setTimer() {
	if timer == nil {
		timer = time.AfterFunc(interval, logBuffs)
	}
}

func LogInfoDef(e InfoEntry) {
	e.timestamp = time.Now().Format(time.DateTime)
	infoMutex.Lock()
	infoBuffer.Write(e.Values())
	infoMutex.Unlock()
	setTimer()
}

func LogWarningDef(e WarningEntry) {
	e.timestamp = time.Now().Format(time.DateTime)
	warningMutex.Lock()
	warningBuffer.Write(e.Values())
	warningMutex.Unlock()
	setTimer()
}

func logBuffs() {
	infoMutex.Lock()
	infoBuffer.Flush()
	infoMutex.Unlock()

	warningMutex.Lock()
	warningBuffer.Flush()
	warningMutex.Unlock()

	timer = nil
}

func StopLogging() {
	LogInfoDef(InfoEntry{Description: "Stopped logging"})
	if timer != nil {
		timer.Stop()
		timer = nil
	}

	logBuffs()
	infoFile.Close()
	warningFile.Close()
}
