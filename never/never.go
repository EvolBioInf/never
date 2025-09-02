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
type Taxon struct {
	Taxid      int    `json:"taxid"`
	Parent     int    `json:"parent"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
}
type Accession struct {
	Accession string `json:"accession"`
	Level     string `json:"level"`
}
type Name struct {
	Taxid      int    `json:"taxid"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
}
type Rank struct {
	Taxid int    `json:"taxid"`
	Rank  string `json:"rank"`
}
type Taxid struct {
	Taxid int `json:"taxid"`
}
type Child struct {
	Taxid      int    `json:"taxid"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
}
type Node struct {
	Taxid      int    `json:"taxid"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
	Parent     int    `json:"parent"`
}
type Level struct {
	Accession string `json:"accession"`
	Level     string `json:"level"`
}
type GenomeCount struct {
	Level string `json:"level"`
	Count int    `json:"count"`
}
type TaxonInfo struct {
	Taxid      int           `json:"taxid"`
	Parent     int           `json:"parent"`
	IsLeaf     bool          `json:"is_leaf"`
	Name       string        `json:"name"`
	CommonName string        `json:"common_name"`
	Rank       string        `json:"rank"`
	RawCounts  []GenomeCount `json:"raw_genome_counts"`
	RecCounts  []GenomeCount `json:"rec_genome_counts"`
}

var host, port string
var neidb *tdb.TaxonomyDB
var dateFile string
var services []Service
var templates = template.New("templates")
var templateFuncs = make(template.FuncMap)

