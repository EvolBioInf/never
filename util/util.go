// Package util provides auxiliary functions for the never package.
package util

import (
	"bytes"
	"fmt"
	"github.com/evolbioinf/clio"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var program string
var date, version string

// Check takes an error as argument. If the error isn't nil, it is printed to the standard error stream.
func Check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR[never]: %s\n", err)
	}
}

// CheckHTTP takes as arguments a HTTP respose writer and an eror. If the error is not nil, it is printed, unless it corresponds to one of the two standard messages that crop up in never, in which case the error is ignored.
func CheckHTTP(w http.ResponseWriter, err error) {
	m1 := "sql: Rows closed"
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
	authors := "Bernhard Haubold,Ben Bahnsen"
	email := "haubold@evolbio.mpg.de,bahnsen@evolbio.mpg.de"
	license := "Gnu General Public License, " +
		"https://www.gnu.org/licenses/gpl.html"
	clio.PrintInfo(program, version, date,
		authors, email, license)
	os.Exit(0)
}

// The function SendGetRequest takes as argument an address as a string and the program's options and extra arguments as string slices, as well as miscellaneous arguments as a map. It sends a get request using these values and returns the result.
func SendGetRequest(address string, options, extraArgs []string, miscArgs map[string]string) string {
	qb := new(strings.Builder)
	urlEncodeSlice(qb, options, "options")
	urlEncodeSlice(qb, extraArgs, "extra")
	for m, a := range miscArgs {
		if qb.Len() == 0 {
			qb.WriteRune('?')
		} else {
			qb.WriteRune('&')
		}
		qb.WriteString(url.QueryEscape(m))
		qb.WriteRune('=')
		qb.WriteString(url.QueryEscape(a))
	}
	req, err := http.NewRequest(http.MethodGet, address+qb.String(), nil)
	Check(err)
	resp, err := http.DefaultClient.Do(req)
	Check(err)
	body, err := io.ReadAll(resp.Body)
	Check(err)
	return string(body)
}
func urlEncodeSlice(qb *strings.Builder, slc []string, paramName string) {
	for _, v := range slc {
		if qb.Len() == 0 {
			qb.WriteRune('?')
		} else {
			qb.WriteRune('&')
		}
		qb.WriteString(paramName)
		qb.WriteRune('=')
		qb.WriteString(url.QueryEscape(v))
	}
}

// The function SendQueryRequest takes as argument an address as a string, program options and extra arguments as a slice of strings, as well as files and stdin. It sends a query request using these values and returns the result.
func SendQueryRequest(
	address string,
	options, extraArgs []string,
	miscArgs map[string]string,
	files []*os.File,
	stdin *os.File,
) string {
	qb := new(strings.Builder)
	urlEncodeSlice(qb, options, "options")
	urlEncodeSlice(qb, extraArgs, "extra")
	for m, a := range miscArgs {
		if qb.Len() == 0 {
			qb.WriteRune('?')
		} else {
			qb.WriteRune('&')
		}
		qb.WriteString(url.QueryEscape(m))
		qb.WriteRune('=')
		qb.WriteString(url.QueryEscape(a))
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for i, file := range files {
		fw, err := w.CreateFormFile(strconv.Itoa(i), file.Name())
		Check(err)
		_, err = io.Copy(fw, file)
		Check(err)
	}
	if stdin != nil {
		fw, err := w.CreateFormFile("stdin", "stdin")
		Check(err)
		_, err = io.Copy(fw, stdin)
		Check(err)
	}
	w.Close()
	req, err := http.NewRequest("QUERY", address+qb.String(), &b)
	Check(err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	Check(err)
	body, err := io.ReadAll(resp.Body)
	Check(err)
	return string(body)
}
