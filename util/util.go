// Package util provides auxiliary functions for the never package.
package util

import (
	"github.com/evolbioinf/clio"
	"log"
	"net/http"
	"os"
)

var program string
var date, version string

// Check takes an error as argument and logs a fatal error if the error isn't nil.
func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// CheckHTTP takes as arguments a HTTP respose writer and an eror. If the error is not nil, it is printed, unless it corresponds to one of the two standard messages that crop up in never, in which case the error is ignored.
func CheckHTTP(w http.ResponseWriter, err error) {
	m1 := "sql: Rows are closed"
	m2 := "Empty ID list in tdb.MRCA"
	if err != nil && err.Error() != m1 &&
		err.Error() != m2 {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
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
