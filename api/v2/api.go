package apiv2

import (
	"net/http"

	"log"

	"github.com/evolbioinf/neighbors/tdb"

	"strings"

	"strconv"

	"encoding/json"
	"fmt"

	"github.com/evolbioinf/never/util"

	"slices"
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
	Types    []string
	// PathParams  []string
	// QueryParams []string
}

func (node *Node) makeLink(rel, href string) Link {
	if href != "" {
		node.BasePath = href
	}
	return Link{
		Rel:    rel,
		Href:   node.BasePath,
		Action: node.Action,
		Types:  node.Types,
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

var root Node

var prefix string

func RegisterRoutes(pref, dbPath string) {
	var neidb *tdb.TaxonomyDB
	neidb, err := tdb.OpenTaxonomyDBcheck(dbPath)
	if err != nil {
		log.Fatal("apiV2: " + err.Error())
	}

	prefix = pref

	jsonT := "application/json"
	plainT := "plain/text"
	get := "GET"
	post := "POST"

	rootDocL := Node{
		Links:    make(map[string][]Node),
		Name:     "rootDocument",
		BasePath: prefix,
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(rootDocL, rootDocument, neidb) // new

	accessionsL := Node{
		Links:    make(map[string][]Node),
		Name:     "accessions",
		BasePath: prefix + "/accessions",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(accessionsL, accessions, neidb) // previously known as levels

	accessionL := Node{
		Links:    make(map[string][]Node),
		Name:     "accession",
		BasePath: prefix + "/accessions/{accession_id}",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(accessionL, accession, neidb) // new

	fintacL := Node{
		Links:    make(map[string][]Node),
		Name:     "fintac",
		BasePath: prefix + "/fintac",
		Action:   post,
		Types:    []string{jsonT},
	}
	makeRoute(fintacL, fintac, neidb) // new - calls fintac program

	taxaL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxa",
		BasePath: prefix + "/taxa",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(taxaL, taxa, neidb) // previously known as taxi

	taxonL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxon",
		BasePath: prefix + "/taxa/{taxon_id}",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(taxonL, taxon, neidb) // new

	ancestorsL := Node{
		Links:    make(map[string][]Node),
		Name:     "ancestors",
		BasePath: prefix + "/taxa/{taxon_id}/ancestors",
		Action:   get,
		Types:    []string{jsonT, plainT},
	}
	makeRoute(ancestorsL, ancestors, neidb) // new - calls ants program

	childrenL := Node{
		Links:    make(map[string][]Node),
		Name:     "children",
		BasePath: prefix + "/taxa/{taxon_id}/children",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(childrenL, children, neidb) // previously just children

	genomeCountL := Node{
		Links:    make(map[string][]Node),
		Name:     "genomeCount",
		BasePath: prefix + "/taxa/{taxon_id}/genome_count",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(genomeCountL, genomeCount, neidb) // previously known as num_genomes

	genomeCountRecL := Node{
		Links:    make(map[string][]Node),
		Name:     "genomeCountRec",
		BasePath: prefix + "/taxa/{taxon_id}/genome_count_recursive",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(genomeCountRecL, genomeCountRec, neidb) // previously known as num_genomes_rec

	parentL := Node{
		Links:    make(map[string][]Node),
		Name:     "parent",
		BasePath: prefix + "/taxa/{taxon_id}/parent",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(parentL, parent, neidb) // previously known as parent

	rankDistL := Node{
		Links:    make(map[string][]Node),
		Name:     "rankDistribution",
		BasePath: prefix + "/taxa/{taxon_id}/rank_distribution",
		Action:   post,
		Types:    []string{jsonT, plainT},
	}
	makeRoute(rankDistL, rankDistribution, neidb) // new - calls ranks program

	subtreeL := Node{
		Links:    make(map[string][]Node),
		Name:     "subtree",
		BasePath: prefix + "/taxa/{taxon_id}/subtree",
		Action:   get,
		Types:    []string{jsonT, plainT},
	}
	makeRoute(subtreeL, subtree, neidb) // previously just subtree

	taxonomyL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxonomy",
		BasePath: prefix + "/taxonomy",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(taxonomyL, taxonomy, neidb) // new

	mrcaL := Node{
		Links:    make(map[string][]Node),
		Name:     "mrca",
		BasePath: prefix + "/taxonomy/most_recent_common_ancestor",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(mrcaL, mrca, neidb) // previously just mrca

	neighborsL := Node{
		Links:    make(map[string][]Node),
		Name:     "neighbors",
		BasePath: prefix + "/taxonomy/neighbors",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(neighborsL, neighbors, neidb) // new - calls neighbors program

	pathL := Node{
		Links:    make(map[string][]Node),
		Name:     "path",
		BasePath: prefix + "/taxonomy/path",
		Action:   get,
		Types:    []string{jsonT},
	}
	makeRoute(pathL, path, neidb) // previously just path

	rootDocL.Links["self"] = append(rootDocL.Links["self"], rootDocL)
	accessionsL.Links["self"] = append(accessionsL.Links["self"], accessionsL)
	accessionL.Links["self"] = append(accessionL.Links["self"], accessionL)
	fintacL.Links["self"] = append(fintacL.Links["self"], fintacL)
	taxaL.Links["self"] = append(taxaL.Links["self"], taxaL)
	taxonL.Links["self"] = append(taxonL.Links["self"], taxonL)
	ancestorsL.Links["self"] = append(ancestorsL.Links["self"], ancestorsL)
	childrenL.Links["self"] = append(childrenL.Links["self"], childrenL)
	genomeCountL.Links["self"] = append(genomeCountL.Links["self"], genomeCountL)
	genomeCountRecL.Links["self"] = append(genomeCountRecL.Links["self"], genomeCountRecL)
	parentL.Links["self"] = append(parentL.Links["self"], parentL)
	rankDistL.Links["self"] = append(rankDistL.Links["self"], rankDistL)
	subtreeL.Links["self"] = append(subtreeL.Links["self"], subtreeL)
	taxonomyL.Links["self"] = append(taxonomyL.Links["self"], taxonomyL)
	mrcaL.Links["self"] = append(mrcaL.Links["self"], mrcaL)
	neighborsL.Links["self"] = append(neighborsL.Links["self"], neighborsL)
	pathL.Links["self"] = append(pathL.Links["self"], pathL)

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

	accessionsL.Links["previous"] = append(accessionsL.Links["previous"], accessionsL)
	taxaL.Links["previous"] = append(taxaL.Links["previous"], taxaL)
	ancestorsL.Links["previous"] = append(ancestorsL.Links["previous"], ancestorsL)
	childrenL.Links["previous"] = append(childrenL.Links["previous"], childrenL)
	subtreeL.Links["previous"] = append(subtreeL.Links["previous"], subtreeL)
	neighborsL.Links["previous"] = append(neighborsL.Links["previous"], neighborsL)
	pathL.Links["previous"] = append(pathL.Links["previous"], pathL)
	accessionsL.Links["next"] = append(accessionsL.Links["next"], accessionsL)
	taxaL.Links["next"] = append(taxaL.Links["next"], taxaL)
	ancestorsL.Links["next"] = append(ancestorsL.Links["next"], ancestorsL)
	childrenL.Links["next"] = append(childrenL.Links["next"], childrenL)
	subtreeL.Links["next"] = append(subtreeL.Links["next"], subtreeL)
	neighborsL.Links["next"] = append(neighborsL.Links["next"], neighborsL)
	pathL.Links["next"] = append(pathL.Links["next"], pathL)

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
	root = rootDocL

}
func makeRoute(node Node, fn func(http.ResponseWriter, *http.Request, ...any), args ...any) {
	http.HandleFunc(node.Action+" "+node.BasePath, func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, args...)
	})
}

func rootDocument(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[*any]{}
	out.Links = append(out.Links, root.Links["self"][0].makeLink("self", ""))
	for _, v := range root.Links["service"] {
		out.Links = append(out.Links, v.makeLink("service", ""))
	}

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
			if !plain {
				accNode := *root.getService("accession")
				acc.Links = append(acc.Links, accNode.makeLink("self", fillTemplate(accNode, map[string]string{"accession_id": acc.Accession})))

			}
			data = append(data, acc)
		}
	}

	var out any
	if plain {
		out = data
	} else {
		var links []Link
		self := root.getService("accessions")
		for link := range self.Links {
			for _, node := range self.Links[link] {
				links = append(links, node.makeLink(link, ""))
			}
		}
		out = ResponseBody[[]Accession]{Data: data, Links: links}
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

func fillTemplate(node Node, vals map[string]string) string {
	for k := range vals {
		node.BasePath = strings.Replace(node.BasePath, "{"+k+"}", vals[k], 1)
	}
	return node.BasePath
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

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxon(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func ancestors(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func children(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func genomeCount(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func genomeCountRec(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func parent(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func rankDistribution(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func subtree(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func taxonomy(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func mrca(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func neighbors(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
func path(w http.ResponseWriter, r *http.Request, args ...any) {
	out := ResponseBody[string]{Data: "Not implemented"}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
