package apiv2

import (
	"github.com/elnormous/contenttype"

	"net/http"

	"github.com/evolbioinf/neighbors/tdb"
	"log"

	"strings"

	"bytes"

	"strconv"

	"errors"
	"runtime/debug"

	"encoding/json"
	"fmt"
	"github.com/evolbioinf/never/util"

	"slices"

	"context"
	"time"

	"os/exec"

	"sort"

	"mime/multipart"

	"os"

	"mime"
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
type Rank struct {
	TaxId int    `json:"tax_id"`
	Rank  string `json:"rank"`
	Links []Link `json:"links,omitempty"`
}
type Image struct {
	Id          int    `json:"id"`
	Url         string `json:"url"`
	Attribution string `json:"attribution"`
}
type Taxon struct {
	TaxId          int           `json:"tax_id"`
	Parent         int           `json:"parent,omitempty"`
	IsLeaf         bool          `json:"is_leaf,omitempty"`
	Name           string        `json:"name,omitempty"`
	CommonName     string        `json:"common_name,omitempty"`
	Rank           string        `json:"rank,omitempty"`
	Accessions     []Accession   `json:"accessions,omitempty"`
	RawGenomeCount []GenomeCount `json:"raw_genome_counts,omitempty"`
	RecGenomeCount []GenomeCount `json:"rec_genome_counts,omitempty"`
	Images         []Image       `json:"images,omitempty"`
	Links          []Link        `json:"links,omitempty"`
}
type TaxonAccession struct {
	TaxId      int         `json:"tax_id"`
	Accessions []Accession `json:"accessions,omitempty"`
	Links      []Link      `json:"links,omitempty"`
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
		Href:   serverAddress + node.BasePath,
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

var root Node

var prefix string
var serverAddress string

var jsonCt = contenttype.MediaType{Type: "application", Subtype: "json", Parameters: contenttype.Parameters{"charset": "utf-8"}}
var graphvizCt = contenttype.MediaType{Type: "text", Subtype: "vnd.graphviz", Parameters: contenttype.Parameters{"charset": "utf-8"}}
var plainCt = contenttype.MediaType{Type: "text", Subtype: "plain", Parameters: contenttype.Parameters{"charset": "utf-8"}}

var multipartCt = contenttype.MediaType{Type: "multipart", Subtype: "form-data"}

func RegisterRoutes(pref, dbPath, serverAdr string) {
	var neidb *tdb.TaxonomyDB
	neidb, err := tdb.OpenTaxonomyDBcheck(dbPath)
	if err != nil {
		log.Fatal("apiV2: error while opening the database: ", err.Error())
	}

	prefix = pref
	serverAddress = serverAdr

	rootDocL := Node{
		Links:    make(map[string][]Node),
		Name:     "rootDocument",
		BasePath: prefix,
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&rootDocL, rootDocument, neidb) // new

	accessionsL := Node{
		Links:    make(map[string][]Node),
		Name:     "accessions",
		BasePath: prefix + "/accessions",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&accessionsL, accessions, neidb) // previously known as levels

	accessionL := Node{
		Links:    make(map[string][]Node),
		Name:     "accession",
		BasePath: prefix + "/accessions/{accession_id}",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&accessionL, accession, neidb) // new

	taxaL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxa",
		BasePath: prefix + "/taxa",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&taxaL, taxa, neidb, dbPath) // previously known as taxi

	taxonL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxon",
		BasePath: prefix + "/taxa/{taxon_id}",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt, plainCt},
	}
	makeRoute(&taxonL, taxon, neidb, dbPath) // new

	childrenL := Node{
		Links:    make(map[string][]Node),
		Name:     "children",
		BasePath: prefix + "/taxa/{taxon_id}/children",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&childrenL, children, neidb) // previously just children

	parentL := Node{
		Links:    make(map[string][]Node),
		Name:     "parent",
		BasePath: prefix + "/taxa/{taxon_id}/parent",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&parentL, parent, neidb) // previously known as parent

	subtreeL := Node{
		Links:    make(map[string][]Node),
		Name:     "subtree",
		BasePath: prefix + "/taxa/{taxon_id}/subtree",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&subtreeL, subtree, neidb, dbPath) // previously just subtree

	taxonomyL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxonomy",
		BasePath: prefix + "/taxonomy",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&taxonomyL, taxonomy, neidb) // new

	taxonAccessionsL := Node{
		Links:    make(map[string][]Node),
		Name:     "taxonAccessions",
		BasePath: prefix + "/taxonomy/accessions",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&taxonAccessionsL, taxonAccessions, neidb) // previously called accessions

	mrcaL := Node{
		Links:    make(map[string][]Node),
		Name:     "mrca",
		BasePath: prefix + "/taxonomy/mrca",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&mrcaL, mrca, neidb) // previously just mrca

	pathL := Node{
		Links:    make(map[string][]Node),
		Name:     "path",
		BasePath: prefix + "/taxonomy/path",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&pathL, path, neidb) // previously just path

	programsDocL := Node{
		Links:    make(map[string][]Node),
		Name:     "programs",
		BasePath: prefix + "/programs",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{jsonCt},
	}
	makeRoute(&programsDocL, programsDoc)

	progAntsL := Node{
		Links:    make(map[string][]Node),
		Name:     "progAnts",
		BasePath: prefix + "/programs/ants",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progAntsL, programEndpoint, dbPath, "ants")

	progDreeL := Node{
		Links:    make(map[string][]Node),
		Name:     "progDree",
		BasePath: prefix + "/programs/dree",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progDreeL, programEndpoint, dbPath, "dree")

	progFintacL := Node{
		Links:    make(map[string][]Node),
		Name:     "progFintac",
		BasePath: prefix + "/programs/fintac",
		Action:   http.MethodPost,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progFintacL, programEndpoint, dbPath)

	progNeighborsGetL := Node{
		Links:    make(map[string][]Node),
		Name:     "progNeighbors",
		BasePath: prefix + "/programs/neighbors",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progNeighborsGetL, programEndpoint, dbPath, "neighbors")
	progNeighborsPostL := Node{
		Links:    make(map[string][]Node),
		Name:     "progNeighbors",
		BasePath: prefix + "/programs/neighbors",
		Action:   http.MethodPost,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progNeighborsPostL, programEndpoint, dbPath, "neighbors")

	progRanksGetL := Node{
		Links:    make(map[string][]Node),
		Name:     "progRanks",
		BasePath: prefix + "/programs/ranks",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progRanksGetL, programEndpoint, dbPath, "ranks")
	progRanksPostL := Node{
		Links:    make(map[string][]Node),
		Name:     "progRanks",
		BasePath: prefix + "/programs/ranks",
		Action:   http.MethodPost,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progRanksPostL, programEndpoint, dbPath, "ranks")

	progTaxiL := Node{
		Links:    make(map[string][]Node),
		Name:     "progTaxi",
		BasePath: prefix + "/programs/taxi",
		Action:   http.MethodGet,
		Types:    []contenttype.MediaType{plainCt},
	}
	makeRoute(&progTaxiL, programEndpoint, dbPath, "taxi")

	rootDocL.Links["service"] = append(rootDocL.Links["service"],
		accessionsL,
		accessionL,
		taxaL,
		taxonL,
		childrenL,
		parentL,
		subtreeL,
		taxonomyL,
		taxonAccessionsL,
		mrcaL,
		pathL,
		progAntsL,
		progDreeL,
		progFintacL,
		progNeighborsGetL,
		progNeighborsPostL,
		progRanksGetL,
		progRanksPostL,
		progTaxiL,
	)
	accessionsL.Links["service"] = append(accessionsL.Links["service"], accessionL)
	taxaL.Links["service"] = append(taxaL.Links["service"],
		taxonL,
		childrenL,
		parentL,
		subtreeL,
	)
	taxonomyL.Links["service"] = append(taxonomyL.Links["service"],
		taxonAccessionsL,
		mrcaL,
		pathL,
	)
	programsDocL.Links["service"] = append(programsDocL.Links["service"],
		progAntsL,
		progDreeL,
		progFintacL,
		progNeighborsGetL,
		progNeighborsPostL,
		progRanksGetL,
		progRanksPostL,
		progTaxiL,
	)

	accessionsL.Links["entities"] = append(accessionsL.Links["entities"], accessionL)
	taxaL.Links["entities"] = append(taxaL.Links["entities"], taxonL)
	childrenL.Links["entities"] = append(childrenL.Links["entities"], taxonL)
	subtreeL.Links["entities"] = append(subtreeL.Links["entities"], taxonL)
	pathL.Links["entities"] = append(pathL.Links["entities"], taxonL)

	accessionL.Links["part-of"] = append(accessionL.Links["part-of"], accessionsL)
	taxonL.Links["part-of"] = append(taxonL.Links["part-of"], taxaL)

	mrcaL.Links["path"] = append(mrcaL.Links["path"], pathL)
	childrenL.Links["all descendants"] = append(childrenL.Links["all descendants"], subtreeL)
	root = rootDocL

	services := root.Links["service"]
	grouped := make(map[string][]Node)
	for _, service := range services {
		grouped[service.BasePath] = append(grouped[service.BasePath], service)
	}

	for path := range grouped {
		makeOptionsRoute(grouped[path])
	}

}
func makeRoute(node *Node, fn func(http.ResponseWriter, *http.Request, *Node, ...any), args ...any) {
	http.HandleFunc(node.Action+" "+node.BasePath, func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, node, args...)
	})
}

