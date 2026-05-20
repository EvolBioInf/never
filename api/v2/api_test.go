package apiv2

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestApi(t *testing.T) {
	tests := []*exec.Cmd{}
	prog := "../../bin/fetch"
	url := "http://localhost:8080/api/v2/"
	tmpl := "%s/%s%s"
	service := "taxa/9606/children"
	u := fmt.Sprintf(tmpl, url, service, "")
	test := exec.Command(prog, u)
	tests = append(tests, test)
	service = "taxa/562/genome-count"
	u = fmt.Sprintf(tmpl, url, service, "")
	test = exec.Command(prog, u)
	tests = append(tests, test)
	u = fmt.Sprintf(tmpl, url, service+"-recursive", "")
	test = exec.Command(prog, u)
	tests = append(tests, test)
	service = "taxa/562/parent"
	u = fmt.Sprintf(tmpl, url, service, "")
	test = exec.Command(prog, u)
	tests = append(tests, test)
	service = "taxa/9606/subtree"
	u = fmt.Sprintf(tmpl, url, service, "")
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query := "?taxon_ids=9606,9605"
	u = fmt.Sprintf(tmpl, url, "ranks", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "?taxon_ids=278148,602633"
	u = fmt.Sprintf(tmpl, url, "taxa-accessions", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "?taxon_ids=562,9606"
	u = fmt.Sprintf(tmpl, url, "taxa-info", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "?taxon_ids=9606,9605"
	u = fmt.Sprintf(tmpl, url, "taxa-names", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	u = fmt.Sprintf(tmpl, url, "/taxa/9606,741158,63221/mrca", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	u = fmt.Sprintf(tmpl, url, "taxa/9606/path/40674", "")
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "?name=dolph&page_size=10&page=2"
	u = fmt.Sprintf(tmpl, url, "taxa", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "?accession_ids=GCF_000001405.40,GCA_000002115.2"
	u = fmt.Sprintf(tmpl, url, "accessions", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	for i, test := range tests {
		get, err := test.Output()
		if err != nil {
			t.Error(err)
		}
		f := "r" + strconv.Itoa(i+1) + ".txt"
		want, err := os.ReadFile(f)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(get, want) {
			t.Errorf("%s - get:\n%s\nwant:\n%s\n", f, get, want)
		}
	}
}
