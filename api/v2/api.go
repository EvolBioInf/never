package apiv2

import (
	"net/http"

	"github.com/evolbioinf/neighbors/tdb"

	"log"

	"cmp"
	"slices"

	"strings"

	"strconv"

	"encoding/json"
	"fmt"

	"github.com/evolbioinf/never/util"
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
type Link struct {
	Rel    string   `json:"rel"`
	Href   string   `json:"href"`
	Action string   `json:"action"`
	Types  []string `json:"types"`
}
type ResponseBody[T any] struct {
	Data  T      `json:"data,omitempty"`
	Links []Link `json:"links,omitempty"`
}

type Node struct {
	Links    map[string][]Node
	BasePath string
	// PathParams  []string
	// QueryParams []string
	Action string
	Types  []string
}

var prefix string
var root Node

func initLinkGraph() {
	dbPath := ""

	var neidb *tdb.TaxonomyDB
	neidb, err := tdb.OpenTaxonomyDBcheck(dbPath)
	if err != nil {
		log.Fatal("apiV2: " + err.Error())
	}

	jsonT := "application/json"
	plainT := "plain/text"
	get := "GET"
	post := "POST"

	rootDocL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix,
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute("GET", "", rootDocument, neidb) // new

	accessionsL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/accessions",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(accessionsL.Action, accessionsL.BasePath, accessions, neidb) // previously known as levels

	accessionL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/accessions/accession_id",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(accessionL.Action, accessionL.BasePath, accession, neidb) // new

	fintacL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/fintac",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(fintacL.Action, fintacL.BasePath, accession, neidb) // new - calls fintac program

	taxaL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(taxaL.Action, taxaL.BasePath, accession, neidb) // previously known as taxi

	taxonL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(taxonL.Action, taxonL.BasePath, accession, neidb) // new

	ancestorsL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}/ancestors",
		Action:   get,
		Types:    []string{jsonT, "plain/text"},
	}
	makeRoute(ancestorsL.Action, ancestorsL.BasePath, accession, neidb) // new - calls ants program

	childrenL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}/children",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(childrenL.Action, childrenL.BasePath, accession, neidb) // previously just children

	genome_countL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}/genome_count",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(genome_countL.Action, genome_countL.BasePath, accession, neidb) // previously known as num_genomes

	genome_count_recL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}/genome_count_recursive",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(genome_count_recL.Action, genome_count_recL.BasePath, accession, neidb) // previously known as num_genomes_rec

	parentL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}/parent",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(parentL.Action, parentL.BasePath, accession, neidb) // previously known as parent

	ranksL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}/rank_distribution",
		Action:   get,
		Types:    []string{jsonT, "plain/text"},
	}
	makeRoute(ranksL.Action, ranksL.BasePath, accession, neidb) // new - calls ranks program

	subtreeL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxa/{taxon_id}/rank_distribution",
		Action:   get,
		Types:    []string{jsonT, "plain/text"},
	}
	makeRoute(subtreeL.Action, subtreeL.BasePath, accession, neidb) // previously just subtree

	taxonomyL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxonomy",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(taxonomyL.Action, taxonomyL.BasePath, accession, neidb) // new

	mrcaL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxonomy/most_recent_common_ancestor",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(mrcaL.Action, mrcaL.BasePath, accession, neidb) // previously just mrca

	neighborsL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxonomy/neighbors",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(neighborsL.Action, neighborsL.BasePath, accession, neidb) // new - calls neighbors program

	pathL := Node{
		Links:    make(map[string][]Node),
		BasePath: prefix + "/taxonomy/path",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(pathL.Action, pathL.BasePath, accession, neidb) // previously just path

	rootDocL.Links["service"] = append(rootDocL.Links["service"], accessionsL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], accessionL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], fintacL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], taxaL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], taxonL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], ancestorsL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], childrenL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], genome_countL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], genome_count_recL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], parentL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], ranksL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], subtreeL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], taxonomyL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], mrcaL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], neighborsL)
	rootDocL.Links["service"] = append(rootDocL.Links["service"], pathL)

	rootDocL.Links["self"] = append(rootDocL.Links["self"], rootDocL)
	accessionsL.Links["self"] = append(accessionsL.Links["self"], accessionsL)
	accessionL.Links["self"] = append(accessionL.Links["self"], accessionL)
	fintacL.Links["self"] = append(fintacL.Links["self"], fintacL)
	taxaL.Links["self"] = append(taxaL.Links["self"], taxaL)
	taxonL.Links["self"] = append(taxonL.Links["self"], taxonL)
	ancestorsL.Links["self"] = append(ancestorsL.Links["self"], ancestorsL)
	childrenL.Links["self"] = append(childrenL.Links["self"], childrenL)
	genome_countL.Links["self"] = append(genome_countL.Links["self"], genome_countL)
	genome_count_recL.Links["self"] = append(genome_count_recL.Links["self"], genome_count_recL)
	parentL.Links["self"] = append(parentL.Links["self"], parentL)
	ranksL.Links["self"] = append(ranksL.Links["self"], ranksL)
	subtreeL.Links["self"] = append(subtreeL.Links["self"], subtreeL)
	taxonomyL.Links["self"] = append(taxonomyL.Links["self"], taxonomyL)
	mrcaL.Links["self"] = append(mrcaL.Links["self"], mrcaL)
	neighborsL.Links["self"] = append(neighborsL.Links["self"], neighborsL)
	pathL.Links["self"] = append(pathL.Links["self"], pathL)

	accessionsL.Links["contains"] = append(accessionsL.Links["contains"], accessionL)
	taxaL.Links["contains"] = append(taxaL.Links["contains"], taxonL)
	ancestorsL.Links["contains"] = append(ancestorsL.Links["contains"], taxonL)
	childrenL.Links["contains"] = append(childrenL.Links["contains"], taxonL)
	subtreeL.Links["contains"] = append(subtreeL.Links["contains"], taxonL)
	neighborsL.Links["contains"] = append(neighborsL.Links["contains"], taxonL)
	pathL.Links["contains"] = append(pathL.Links["contains"], taxonL)

	accessionL.Links["part-of"] = append(accessionL.Links["part-of"], accessionsL)
	taxonL.Links["part-of"] = append(taxonL.Links["part-of"], taxaL)

	accessionsL.Links["previous"] = append(accessionsL.Links["previous"], accessionsL)
	taxaL.Links["previous"] = append(taxaL.Links["previous"], taxaL)
	ancestorsL.Links["previous"] = append(ancestorsL.Links["previous"], ancestorsL)
	childrenL.Links["previous"] = append(childrenL.Links["previous"], childrenL)
	subtreeL.Links["previous"] = append(subtreeL.Links["previous"], subtreeL)
	neighborsL.Links["previous"] = append(neighborsL.Links["previous"], neighborsL)
	pathL.Links["previous"] = append(pathL.Links["previous"], pathL)

	taxaL.Links["more fields"] = append(taxaL.Links["more fields"], taxaL)
	taxonL.Links["more fields"] = append(taxonL.Links["more fields"], taxonL)
	ancestorsL.Links["more fields"] = append(ancestorsL.Links["more fields"], ancestorsL)
	childrenL.Links["more fields"] = append(childrenL.Links["more fields"], childrenL)
	parentL.Links["more fields"] = append(parentL.Links["more fields"], parentL)
	subtreeL.Links["more fields"] = append(subtreeL.Links["more fields"], subtreeL)
	mrcaL.Links["more fields"] = append(mrcaL.Links["more fields"], mrcaL)
	pathL.Links["more fields"] = append(pathL.Links["more fields"], pathL)

	mrcaL.Links["path"] = append(mrcaL.Links["path"], pathL)
	parentL.Links["all ancestors"] = append(parentL.Links["all ancestors"], ancestorsL)
	childrenL.Links["all descendants"] = append(childrenL.Links["all descendants"], subtreeL)
	childrenL.Links["all descendants"] = append(childrenL.Links["all descendants"], subtreeL)

	root = rootDocL
}

