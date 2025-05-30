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
	"slices"
	"strconv"
	"strings"
)

type PageData struct {
	Title, URL, Date string
	Services         []Service
	Ngenomes, Ntaxa  int
}
type Service struct {
	Name, Query string
}
type TaxiOut struct {
	Taxid  int    `json:"taxid"`
	Parent int    `json:"parent"`
	Name   string `json:"name"`
}
type Accession struct {
	Accession string `json:"accession"`
}
type Name struct {
	Taxid int    `json:"taxid"`
	Name  string `json:"name"`
}
type Rank struct {
	Taxid int    `json:"taxid"`
	Rank  string `json:"rank"`
}
type Taxid struct {
	Taxid int `json:"taxid"`
}
type Level struct {
	Level string `json:"level"`
}

var host, port string
var neidb *tdb.TaxonomyDB
var dateFile string
var services []Service
var templates = template.New("templates")
var templateFuncs = make(template.FuncMap)

func index(w http.ResponseWriter, r *http.Request, p *PageData) {
	p.Title = "Neighbors"
	p.Services = services
	slices.SortFunc(p.Services, func(a, b Service) int {
		return strings.Compare(a.Name, b.Name)
	})
	p.Ntaxa = neidb.NumTaxa()
	p.Ngenomes = neidb.NumGenomes()
	date, err := os.ReadFile(dateFile)
	util.CheckHTTP(w, err)
	fields := strings.Fields(string(date))
	p.Date = fmt.Sprintf("%s %s %s at %s %s %s",
		fields[1],
		fields[2],
		fields[6],
		fields[3],
		fields[4],
		fields[5])

	err = templates.ExecuteTemplate(w, "index", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func init() {
	query := "?t=Homo+sapiens&e=1"
	service := Service{Name: "taxi", Query: query}
	services = append(services, service)
	query = "?t=9606"
	service = Service{Name: "accessions",
		Query: query}
	services = append(services, service)
	queries := query + ",9605"
	service = Service{Name: "names",
		Query: queries}
	services = append(services, service)
	service = Service{Name: "ranks",
		Query: queries}
	services = append(services, service)
	service = Service{Name: "parent",
		Query: query}
	services = append(services, service)
	service = Service{Name: "children",
		Query: query}
	services = append(services, service)
	query = "?t=Homo+sapiens"
	service = Service{Name: "taxids",
		Query: query}
	services = append(services, service)

	query = "?t=9606,741158,63221"
	service = Service{Name: "mrca",
		Query: query}
	services = append(services, service)
	query = "?a=GCF_000001405.40,GCA_000002115.2"
	service = Service{Name: "levels",
		Query: query}
	services = append(services, service)
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
			Taxid:  id,
			Parent: parent,
			Name:   sciName}
		out = append(out, tout)
	}
	b, err := json.Marshal(out)
	util.Check(err)
	fmt.Fprintf(w, "%s", string(b))
}
func accessions(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxid := getTaxa(w, r)[0]
	out := []Accession{}
	accs := neidb.Accessions(taxid)
	for _, acc := range accs {
		o := Accession{acc}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func getTaxa(w http.ResponseWriter, r *http.Request) []int {
	taxa := []int{}
	t := r.URL.Query().Get("t")
	tokes := strings.Split(t, ",")
	for _, token := range tokes {
		taxon, err := strconv.Atoi(token)
		util.CheckHTTP(w, err)
		taxa = append(taxa, taxon)
	}
	return taxa
}
func names(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxa := getTaxa(w, r)
	names := neidb.Names(taxa)
	out := []Name{}
	for i, name := range names {
		o := Name{Taxid: taxa[i], Name: name}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func ranks(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxa := getTaxa(w, r)
	ranks := neidb.Ranks(taxa)
	out := []Rank{}
	for i, rank := range ranks {
		o := Rank{Taxid: taxa[i], Rank: rank}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func parent(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxon := getTaxa(w, r)[0]
	parent := neidb.Parent(taxon)
	out := Taxid{parent}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func children(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxon := getTaxa(w, r)[0]
	out := []Taxid{}
	children := neidb.Children(taxon)
	for _, child := range children {
		o := Taxid{child}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func subtree(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxon := getTaxa(w, r)[0]
	taxids := neidb.Subtree(taxon)
	out := []Taxid{}
	for _, taxid := range taxids {
		o := Taxid{taxid}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func taxids(w http.ResponseWriter, r *http.Request, p *PageData) {
	name := r.URL.Query().Get("t")
	out := []Taxid{}
	taxids := neidb.Taxids(name)
	for _, taxid := range taxids {
		o := Taxid{taxid}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func mrca(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxa := getTaxa(w, r)
	mrca := neidb.MRCA(taxa)
	out := Taxid{mrca}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func levels(w http.ResponseWriter, r *http.Request, p *PageData) {
	accessions := getAccessions(w, r)
	levels := neidb.Levels(accessions)
	out := []Level{}
	for _, level := range levels {
		o := Level{level}
		out = append(out, o)
	}
	b, err := json.Marshal(out)
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s", string(b))
}
func getAccessions(w http.ResponseWriter, r *http.Request) []string {
	accessions := []string{}
	a := r.URL.Query().Get("a")
	accs := strings.Split(a, ",")
	for _, accession := range accs {
		accessions = append(accessions, accession)
	}
	return accessions
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
	dateFile = *flagU
	files := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", files))
	http.HandleFunc("/", makeHandler(index))
	http.HandleFunc("/taxi/", makeHandler(taxi))
	http.HandleFunc("/accessions/", makeHandler(accessions))
	http.HandleFunc("/names/", makeHandler(names))
	http.HandleFunc("/ranks/", makeHandler(ranks))
	http.HandleFunc("/parent/", makeHandler(parent))
	http.HandleFunc("/children/", makeHandler(children))
	http.HandleFunc("/subtree/", makeHandler(subtree))
	http.HandleFunc("/taxids/", makeHandler(taxids))
	http.HandleFunc("/mrca/", makeHandler(mrca))
	http.HandleFunc("/levels/", makeHandler(levels))
	host := *flagO + ":" + *flagP
	if *flagC != "" && *flagK != "" {
		log.Fatal(http.ListenAndServeTLS(host, *flagC,
			*flagK, nil))
	} else {
		log.Fatal(http.ListenAndServe(host, nil))
	}
}
