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

type TaxonInfo struct {
	TaxId          int           `json:"tax_id"`
	Parent         int           `json:"parent"`
	IsLeaf         bool          `json:"is_leaf"`
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

	makeRoute(prefix+"/accessions", accessions, neidb)                                     // formerly known as levels
	makeRoute(prefix+"/assembly-levels", assemblyLevels, neidb)                            // new
	makeRoute(prefix+"/taxa-accessions", taxaAccessions, neidb)                            // formerly known as accessions
	makeRoute(prefix+"/ranks", ranks, neidb)                                               // same as before
	makeRoute(prefix+"/taxa", taxa, neidb)                                                 // formerly known as taxi
	makeRoute(prefix+"/taxa-count", taxaCount, neidb)                                      // new
	makeRoute(prefix+"/taxa-info", taxaInfo, neidb)                                        // same as before
	makeRoute(prefix+"/taxa-names", taxaNames, neidb)                                      // formerly known as names
	makeRoute(prefix+"/taxa/{start_id}/path/{end_id}", taxaPath, neidb)                    // formerly just path
	makeRoute(prefix+"/taxa/{taxon_id}/children", taxaChildren, neidb)                     // formerly just children
	makeRoute(prefix+"/taxa/{taxon_id}/images", taxaImages, neidb)                         // new
	makeRoute(prefix+"/taxa/{taxon_id}/genome-count", taxaGenomeCount, neidb)              // formerly known as num_genomes
	makeRoute(prefix+"/taxa/{taxon_id}/genome-count-recursive", taxaGenomeCountRec, neidb) // formerly known as num_genomes_rec
	makeRoute(prefix+"/taxa/{taxon_id}/parent", taxaParent, neidb)                         // formerly known as parent
	makeRoute(prefix+"/taxa/{taxon_id}/subtree", taxaSubtree, neidb)                       // formerly just subtree
	makeRoute(prefix+"/taxa/{taxon_ids}/mrca", taxaMRCA, neidb)                            // formerly just mrca

}

func makeRoute(path string, fn func(http.ResponseWriter, *http.Request, ...any), args ...any) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) { fn(w, r, args...) })
}