func index(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	p.Title = "Neighbors"
	p.Services = services
	slices.SortFunc(p.Services, func(a, b Service) int {
		return strings.Compare(a.Name, b.Name)
	})
	nt, err := neidb.NumTaxa()
	util.Check(err)
	p.Ntaxa = humanize.Comma(int64(nt))
	ng := 0
	for _, level := range tdb.AssemblyLevels() {
		n, err := neidb.NumGenomesRec(1, level)
		util.Check(err)
		ng += n
	}
	p.Ngenomes = humanize.Comma(int64(ng))
	date, err := os.ReadFile(dateFile)
	util.Check(err)
	fields := strings.Fields(string(date))
	p.Date = fmt.Sprintf("%s %s %s at %s %s %s",
		fields[1],
		fields[2],
		fields[6],
		fields[3],
		fields[4],
		fields[5])

	err = templates.ExecuteTemplate(w, "index", p)
	util.Check(err)
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
	query = "?t=562,9606"
	service = Service{Name: "taxa_info",
		Query: query}
	services = append(services, service)
	query = "?t=9606,40674"
	service = Service{Name: "path",
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
	if err != nil {
		limit = 0
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		pageNum = 1
	}
	offset = (pageNum - 1) * limit
	ids, err := neidb.Taxids(name, limit, offset)
	util.Check(err)
	out := []Taxon{}
	for _, id := range ids {
		sciName, err := neidb.Name(id)
		util.Check(err)
		comName, err := neidb.CommonName(id)
		util.Check(err)
		tout := Taxon{}
		parent, err := neidb.Parent(id)
		if err == nil {
			tout = Taxon{Taxid: id, Parent: parent,
				Name: sciName, CommonName: comName}
		}
		if err == nil {
			out = append(out, tout)
		}
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func accessions(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	taxid := 0
	if len(taxa) > 0 {
		taxid = taxa[0]
	}
	out := []Accession{}
	accs, err := neidb.Accessions(taxid)
	util.Check(err)
	for _, acc := range accs {
		level, err := neidb.Level(acc)
		util.Check(err)
		o := Accession{acc, level}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func getTaxa(w http.ResponseWriter, r *http.Request) []int {
	taxa := []int{}
	t := r.URL.Query().Get("t")
	tokes := strings.Split(t, ",")
	for _, token := range tokes {
		taxon := 0
		taxon, err := strconv.Atoi(token)
		if err != nil {
			continue
		}
		_, err = neidb.Name(taxon)
		if err != nil {
			continue
		}
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
		util.Check(err)
		cname, err := neidb.CommonName(taxon)
		util.Check(err)
		o := Name{Taxid: taxa[i], Name: name,
			CommonName: cname}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func ranks(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	out := []Rank{}
	for i, taxon := range taxa {
		rank, err := neidb.Rank(taxon)
		util.Check(err)
		o := Rank{Taxid: taxa[i], Rank: rank}

		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func parent(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	taxid := 0
	if len(taxa) > 0 {
		taxid = taxa[0]
	}
	parent, err := neidb.Parent(taxid)
	out := Taxid{0}
	if err == nil {
		out = Taxid{parent}
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func children(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	taxid := 0
	if len(taxa) > 0 {
		taxid = taxa[0]
	}
	children, err := neidb.Children(taxid)
	util.Check(err)
	out := []Child{}
	for _, child := range children {
		name, err := neidb.Name(child)
		util.Check(err)
		cname, err := neidb.CommonName(child)
		util.Check(err)
		o := Child{child, name, cname}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func subtree(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	taxid := 0
	if len(taxa) > 0 {
		taxid = taxa[0]
	}
	taxa, err := neidb.Subtree(taxid)
	util.Check(err)
	out := []Node{}
	for _, taxon := range taxa {
		parent := taxon
		parent, err := neidb.Parent(taxon)
		util.Check(err)
		if err != nil {
			continue
		}
		name := ""
		cname := ""
		name, err = neidb.Name(taxon)
		util.Check(err)
		if err != nil {
			continue
		}
		cname, err = neidb.CommonName(taxon)
		util.Check(err)
		if err != nil {
			continue
		}
		o := Node{Taxid: taxon, Parent: parent, Name: name,
			CommonName: cname}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func taxids(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	name := r.URL.Query().Get("t")
	out := []Taxid{}
	taxids, err := neidb.CommonTaxids(name, -1, 0)
	util.Check(err)
	for _, taxid := range taxids {
		o := Taxid{taxid}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func mrca(w http.ResponseWriter, r *http.Request, p *PageData) {
	taxa := getTaxa(w, r)
	out := Taxid{0}
	if len(taxa) > 0 {
		mrca, err := neidb.MRCA(taxa)
		if err == nil {
			out = Taxid{mrca}
		}
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func levels(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	str := r.URL.Query().Get("a")
	accessions := strings.Split(str, ",")
	out := []Level{}
	for _, accession := range accessions {
		level, err := neidb.Level(accession)
		if err == nil {
			o := Level{Accession: accession, Level: level}
			out = append(out, o)
		}
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func num_genomes(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	taxid := 0
	if len(taxa) > 0 {
		taxid = taxa[0]
	}
	out := []GenomeCount{}
	for _, level := range tdb.AssemblyLevels() {
		n, err := neidb.NumGenomes(taxid, level)
		if err == nil {
			o := GenomeCount{Count: n, Level: level}
			out = append(out, o)
		}
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func num_genomes_rec(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	taxid := 0
	if len(taxa) > 0 {
		taxid = taxa[0]
	}
	out := []GenomeCount{}
	for _, level := range tdb.AssemblyLevels() {
		n, err := neidb.NumGenomesRec(taxid, level)
		if err == nil {
			o := GenomeCount{Count: n, Level: level}
			out = append(out, o)
		}
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func taxa_info(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	out := []TaxonInfo{}
	for _, taxon := range taxa {
		parent, err := neidb.Parent(taxon)
		util.Check(err)
		isLeaf, err := neidb.IsLeaf(taxon)
		util.Check(err)
		name, err := neidb.Name(taxon)
		util.Check(err)
		cname, err := neidb.CommonName(taxon)
		util.Check(err)
		rank, err := neidb.Rank(taxon)
		util.Check(err)
		var raw, rec []GenomeCount
		for _, level := range tdb.AssemblyLevels() {
			count, err := neidb.NumGenomes(taxon, level)
			util.Check(err)
			gc := GenomeCount{Count: count, Level: level}
			raw = append(raw, gc)
			count, err = neidb.NumGenomesRec(taxon, level)
			util.Check(err)
			gc = GenomeCount{Count: count, Level: level}
			rec = append(rec, gc)
		}
		o := TaxonInfo{
			Taxid:      taxon,
			Parent:     parent,
			IsLeaf:     isLeaf,
			Name:       name,
			CommonName: cname,
			Rank:       rank,
			RawCounts:  raw,
			RecCounts:  rec}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))
}
func path(w http.ResponseWriter, r *http.Request,
	p *PageData) {
	taxa := getTaxa(w, r)
	out := []Taxon{}
	if len(taxa) != 2 {
		b, err := json.MarshalIndent(out, "", "    ")
		util.Check(err)
		fmt.Fprintf(w, "%s\n", string(b))
		return
	}
	start := taxa[0]
	end := taxa[1]
	parent, err := neidb.Parent(start)
	util.Check(err)
	if parent == start && start != end {
		b, err := json.MarshalIndent(out, "", "    ")
		util.Check(err)
		fmt.Fprintf(w, "%s\n", string(b))
		return
	}
	name, err := neidb.Name(start)
	o := Taxon{Taxid: start, Parent: parent, Name: name}
	out = append(out, o)
	for start != end {
		parent, err := neidb.Parent(start)
		util.Check(err)
		if start == parent {
			out = out[:0]
			break
		}
		start = parent
		name, err := neidb.Name(start)
		util.Check(err)
		cname, err := neidb.CommonName(start)
		util.Check(err)
		parent, err = neidb.Parent(start)
		util.Check(err)
		o := Taxon{Taxid: start, Parent: parent, Name: name,
			CommonName: cname}
		out = append(out, o)
	}
	b, err := json.MarshalIndent(out, "", "    ")
	util.Check(err)
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
	staticFiles := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFiles))
	vitaxFiles := http.FileServer(http.Dir("vitax"))
	http.Handle("/vitax/", http.StripPrefix("/vitax/", vitaxFiles))
	dataFiles := http.FileServer(http.Dir("data"))
	http.Handle("/data/", http.StripPrefix("/data/", dataFiles))
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
	http.HandleFunc("/taxa_info/", makeHandler(taxa_info))
	http.HandleFunc("/path/", makeHandler(path))
	host := *flagO + ":" + *flagP
	if *flagC != "" && *flagK != "" {
		log.Fatal(http.ListenAndServeTLS(host, *flagC,
			*flagK, nil))
	} else {
		log.Fatal(http.ListenAndServe(host, nil))
	}
}
