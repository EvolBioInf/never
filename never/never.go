package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
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
	Services []Service
	Title    string
	Ntaxa    string
	Ngenomes string
	Date     string
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
type Node struct {
	Taxid  int `json:"taxid"`
	Parent int `json:"parent"`
}
type Level struct {
	Accession string `json:"accession"`
	Level     string `json:"level"`
}
type GenomeCount struct {
	Count int    `json:"count"`
	Level string `json:"level"`
}

var host, port string
var neidb *tdb.TaxonomyDB
var dateFile string
var services []Service
var templates = template.New("templates")
var templateFuncs = make(template.FuncMap)
var assemblyLevels = []string{
	"complete",
	"chromosome",
	"scaffold",
	"contig"}

func index(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	p.Title = "Neighbors"
	p.Services = services
	slices.SortFunc(p.Services, func(a, b Service) int {
		return strings.Compare(a.Name, b.Name)
	})
	nt, err := neidb.NumTaxa()
	util.CheckHTTP(w, err)
	p.Ntaxa = humanize.Comma(int64(nt))
	n1, err := neidb.NumGenomesRec(1, "complete")
	util.CheckHTTP(w, err)
	n2, err := neidb.NumGenomesRec(1, "chromosome")
	util.CheckHTTP(w, err)
	n3, err := neidb.NumGenomesRec(1, "scaffold")
	util.CheckHTTP(w, err)
	n4, err := neidb.NumGenomesRec(1, "contig")
	util.CheckHTTP(w, err)
	ng := n1 + n2 + n3 + n4
	p.Ngenomes = humanize.Comma(int64(ng))
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
	util.CheckHTTP(w, err)
}
func init() {
	var service Service
	var query string
	query = "?t=E&n=10&p=2"
	service = Service{Name: "taxi", Query: query}
	services = append(services, service)
	query = "?t=9606"
	service = Service{Name: "accessions",
		Query: query}
	services = append(services, service)
	query = "?t=9606,9605"
	service = Service{Name: "names",
		Query: query}
	services = append(services, service)
	query = "?t=9606,9605"
	service = Service{Name: "ranks",
		Query: query}
	services = append(services, service)
	query = "?t=9606"
	service = Service{Name: "parent",
		Query: query}
	services = append(services, service)
	query = "?t=9606"
	service = Service{Name: "children",
		Query: query}
	services = append(services, service)
	query = "?t=9606"
	service = Service{Name: "subtree",
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
	query = "?t=562"
	service = Service{Name: "num_genomes",
		Query: query}
	services = append(services, service)
	query = "?t=562"
	service = Service{Name: "num_genomes_rec",
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
		w.Header().Set("Access-Control-Allow-Origin",
			"*")
		fn(w, r, p)
	}
}
func taxi(w http.ResponseWriter, r *http.Request, p *PageData) {
	name := r.URL.Query().Get("t")
	sstr := r.URL.Query().Get("e")
	page := r.URL.Query().Get("p")
	size := r.URL.Query().Get("n")
	if sstr != "1" && len(name) > 0 {
		name = strings.ReplaceAll(name, " ", "% %")
		name = "%" + name + "%"
	}
	var limit, offset int
	limit, err := strconv.Atoi(size)
	util.CheckHTTP(w, err)
	pageNum, err := strconv.Atoi(page)
	util.CheckHTTP(w, err)
	offset = (pageNum - 1) * limit
	ids, err := neidb.Taxids(name, limit, offset)
	util.CheckHTTP(w, err)
	out := []TaxiOut{}
	for _, id := range ids {
		sciName, err := neidb.Name(id)
		util.CheckHTTP(w, err)
		parent, err := neidb.Parent(id)
		util.CheckHTTP(w, err)
		tout := TaxiOut{
			Taxid:  id,
			Parent: parent,
			Name:   sciName}
		out = append(out, tout)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func accessions(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxid := getTaxa(w, r)[0]
	out := []Accession{}
	accs, err := neidb.Accessions(taxid)
	util.CheckHTTP(w, err)
	for _, acc := range accs {
		o := Accession{acc}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
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
func names(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	out := []Name{}
	for i, taxon := range taxa {
		name, err := neidb.Name(taxon)
		util.CheckHTTP(w, err)
		o := Name{Taxid: taxa[i], Name: name}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func ranks(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	out := []Rank{}
	for i, taxon := range taxa {
		rank, err := neidb.Rank(taxon)
		util.CheckHTTP(w, err)
		o := Rank{Taxid: taxa[i], Rank: rank}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func parent(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxon := getTaxa(w, r)[0]
	parent, err := neidb.Parent(taxon)
	util.CheckHTTP(w, err)
	out := Taxid{parent}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func children(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxon := getTaxa(w, r)[0]
	out := []Taxid{}
	children, err := neidb.Children(taxon)
	util.CheckHTTP(w, err)
	for _, child := range children {
		o := Taxid{child}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func subtree(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxon := getTaxa(w, r)[0]
	taxids, err := neidb.Subtree(taxon)
	util.CheckHTTP(w, err)
	out := []Node{}
	for _, taxid := range taxids {
		p := taxon
		if taxid != taxon {
			p, err = neidb.Parent(taxid)
		}
		util.CheckHTTP(w, err)
		o := Node{Taxid: taxid, Parent: p}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func taxids(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	name := r.URL.Query().Get("t")
	out := []Taxid{}
	taxids, err := neidb.Taxids(name, -1, 0)
	util.CheckHTTP(w, err)
	for _, taxid := range taxids {
		o := Taxid{taxid}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func mrca(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxa := getTaxa(w, r)
	mrca, err := neidb.MRCA(taxa)
	util.CheckHTTP(w, err)
	out := Taxid{mrca}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func levels(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	str := r.URL.Query().Get("a")
	accessions := strings.Split(str, ",")
	out := []Level{}
	for _, accession := range accessions {
		level, err := neidb.Level(accession)
		util.CheckHTTP(w, err)
		o := Level{Accession: accession, Level: level}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func num_genomes(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxid := getTaxa(w, r)[0]
	out := []GenomeCount{}
	for _, level := range assemblyLevels {
		n, err := neidb.NumGenomes(taxid, level)
		util.CheckHTTP(w, err)
		o := GenomeCount{Count: n, Level: level}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func num_genomes_rec(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxid := getTaxa(w, r)[0]
	out := []GenomeCount{}
	for _, level := range assemblyLevels {
		n, err := neidb.NumGenomesRec(taxid, level)
		util.CheckHTTP(w, err)
		o := GenomeCount{Count: n, Level: level}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.CheckHTTP(w, err)
	fmt.Fprintf(w, "%s\n", string(b))
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
	e := "never -o 10.254.1.21 -c Cert_bundle.pem -k privateKey.pem"
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
	http.HandleFunc("/num_genomes/",
		makeHandler(num_genomes))
	http.HandleFunc("/num_genomes_rec/", makeHandler(num_genomes_rec))
	host := *flagO + ":" + *flagP
	if *flagC != "" && *flagK != "" {
		log.Fatal(http.ListenAndServeTLS(host, *flagC,
			*flagK, nil))
	} else {
		log.Fatal(http.ListenAndServe(host, nil))
	}
}
