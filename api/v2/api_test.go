package apiv2

import (
	"bytes"
	"fmt"
	util "github.com/evolbioinf/never/util"
	"os"
	"os/exec"
	"strconv"
	"testing"
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
	prog := "../../fetch/fetch"
	url := "http://localhost:8008/api/v2"
	tmpl := "%s/%s%s"
	u := url
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
	u = fmt.Sprintf(tmpl, url, service, "?name=Canis+lupus+familiaris&exact=true&scientific=true&limit=1&plain_data=true")
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
	u = fmt.Sprintf(tmpl, url, service, "?limit=5&offset=3&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 8}
	exTests = append(exTests, exTest)
	service = "taxa/9615/parent"
	u = fmt.Sprintf(tmpl, url, service, "?plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 9}
	exTests = append(exTests, exTest)
	service = "taxa/9605/subtree"
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
	u = fmt.Sprintf(tmpl, url, service, "?taxon_ids=56312,28725,8825&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 13}
	exTests = append(exTests, exTest)
	service = "taxonomy/path"
	u = fmt.Sprintf(tmpl, url, service, "?start_id=56312&end_id=3078114&plain_data=true")
	exTest = ExTest{t: exec.Command(prog, u), r: 14}
	exTests = append(exTests, exTest)
	service = "programs"
	u = fmt.Sprintf(tmpl, url, service, "")
	exTest = ExTest{t: exec.Command(prog, u), r: 15}
	exTests = append(exTests, exTest)
	service = "programs/ants"
	u = fmt.Sprintf(tmpl, url, service, "?options=56312")
	exTest = ExTest{t: exec.Command(prog, u), r: 16}
	exTests = append(exTests, exTest)
	service = "programs/dree"
	u = fmt.Sprintf(tmpl, url, service, "?options=-r&options=-n&options=3073808")
	exTest = ExTest{t: exec.Command(prog, u), r: 17}
	exTests = append(exTests, exTest)
	service = "programs/fintac"
	fu := fmt.Sprintf(tmpl, url, service, "")
	e1, err := os.Open("testing/eco7k.nwk")
	util.Check(err)
	defer e1.Close()
	libTest := LibTest{t: func() string {
		return util.SendPostRequest(
			fu,
			[]string{"-t", "991910_", "-u", "562_"},
			[]string{"testing/eco7k.nwk"},
			[]*os.File{e1},
			nil,
		)
	},
		r: 18}
	libTests = append(libTests, libTest)
	e2, err := os.Open("testing/eco7k.nwk")
	util.Check(err)
	defer e2.Close()
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			fu,
			[]string{"-t", "991910_", "-u", "562_"},
			[]string{},
			nil,
			e2,
		)
	},
		r: 18}
	libTests = append(libTests, libTest)
	e3, err := os.Open("testing/eco7k.nwk")
	util.Check(err)
	defer e3.Close()
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			fu,
			[]string{"-t", "991910_"},
			[]string{},
			nil,
			e3,
		)
	},
		r: 19}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			fu,
			[]string{},
			[]string{},
			nil,
			nil,
		)
	},
		r: 31}
	libTests = append(libTests, libTest)
	service = "programs/neighbors"
	nu := fmt.Sprintf(tmpl, url, service, "")
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			nu,
			[]string{"-t", "9606", "-L", "complete"},
			[]string{},
		)
	},
		r: 21}
	libTests = append(libTests, libTest)
	t1, err := os.Open("testing/targets.txt")
	util.Check(err)
	defer t1.Close()
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			nu,
			[]string{"-L", "complete"},
			[]string{"testing/targets.txt"},
			[]*os.File{t1},
			nil,
		)
	},
		r: 21}
	libTests = append(libTests, libTest)
	t2, err := os.Open("testing/targets.txt")
	util.Check(err)
	defer t2.Close()
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			nu,
			[]string{"-L", "complete"},
			[]string{},
			nil,
			t2,
		)
	},
		r: 21}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			nu,
			[]string{},
			[]string{},
		)
	},
		r: 20}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			nu,
			[]string{"-t", "9685", "-l"},
			[]string{},
		)
	},
		r: 22}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			nu,
			[]string{"-t", "9685", "-l", "-o"},
			[]string{},
		)
	},
		r: 23}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			nu,
			[]string{"-l", "-t", "9612,9615", "-T"},
			[]string{},
		)
	},
		r: 24}
	libTests = append(libTests, libTest)
	service = "programs/ranks"
	ru := fmt.Sprintf(tmpl, url, service, "")
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			ru,
			[]string{"9612"},
			[]string{},
		)
	},
		r: 25}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			ru,
			[]string{"-L", "chromosome", "9612"},
			[]string{},
		)
	},
		r: 26}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			ru,
			[]string{"-t", "-L", "chromosome", "9612"},
			[]string{},
		)
	},
		r: 27}
	libTests = append(libTests, libTest)
	g1, err := os.Open("testing/myGenomeList.txt")
	util.Check(err)
	defer g1.Close()
	libTest = LibTest{t: func() string {
		return util.SendPostRequest(
			ru,
			[]string{"-g", "testing/myGenomeList.txt", "9612"},
			[]string{},
			[]*os.File{g1},
			nil,
		)
	},
		r: 28}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			ru,
			[]string{},
			[]string{},
		)
	},
		r: 20}
	libTests = append(libTests, libTest)
	service = "programs/taxi"
	tu := fmt.Sprintf(tmpl, url, service, "")
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			tu,
			[]string{"-t", "9612"},
			[]string{},
		)
	},
		r: 29}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			tu,
			[]string{"homo sapiens"},
			[]string{},
		)
	},
		r: 30}
	libTests = append(libTests, libTest)
	libTest = LibTest{t: func() string {
		return util.SendGetRequest(
			tu,
			[]string{},
			[]string{},
		)
	},
		r: 20}
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
		get := test.t()
		f := "testing/r" + strconv.Itoa(test.r) + ".txt"
		want, err := os.ReadFile(f)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal([]byte(get), want) {
			t.Errorf("%s - get:\n%s\nwant:\n%s\n", f, get, want)
		}
	}
}
