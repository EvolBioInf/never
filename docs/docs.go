package docs

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"
)

type Content struct {
	Endpoints []Endpoint
}

type ResponseCode struct {
	Code        int
	Description string
}

type Endpoint struct {
	Route         string
	Description   string
	Example       string
	RequestType   string
	Extended      bool
	PathParams    []Argument
	QueryParams   []Argument
	ResponseCodes []ResponseCode
}

type Argument struct {
	Name        string
	Datatype    string
	Example     string
	Description string
	Required    bool
}

const (
	StatusCode200 = "OK"
	StatusCode400 = "Bad Request"
	StatusCode401 = "Unauthorized"
	StatusCode403 = "Forbidden"
	StatusCode404 = "Not Found"
	StatusCode500 = "Internal Server Error"
)

func toCSSId(args ...string) string {
	noSlash := make([]string, len(args))
	for i, str := range args {
		noSlash[i] = strings.ReplaceAll(str, "/", "-")
	}

	joined := strings.ReplaceAll(strings.Join(noSlash, "-"), "--", "-")
	if len(joined) > 1 && joined[0] == '-' {
		joined = joined[1:]
	}

	return joined
}

func RegisterRoutes(prefix string) {
	var err error
	fmt.Println("Creating template")
	tmpl := template.New("app")

	fmt.Println("Registering custom functions")
	tmpl.Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"toCSSId": toCSSId,
	})

	fmt.Println("Reading files")
	files := []string{
		path.Join("docs", "pages", "*.html"),
		path.Join("docs", "components", "*.html"),
		path.Join("docs", "components", "*", "*.html"),
	}

	for _, file := range files {
		tmpl, err = tmpl.ParseGlob(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Internal Server error: %s\n %d\n", err.Error(), http.StatusInternalServerError)
		}
	}

	fmt.Println("Registering handlers")
	http.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) { defaultHandler(tmpl, w, r) })

	http.Handle("/docs/v2/static/", http.StripPrefix("/docs/v2/static/", http.FileServer(http.Dir("docs/static"))))
}

func defaultHandler(tmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Served default")

	tmpl.ExecuteTemplate(w, "app.html",
		&Content{Endpoints: []Endpoint{
			{
				Route:       "/taxon/{from}/{to}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/children/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: true,
			},
			{
				Route:       "/parent/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent21/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent20/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent19/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent18/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent17/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent16/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent15/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent14/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent13/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent12/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent10/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent9/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent8/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent7/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent6/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent5/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent4/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent3/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent2/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
			{
				Route:       "/parent1/{tax-id}",
				Description: "Returns path connection start and end",
				Example:     "/accessions/278148",
				RequestType: "GET",
				PathParams: []Argument{
					{
						Name:        "from",
						Datatype:    "int",
						Example:     "9234",
						Description: "the id of the start taxon",
						Required:    true,
					},
					{
						Name:        "to",
						Datatype:    "int",
						Example:     "526",
						Description: "the id of the end taxon",
						Required:    true,
					},
				},
				ResponseCodes: []ResponseCode{
					{Code: 200, Description: StatusCode200},
					{Code: 400, Description: StatusCode400},
				},
				Extended: false,
			},
		}})
}
