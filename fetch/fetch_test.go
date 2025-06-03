package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []*exec.Cmd{}
	url := "https://neighbors.evolbio.mpg.de"
	test := exec.Command("./fetch", url)
	tests = append(tests, test)
	query := "t=9606"
	tmpl := "%s/%s/?%s"
	eURL := fmt.Sprintf(tmpl, url, "accessions", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	eURL = fmt.Sprintf(tmpl, url, "children", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	eURL = fmt.Sprintf(tmpl, url, "parent", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	query = "t=9606,741158,63221"
	eURL = fmt.Sprintf(tmpl, url, "mrca", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	eURL = fmt.Sprintf(tmpl, url, "names", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	eURL = fmt.Sprintf(tmpl, url, "ranks", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	query = "t=Homo sapiens"
	eURL = fmt.Sprintf(tmpl, url, "taxi", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	eURL = fmt.Sprintf(tmpl, url, "taxids", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	query = "a=GCF_000001405.40,GCA_000002115.2"
	eURL = fmt.Sprintf(tmpl, url, "levels", query)
	test = exec.Command("./fetch", eURL)
	tests = append(tests, test)
	for i, test := range tests {
		get, err := test.Output()
		if err != nil {
			t.Error(err)
		}
		name := fmt.Sprintf("r%d.txt", i+1)
		want, err := os.ReadFile(name)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(get, want) {
			t.Errorf("get:\n%s\nwant:\n%s\n", get, want)
		}
	}
}
