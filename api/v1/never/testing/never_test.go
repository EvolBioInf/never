package neverV1Test

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func TestNever(t *testing.T) {
	tests := []*exec.Cmd{}
	prog := "../../../../fetch/fetch"
	url := "http://localhost:8008"
	test := exec.Command(prog, url+"/docs/api/v1")
	tests = append(tests, test)
	tmpl := "%s/api/v1/%s/?%s"
	query := "t=9606"
	u := fmt.Sprintf(tmpl, url, "children", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=562"
	u = fmt.Sprintf(tmpl, url, "num_genomes", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	u = fmt.Sprintf(tmpl, url, "num_genomes_rec", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	u = fmt.Sprintf(tmpl, url, "parent", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=9606"
	u = fmt.Sprintf(tmpl, url, "subtree", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=9685,61379"
	u = fmt.Sprintf(tmpl, url, "accessions", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=9606,741158,63221"
	u = fmt.Sprintf(tmpl, url, "mrca", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=9606,9605"
	u = fmt.Sprintf(tmpl, url, "names", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=9606,40674"
	u = fmt.Sprintf(tmpl, url, "path", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=9606,9605"
	u = fmt.Sprintf(tmpl, url, "ranks", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=562,9606"
	u = fmt.Sprintf(tmpl, url, "taxa_info", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=homo+sapiens"
	u = fmt.Sprintf(tmpl, url, "taxids", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "t=cat&n=4&p=2"
	u = fmt.Sprintf(tmpl, url, "taxi", query)
	test = exec.Command(prog, u)
	tests = append(tests, test)
	query = "a=GCF_000001405.40,GCA_000002115.2"
	u = fmt.Sprintf(tmpl, url, "levels", query)
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
		wantLines := strings.Split(string(want), "\n")
		getLines := strings.Split(string(get), "\n")
		mismatched := false
		for i := range wantLines {
			if i < len(getLines) {
				if strings.TrimSpace(getLines[i]) != strings.TrimSpace(wantLines[i]) {
					mismatched = true
					t.Errorf("Differed in line %d\nget:\n%s\nwant:\n%s", i, getLines[i], wantLines[i])
				}
			}
		}
		if len(getLines) != len(wantLines) {
			t.Errorf("Line length mismatch\nget:\n%d\nwant:\n%d\n", len(getLines), len(wantLines))
		}

		if mismatched {
			t.Errorf("total: %s\n get:\n%s\nwant:\n%s\n", f, get, want)
		}
	}
}
