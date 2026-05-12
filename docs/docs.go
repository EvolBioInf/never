package docs

import (
	"fmt"
	"html/template"

	"net/http"
	"path"

	"os"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"strconv"
)

type Content struct {
	Version     string
	Description string
	ApiVersion  string
	ServerURL   string
	Title       string
	Tags        []Tag
	Paths       []Path
}

type Tag struct {
	Name        string
	Description string
}

type Path struct {
	Name       string
	Operations []Operation
}

type Operation struct {
	Verb            string
	Description     string
	Tags            []string
	PathParameters  []Parameter
	QueryParameters []Parameter
	Responses       []Response
	Example         string
}

type Parameter struct {
	ParamType   string
	Name        string
	Description string
	Required    bool
	Example     string
	Schema      string
}

type Response struct {
	Code        int
	Description string
	Mime        string
	Schema      string
	Example     string
}

func RegisterRoutes(prefix string) {
	fmt.Println("Creating template")
	tmpl := template.New("app")

	fmt.Println("Registering custom functions")
	tmpl.Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"toCSSId": func(args ...string) string {
			noSlash := make([]string, len(args))
			for i, str := range args {
				noSlash[i] = strings.ReplaceAll(str, "/", "-")
			}

			joined := strings.ReplaceAll(strings.Join(noSlash, "-"), "--", "-")
			if len(joined) > 1 && joined[0] == '-' {
				joined = joined[1:]
			}

			return joined
		},

		"dict": func(args ...any) map[string]any {
			dict := make(map[string]any)
			if len(args)%2 != 0 {
				panic("Cannot create dictionary in template. Number of parameters is odd.\n")
			}

			for i := 0; i < len(args); i += 2 {
				key, ok := args[i].(string)
				if !ok {
					panic("Cannot create dictionary in template. Key argument is not a string.\n")
				}

				dict[key] = args[i+1]
			}
			return dict
		},
	})

	fmt.Println("Reading files")
	files := []string{
		path.Join("docs", "pages", "*.html"),
		path.Join("docs", "components", "*.html"),
		path.Join("docs", "components", "*", "*.html"),
	}

	var err error
	for _, file := range files {
		tmpl, err = tmpl.ParseGlob(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Internal Server error: %s\n %d\n", err.Error(), http.StatusInternalServerError)
		}
	}

	content := retrieveData("docs/api_spec.json")

	http.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) { defaultHandler(tmpl, &content, w, r) })

	http.Handle("/docs/v2/static/", http.StripPrefix("/docs/v2/static/", http.FileServer(http.Dir("docs/static"))))

}

func retrieveData(filepath string) Content {
	fmt.Println("reading spec")
	spec, _ := os.ReadFile(filepath)

	fmt.Println("creating document")
	document, err := libopenapi.NewDocument(spec)
	if err != nil {
		panic(fmt.Sprintf("cannot create new document: %e", err))
	}

	fmt.Println("building model")
	docModel, err := document.BuildV3Model()
	if err != nil {
		panic(fmt.Sprintf("cannot create v3 model from document: %e", err))
	}

	content := Content{
		ApiVersion:  docModel.Model.Version,
		Title:       docModel.Model.Info.Title,
		Version:     docModel.Model.Info.Version,
		Description: docModel.Model.Info.Description,
		ServerURL:   docModel.Model.Servers[0].URL,
	}

	for _, tag := range docModel.Model.Tags {
		content.Tags = append(content.Tags, Tag{
			Name:        tag.Name,
			Description: tag.Description,
		})
	}

	for pathName, pathItem := range docModel.Model.Paths.PathItems.FromOldest() {
		newPath := Path{Name: pathName}

		pathParams := extractParameters(pathItem.Parameters)

		for verb, operation := range pathItem.GetOperations().FromOldest() {
			newOperation := Operation{
				Verb:        strings.ToUpper(verb),
				Description: operation.Description,
				Tags:        operation.Tags,
			}
			newOperation.PathParameters = append(newOperation.PathParameters, pathParams...)

			queryParams := extractParameters(operation.Parameters)
			newOperation.QueryParameters = append(newOperation.QueryParameters, queryParams...)

			examplePath := newPath.Name
			for _, param := range pathParams {
				examplePath = strings.ReplaceAll(examplePath, param.Name, param.Example)
			}

			examplePath = strings.ReplaceAll(examplePath, "{", "")
			newOperation.Example = strings.ReplaceAll(examplePath, "}", "")

			needsPlus := false
			for _, param := range queryParams {
				if param.Example != "" {
					if needsPlus {
						examplePath += "&"
					} else {
						examplePath += "?"
						needsPlus = true
					}

					examplePath += fmt.Sprintf("%s=%s", param.Name, param.Example)
				}
			}

			for strCode, response := range operation.Responses.Codes.FromOldest() {
				code, err := strconv.Atoi(strCode)
				if err != nil {
					panic(fmt.Sprintf("cannot convert HTTP response code to int: %e", err))
				}

				if !response.Content.IsZero() {
					responseContent := response.Content.First()

					newResponse := Response{
						Code:        code,
						Description: response.Description,
						Mime:        responseContent.Key(),
					}

					contentValue := responseContent.Value()
					if contentValue != nil {
						if contentValue.Example != nil {
							newResponse.Example = contentValue.Example.Value
						}

						if contentValue.Schema != nil {
							schema, err := contentValue.Schema.BuildSchema()
							if err != nil {
								panic(fmt.Sprintf("cannot build schema: %e", err))
							}

							renderedSchema, err := schema.RenderInline()
							if err != nil {
								panic(fmt.Sprintf("cannot render schema: %e", err))
							}

							newResponse.Schema = string(renderedSchema)
						}
					}

					newOperation.Responses = append(newOperation.Responses, newResponse)
				}
			}

			newPath.Operations = append(newPath.Operations, newOperation)
		}

		content.Paths = append(content.Paths, newPath)
	}

	return content

}

func extractParameters(parameters []*v3.Parameter) []Parameter {
	var parsed []Parameter
	for _, parameter := range parameters {
		schema, err := parameter.Schema.BuildSchema()

		if err != nil {
			panic(fmt.Sprintf("cannot build schema: %e", err))
		}

		param := Parameter{
			ParamType:   parameter.In,
			Name:        parameter.Name,
			Description: parameter.Description,
		}

		renderedSchema, err := schema.RenderInline()

		if err != nil {
			panic(fmt.Sprintf("cannot render schema: %e", err))
		}

		param.Schema = string(renderedSchema)

		if parameter.Required != nil {
			param.Required = *parameter.Required
		}

		if schema.Example != nil {
			param.Example = schema.Example.Value
		}

		parsed = append(parsed, param)
	}

	return parsed
}

func defaultHandler(tmpl *template.Template, content *Content, w http.ResponseWriter, _ *http.Request) {
	fmt.Println("served default")
	tmpl.ExecuteTemplate(w, "app.html", content)
}
