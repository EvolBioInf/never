package apiv2

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/evolbioinf/neighbors/util"
)

type ExTest struct {
	t *exec.Cmd
	r int
}
type LibTest struct {
	t func() string
	r int
}

func TestApi(t *testing.T) {
	exTests := []ExTest{}
	libTests := []LibTest{}
	prog := "../../bin/fetch"
	url := "http://localhost:8080/api/v2/"
	tmpl := "%s/%s%s"
	u := fmt.Sprintf(tmpl, url, "", "")
	exTest := ExTest{t: exec.Command(prog, u), r: 1}
	exTests = append(exTests, exTest)
	service := "accessions"
	u = fmt.Sprintf(tmpl, url, service, "?accession_ids=GCF_000001405.40,GCA_000002115.2&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 2}
	exTests = append(exTests, exTest)
	service = "accessions/GCF_000001405.40"
	u = fmt.Sprintf(tmpl, url, service, "?plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 3}
	exTests = append(exTests, exTest)
	service = "taxa"
	u = fmt.Sprintf(tmpl, url, service, "?name=Canis+lupus+familiaris&exact=true&scientific=true&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 4}
	exTests = append(exTests, exTest)
	service = "taxa/9685"
	u = fmt.Sprintf(tmpl, url, service, "?plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 5}
	exTests = append(exTests, exTest)
	u = fmt.Sprintf(tmpl, url, service, "?field_composite=id&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 6}
	exTests = append(exTests, exTest)
	u = fmt.Sprintf(tmpl, url, service, "?field_composite=rank&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 7}
	exTests = append(exTests, exTest)
	service = "taxa/9612/children"
	u = fmt.Sprintf(tmpl, url, service, "?plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 8}
	exTests = append(exTests, exTest)
	service = "taxa/9615/parent"
	u = fmt.Sprintf(tmpl, url, service, "?plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 9}
	exTests = append(exTests, exTest)
	service = "taxa/9606/subtree"
	u = fmt.Sprintf(tmpl, url, service, "?plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 10}
	exTests = append(exTests, exTest)
	service = "taxonomy"
	u = fmt.Sprintf(tmpl, url, service, "")
	exTest = ExTest{t: exec.Command(prog, u), r: 11}
	exTests = append(exTests, exTest)
	service = "taxonomy/accessions"
	u = fmt.Sprintf(tmpl, url, service, "?taxon_ids=338152,338153&limit=5&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 12}
	exTests = append(exTests, exTest)
	service = "taxonomy/mrca"
	u = fmt.Sprintf(tmpl, url, service, "?taxon_ids=56313,111818,507991&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 13}
	exTests = append(exTests, exTest)
	service = "taxonomy/path"
	u = fmt.Sprintf(tmpl, url, service, "?start_id=56313&end_id=30458&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 14}
	exTests = append(exTests, exTest)
	service = "programs"
	u = fmt.Sprintf(tmpl, url, service, "")
	exTest = ExTest{t: exec.Command(prog, u), r: 15}
	exTests = append(exTests, exTest)
	service = "programs/ants"
	u = fmt.Sprintf(tmpl, url, service, "?options=30422")
	exTest = ExTest{t: exec.Command(prog, u), r: 16}
	exTests = append(exTests, exTest)
	service = "programs/dree"
	u = fmt.Sprintf(tmpl, url, service, "?options=-r&options=-n&options=-g&options=28725")
	exTest = ExTest{t: exec.Command(prog, u), r: 17}
	exTests = append(exTests, exTest)
	service = "programs/fintac"
	u = fmt.Sprintf(tmpl, url, service, "")
	eco7k := util.Open("eco7k.nwk")
	defer eco7k.Close()
	libTest := LibTest{t: func() string {
		return util.SendPostRequest(
			u,
			[]string{"-t", "991910_", "-u", "562_"},
			[]string{"eco7k.nwk"},
			[]*os.File{eco7k},
			nil,
		)
	},
		r: 18}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			u,
			[]string{"-t", "991910_", "-u", "562_"},
			[]string{},
			nil,
			eco7k,
		)
	},
		r: 18}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			u,
			[]string{"-t", "991910_"},
			[]string{},
			nil,
			eco7k,
		)
	},
		r: 19}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			u,
			[]string{},
			[]string{},
			nil,
			nil,
		)
	},
		r: 20}
	libTests = append(libTests, libTest)
	service = "programs/neighbors"
	u = fmt.Sprintf(tmpl, url, service, "")
	targets := util.Open("testig/targets.txt")
	defer targets.Close()
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"-t", "9606", "-L", "complete"},
			[]string{},
		)
	},
		r: 21}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			u,
			[]string{"-L", "complete"},
			[]string{"targets.txt"},
			[]*os.File{targets},
			nil,
		)
	},
		r: 21}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			u,
			[]string{"-L", "complete"},
			nil,
			nil,
			targets,
		)
	},
		r: 21}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			nil,
			nil,
			nil,
		)
	},
		r: 19}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"-t", "9685", "-l"},
			[]string{},
		)
	},
		r: 22}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"-t", "9685", "-l", "-o"},
			[]string{},
		)
	},
		r: 23}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"-l", "-t", "9612,9615", "-T"},
			[]string{},
		)
	},
		r: 24}
	libTests = append(libTests, libTest)
	service = "programs/ranks"
	u = fmt.Sprintf(tmpl, url, service, "")
	genomeList := util.Open("testig/myGenomeList.txt")
	defer genomeList.Close()
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"9612"},
			[]string{},
		)
	},
		r: 25}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"-L", "chromosome", "9612"},
			[]string{},
		)
	},
		r: 26}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"-t", "-L", "chromosome", "9612"},
			[]string{},
		)
	},
		r: 27}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			u,
			[]string{"-g", "testig/myGenomeList.txt", "9612"},
			[]string{},
			[]*os.File{genomeList},
			nil,
		)
	},
		r: 28}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{},
			[]string{},
		)
	},
		r: 19}
	libTests = append(libTests, libTest)
	service = "programs/taxi"
	u = fmt.Sprintf(tmpl, url, service, "")
	defer genomeList.Close()
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"-t", "9612"},
			[]string{},
		)
	},
		r: 29}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{"homo sapiens"},
			[]string{},
		)
	},
		r: 30}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			u,
			[]string{},
			[]string{},
		)
	},
		r: 19}
	libTests = append(libTests, libTest)

	for _, test := range exTests {
		get, err := test.t.Output()
		if err != nil {
			t.Error(err)
		}
		f := "testing/r" + strconv.Itoa(test.r) + ".txt"
		want, err := os.ReadFile(f)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(get, want) {
			t.Errorf("%s - get:\n%s\nwant:\n%s\n", f, get, want)
		}
	}
	for _, test := range libTests {
		get, err := test.t()
		if err != nil {
			t.Error(err)
		}
		f := "testing/r" + strconv.Itoa(test.r) + ".txt"
		want, err := os.ReadFile(f)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(get, want) {
			t.Errorf("%s - get:\n%s\nwant:\n%s\n", f, get, want)
		}
	}
}
