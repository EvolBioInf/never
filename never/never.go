package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/evolbioinf/clio"
	"github.com/evolbioinf/neighbors/tdb"
	"github.com/evolbioinf/never/util"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type PageData struct {
	Title, URL string
	Functions  []TableRow
	Programs   []TableRow
}
type TableRow struct {
	Name, Query string
}
type TaxiOut struct {
	Id     int    `json:"id"`
	Parent int    `json:"parent"`
	Name   string `json:"name"`
}
type AccessionsOut struct {
	Accession string `json:"accession"`
}

var host, port string
var neidb *tdb.TaxonomyDB
var functions, programs []TableRow
var templates = template.New("templates")
var templateFuncs = make(template.FuncMap)

func index(w http.ResponseWriter, r *http.Request, p *PageData) {
	p.Title = "Neighbors"
	p.Functions = functions
	p.Programs = programs
	err := templates.ExecuteTemplate(w, "index", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func init() {
	query := "?t=Homo+sapiens&e=1"
	row := TableRow{Name: "taxi", Query: query}
	programs = append(programs, row)
	query = "?t=9606"
	row = TableRow{Name: "accessions",
		Query: query}
	functions = append(functions, row)
}
func inc(i int) int {
	return i + 1
}
func init() {
	templateFuncs["inc"] = inc
	templates = templates.Funcs(templateFuncs)
	path := "./static/templates.html"
	templates = template.Must(templates.ParseFiles(path))
}
func makeHandler(fn func(http.ResponseWriter, *http.Request,
	*PageData)) http.HandlerFunc {
	p := new(PageData)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fn(w, r, p)
	}
}
func taxi(w http.ResponseWriter, r *http.Request, p *PageData) {
	name := r.URL.Query().Get("t")
	sstr := r.URL.Query().Get("e")
	if sstr != "1" && len(name) > 0 {
		name = strings.ReplaceAll(name, " ", "%")
		name = "%" + name + "%"
	}
	ids := neidb.Taxids(name)
	out := []TaxiOut{}
	for _, id := range ids {
		sciName := neidb.Name(id)
		parent := neidb.Parent(id)
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
func accessions(w http.ResponseWriter, r *http.Request, p *PageData) {
	t := r.URL.Query().Get("t")
	n, err := strconv.Atoi(t)
	util.Check(err)
	out := []AccessionsOut{}
	accs := neidb.Accessions(n)
	for _, acc := range accs {
		o := AccessionsOut{acc}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	fmt.Fprintf(w, "%s", string(b))
}
func main() {
	util.PrepLog("never")
	flagV := flag.Bool("v", false, "version")
	flagO := flag.String("o", "localhost", "host")
	flagP := flag.String("p", "443", "port")
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
	host = *flagO
	port = *flagP
	neidb = tdb.OpenTaxonomyDB(*flagD)
	date, err := os.ReadFile(*flagU)
	util.Check(err)
	tmpFields := bytes.Fields(date)
	if len(tmpFields) != 7 {
		log.Fatalf("%q doesn't look like a date",
			string(date))
	}
	files := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", files))
	http.HandleFunc("/", makeHandler(index))
	http.HandleFunc("/taxi/", makeHandler(taxi))
	http.HandleFunc("/accessions/", makeHandler(accessions))
	host := *flagO + ":" + *flagP
	if *flagC != "" && *flagK != "" {
		log.Fatal(http.ListenAndServeTLS(host, *flagC,
			*flagK, nil))
	} else {
		log.Fatal(http.ListenAndServe(host, nil))
	}

}
