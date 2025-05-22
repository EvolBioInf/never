// Package uti provides auxiliary functions for the never  package.
package util

import (
	"github.com/evolbioinf/clio"
	"log"
	"os"
)

var program string
var date, version string

// Check takes an error as argument and logs a fatal error if the  error isn't nil.
func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// PrepLog takes as argument the program name and uses it as  prefix for the log message.
func PrepLog(progName string) {
	log.SetPrefix(progName + ": ")
	program = progName
}

// PrintInfo prints information about the version, the  author(s), and the license of never.
func PrintInfo() {
	authors := "Bernhard Haubold"
	email := "haubold@evolbio.mpg.de"
	license := "Gnu General Public License, " +
		"https://www.gnu.org/licenses/gpl.html"
	clio.PrintInfo(program, version, date,
		authors, email, license)
	os.Exit(0)
}