func makeOptionsRoute(nodes []Node) {
	http.HandleFunc("OPTIONS "+nodes[0].BasePath, func(w http.ResponseWriter, r *http.Request) {
		methods := []string{"OPTIONS"}
		for _, node := range nodes {
			methods = append(methods, node.Action)
			if node.Action == "GET" {
				methods = append(methods, "HEAD")
			}
		}

		w.Header().Set("Allow", strings.Join(methods, ", "))
		w.WriteHeader(http.StatusNoContent)
	})
}

func rootDocument(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	out := ResponseBody[*any]{}
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	out.Links = append(out.Links, root.makeLink("self", ""))
	for _, v := range root.Links["service"] {
		out.Links = append(out.Links, v.makeLink("service", ""))
	}

	writeJsonOutput(w, out)

}

func writeJsonOutput(w http.ResponseWriter, out any) {
	w.Header().Set("Content-Type", jsonCt.String())
	w.WriteHeader(http.StatusOK)
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(out)
	util.Check(err)
	w.Write(buf.Bytes())
}

func accessions(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	valid := checkParams(w, r, "accession_ids")
	if !valid {
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	str := r.URL.Query().Get("accession_ids")
	accessions := strings.Split(str, ",")

	offset, limit := extractPaging(r)

	var data = []Accession{}

	if limit == -1 {
		limit = len(accessions)
	}
	for i := offset; i < min(offset+limit, len(accessions)); i++ {
		accession := accessions[i]
		level, err := neidb.Level(accession)
		if err != nil {
			writeServerError(w, "fn accessions - Error while accessing neidb.Level", "", err)
			return
		}
		acc := Accession{Accession: accession, Level: level}
		if !plain {
			accNode := *root.getService("accession")
			acc.Links = append(acc.Links, accNode.makeLink("self",
				fillTemplate(accNode, map[string]string{}, map[string]string{})))

		}
		data = append(data, acc)
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
			writeBadRequestResp(w, "Missing required parameter.")

			return false
		}
	}
	return true
}

