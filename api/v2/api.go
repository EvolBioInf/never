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

	makeRoute(prefix+"/accessions", accessions, neidb)
	makeRoute(prefix+"/taxa-accessions", taxaAccessions, neidb)

}

func makeRoute(path string, fn func(http.ResponseWriter, *http.Request, ...any), args ...any) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) { fn(w, r, args...) })
}

func accessions(w http.ResponseWriter, r *http.Request, args ...any) {
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

func taxaAccessions(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	str := r.URL.Query().Get("taxon-ids")
	strTaxa := strings.Split(str, ",")
	var taxIds []int
	for _, strT := range strTaxa {
		t, err := strconv.Atoi(strT)
		if err != nil {
			taxIds = append(taxIds, t)
		}
	}

	offset, size := extractPaging(r)

	if size == -1 {
		size = len(taxIds) - 1
	}

	taxa := getTaxa(w, r, taxIds[offset:min(size, len(taxIds)-1)], neidb)
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

func getTaxa(w http.ResponseWriter, r *http.Request, taxonIds []int, neidb *tdb.TaxonomyDB) []int {
	taxa := []int{}
	for _, taxon := range taxonIds {
		_, err := neidb.Name(taxon)
		if err == nil {
			taxa = append(taxa, taxon)
		}
	}

	return taxa
}
