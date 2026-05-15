package apiv2

import (
	"net/http"

	"github.com/evolbioinf/neighbors/tdb"

	"strings"

	"strconv"

	"encoding/json"
	"fmt"
	"github.com/evolbioinf/never/util"
)

type Accession struct {
	Accession string `json:"accession"`
	Level     string `json:"level"`
}

type GenomeCount struct {
	Level string `json:"level"`
	Count int    `json:"count"`
}

type Image struct {
	Id          int    `json:"id"`
	Url         string `json:"url"`
	Attribution string `json:"attribution"`
}

type Rank struct {
	TaxId int    `json:"tax_id"`
	Rank  string `json:"rank"`
}

type TaxonAccessions struct {
	TaxId      int         `json:"tax_id"`
	Accessions []Accession `json:"accessions"`
}

type TaxId struct {
	TaxId int `json:"tax_id"`
}

type TaxInfo struct {
	TaxId          int           `json:"tax_id"`
	Parent         int           `json:"parent"`
	IsLeaf         int           `json:"is_leaf"`
	Name           string        `json:"name"`
	CommonName     string        `json:"common_name"`
	Rank           string        `json:"rank"`
	RawGenomeCount []GenomeCount `json:"raw_genome_count"`
	RecGenomeCount []GenomeCount `json:"rec_genome_count"`
	Images         []Image       `json:"images"`
}

type Taxon struct {
	TaxId      int    `json:"tax_id"`
	Parent     int    `json:"parent"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
}

type TaxonName struct {
	TaxId      int    `json:"tax_id"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
}

func RegisterRoutes(prefix string) {
	neidb := tdb.OpenTaxonomyDB("neidb")

	makeRoute(prefix+"/accessions", accessions, neidb)          // formerly known as levels
	makeRoute(prefix+"/assembly-levels", assemblyLevels, neidb) // new
	makeRoute(prefix+"/taxa-accessions", taxaAccessions, neidb) // formerly known as accessions
	makeRoute(prefix+"/ranks", ranks, neidb)                    // same as before
	makeRoute(prefix+"/taxa", taxa, neidb)                      // formerly known as taxi

}

func makeRoute(path string, fn func(http.ResponseWriter, *http.Request, ...any), args ...any) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) { fn(w, r, args...) })
}

func accessions(w http.ResponseWriter, r *http.Request, args ...any) {
	valid := checkParams(w, r, "accession-ids")
	if !valid {
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	str := r.URL.Query().Get("accession-ids")
	accessions := strings.Split(str, ",")

	offset, size := extractPaging(r)

	out := []Accession{}

	if size == -1 {
		size = len(accessions)
	}

	for i := offset; i < min(offset+size, len(accessions)); i++ {
		accession := accessions[i]
		level, err := neidb.Level(accession)
		if err == nil {
			out = append(out, Accession{Accession: accession, Level: level})
		}
	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
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
	strPageSize := r.URL.Query().Get("page-size")

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

func assemblyLevels(w http.ResponseWriter, r *http.Request, args ...any) {
	out := tdb.AssemblyLevels()
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))

}

func ranks(w http.ResponseWriter, r *http.Request, args ...any) {
	valid := checkParams(w, r, "taxon-ids")
	if !valid {
		return
	}
	neidb := args[0].(*tdb.TaxonomyDB)

	str := r.URL.Query().Get("taxon-ids")
	strTaxa := strings.Split(str, ",")
	var taxIds []int
	for _, strT := range strTaxa {
		t, err := strconv.Atoi(strT)
		if err == nil {
			taxIds = append(taxIds, t)
		}
	}

	offset, size := extractPaging(r)

	if size == -1 {
		size = len(taxIds)
	}

	taxa := getTaxa(taxIds[offset:min(size, len(taxIds))], neidb)

	out := []Rank{}
	for i, taxon := range taxa {
		rank, err := neidb.Rank(taxon)
		util.Check(err)
		o := Rank{TaxId: taxa[i], Rank: rank}

		out = append(out, o)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	fmt.Fprintf(w, "%s\n", string(b))

}

func getTaxa(taxonIds []int, neidb *tdb.TaxonomyDB) []int {
	taxa := []int{}
	for _, taxon := range taxonIds {
		_, err := neidb.Name(taxon)
		if err == nil {
			taxa = append(taxa, taxon)
		}
	}

	return taxa
}

func taxaAccessions(w http.ResponseWriter, r *http.Request, args ...any) {
	valid := checkParams(w, r, "taxon-ids")
	if !valid {
		return
	}
	neidb := args[0].(*tdb.TaxonomyDB)

	str := r.URL.Query().Get("taxon-ids")
	strTaxa := strings.Split(str, ",")
	var taxIds []int
	for _, strT := range strTaxa {
		t, err := strconv.Atoi(strT)
		if err == nil {
			taxIds = append(taxIds, t)
		}
	}

	offset, size := extractPaging(r)

	if size == -1 {
		size = len(taxIds)
	}

	taxa := getTaxa(taxIds[offset:min(size, len(taxIds))], neidb)

	out := []TaxonAccessions{}
	for len(taxa) > 0 {
		taxId := taxa[0]
		taxa = taxa[1:]
		accs, err := neidb.Accessions(taxId)
		util.Check(err)
		if len(accs) > 0 {
			o := TaxonAccessions{TaxId: taxId}
			for _, acc := range accs {
				level, err := neidb.Level(acc)
				util.Check(err)
				accession := Accession{Accession: acc, Level: level}
				o.Accessions = append(o.Accessions, accession)
			}

			out = append(out, o)

		}
		children, err := neidb.Children(taxId)
		for _, child := range children {
			taxa = append(taxa, child)
		}

	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
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
	fmt.Fprintf(w, "%s\n", string(b))

}