func writeBadRequestResp(w http.ResponseWriter, text string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(text))
}

func extractPaging(r *http.Request) (offset, limit int) {
	strOffset := r.URL.Query().Get("offset")
	strLimit := r.URL.Query().Get("limit")

	if strOffset != "" {
		cOff, err := strconv.Atoi(strOffset)
		if err != nil {
			offset = 0
		} else {
			offset = cOff
		}
	} else {
		offset = 0
	}

	if strLimit != "" {
		cLim, err := strconv.Atoi(strLimit)
		if err != nil {
			limit = -1
		} else {
			limit = cLim
		}
	} else {
		limit = -1
	}
	return
}

func writeServerError(w http.ResponseWriter, internalMsg string, responseMsg string, err error) {
	if err == nil {
		err = errors.New("No error passed")
	}
	util.LogWarningDef(
		util.WarningEntry{
			Warning: fmt.Sprintf("Apiv2: Internal server error.\nMessage: %s\nErr: %s\nTrace: %s", internalMsg, err.Error(), debug.Stack()),
		})
	w.WriteHeader(http.StatusInternalServerError)
	if responseMsg != "" {
		w.Write([]byte("An error occurred: " + responseMsg))
	} else {
		w.Write([]byte("Internal server error."))
	}
}

func fillTemplate(node Node, pathParams map[string]string, queryParams map[string]string) string {
	for k := range pathParams {
		node.BasePath = strings.Replace(node.BasePath, "{"+k+"}", pathParams[k], 1)
	}
	i := 0
	for k := range queryParams {
		str := ""
		if i == 0 {
			str += "?"
		} else {
			str += "&"
		}
		node.BasePath = fmt.Sprintf("%s%s%s=%s", node.BasePath, str, k, queryParams[k])
		i++
	}
	return node.BasePath
}