func accessions(w http.ResponseWriter, r *http.Request, args ...any) {
	valid := checkParams(w, r, "accession_ids")
	if !valid {
		return
	}

	neidb := args[0].(*tdb.TaxonomyDB)

	str := r.URL.Query().Get("accession_ids")
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

func assemblyLevels(w http.ResponseWriter, r *http.Request, args ...any) {
	out := tdb.AssemblyLevels()
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func ranks(w http.ResponseWriter, r *http.Request, args ...any) {
	valid := checkParams(w, r, "taxon_ids")
	if !valid {
		return
	}
	neidb := args[0].(*tdb.TaxonomyDB)

	str := r.URL.Query().Get("taxon_ids")
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

	taxa := getTaxa(taxIds[offset:min(offset+size, len(taxIds))], neidb)

	out := []Rank{}
	for i, taxon := range taxa {
		rank, err := neidb.Rank(taxon)
		util.Check(err)
		o := Rank{TaxId: taxa[i], Rank: rank}

		out = append(out, o)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
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
	valid := checkParams(w, r, "taxon_ids")
	if !valid {
		return
	}
	neidb := args[0].(*tdb.TaxonomyDB)

	str := r.URL.Query().Get("taxon_ids")
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

	taxa := getTaxa(taxIds[offset:min(offset+size, len(taxIds))], neidb)

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

func taxaCount(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	res, err := neidb.NumTaxa()
	out := struct {
		NumTaxa int `json:"num_taxa"`
	}{NumTaxa: res}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaInfo(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	valid := checkParams(w, r, "taxon_ids")
	if !valid {
		return
	}

	offset, size := extractPaging(r)

	str := r.URL.Query().Get("taxon_ids")
	strTaxa := strings.Split(str, ",")
	var taxIds []int
	for _, strT := range strTaxa {
		t, err := strconv.Atoi(strT)
		if err == nil {
			taxIds = append(taxIds, t)
		}
	}

	if size == -1 {
		size = len(taxIds)
	}

	taxa := getTaxa(taxIds[offset:min(offset+size, len(taxIds))], neidb)

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

		var neiImages []Image
		images, err := neidb.Images(taxon)
		util.Check(err)

		for _, image := range images {
			i := Image{Id: image.Id,
				Url:         image.Url,
				Attribution: image.Attribution}
			neiImages = append(neiImages, i)
		}

		o := TaxonInfo{
			TaxId:          taxon,
			Parent:         parent,
			IsLeaf:         isLeaf,
			Name:           name,
			CommonName:     cname,
			Rank:           rank,
			RawGenomeCount: raw,
			RecGenomeCount: rec,
			Images:         neiImages,
		}

		out = append(out, o)

	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaNames(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	valid := checkParams(w, r, "taxon_ids")
	if !valid {
		return
	}

	offset, size := extractPaging(r)

	str := r.URL.Query().Get("taxon_ids")
	strTaxa := strings.Split(str, ",")
	var taxIds []int
	for _, strT := range strTaxa {
		t, err := strconv.Atoi(strT)
		if err == nil {
			taxIds = append(taxIds, t)
		}
	}

	if size == -1 {
		size = len(taxIds)
	}

	taxa := getTaxa(taxIds[offset:min(offset+size, len(taxIds))], neidb)

	out := []TaxonName{}
	for _, taxon := range taxa {
		name, err := neidb.Name(taxon)
		util.Check(err)
		cname, err := neidb.CommonName(taxon)
		util.Check(err)
		o := TaxonName{TaxId: taxon, Name: name, CommonName: cname}
		out = append(out, o)

	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaPath(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	offset, size := extractPaging(r)

	strStartTaxon := r.PathValue("start_id")
	start, err := strconv.Atoi(strStartTaxon)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Can't find parameter's taxa."))
		return
	}

	strEndTaxon := r.PathValue("end_id")
	end, err := strconv.Atoi(strEndTaxon)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Can't find parameter's taxa."))
		return
	}

	parent, err := neidb.Parent(start)
	util.Check(err)
	out := []Taxon{}

	if parent == start && start != end {
		b, err := json.MarshalIndent(out, "", "  ")
		util.Check(err)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s\n", string(b))

		return
	}

	name, err := neidb.Name(start)
	util.Check(err)
	cn, err := neidb.CommonName(start)
	util.Check(err)
	o := Taxon{
		TaxId:      start,
		Parent:     parent,
		Name:       name,
		CommonName: cn,
	}
	out = append(out, o)

	for i := 0; (i < offset+size || size == -1) && start != end; i++ {
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
		o := Taxon{
			TaxId:      start,
			Parent:     parent,
			Name:       name,
			CommonName: cname,
		}

		out = append(out, o)

	}

	out = out[1:]

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaChildren(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	strTaxonId := r.PathValue("taxon_id")
	taxId, _ := strconv.Atoi(strTaxonId)

	offset, size := extractPaging(r)

	children, err := neidb.Children(taxId)
	util.Check(err)
	out := []TaxonName{}
	if size == -1 {
		size = len(children)
	}

	for i := offset; i < min(offset+size, len(children)); i++ {
		child := children[i]
		name, err := neidb.Name(child)
		util.Check(err)
		cname, err := neidb.CommonName(child)
		util.Check(err)
		o := TaxonName{child, name, cname}
		out = append(out, o)

	}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaImages(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	strTaxonId := r.PathValue("taxon_id")
	taxId, _ := strconv.Atoi(strTaxonId)

	offset, size := extractPaging(r)

	images, err := neidb.Images(taxId)
	util.Check(err)

	if size == -1 {
		size = len(images)
	}

	out := []Image{}
	for i := offset; i < min(offset+size, len(images)); i++ {
		image := images[i]
		o := Image{
			Id:          image.Id,
			Url:         image.Url,
			Attribution: image.Attribution,
		}
		out = append(out, o)

	}
	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaGenomeCount(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	strTaxonId := r.PathValue("taxon_id")
	taxId, _ := strconv.Atoi(strTaxonId)

	offset, size := extractPaging(r)

	out := []GenomeCount{}
	for _, level := range tdb.AssemblyLevels() {
		n, err := neidb.NumGenomes(taxId, level)
		if err == nil {
			o := GenomeCount{Count: n, Level: level}
			out = append(out, o)
		}
	}

	if size == -1 {
		size = len(out)
	}

	out = out[offset:min(offset+size, len(out))]

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaGenomeCountRec(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	strTaxonId := r.PathValue("taxon_id")
	taxId, _ := strconv.Atoi(strTaxonId)

	offset, size := extractPaging(r)

	out := []GenomeCount{}
	for _, level := range tdb.AssemblyLevels() {
		n, err := neidb.NumGenomesRec(taxId, level)
		if err == nil {
			o := GenomeCount{Count: n, Level: level}
			out = append(out, o)
		}
	}

	if size == -1 {
		size = len(out)
	}

	out = out[offset:min(offset+size, len(out))]

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaParent(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	strTaxonId := r.PathValue("taxon_id")
	taxId, _ := strconv.Atoi(strTaxonId)

	parent, err := neidb.Parent(taxId)
	out := TaxId{0}
	if err == nil {
		out = TaxId{parent}
	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaSubtree(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	strTaxonId := r.PathValue("taxon_id")
	taxId, _ := strconv.Atoi(strTaxonId)

	offset, size := extractPaging(r)

	taxa, err := neidb.Subtree(taxId)
	util.Check(err)

	if size == -1 {
		size = len(taxa)
	}

	out := []Taxon{}
	for i := offset; i < min(offset+size, len(taxa)); i++ {
		taxon := taxa[i]
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

		o := Taxon{TaxId: taxon, Parent: parent, Name: name,
			CommonName: cname}
		out = append(out, o)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}

func taxaMRCA(w http.ResponseWriter, r *http.Request, args ...any) {
	neidb := args[0].(*tdb.TaxonomyDB)

	out := TaxId{0}
	strTaxonIds := r.PathValue("taxon_ids")
	split := strings.Split(strTaxonIds, ",")
	taxa := []int{}
	for _, str := range split {
		id, err := strconv.Atoi(str)
		if err != nil {
			return
		}

		taxa = append(taxa, id)
	}
	if len(taxa) > 0 {
		mrca, err := neidb.MRCA(taxa)
		if err == nil {
			out = TaxId{mrca}
		}
	}

	b, err := json.MarshalIndent(out, "", "  ")
	util.Check(err)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(b))

}
