package docs

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
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

func RegisterRoutes(prefix string) {
	var err error
	fmt.Println("Creating template")
	tmpl := template.New("app")

	fmt.Println("Registering custom functions")
	tmpl.Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
	})

	fmt.Println("Reading files")
	files := []string{path.Join("docs", "pages", "*.html"), path.Join("docs", "components", "*", "*.html")}
	for _, file := range files {
		tmpl, err = tmpl.ParseGlob(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Internal Server error: %s\n %d\n", err.Error(), http.StatusInternalServerError)
		}
	}

	fmt.Println("Registering handlers")
	http.HandleFunc(prefix+"/", func(w http.ResponseWriter, r *http.Request) { defaultHandler(tmpl, w, r) })

	http.Handle("/docs/v2/static/", http.StripPrefix("/docs/v2/static/", http.FileServer(http.Dir("docs/static"))))
}

func defaultHandler(tmpl *template.Template, w http.ResponseWriter, _ *http.Request) {
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
			},
		}})
}