func accession(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	accession := r.PathValue("accession_id")
	level, err := neidb.Level(accession)
	if err != nil {
		writeServerError(w, "fn accession - Error while accessing neidb.Level", "", err)
		return
	}
	var data Accession
	var links []Link
	data = Accession{Accession: accession, Level: level}
	if !plain {
		links = append(links, selfNode.makeLink("self", r.URL.String()))

		for link := range selfNode.Links {
			for _, node := range selfNode.Links[link] {
				links = append(links, node.makeLink(link, ""))
			}
		}

	}

	out := ResponseBody[Accession]{Data: data, Links: links}

	if plain {
		writeJsonOutput(w, out.Data)
	} else {
		writeJsonOutput(w, out)
	}

}

func taxa(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	valid := checkParams(w, r, "name")
	if !valid {
		return
	}
	offset, limit := extractPaging(r)

	name := r.URL.Query().Get("name")
	strExact := r.URL.Query().Get("exact")
	strScientific := r.URL.Query().Get("scientific")
	fieldComposite := r.URL.Query().Get("field_composite")
	fieldComposite = parseFieldComposite(fieldComposite)

	exact, err := strconv.ParseBool(strExact)
	if strExact != "" && err != nil {
		writeBadRequestResp(w, "exact is not a bool.")
		return
	}
	scientific, err := strconv.ParseBool(strScientific)
	if strScientific != "" && err != nil {
		writeBadRequestResp(w, "scientific is not a bool.")
		return
	}
	if !exact {
		name = strings.ReplaceAll(name, " ", "% %")
		name = "%" + name + "%"
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	var ids []int
	if scientific {
		ids, err = neidb.Taxids(name, limit, offset)
		if err != nil {
			writeServerError(w, "fn taxa - Error while accessing neidb.Taxids", "", err)
			return
		}
	} else {
		ids, err = neidb.CommonTaxids(name, limit, offset)
		if err != nil {
			writeServerError(w, "fn taxa - Error while accessing neidb.CommonTaxids", "", err)
			return
		}
	}
	data := []Taxon{}
	for _, id := range ids {
		tax, err := getTaxonData(id, plain, fieldComposite, neidb)
		if err != nil {
			writeServerError(w, "Error from getTaxonData", "", err)
			return
		}
		data = append(data, tax)

	}

	out := ResponseBody[[]Taxon]{Data: data}

	if !plain {
		var links []Link
		links = append(links, selfNode.makeLink("self", r.URL.String()))

		for link := range selfNode.Links {
			for _, node := range selfNode.Links[link] {
				links = append(links, node.makeLink(link, ""))
			}
		}

		links = append(links, getTaxonCompositeLinks(selfNode, fieldComposite, r)...)

		out.Links = links
	}

	if plain {
		writeJsonOutput(w, out.Data)
	} else {
		writeJsonOutput(w, out)
	}

}

func parseFieldComposite(arg string) string {
	availableComps := []string{"all", "default", "gen_count", "gen_count_rec", "id", "rank"}
	_, found := slices.BinarySearch(availableComps, arg)
	if !found {
		arg = "default"
	}
	return arg
}

func getTaxonData(id int, plain bool, fieldComposite string, neidb *tdb.TaxonomyDB) (tax Taxon, err error) {
	tax.TaxId = id
	if fieldComposite == "gen_count" || fieldComposite == "all" {
		var raw []GenomeCount
		for _, level := range tdb.AssemblyLevels() {
			var count int
			count, err = neidb.NumGenomes(id, level)
			if err != nil {
				return
			}
			gc := GenomeCount{Count: count, Level: level}
			raw = append(raw, gc)
		}
		tax.RawGenomeCount = raw

	}
	if fieldComposite == "gen_count_rec" || fieldComposite == "all" {
		var rec []GenomeCount
		for _, level := range tdb.AssemblyLevels() {
			var count int
			count, err = neidb.NumGenomesRec(id, level)
			if err != nil {
				return
			}
			gc := GenomeCount{Count: count, Level: level}
			rec = append(rec, gc)
		}
		tax.RecGenomeCount = rec

	}
	if fieldComposite == "rank" || fieldComposite == "all" {
		var rank string
		rank, err = neidb.Rank(id)
		if err != nil {
			return
		}
		tax.Rank = rank

	}
	if fieldComposite == "default" || fieldComposite == "all" {
		var sciName string
		sciName, err = neidb.Name(id)
		if err != nil {
			return
		}
		tax.Name = sciName
		var comName string
		comName, err = neidb.CommonName(id)
		if err != nil {
			return
		}
		tax.CommonName = comName

		var parent int
		parent, err = neidb.Parent(id)
		if err != nil {
			return
		}
		tax.Parent = parent

	}
	if fieldComposite == "all" {
		var isLeaf bool
		isLeaf, err = neidb.IsLeaf(id)
		if err != nil {
			return
		}
		tax.IsLeaf = isLeaf

		var neiImages []Image
		var images []tdb.Image
		images, err = neidb.Images(id)
		if err != nil {
			return
		}
		for _, image := range images {
			i := Image{Id: image.Id,
				Url:         image.Url,
				Attribution: image.Attribution}
			neiImages = append(neiImages, i)
		}

	}
	if !plain {
		taxNode := *root.getService("taxon")
		tax.Links = append(tax.Links, taxNode.makeLink("self", fillTemplate(
			taxNode,
			map[string]string{"taxon_id": strconv.Itoa(id)},
			map[string]string{"field_composite": fieldComposite},
		)))

	}
	return
}

func getTaxonCompositeLinks(node *Node, fieldComposite string, r *http.Request) (links []Link) {
	available := []string{"all", "default", "gen_count", "gen_count_rec", "id", "rank"}
	queryName := "field_composite"
	u := *r.URL
	for _, v := range available {
		if v != fieldComposite {
			q := u.Query()
			q.Set(queryName, v)
			u.RawQuery = q.Encode()
			links = append(links, node.makeLink("more fields", u.String()))

		}
	}
	return
}

func taxon(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	fieldComposite := r.URL.Query().Get("field_composite")
	fieldComposite = parseFieldComposite(fieldComposite)

	taxIdStr := r.PathValue("taxon_id")
	taxId, err := strconv.Atoi(taxIdStr)
	if err != nil {
		writeBadRequestResp(w, "Taxon id is not an integer.")
		return
	}

	tax, err := getTaxonData(taxId, true, fieldComposite, neidb)
	if err != nil {
		writeServerError(w, "fn taxon - Error from getTaxonData", "", err)
		return
	}
	var out ResponseBody[Taxon]
	out.Data = tax

	var links []Link
	links = append(links, getTaxonCompositeLinks(selfNode, fieldComposite, r)...)
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

func children(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	taxIdStr := r.PathValue("taxon_id")
	taxId, err := strconv.Atoi(taxIdStr)
	if err != nil {
		writeBadRequestResp(w, "Taxon id is not an integer.")
		return
	}

	offset, limit := extractPaging(r)

	fieldComposite := r.URL.Query().Get("field_composite")
	fieldComposite = parseFieldComposite(fieldComposite)

	children, err := neidb.Children(taxId)
	if err != nil {
		writeServerError(w, "fn children - Error while executing neidbChildren", "", err)
		return
	}
	if limit == -1 {
		limit = len(children)
	}
	data := []Taxon{}
	for i := offset; i < min(offset+limit, len(children)); i++ {
		id := children[i]
		tax, err := getTaxonData(id, plain, fieldComposite, neidb)
		if err != nil {
			writeServerError(w, "Error from getTaxonData", "", err)
			return
		}
		data = append(data, tax)

	}

	out := ResponseBody[[]Taxon]{Data: data}

	pathParams := map[string]string{"taxon_id": taxIdStr}
	queryParams := map[string]string{}

	if !plain {
		var links []Link
		links = append(links, selfNode.makeLink("self", r.URL.String()))

		for link := range selfNode.Links {
			for _, node := range selfNode.Links[link] {
				links = append(links, node.makeLink(link, fillTemplate(node, pathParams, queryParams)))
			}
		}
		links = append(links, getTaxonCompositeLinks(selfNode, fieldComposite, r)...)

		out.Links = links
	}

	if plain {
		writeJsonOutput(w, out.Data)
	} else {
		writeJsonOutput(w, out)
	}

}

func parent(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	taxIdStr := r.PathValue("taxon_id")
	taxId, err := strconv.Atoi(taxIdStr)
	if err != nil {
		writeBadRequestResp(w, "Taxon id is not an integer.")
		return
	}

	fieldComposite := r.URL.Query().Get("field_composite")
	fieldComposite = parseFieldComposite(fieldComposite)

	parent, err := neidb.Parent(taxId)
	tax, err := getTaxonData(parent, true, fieldComposite, neidb)
	if err != nil {
		writeServerError(w, "fn parent - Error from getTaxonData", "", err)
		return
	}
	var out ResponseBody[Taxon]
	out.Data = tax

	var links []Link
	links = append(links, getTaxonCompositeLinks(selfNode, fieldComposite, r)...)
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

func subtree(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	taxIdStr := r.PathValue("taxon_id")
	taxId, err := strconv.Atoi(taxIdStr)
	if err != nil {
		writeBadRequestResp(w, "Taxon id is not an integer.")
		return
	}

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	offset, limit := extractPaging(r)

	fieldComposite := r.URL.Query().Get("field_composite")
	fieldComposite = parseFieldComposite(fieldComposite)

	taxa, err := neidb.Subtree(taxId)
	if err != nil {
		writeServerError(w, "fn subtree - Error while executing neidb.Subtree", "", err)
		return
	}

	if limit == -1 {
		limit = len(taxa)
	}

	data := []Taxon{}
	for i := offset; i < min(offset+limit, len(taxa)); i++ {
		id := taxa[i]
		tax, err := getTaxonData(id, plain, fieldComposite, neidb)
		if err != nil {
			writeServerError(w, "Error from getTaxonData", "", err)
			return
		}
		data = append(data, tax)

	}

	out := ResponseBody[[]Taxon]{Data: data}

	if !plain {
		var links []Link
		links = append(links, selfNode.makeLink("self", r.URL.String()))

		for link := range selfNode.Links {
			for _, node := range selfNode.Links[link] {
				links = append(links, node.makeLink(link, ""))
			}
		}

		links = append(links, getTaxonCompositeLinks(selfNode, fieldComposite, r)...)

		out.Links = links
	}

	if plain {
		writeJsonOutput(w, out.Data)
	} else {
		writeJsonOutput(w, out)
	}

}

func taxonomy(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	out := ResponseBody[*any]{}
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	var links []Link
	links = append(links, selfNode.makeLink("self", r.URL.String()))

	for link := range selfNode.Links {
		for _, node := range selfNode.Links[link] {
			links = append(links, node.makeLink(link, ""))
		}
	}

	out.Links = links

	writeJsonOutput(w, out)
}

func taxonAccessions(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	offset, limit := extractPaging(r)

	valid := checkParams(w, r, "taxon_ids")
	if !valid {
		return
	}
	idsStr := r.URL.Query().Get("taxon_ids")
	var taxIds []int
	split := strings.Split(idsStr, ",")
	for _, id := range split {
		p, err := strconv.Atoi(id)
		if err != nil {
			writeBadRequestResp(w, "At least one taxon id is not an integer.")
			return
		}
		taxIds = append(taxIds, p)
	}

	data := []TaxonAccession{}
	taxNode := *root.getService("taxon")
	accNode := *root.getService("accession")

	for len(taxIds) > 0 && (limit == -1 || len(data) < limit) {
		taxId := taxIds[0]
		taxIds = taxIds[1:]
		accs, err := neidb.Accessions(taxId)
		if err != nil {
			writeServerError(w, "fn taxonAccessions - Error while executing neidb.Accessions", "", err)
			return
		}
		if len(accs) > 0 {
			if offset > 0 {
				offset--
			} else {
				d := TaxonAccession{TaxId: taxId}
				if !plain {
					d.Links = append(d.Links, taxNode.makeLink("taxon", fillTemplate(
						taxNode,
						map[string]string{"taxon_id": strconv.Itoa(taxId)},
						map[string]string{},
					)))
				}

				for _, acc := range accs {
					level, err := neidb.Level(acc)
					if err != nil {
						writeServerError(w, "fn taxonAccessions - Error while executing neidb.Level", "", err)
						return
					}
					accession := Accession{Accession: acc, Level: level}
					if !plain {
						accession.Links = append(accession.Links, taxNode.makeLink("self", fillTemplate(
							accNode,
							map[string]string{"accession_id": acc},
							map[string]string{},
						)))
					}

					d.Accessions = append(d.Accessions, accession)
				}
				data = append(data, d)
			}

		}
		children, err := neidb.Children(taxId)
		if err != nil {
			writeServerError(w, "fn taxonAccessions - Error while executing neidb.Children", "", err)
			return
		}
		for _, child := range children {
			taxIds = append(taxIds, child)
		}

	}
	out := ResponseBody[[]TaxonAccession]{Data: data}

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

func mrca(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	fieldComposite := r.URL.Query().Get("field_composite")
	fieldComposite = parseFieldComposite(fieldComposite)

	valid := checkParams(w, r, "taxon_ids")
	if !valid {
		return
	}
	idsStr := r.URL.Query().Get("taxon_ids")
	var taxIds []int
	split := strings.Split(idsStr, ",")
	for _, id := range split {
		p, err := strconv.Atoi(id)
		if err != nil {
			writeBadRequestResp(w, "At least one taxon id is not an integer.")
			return
		}
		taxIds = append(taxIds, p)
	}

	var out ResponseBody[Taxon]
	if len(taxIds) > 0 {
		id, err := neidb.MRCA(taxIds)
		if err != nil {
			writeServerError(w, "fn mrca - Error while executing neidb.MRCA", "", err)
			return
		}

		tax, err := getTaxonData(id, plain, fieldComposite, neidb)
		if err != nil {
			writeServerError(w, "fn mrca - Error from getTaxonData", "", err)
			return
		}
		out.Data = tax

		var links []Link
		links = append(links, getTaxonCompositeLinks(selfNode, fieldComposite, r)...)
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

func path(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	strPlain := r.URL.Query().Get("plain_data")
	plain, err := strconv.ParseBool(strPlain)
	if strPlain != "" && err != nil {
		writeBadRequestResp(w, "plain_data is not a bool.")
		return
	}

	offset, limit := extractPaging(r)

	fieldComposite := r.URL.Query().Get("field_composite")
	fieldComposite = parseFieldComposite(fieldComposite)

	startStr := r.URL.Query().Get("start_id")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		writeBadRequestResp(w, "start id is not an integer.")
		return
	}
	endStr := r.URL.Query().Get("end_id")
	end, err := strconv.Atoi(endStr)
	if err != nil {
		writeBadRequestResp(w, "end id is not an integer.")
		return
	}

	var data []Taxon
	for i := 0; (i < offset+limit || limit == -1) && start != end; i++ {
		parent, err := neidb.Parent(start)

		if err != nil {
			writeServerError(w, "fn path - Error while calling neidb.Parent", "", err)
			return
		}
		if start == parent {
			data = data[:0]
			break
		}

		start = parent
		tax, err := getTaxonData(start, plain, fieldComposite, neidb)
		if err != nil {
			writeServerError(w, "fn path - Error from getTaxonData", "", err)
			return
		}
		data = append(data, tax)

	}
	out := ResponseBody[[]Taxon]{Data: data}

	var links []Link
	links = append(links, getTaxonCompositeLinks(selfNode, fieldComposite, r)...)
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

func programsDoc(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	out := ResponseBody[*any]{}
	_, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Server does not provide any of the accepted content types."))
		return
	}

	var links []Link
	links = append(links, selfNode.makeLink("self", r.URL.String()))

	for link := range selfNode.Links {
		for _, node := range selfNode.Links[link] {
			links = append(links, node.makeLink(link, ""))
		}
	}

	out.Links = links

	writeJsonOutput(w, out)
}

func programEndpoint(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
	dbPath := args[0].(string)
	progName := args[1].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	callArgs := r.URL.Query()["options"]
	if progName == "fintac" {
		callArgs = append(callArgs, "-H")
	}
	callArgs = append(callArgs, "../"+dbPath)

	callArgs = append(callArgs, r.URL.Query()["extra"]...)
	callArgs = slices.DeleteFunc(callArgs, func(w string) bool { return w == "-r" })

	dir := strconv.FormatInt(time.Now().UnixNano(), 10)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		writeServerError(w, "Failed to open directory.", err.Error(), err)
		return
	}
	defer os.RemoveAll(dir)
	var stdinData string
	if r.Method == http.MethodPost {
		ct, err := contenttype.GetMediaType(r)
		if err != nil {
			writeBadRequestResp(w, "Malformed content type.")
			return
		}
		if !ct.EqualsMIME(multipartCt) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("Use multipart/form-data with POST requests."))
			return
		} else {
			stdinData, err = parseFormData(w, r, dir, callArgs)
			if err != nil {
				return
			}
		}
	}

	cmd := exec.CommandContext(ctx, "../prog/"+progName, callArgs...)
	cmd.Stdin = strings.NewReader(stdinData)
	cmd.Dir = dir
	fmt.Println(cmd)
	res, err := cmd.CombinedOutput()

	if err != nil {
		writeServerError(w, "Failed while calling neighbors program.", err.Error(), err)
	} else {
		w.Header().Set("Content-Type", plainCt.String())
		w.Write(res)
	}
}

func parseFormData(w http.ResponseWriter, r *http.Request, dir string, callArgs []string) (stdin string, err error) {
	err = r.ParseMultipartForm(3_000_000)
	if err != nil {
		writeBadRequestResp(w, "Malformed multipart form request.")
		return "", err
	}
	if len(r.MultipartForm.File) == 0 {
		err = errors.New("Please provide files, when sending multipart/form-data request.")
		writeBadRequestResp(w, err.Error())
		return "", err
	} else {
		keys := []string{}
		for key := range r.MultipartForm.File {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for i, key := range keys {
			mFiles := r.MultipartForm.File[key]
			if len(mFiles) > 0 {
				h := *mFiles[0]
				var rf multipart.File
				rf, err = h.Open()
				if err != nil {
					writeServerError(w, "Error while opening multipart file", "", err)
					return
				}
				defer rf.Close()
				b := make([]byte, h.Size)
				rf.Read(b)
				if h.Filename == "stdin" {
					stdin = string(b)
				} else {
					var tf *os.File
					tf, err = os.OpenFile(dir+"/"+strconv.Itoa(i), os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						writeServerError(w, "Error while opening temp file", "", err)
						return
					}
					cd := h.Header.Get("Content-Disposition")
					var params map[string]string
					_, params, err = mime.ParseMediaType(cd)
					if err != nil {
						writeServerError(w, "Error while parsing media type of multipart file", "", err)
						return
					}
					filename := params["filename"]

					for j, arg := range callArgs {
						if j != 0 && (arg == h.Filename || arg == filename) {
							callArgs[j] = strconv.Itoa(i)
						}
					}

					buffer := bytes.NewBuffer(b)
					buffer.WriteTo(tf)
					tf.Close()
				}

			}
		}
		return

	}
}
