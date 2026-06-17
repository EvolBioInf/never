package apiv2

import (
	"github.com/elnormous/contenttype"

	"net/http"

	"github.com/evolbioinf/neighbors/tdb"
	"log"

	"slices"

	"strings"

	"strconv"

	"encoding/json"
	"fmt"
	"github.com/evolbioinf/never/util"

	"bytes"
	"os"
	"os/exec"
	"time"
)

type Accession struct {
	Accession string `json:"accession"`
	Level     string `json:"level"`
	Links     []Link `json:"links,omitempty"`
}
type GenomeCount struct {
	Level string `json:"level"`
	Count int    `json:"count"`
	Links []Link `json:"links,omitempty"`
}
type TaxId struct {
	TaxId int    `json:"tax_id"`
	Links []Link `json:"links,omitempty"`
}
type Rank struct {
	TaxId int    `json:"tax_id"`
	Rank  string `json:"rank"`
	Links []Link `json:"links,omitempty"`
}
type TaxonAccessions struct {
	TaxId      int         `json:"tax_id"`
	Accessions []Accession `json:"accessions"`
	Links      []Link      `json:"links,omitempty"`
}
type TaxonMid struct {
	TaxId      int    `json:"tax_id"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
	Parent     int    `json:"parent"`
	Links      []Link `json:"links,omitempty"`
}
type TaxonNT struct {
	TaxId      int         `json:"tax_id"`
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Accessions []Accession `json:"parent"`
	Links      []Link      `json:"links,omitempty"`
}
type Image struct {
	Id          int    `json:"id"`
	Url         string `json:"url"`
	Attribution string `json:"attribution"`
}
type TaxonInfo struct {
	TaxId          int           `json:"tax_id"`
	Parent         int           `json:"parent"`
	IsLeaf         bool          `json:"is_leaf"`
	Name           string        `json:"name"`
	CommonName     string        `json:"common_name"`
	Rank           string        `json:"rank"`
	RawGenomeCount []GenomeCount `json:"raw_genome_counts"`
	RecGenomeCount []GenomeCount `json:"rec_genome_counts"`
	Images         []Image       `json:"images"`
	Links          []Link        `json:"links,omitempty"`
}

type ResponseBody[T any] struct {
	Data  T      `json:"data,omitempty"`
	Links []Link `json:"links,omitempty"`
}
type Link struct {
	Rel    string   `json:"rel"`
	Href   string   `json:"href"`
	Action string   `json:"action"`
	Types  []string `json:"types"`
}

type Node struct {
	Links    map[string][]Node
	Name     string
	BasePath string
	Action   string
	Types    []contenttype.MediaType
}

func (node *Node) makeLink(rel, href string) Link {
	if href != "" {
		node.BasePath = href
	}
	types := []string{}
	for _, t := range node.Types {
		types = append(types, t.String())
	}
	return Link{
		Rel:    rel,
		Href:   node.BasePath,
		Action: node.Action,
		Types:  types,
	}
}

func (node *Node) getService(name string) *Node {
	services := node.Links["service"]
	idx := slices.IndexFunc(services, func(a Node) bool { return a.Name == name })

	if idx != -1 {
		return &services[idx]
	} else {
		return nil
	}
}

var jsonct = contenttype.MediaType{Type: "application", Subtype: "json", Parameters: contenttype.Parameters{"charset": "utf-8"}}
var plainct = contenttype.MediaType{Type: "text", Subtype: "plain", Parameters: contenttype.Parameters{"charset": "utf-8"}}

var root Node

var prefix string

func RegisterRoutes(pref, dbPath string) {
	var neidb *tdb.TaxonomyDB
	neidb, err := tdb.OpenTaxonomyDBcheck(dbPath)
	if err != nil {
		log.Fatal("apiV2: " + err.Error())
	}

	prefix = pref

	get := "GET"
	post := "POST"

	rootDocL := Node{
		Links:    make(map[string][]Node),
		Name:     "rootDocument",
		BasePath: prefix,
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(rootDocL, rootDocument, neidb) // new

	accessionsL := Node{
		Links:    make(map[string][]Node),
		Name:     "accessions",
		BasePath: prefix + "/accessions",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(accessionsL, accessions, neidb) // previously known as levels

	accessionL := Node{
		Links:    make(map[string][]Node),
		Name:     "accession",
		BasePath: prefix + "/accessions/{accession_id}",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(accessionL, accession, neidb) // new

	taxaL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxa",
		BasePath: prefix + "/taxa",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(taxaL, taxa, neidb) // previously known as taxi

	taxonL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxon",
		BasePath: prefix + "/taxa/{taxon_id}",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(taxonL, taxon, neidb) // new

	ancestorsL := Node{
		Links:    make(map[string][]Node),
		Name:     "ancestors",
		BasePath: prefix + "/taxa/{taxon_id}/ancestors",
		Action:   get,
		Types:    []contenttype.MediaType{plainct},
	}
	makeRoute(ancestorsL, ancestors, neidb) // new - calls ants program

	childrenL := Node{
		Links:    make(map[string][]Node),
		Name:     "children",
		BasePath: prefix + "/taxa/{taxon_id}/children",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(childrenL, children, neidb) // previously just children

	genomeCountL := Node{
		Links:    make(map[string][]Node),
		Name:     "genomeCount",
		BasePath: prefix + "/taxa/{taxon_id}/genome_count",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(genomeCountL, genomeCount, neidb) // previously known as num_genomes

	genomeCountRecL := Node{
		Links:    make(map[string][]Node),
		Name:     "genomeCountRec",
		BasePath: prefix + "/taxa/{taxon_id}/genome_count_recursive",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(genomeCountRecL, genomeCountRec, neidb) // previously known as num_genomes_rec

	parentL := Node{
		Links:    make(map[string][]Node),
		Name:     "parent",
		BasePath: prefix + "/taxa/{taxon_id}/parent",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(parentL, parent, neidb) // previously known as parent

	rankDistL := Node{
		Links:    make(map[string][]Node),
		Name:     "rankDistribution",
		BasePath: prefix + "/taxa/{taxon_id}/rank_distribution",
		Action:   post,
		Types:    []contenttype.MediaType{plainct},
	}
	makeRoute(rankDistL, rankDistribution, neidb) // new - calls ranks program

	subtreeL := Node{
		Links:    make(map[string][]Node),
		Name:     "subtree",
		BasePath: prefix + "/taxa/{taxon_id}/subtree",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct, plainct},
	}
	makeRoute(subtreeL, subtree, neidb) // previously just subtree

	taxonomyL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxonomy",
		BasePath: prefix + "/taxonomy",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(taxonomyL, taxonomy, neidb) // new

	fintacL := Node{
		Links:    make(map[string][]Node),
		Name:     "fintac",
		BasePath: prefix + "/taxonomy/fintac",
		Action:   post,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(fintacL, fintac, neidb) // new - calls fintac program

	mrcaL := Node{
		Links:    make(map[string][]Node),
		Name:     "mrca",
		BasePath: prefix + "/taxonomy/most_recent_common_ancestor",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(mrcaL, mrca, neidb) // previously just mrca

	neighborsL := Node{
		Links:    make(map[string][]Node),
		Name:     "neighbors",
		BasePath: prefix + "/taxonomy/neighbors",
		Action:   get,
		Types:    []contenttype.MediaType{plainct},
	}
	makeRoute(neighborsL, neighbors, neidb) // new - calls neighbors program

	pathL := Node{
		Links:    make(map[string][]Node),
		Name:     "path",
		BasePath: prefix + "/taxonomy/path",
		Action:   get,
		Types:    []contenttype.MediaType{jsonct},
	}
	makeRoute(pathL, path, neidb) // previously just path

	rootDocL.Links["service"] = append(rootDocL.Links["service"], accessionsL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], accessionL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], fintacL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], taxaL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], taxonL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], ancestorsL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], childrenL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], genomeCountL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], genomeCountRecL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], parentL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], rankDistL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], subtreeL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], taxonomyL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], mrcaL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], neighborsL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], pathL)
	accessionsL.Links["service"] = append(accessionsL.Links["service"], accessionL)
	taxaL.Links["service"] = append(taxaL.Links["service"], taxonL)
	taxaL.Links["service"] = append(taxaL.Links["service"], ancestorsL)
	taxaL.Links["service"] = append(taxaL.Links["service"], childrenL)
	taxaL.Links["service"] = append(taxaL.Links["service"], genomeCountL)
	taxaL.Links["service"] = append(taxaL.Links["service"], genomeCountRecL)
	taxaL.Links["service"] = append(taxaL.Links["service"], parentL)
	taxaL.Links["service"] = append(taxaL.Links["service"], rankDistL)
	taxaL.Links["service"] = append(taxaL.Links["service"], subtreeL)
	taxonomyL.Links["service"] = append(taxonomyL.Links["service"], fintacL)
	taxonomyL.Links["service"] = append(taxonomyL.Links["service"], mrcaL)
	taxonomyL.Links["service"] = append(taxonomyL.Links["service"], neighborsL)
	taxonomyL.Links["service"] = append(taxonomyL.Links["service"], pathL)

	accessionsL.Links["entities"] = append(accessionsL.Links["entities"], accessionL)
	taxaL.Links["entities"] = append(taxaL.Links["entities"], taxonL)
	ancestorsL.Links["entities"] = append(ancestorsL.Links["entities"], taxonL)
	childrenL.Links["entities"] = append(childrenL.Links["entities"], taxonL)
	subtreeL.Links["entities"] = append(subtreeL.Links["entities"], taxonL)
	neighborsL.Links["entities"] = append(neighborsL.Links["entities"], taxonL)
	pathL.Links["entities"] = append(pathL.Links["entities"], taxonL)

	accessionL.Links["part-of"] = append(accessionL.Links["part-of"], accessionsL)
	taxonL.Links["part-of"] = append(taxonL.Links["part-of"], taxaL)

	mrcaL.Links["path"] = append(mrcaL.Links["path"], pathL)
	parentL.Links["all ancestors"] = append(parentL.Links["all ancestors"], ancestorsL)
	childrenL.Links["all descendants"] = append(childrenL.Links["all descendants"], subtreeL)
	root = rootDocL

}
func makeRoute(node Node, fn func(http.ResponseWriter, *http.Request, ...any), args ...any) {
	http.HandleFunc(node.Action+" "+node.BasePath, func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, args...)
	})
}

func rootDocument(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[*any]{}
	selfNode := root
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Serverd does not provide any of the accepted content types."))
		return
	}

	out.Links = append(out.Links, root.makeLink("self", ""))
	for _, v := range root.Links["service"] {
		out.Links = append(out.Links, v.makeLink("service", ""))
	}

	writeJsonOutput(w, out)

}

func writeJsonOutput(w http.ResponseWriter, out any) {
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))
}

func accessions(w http.ResponseWriter, r *http.Request, args ...any) {
	selfNode := root.getService("accessions")
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Serverd does not provide any of the accepted content types."))
		return
	}

	valid := checkParams(w, r, "accession_ids")
	if !valid {
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if err != nil {
		plain = false
	}

	str := r.URL.Query().Get("accession_ids")
	accessions := strings.Split(str, ",")

	offset, size := extractPaging(r)

	var data = []Accession{}

	if size == -1 {
		size = len(accessions)
	}
	for i := offset; i < min(offset+size, len(accessions)); i++ {
		accession := accessions[i]
		level, err := neidb.Level(accession)
		if err == nil {
			acc := Accession{Accession: accession, Level: level}
			if !plain {
				accNode := *root.getService("accession")
				acc.Links = append(acc.Links, accNode.makeLink("self", fillTemplate(accNode, map[string]string{"accession_id": acc.Accession})))

			}
			data = append(data, acc)
		}
	}

	out := ResponseBody[[]Accession]{Data: data}
	if !plain {
		var links []Link
		links = append(links, selfNode.makeLink("self", r.URL.String()))

		for link := range selfNode.Links {
			for _, node := range selfNode.Links[link] {
				links = append(links, node.makeLink(link, ""))
			}
		}

		out.Links = links
	}

	if plain {
		writeJsonOutput(w, out.Data)
	} else {
		writeJsonOutput(w, out)
	}

}

func checkParams(w http.ResponseWriter, r *http.Request, args ...string) bool {
	for _, arg := range args {
		p := r.URL.Query().Get(arg)
		if p == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing required parameter."))
			return false
		}
	}
	return true
}

func extractPaging(r *http.Request) (offset, size int) {
	strPage := r.URL.Query().Get("page")
	strPageSize := r.URL.Query().Get("page_size")

	if strPageSize != "" {
		cSize, err := strconv.Atoi(strPageSize)
		size = cSize
		if err != nil {
			size = -1
		}
	} else {
		size = -1
	}
	if size != -1 && strPage != "" {
		page, err := strconv.Atoi(strPage)
		if err != nil {
			offset = 0
		} else {
			offset = page * size
		}
	} else {
		offset = 0
	}
	return
}

func fillTemplate(node Node, vals map[string]string) string {
	for k := range vals {
		node.BasePath = strings.Replace(node.BasePath, "{"+k+"}", vals[k], 1)
	}
	return node.BasePath
}

func accession(w http.ResponseWriter, r *http.Request, args ...any) {
	selfNode := root.getService("accession")
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Serverd does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if err != nil {
		plain = false
	}

	accession := r.PathValue("accession_id")
	level, err := neidb.Level(accession)
	out := ResponseBody[Accession]{}
	if err == nil {
		out.Data = Accession{Accession: accession, Level: level}
		if !plain {
			var links []Link
			links = append(links, selfNode.makeLink("self", r.URL.String()))

			for link := range selfNode.Links {
				for _, node := range selfNode.Links[link] {
					links = append(links, node.makeLink(link, ""))
				}
			}

			out.Links = links

		}
	}

	var links []Link
	links = append(links, selfNode.makeLink("self", r.URL.String()))

	for link := range selfNode.Links {
		for _, node := range selfNode.Links[link] {
			links = append(links, node.makeLink(link, ""))
		}
	}

	out.Links = links

	if plain {
		writeJsonOutput(w, out.Data)
	} else {
		writeJsonOutput(w, out)
	}

}

func taxa(w http.ResponseWriter, r *http.Request, args ...any) {
	valid := checkParams(w, r, "name")
	if !valid {
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	offset, size := extractPaging(r)

	name := r.URL.Query().Get("name")
	strExact := r.URL.Query().Get("exact")
	strScientific := r.URL.Query().Get("scientific")

	exact, err := strconv.ParseBool(strExact)
	if err != nil {
		exact = false
	}

	scientific, err := strconv.ParseBool(strScientific)
	if err != nil {
		scientific = false
	}

	if !exact {
		name = strings.ReplaceAll(name, " ", "% %")
		name = "%" + name + "%"
	}

	var ids []int
	if scientific {
		ids, err = neidb.Taxids(name, size, offset)
	} else {
		ids, err = neidb.CommonTaxids(name, size, offset)
	}

	util.Check(err)
	out := []TaxonMid{}
	for _, id := range ids {
		sciName, err := neidb.Name(id)
		util.Check(err)
		comName, err := neidb.CommonName(id)
		util.Check(err)
		tout := TaxonMid{}
		parent, err := neidb.Parent(id)
		if err == nil {
			tout = TaxonMid{TaxId: id, Parent: parent,
				Name: sciName, CommonName: comName}
		}

		if err == nil {
			out = append(out, tout)
		}

	}

	writeJsonOutput(w, out)
}

func taxon(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func ancestors(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func children(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func genomeCount(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func genomeCountRec(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func parent(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func rankDistribution(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func subtree(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func taxonomy(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}

func fintac(w http.ResponseWriter, r *http.Request, args ...any) {
	a := r.URL.Query().Get("a")
	n := r.URL.Query().Get("n")
	t := r.URL.Query().Get("t")
	u := r.URL.Query().Get("u")

	fintacArgs := []string{}

	if a != "" {
		fintacArgs = append(fintacArgs, "-a", a)
	}
	if n != "" {
		fintacArgs = append(fintacArgs, "-n", n)
	}
	if t != "" {
		fintacArgs = append(fintacArgs, "-t", t)
	}
	if u != "" {
		fintacArgs = append(fintacArgs, "-u", u)
	}

	fintacArgs = append(fintacArgs, "-N", "neidb")

	r.ParseMultipartForm(100_000_000)

	paths := []string{}
	i := 0
	for key := range r.MultipartForm.File {
		files := r.MultipartForm.File[key]
		if len(files) > 0 {
			h := *files[0]
			rf, err := h.Open()
			if err != nil {
				log.Fatal("Err while reading file from fintac request: ", err)
			}
			b := make([]byte, h.Size)
			rf.Read(b)
			rf.Close()

			buffer := bytes.NewBuffer(b)
			path := "apiv2_temp_" + strconv.Itoa(i) + strconv.Itoa(time.Now().Nanosecond())
			defer os.Remove(path)
			paths = append(paths, path)
			tf, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(fmt.Sprintf("Could not open %s: %s\n", path, err))
			}
			buffer.WriteTo(tf)
			tf.Close()
			i++
		}
	}

	fintacArgs = append(fintacArgs, paths...)
	fmt.Println("fintac", fintacArgs)
	w.WriteHeader(http.StatusOK)
	for _, path := range paths {
		cmd := exec.Command("cat", path)
		outt, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte(outt))
	}
}

func mrca(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func neighbors(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
func path(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	writeJsonOutput(w, out)
}
