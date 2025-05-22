package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/evolbioinf/clio"
	"github.com/evolbioinf/neighbors/tdb"
	"github.com/evolbioinf/never/util"
	"log"
	"net/http"
	"os"
	"strings"
)

type TaxiOut struct {
	Id     int    `json:"id"`
	Parent int    `json:"parent"`
	Name   string `json:"name"`
}
type MyDB struct {
	db *tdb.TaxonomyDB
}

func (m MyDB) taxi(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("t")
	sstr := r.URL.Query().Get("s")
	if sstr == "1" && len(name) > 0 {
		name = strings.ReplaceAll(name, " ", "%")
		name = "%" + name + "%"
	}
	ids := m.db.Taxids(name)
	out := []TaxiOut{}
	for _, id := range ids {
		sciName := m.db.Name(id)
		parent := m.db.Parent(id)
		tout := TaxiOut{
			Id:     id,
			Parent: parent,
			Name:   sciName}
		out = append(out, tout)
	}
	b, err := json.Marshal(out)
	util.Check(err)
	fmt.Fprintf(w, "%s", string(b))
}
func main() {
	util.PrepLog("never")
	flagV := flag.Bool("v", false, "version")
	flagO := flag.String("o", "localhost", "host")
	flagP := flag.String("p", "80", "port")
	flagC := flag.String("c", "", "certificate")
	flagK := flag.String("k", "", "private key")
	flagD := flag.String("d", "neidb", "database")
	flagU := flag.String("u", "updated.txt", "last updated")
	u := "never [flag]..."
	p := "The program never is a web server " +
		"providing a REST API for the Neighbors package."
	e := "never -o 10.254.1.21 -c Cert.pem -k privateKey.pem"
	clio.Usage(u, p, e)
	flag.Parse()
	if *flagV {
		util.PrintInfo()
	}
	db := tdb.OpenTaxonomyDB(*flagD)
	date, err := os.ReadFile(*flagU)
	util.Check(err)
	tmpFields := bytes.Fields(date)
	if len(tmpFields) != 7 {
		log.Fatalf("%q doesn't look like a date",
			string(date))
	}
	var myDB MyDB
	myDB.db = db
	http.HandleFunc("/taxi/", myDB.taxi)
	fmt.Println("TO DO: Handle calls to tdb functions")
	host := *flagO + ":" + *flagP
	if *flagC != "" && *flagK != "" {
		log.Fatal(http.ListenAndServeTLS(host, *flagC,
			*flagK, nil))
	} else {
		log.Fatal(http.ListenAndServe(host, nil))
	}

}
