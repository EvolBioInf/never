// The package clui helps building command-line user interfaces.
package clui

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
)

type ParseFunc func(r io.Reader, args ...any)

// Usage sets the usage statement.
func Usage(programDescription, programExampleArguments, programExampleText string) {
	var programName = path.Base(os.Args[0])

	var usage = fmt.Sprintf("Usage for: %s\n", programName)
	usage += fmt.Sprintln(programDescription)
	usage += fmt.Sprintf("Example: ./%s %s\n", programName, programExampleArguments)
	usage += fmt.Sprintln(programExampleText)
	fmt.Fprintf(flag.CommandLine.Output(), "%sFlags:\n", usage)
	flag.PrintDefaults()
}

func Info(version, date string) {
	const author = "Ben Bahnsen"
	const contact = "github.com/BenBahnsen"
	const license = "BSD-3"

	var programName = path.Base(os.Args[0])

	var info = fmt.Sprintln(programName, version)
	info += fmt.Sprintf("Last modified: %s\n", date)
	info += fmt.Sprintf("Author: %s\n", author)
	info += fmt.Sprintf("Contact: %s\n", contact)
	info += fmt.Sprintf("Licensed under: %s\n", license)
	fmt.Println(info)
	os.Exit(0)
}
