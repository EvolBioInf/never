package apiv2

import "net/http"

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

func v(p string, fn func(func(http.ResponseWriter, *http.Request, ...any)), sdfsdf int)