func RegisterRoutes(pref, dbPath string) {

	prefix = pref

	// makeRoute("GET", "/accessions/{accession_id}", "base-accession", accession, neidb) // new
	// makeRoute("POST", "/fintac", "base-fintac", fintac, neidb)                         // new - calls fintac program
	// makeRoute("GET", "/taxa", "base-taxa", taxa, neidb)                                // previously known as taxi

}
func makeRoute(method, path string, fn func(http.ResponseWriter, *http.Request, ...any), args ...any) {
	http.HandleFunc(method+" "+path, func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, args...)
	})
}

func rootDocument(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[*any]{}
	for k := range LinkMap {
		curr := LinkMap[k]
		curr.Rel = "service"
		out.Links = append(out.Links, curr)
	}

	slices.SortFunc(out.Links, func(a, b Link) int {
		return cmp.Compare(a.Href, b.Href)
	})

	out.Links[0].Rel = "self"

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func accessions(w http.ResponseWriter, r *http.Request, args ...any) {
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
			acc.Links = []Link{}

			data = append(data)
		}
	}

	var out any
	if plain {
		out = data
	} else {
		out = ResponseBody[[]Accession]{Data: data}
		links := []Link{}

	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

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

func accession(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	out := ResponseBody[Accession]{}
	accession := r.PathValue("accession_id")
	level, err := neidb.Level(accession)
	if err == nil {
		out.Data = Accession{Accession: accession, Level: level}
	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func fintac(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

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
	out := []Taxon{}
	for _, id := range ids {
		sciName, err := neidb.Name(id)
		util.Check(err)
		comName, err := neidb.CommonName(id)
		util.Check(err)
		tout := Taxon{}
		parent, err := neidb.Parent(id)
		if err == nil {
			tout = Taxon{TaxId: id, Parent: parent,
				Name: sciName, CommonName: comName}
		}

		if err == nil {
			out = append(out, tout)
		}

	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
