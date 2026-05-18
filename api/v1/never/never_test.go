package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestNever(t *testing.T) {
	tests := []*exec.Cmd{}
	prog := "../bin/fetch"
	url := "http://localhost:8080"
	test := exec.Command(prog, url)
	tests = append(tests, test)
	tmpl := "%s/%s/?%s"
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
	query = "t=278148,602633"
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
	query = "t=dolph&n=10&p=2"
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
		if !bytes.Equal(get, want) {
			t.Errorf("%s - get:\n%s\nwant:\n%s\n", f, get, want)
		}
	}
}
