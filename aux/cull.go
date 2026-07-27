package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"slices"
	"strings"
)

func main() {
	dbDir := "../never/databases/"

	// Read directory
	entries, err := os.ReadDir(dbDir)
	if err != nil {
		log.Fatal("error while trying to read database directory", err.Error())
	}

	// Get information on files
	var fileInfos []fs.FileInfo
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			log.Fatal("error while trying to read info", err.Error())
		}
		fileInfos = append(fileInfos, info)
	}

	if len(fileInfos) == 0 {
		log.Fatal(errors.New("No files in provided directory."))
	}

	// Sort by modification date
	slices.SortFunc(fileInfos, func(a, b fs.FileInfo) int {
		return b.ModTime().Compare(a.ModTime())
	})

	fmt.Println("Found files:")
	for _, f := range fileInfos {
		fmt.Println(f.Name())
	}

	// Filter by db checking
	fileInfos = slices.DeleteFunc(fileInfos,
		func(f fs.FileInfo) bool { return !isDb(dbDir + f.Name()) })

	// Filter by whitelist
	f, err := os.Open(dbDir + "whitelist")
	if err != nil {
		log.Fatal(errors.New("Could not open db whitelist."))
	}
	defer f.Close()

	fStat, err := f.Stat()
	if err != nil {
		log.Fatal(errors.New("Could not stat db whitelist."))
	}

	buf := make([]byte, fStat.Size())
	f.Read(buf)
	whitelisted := strings.SplitSeq(string(buf), "\n")
	for w := range whitelisted {
		fileInfos = slices.DeleteFunc(fileInfos,
			func(a fs.FileInfo) bool { return a.Name() == w })
	}

	// Remove elements older than the tenth db
	if len(fileInfos) > 12 {
		fileInfos = fileInfos[12:]

		// Remove dbs and creation components
		for _, f := range fileInfos {
			dbName := dbDir + f.Name()
			arcName, _ := strings.CutSuffix(dbName, ".db")
			arcName = arcName + ".tgz"
			fmt.Printf("deleting: %s and %s\n", dbName, arcName)

			os.Remove(dbName)
			os.Remove(arcName)
		}
	}
}

// A file is an sqlite file if it begins with a fixed set of 16 bytes.
func isDb(path string) bool {
	f, err := os.Open(path)

	if err != nil {
		fmt.Println(path + " error?")
		return false
	}

	var buf = make([]byte, 16)
	f.Read(buf)
	return bytes.Equal(buf, []byte{0x53, 0x51, 0x4c, 0x69, 0x74, 0x65, 0x20, 0x66,
		0x6f, 0x72, 0x6d, 0x61, 0x74, 0x20, 0x33, 0x00})
}
