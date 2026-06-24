package apiv2
import (
          "github.com/elnormous/contenttype"

  "net/http"

  "github.com/evolbioinf/neighbors/tdb"
  "log"

  "strings"

  "strconv"

  "errors"
          "runtime/debug"

  "encoding/json"
  "fmt"
  "github.com/evolbioinf/never/util"

  "slices"

  "github.com/evolbioinf/neighbors/ants"

  "sync"

  "flag"

  "io"

  "sort"

          "mime/multipart"

  "os"
  "time"

  "bytes"

          "github.com/evolbioinf/neighbors/ranks"



          "github.com/evolbioinf/neighbors/dree"

          fintacPack "github.com/evolbioinf/neighbors/fintac"



          neighborsPack "github.com/evolbioinf/neighbors/neighbors"




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
    Rel: rel,
    Href: node.BasePath,
    Action: node.Action,
    Types: types,
  }
}

func (node *Node) getService (name string) *Node {
  services := node.Links["service"]
  idx := slices.IndexFunc(services, func(a Node) bool { return a.Name == name})

  if idx != -1 {
    return &services[idx]
  } else {
    return nil
  }
}

var root Node

var prefix string

var progMu sync.Mutex

        var jsonCt = contenttype.MediaType{Type: "application", Subtype: "json", Parameters: contenttype.Parameters{"charset": "utf-8"}}
        var graphvizCt = contenttype.MediaType{Type: "text", Subtype: "vnd.graphviz", Parameters: contenttype.Parameters{"charset": "utf-8"}}
        var plainCt = contenttype.MediaType{Type: "text", Subtype: "plain", Parameters: contenttype.Parameters{"charset": "utf-8"}}

var multipartCt = contenttype.MediaType{Type: "multipart", Subtype: "form-data"}

func RegisterRoutes(pref, dbPath string) {
  var neidb *tdb.TaxonomyDB
  neidb, err := tdb.OpenTaxonomyDBcheck(dbPath)
  if err != nil {
    log.Fatal("apiV2: error while opening the database: ",  err.Error())
  }

  prefix = pref

          get := "GET"
          post := "POST"

          rootDocL := Node{
                  Links:    make(map[string][]Node),
    Name:     "rootDocument",
                  BasePath: prefix,
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&rootDocL, rootDocument, neidb) // new

          accessionsL := Node{
                  Links:    make(map[string][]Node),
    Name:     "accessions",
                  BasePath: prefix + "/accessions",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&accessionsL, accessions, neidb) // previously known as levels

          accessionL := Node{
                  Links:    make(map[string][]Node),
    Name:     "accession",
                  BasePath: prefix + "/accessions/{accession_id}",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&accessionL, accession, neidb) // new

          taxaL := Node{
                  Links:    make(map[string][]Node),
    Name:     "taxa",
                  BasePath: prefix + "/taxa",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&taxaL, taxa, neidb) // previously known as taxi

          taxonL := Node{
                  Links:    make(map[string][]Node),
    Name:     "taxon",
                  BasePath: prefix + "/taxa/{taxon_id}",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&taxonL, taxon, neidb) // new

          ancestorsL := Node{
                  Links:    make(map[string][]Node),
    Name:     "ancestors",
                  BasePath: prefix + "/taxa/{taxon_id}/ancestors",
                  Action:   get,
                  Types:    []contenttype.MediaType{plainCt},
          }
          makeRoute(&ancestorsL, ancestors, dbPath) // new - calls ants program

          childrenL := Node{
                  Links:    make(map[string][]Node),
    Name:     "children",
                  BasePath: prefix + "/taxa/{taxon_id}/children",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&childrenL, children, neidb) // previously just children

          parentL := Node{
                  Links:    make(map[string][]Node),
    Name:     "parent",
                  BasePath: prefix + "/taxa/{taxon_id}/parent",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&parentL, parent, neidb) // previously known as parent

          rankDistL := Node{
                  Links:    make(map[string][]Node),
    Name:     "rankDistribution",
                  BasePath: prefix + "/taxa/{taxon_id}/rank_distribution",
                  Action:   post,
                  Types:    []contenttype.MediaType{plainCt},
          }
          makeRoute(&rankDistL, rankDistribution, dbPath) // new - calls ranks program

          subtreeL := Node{
                  Links:    make(map[string][]Node),
    Name:     "subtree",
                  BasePath: prefix + "/taxa/{taxon_id}/subtree",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt, graphvizCt, plainCt},
          }
          makeRoute(&subtreeL, subtree, neidb, dbPath) // previously just subtree

          taxonomyL := Node{
                  Links:    make(map[string][]Node),
    Name:     "taxonomy",
                  BasePath: prefix + "/taxonomy",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&taxonomyL, taxonomy, neidb) // new

          taxonAccessionsL := Node{
                  Links:    make(map[string][]Node),
    Name:     "taxonAccessions",
                  BasePath: prefix + "/taxonomy/accessions",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&taxonAccessionsL, taxonAccessions, neidb) // previously called accessions

          fintacL := Node{
                  Links:    make(map[string][]Node),
    Name:     "fintac",
                  BasePath: prefix + "/taxonomy/fintac",
                  Action:   post,
                  Types:    []contenttype.MediaType{plainCt},
          }
          makeRoute(&fintacL, fintac, dbPath) // new - calls fintac program

          mrcaL := Node{
                  Links:    make(map[string][]Node),
    Name:     "mrca",
                  BasePath: prefix + "/taxonomy/mrca",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&mrcaL, mrca, neidb) // previously just mrca

          neighborsL := Node{
                  Links:    make(map[string][]Node),
    Name:     "neighbors",
                  BasePath: prefix + "/taxonomy/neighbors",
                  Action:   get,
                  Types:    []contenttype.MediaType{plainCt},
          }
          makeRoute(&neighborsL, neighbors, dbPath) // new - calls neighbors program

          pathL := Node{
                  Links:    make(map[string][]Node),
    Name:     "path",
                  BasePath: prefix + "/taxonomy/path",
                  Action:   get,
                  Types:    []contenttype.MediaType{jsonCt},
          }
          makeRoute(&pathL, path, neidb) // previously just path


          rootDocL.Links["service"] = append(rootDocL.Links["service"], accessionsL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], accessionL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], taxaL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], taxonL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], ancestorsL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], childrenL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], parentL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], rankDistL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], subtreeL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], taxonomyL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], taxonAccessionsL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], fintacL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], mrcaL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], neighborsL)
          rootDocL.Links["service"] = append(rootDocL.Links["service"], pathL)
          accessionsL.Links["service"] = append(accessionsL.Links["service"], accessionL)
          taxaL.Links["service"] = append(taxaL.Links["service"], taxonL)
          taxaL.Links["service"] = append(taxaL.Links["service"], ancestorsL)
          taxaL.Links["service"] = append(taxaL.Links["service"], childrenL)
          taxaL.Links["service"] = append(taxaL.Links["service"], parentL)
          taxaL.Links["service"] = append(taxaL.Links["service"], rankDistL)
          taxaL.Links["service"] = append(taxaL.Links["service"], subtreeL)
          taxonomyL.Links["service"] = append(taxonomyL.Links["service"], taxonAccessionsL)
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
func makeRoute(node *Node, fn func(http.ResponseWriter, *http.Request, *Node, ...any), args ...any) {
  http.HandleFunc(node.Action + " " + node.BasePath, func(w http.ResponseWriter, r *http.Request) {
                        fn(w, r, node, args...)
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
  for i := offset; i < min(offset + size, len(accessions)); i++ {
    accession := accessions[i]
    level, err := neidb.Level(accession)
    if err != nil {
      writeServerError(w, "fn accessions - Error while accessing neidb.Level", err)
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

func writeServerError(w http.ResponseWriter, internalMsg string, err error) {
  if err == nil {
    err = errors.New("No error passed")
  }
        util.LogWarningDef(
                  util.WarningEntry{
                          Warning: fmt.Sprintf("Apiv2: Internal server error.\nMessage: %s\nErr: %s\nTrace: %s", internalMsg, err.Error(), debug.Stack()),
                })
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error."))
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
          if err != nil {
                  plain = false
          }

  accession := r.PathValue("accession_id")
  level, err := neidb.Level(accession)
  if err != nil {
    writeServerError(w, "fn accession - Error while accessing neidb.Level", err)
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
  neidb := args[0].(*tdb.TaxonomyDB)

          strPlain := r.URL.Query().Get("plain_data")
          plain, err := strconv.ParseBool(strPlain)
          if err != nil {
                  plain = false
          }

  offset, size := extractPaging(r)

  name := r.URL.Query().Get("name")
  strExact := r.URL.Query().Get("exact")
  strScientific := r.URL.Query().Get("scientific")
  fieldComposite := r.URL.Query().Get("field_composite")
  fieldComposite = parseFieldComposite(fieldComposite)


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
    if err != nil {
      writeServerError(w, "fn taxa - Error while accessing neidb.Taxids", err)
      return
    }
  } else {
    ids, err = neidb.CommonTaxids(name, size, offset)
    if err != nil {
      writeServerError(w, "fn taxa - Error while accessing neidb.CommonTaxids", err)
      return
    }
  }
  data := []Taxon{}
  for _, id := range ids {
    tax, err := getTaxonData(id, plain, fieldComposite, neidb)
    if err != nil {
      writeServerError(w, "Error from getTaxonData", err)
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
  sort.Strings(availableComps)
  fmt.Println(availableComps)
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
      count, err := neidb.NumGenomes(id, level)
      if err != nil {
        
      }
      gc := GenomeCount{Count: count, Level: level}
      raw = append(raw, gc)
    }
    tax.RawGenomeCount = raw

  }
  if fieldComposite == "gen_count_rec" || fieldComposite == "all" {
    var rec []GenomeCount
    for _, level := range tdb.AssemblyLevels() {
      count, err := neidb.NumGenomesRec(id, level)
      if err != nil {
        
      }
      gc := GenomeCount{Count: count, Level: level}
      rec = append(rec, gc)
    }
    tax.RecGenomeCount = rec

  }
  if fieldComposite == "rank" || fieldComposite == "all" {
    rank, err := neidb.Rank(id)
    if err != nil {
      
    }
    tax.Rank = rank

  }
  if fieldComposite == "default" || fieldComposite == "all" {
    sciName, err := neidb.Name(id)
    if err != nil {
      
    }
    tax.Name = sciName
    comName, err := neidb.CommonName(id)
    if err != nil {
      
    }
    tax.CommonName = comName

    parent, err := neidb.Parent(id)
    if err != nil {
      
    }
    tax.Parent = parent


  }
  if fieldComposite == "all" {
    isLeaf, err := neidb.IsLeaf(id)
    if err != nil {
      
    }
    tax.IsLeaf = isLeaf

    var neiImages []Image
    images, err := neidb.Images(id)
    if err != nil {
      
    }
    for _, image := range images {
      i := Image{Id: image.Id,
        Url: image.Url,
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
  available := []string{"id", "rank", "gen_count", "gen_count_rec", "default", "all"}
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
          if err != nil {
                  plain = false
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
    writeServerError(w, "fn taxon - Error from getTaxonData", err)
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

func ancestors(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
  _, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
  if err != nil {
    w.WriteHeader(http.StatusNotAcceptable)
                  w.Write([]byte("Server does not provide any of the accepted content types."))
    return
  }


  taxIdStr := r.PathValue("taxon_id")

  dbPath := args[0].(string)

  callArgs := []string{"./ants", taxIdStr, dbPath}
  out, errMsg := callPackage(callArgs, ants.Run)
  if len(errMsg) > 0 {
    writeServerError(w, "fn ancestors - Error while calling package ants: " + string(errMsg), nil)
    return
  }

  writePlainOutput(w, out)



}

func checkParamInt(w http.ResponseWriter, param string) bool {
  _, err := strconv.Atoi(param)
  if err != nil {
    writeBadRequestResp(w, "Malformed integer.")
  }
  return err != nil
}

func callPackage(callArgs []string, runFn func()) (out, errMsg []byte) {
  progMu.Lock()

  serverArgs := os.Args
  os.Args = callArgs

  oldFlags := flag.CommandLine
  flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

          prevOut := os.Stdout
          rOut, wOut, err := os.Pipe()
          if err != nil {
                  errMsg = append(errMsg, []byte("Error while creating stdout pipe: " + err.Error())...)
          } else {
            os.Stdout = wOut
  }

  prevErr := os.Stderr
          rErr, wErr, err := os.Pipe()
          if err != nil {
                  errMsg = append(errMsg, []byte("Error while creating stderr pipe: " + err.Error())...)
          }
          os.Stderr = wErr

  if err == nil {
    runFn()
  }

  os.Stdout = prevOut
  os.Stderr = prevErr
  if wOut != nil {
    err = wOut.Close()
    if err != nil {
      errMsg = append(errMsg, []byte("Error while closing stdout pipe: " + err.Error())...)
    }
  }
  if wErr != nil {
    wErr.Close()
    if err != nil {
      errMsg = append(errMsg, []byte("Error while closing stderr pipe: " + err.Error())...)
    }
  }

  var outBuf, errBuf bytes.Buffer
          _, err = io.Copy(&outBuf, rOut)
          _, err = io.Copy(&errBuf, rErr)
          out = outBuf.Bytes()
          errMsg = append(errMsg, errBuf.Bytes()...)

          flag.CommandLine = oldFlags

  os.Args = serverArgs

  progMu.Unlock()


  return
}

func writePlainOutput(w http.ResponseWriter, out []byte) {
  w.Header().Set("Content-Type", plainCt.String())
  fmt.Fprintf(w, "%s", out)
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
          if err != nil {
                  plain = false
          }

  taxIdStr := r.PathValue("taxon_id")
  taxId, err := strconv.Atoi(taxIdStr)
  if err != nil {
    writeBadRequestResp(w, "Taxon id is not an integer.")
    return
  }

  offset, size := extractPaging(r)

  fieldComposite := r.URL.Query().Get("field_composite")
  fieldComposite = parseFieldComposite(fieldComposite)

    children, err := neidb.Children(taxId)
            if err != nil {
      writeServerError(w, "fn children - Error while executing neidbChildren", err)
      return
    }
    if size == -1 {
      size = len(children)
    }
    data := []Taxon{}
    for i := offset; i < min(offset + size, len(children)); i++ {
      id := children[i]
                    tax, err := getTaxonData(id, plain, fieldComposite, neidb)
                    if err != nil {
                      writeServerError(w, "Error from getTaxonData", err)
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
          if err != nil {
                  plain = false
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
    writeServerError(w, "fn parent - Error from getTaxonData", err)
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

func rankDistribution(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
  _, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
  if err != nil {
    w.WriteHeader(http.StatusNotAcceptable)
                  w.Write([]byte("Server does not provide any of the accepted content types."))
    return
  }


  taxIdStr := r.PathValue("taxon_id")

  dbPath := args[0].(string)

          ranksArgs := []string{"./ranks"}
          levels := r.URL.Query().Get("assembly_levels")
  if levels != "" {
                  ranksArgs = append(ranksArgs, "-L", levels)
  }

          listStr := r.URL.Query().Get("list_genomes")
          l, err := strconv.ParseBool(listStr)
  if listStr != "" && err != nil {
    writeBadRequestResp(w, "list_genomes argument is not a bool.")
    return
  } else if l {
                  ranksArgs = append(ranksArgs, "-l")
  }

  tabStr := r.URL.Query().Get("tabular_output")
          t, err := strconv.ParseBool(tabStr)
  if tabStr != "" && err != nil {
    writeBadRequestResp(w, "tabular_output argument is not a bool.")
    return
  } else if t {
                  ranksArgs = append(ranksArgs, "-t")
  }

          ct, err := contenttype.GetMediaType(r)
          if ct.EqualsMIME(multipartCt) {
    paths, err := filesFromFormData(w, r, 0, 1)
    if err != nil {
      return
    }
    if len(paths) > 0 {
                  ranksArgs = append(ranksArgs, "-g", paths[0])
      for _, p := range paths {
        defer os.Remove(p)
      }
    }
          }


  valid := checkParamInt(w, taxIdStr)
  if !valid {
    return
  }
  valid = checkParamLevels(w, levels)
  if !valid {
    return
  }

          ranksArgs = append(ranksArgs, taxIdStr, dbPath)
  out, errMsg := callPackage(ranksArgs, ranks.Run)
  if len(errMsg) > 0 {
    writeServerError(w, "Error while calling package ranks: " + string(errMsg), nil)
    return
  }

  writePlainOutput(w, out)



}

func filesFromFormData(w http.ResponseWriter, r *http.Request, minFiles, maxFiles int) (paths []string, err error) {
  err = validateMultipartForm(w, r, minFiles, maxFiles)
  if err != nil || len(r.MultipartForm.File) == 0 {
    return
  }

          keys := []string{}
          for key := range r.MultipartForm.File {
                  keys = append(keys, key)
          }
          sort.Strings(keys)

  for i, key := range keys {
                  files := r.MultipartForm.File[key]
                  if len(files) > 0 {
                    h := *files[0]
            var rf multipart.File
                    rf, err = h.Open()
            if err != nil {
              writeServerError(w, "Error while opening multipart file", err)
              return
            }
                    b := make([]byte, h.Size)
                    rf.Read(b)
                    rf.Close()

                    path := "apiv2_temp_" + strconv.Itoa(i) + strconv.Itoa(time.Now().Nanosecond())
            paths = append(paths, path)
                    var tf * os.File
            tf, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
            if err != nil {
              writeServerError(w, "Error while opening temporary file", err)
              return
            }

                    buffer := bytes.NewBuffer(b)
                    buffer.WriteTo(tf)
                    tf.Close()

                  }
          }

  return
}

func validateMultipartForm(w http.ResponseWriter, r *http.Request, minFiles, maxFiles int) (err error) {
          err = r.ParseMultipartForm(3_000_000)
          if err != nil {
    writeBadRequestResp(w, "Malformed multipart form request.")
                  return err
          }

          if len(r.MultipartForm.File) < minFiles {
    writeBadRequestResp(w, fmt.Sprintf("Provide at least %d file(s).", minFiles))
                  return errors.New("not enough files in body")
          }
  if maxFiles != -1 && len(r.MultipartForm.File) > maxFiles {
    writeBadRequestResp(w, fmt.Sprintf("Too many files. This service takes a maximum of %d file(s) per request.", maxFiles))
                  return errors.New("too many files in body")
          }

  return nil
}

func checkParamLevels(w http.ResponseWriter, levels string) bool {
  if levels != "" {
            levelsSplit := strings.Split(levels, ",")
    availableLevels := tdb.AssemblyLevels()
    sort.Strings(availableLevels)
    for _, level := range levelsSplit {
      _, found := slices.BinarySearch(availableLevels, level)
      if !found {
        writeBadRequestResp(w, "Malformed assembly level.")
        return false
      }
    }
  }
  return true
}

func subtree(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
  ct, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
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

  if ct.EqualsMIME(jsonCt) {
            strPlain := r.URL.Query().Get("plain_data")
            plain, err := strconv.ParseBool(strPlain)
            if err != nil {
                    plain = false
            }

    offset, size := extractPaging(r)

    fieldComposite := r.URL.Query().Get("field_composite")
    fieldComposite = parseFieldComposite(fieldComposite)

    taxa, err := neidb.Subtree(taxId)
    if err != nil {
      writeServerError(w, "fn subtree - Error while executing neidb.Subtree", err)
      return
    }

    if size == -1 {
      size = len(taxa)
    }

    data := []Taxon{}
    for i := offset; i < min(offset + size, len(taxa)); i++ {
      id := taxa[i]
      tax, err := getTaxonData(id, plain, fieldComposite, neidb)
      if err != nil {
        writeServerError(w, "Error from getTaxonData", err)
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



  } else if ct.EqualsMIME(plainCt) || ct.EqualsMIME(graphvizCt) {
    dreeArgs := []string{"./dree"}
            levels := r.URL.Query().Get("assembly_levels")
    if levels != "" {
                    dreeArgs = append(dreeArgs, "-L", levels)
    }

    genStr := r.URL.Query().Get("genomes_only")
            g, err := strconv.ParseBool(genStr)
    if genStr != "" && err != nil {
      writeBadRequestResp(w, "genomes_only argument is not a bool.")
      return
    } else if g {
                    dreeArgs = append(dreeArgs, "-g")
    }

    namesStr := r.URL.Query().Get("print_names")
            n, err := strconv.ParseBool(namesStr)
    if namesStr != "" && err != nil {
      writeBadRequestResp(w, "print_names argument is not a bool.")
      return
    } else if n {
                    dreeArgs = append(dreeArgs, "-n")
    }

    depthStr := r.URL.Query().Get("max_depth")
    if depthStr != "" {
                    dreeArgs = append(dreeArgs, "-m", depthStr)
    }

    if ct.EqualsMIME(plainCt) {
                    dreeArgs = append(dreeArgs, "-l")
    }

    valid := checkParamLevels(w, levels)
    if !valid {
      return
    }
    valid = checkParamInt(w, depthStr)
    if !valid {
      return
    }

    dbPath := args[1].(string)
    dreeArgs = append(dreeArgs, taxIdStr, dbPath)
    out, errMsg := callPackage(dreeArgs, dree.Run)
    if len(errMsg) > 0 {
      writeServerError(w, "Error while calling package dree: " + string(errMsg), nil)
      return
    }

    if ct.EqualsMIME(plainCt) {
      writePlainOutput(w, out)



    } else {
      writeGraphvizOutput(w, out)


    }

  }

}

func writeGraphvizOutput(w http.ResponseWriter, out []byte) {
  w.Header().Set("Content-Type", graphvizCt.String())
  fmt.Fprintf(w, "%s", out)
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
          if err != nil {
                  plain = false
          }

  offset, size := extractPaging(r)

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


  for len(taxIds) > 0 && (size == -1 || len(data) < size) {
    taxId := taxIds[0]
    taxIds = taxIds[1:]
    accs, err := neidb.Accessions(taxId)
    if err != nil {
      writeServerError(w, "fn taxonAccessions - Error while executing neidb.Accessions", err)
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
            writeServerError(w, "fn taxonAccessions - Error while executing neidb.Level", err)
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
      writeServerError(w, "fn taxonAccessions - Error while executing neidb.Children", err)
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

func fintac(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
  _, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
  if err != nil {
    w.WriteHeader(http.StatusNotAcceptable)
                  w.Write([]byte("Server does not provide any of the accepted content types."))
    return
  }


          ct, err := contenttype.GetMediaType(r)
          if !ct.EqualsMIME(multipartCt) {
                  w.WriteHeader(http.StatusUnsupportedMediaType)
                  w.Write([]byte("Use multipart/form-data"))
                  return
          }

          fintacArgs := []string{"./fintac"}
          aStr := r.URL.Query().Get("all_splits")
  if aStr != "" {
          a, err := strconv.ParseBool(aStr)
    if err != nil {
              writeBadRequestResp(w, "all_splits argument is not a bool.")
      return
    }
    if a {
                  fintacArgs = append(fintacArgs, "-a")
            }
  }
          n := r.URL.Query().Get("neighbor")
          t := r.URL.Query().Get("target")
          u := r.URL.Query().Get("unknown")
          if n != "" {
                  fintacArgs = append(fintacArgs, "-n", n)
          }
          if t != "" {
                  fintacArgs = append(fintacArgs, "-t", t)
          }
          if u != "" {
                  fintacArgs = append(fintacArgs, "-u", u)
          }

  dbPath := args[0].(string)
          fintacArgs = append(fintacArgs, "-H", dbPath)

  paths, err := filesFromFormData(w, r, 1, -1)
  if err != nil {
    return
  }
  for _, p := range paths {
    defer os.Remove(p)
    fintacArgs = append(fintacArgs, p)
  }


  out, errMsg := callPackage(fintacArgs, fintacPack.Run)
  if len(errMsg) > 0 {
    writeServerError(w, "Error while calling package fintac: " + string(errMsg), nil)
    return
  }

  writePlainOutput(w, out)



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
          if err != nil {
                  plain = false
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
      writeServerError(w, "fn mrca - Error while executing neidb.MRCA", err)
      return
    }

  tax, err := getTaxonData(id, plain, fieldComposite, neidb)
  if err != nil {
    writeServerError(w, "fn mrca - Error from getTaxonData", err)
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

func neighbors(w http.ResponseWriter, r *http.Request, selfNode *Node, args ...any) {
  _, _, err := contenttype.GetAcceptableMediaType(r, selfNode.Types)
  if err != nil {
    w.WriteHeader(http.StatusNotAcceptable)
                  w.Write([]byte("Server does not provide any of the accepted content types."))
    return
  }


  neighborsArgs := []string{"./neighbors"}
          levels := r.URL.Query().Get("assembly_levels")
  if levels != "" {
                  neighborsArgs = append(neighborsArgs, "-L", levels)
  }

  tfStr := r.URL.Query().Get("targets_only")
          tf, err := strconv.ParseBool(tfStr)
  if tfStr != "" && err != nil {
    writeBadRequestResp(w, "targets_only argument is not a bool.")
    return
  } else if tf {
                  neighborsArgs = append(neighborsArgs, "-o")
  }

  gsStr := r.URL.Query().Get("with_genomes_seq_only")
          gs, err := strconv.ParseBool(gsStr)
  if gsStr != "" && err != nil {
    writeBadRequestResp(w, "with_genomes_seq_only argument is not a bool.")
    return
  } else if gs {
                  neighborsArgs = append(neighborsArgs, "-g")
  }

          tabStr := r.URL.Query().Get("tab_output")
          tab, err := strconv.ParseBool(tabStr)
  if tabStr != "" && err != nil {
    writeBadRequestResp(w, "tab_output argument is not a bool.")
    return
  } else if tab {
                  neighborsArgs = append(neighborsArgs, "-T")
  }

  gtStr := r.URL.Query().Get("genomes_and_taxa")
          gt, err := strconv.ParseBool(gtStr)
  if gtStr != "" && err != nil {
    writeBadRequestResp(w, "genomes_and_taxa argument is not a bool.")
    return
  } else if gt {
                  neighborsArgs = append(neighborsArgs, "-l")
  }

  targetIds := r.URL.Query().Get("target_ids")
  neighborsArgs = append(neighborsArgs, "-t", targetIds)


  dbPath := args[0].(string)
  neighborsArgs = append(neighborsArgs , dbPath)

          ct, err := contenttype.GetMediaType(r)
          if ct.EqualsMIME(multipartCt) {
    paths, err := filesFromFormData(w, r, 0, -1)
    if err != nil {
      return
    }

    if len(paths) > 0  && targetIds != "" {
      writeBadRequestResp(w, "Can't provide files and target_ids via url.")
      return
    }

    for _, p := range paths {
                  neighborsArgs = append(neighborsArgs, p)
      defer os.Remove(p)
    }
          }


  valid := checkParamLevels(w, levels)
  if !valid {
    return
  }
  valid = checkParamLevels(w, targetIds)
  if !valid {
    return
  }

  out, errMsg := callPackage(neighborsArgs, neighborsPack.Run)
  if len(errMsg) > 0 {
    writeServerError(w, "fn neighbors - Error while calling package neighbors: " + string(errMsg), nil)
    return
  }

  writePlainOutput(w, out)



}


func checkParamInts(w http.ResponseWriter, param string) bool {
  if param != "" {
    split := strings.Split(param, ",")
    for _, id := range split {
      _, err := strconv.Atoi(id)
      if err != nil {
        writeBadRequestResp(w, "At least one target id is not an integer.")
        return false
      }
    }
  }
  return true
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
          if err != nil {
                  plain = false
          }

  offset, size := extractPaging(r)

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
  for i := 0; (i < offset + size || size == -1) && start != end; i++ {
            parent, err := neidb.Parent(start)

    if err != nil {
      writeServerError(w, "fn path - Error while calling neidb.Parent", err)
      return
    }
            if start == parent {
                      data = data[:0]
                      break
            }

            start = parent
            tax, err := getTaxonData(start, plain, fieldComposite, neidb)
            if err != nil {
              writeServerError(w, "fn path - Error from getTaxonData", err)
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


