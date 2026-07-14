package docsv2

import (
	"fmt"
	"html/template"

	"embed"

	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"strconv"

	"encoding/json"

	"bytes"

	"net/http"
)

type Content struct {
	Version     string
	Description string
	ApiVersion  string
	ServerURL   string
	Prefix      string
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
	ExampleRequest  string
	ExampleResponse string
	Schema          string
	AcceptHeader    []string
}

type Parameter struct {
	ParamType   string
	Name        string
	Description string
	Required    bool
	Extra       string
	Example     string
	Schema      string
}

type Response struct {
	Code        int
	Description string
	Mime        string
}

//go:embed pages/*
var pagesFS embed.FS

//go:embed components/*
var componentsFS embed.FS

//go:embed static/*
var staticFS embed.FS

func RegisterRoutes(prefix string, local bool, port int) {
	fmt.Println("docsV2: Creating template")
	tmpl := template.New("app")

	fmt.Println("docsV2: Registering custom functions")
	tmpl.Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"toKebab": func(args ...string) string {
			noSlash := make([]string, len(args))
			for i, str := range args {
				noSlash[i] = strings.ReplaceAll(str, "/", "-")
			}

			joined := strings.ReplaceAll(strings.Join(noSlash, "-"), "--", "-")
			if len(joined) > 1 && joined[0] == '-' {
				joined = joined[1:]
			}

			joined = strings.ReplaceAll(joined, "{", "")
			joined = strings.ReplaceAll(joined, "}", "")

			return strings.ToLower(joined)
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

	fmt.Println("docsV2: Reading html files")
	tmpl = template.Must(tmpl.ParseFS(pagesFS, "pages/*.html"))
	tmpl = template.Must(tmpl.ParseFS(componentsFS, "components/*.html", "components/*/*.html"))

	content := retrieveData(local, port)
	content.Prefix = prefix

	http.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) { defaultHandler(tmpl, &content, w, r) })

	// The route within in FileServer is a local one, from my filesystem. Files may be queried and served.
	// The path before that are the ones I may use within the browser to ask for these files from the file server.
	http.Handle(prefix+"/static/", http.StripPrefix(prefix, http.FileServer(http.FS(staticFS))))
}

func retrieveData(local bool, port int) Content {
	fmt.Println("docsV2: reading api document")
	spec, _ := staticFS.ReadFile("static/api_spec.json")
	fmt.Println("docsV2: creating document")
	document, err := libopenapi.NewDocument(spec)
	if err != nil {
		panic(fmt.Sprintf("cannot create new document: %e", err))
	}

	fmt.Println("docsV2: building model")
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

			examplePath := content.ServerURL

			if local {
				examplePath = fmt.Sprintf("http://localhost:%d/api/v2", port)
			}

			examplePath += newPath.Name

			for _, param := range pathParams {
				examplePath = strings.ReplaceAll(examplePath, param.Name, param.Example)
			}

			examplePath = strings.ReplaceAll(examplePath, "{", "")
			examplePath = strings.ReplaceAll(examplePath, "}", "")

			needsAnd := false
			for _, param := range queryParams {
				if param.Example != "" {
					if needsAnd {
						examplePath += "&"
					} else {
						examplePath += "?"
						needsAnd = true
					}

					examplePath += fmt.Sprintf("%s=%s", param.Name, param.Example)
				}
			}

			newOperation.ExampleRequest = examplePath

			for strCode, response := range operation.Responses.Codes.FromOldest() {
				code, err := strconv.Atoi(strCode)
				if err != nil {
					panic(fmt.Sprintf("cannot convert HTTP response code to int: %e", err))
				}

				newResponse := Response{
					Code:        code,
					Description: response.Description,
				}

				if !response.Content.IsZero() {
					for mime, contentValue := range response.Content.FromOldest() {
						newResponse.Mime = mime
						if newResponse.Code == 200 && contentValue != nil {
							newOperation.AcceptHeader = append(newOperation.AcceptHeader, mime)
							if contentValue.Examples.Len() > 0 {
								ex := contentValue.Examples.First().Value()
								marhshalled, err := ex.MarshalJSON()

								if err != nil {
									panic(fmt.Sprintf("cannot marshal example: %e", err))
								}

								var wrapped map[string]any
								json.Unmarshal(marhshalled, &wrapped)
								unwrapped, err := json.MarshalIndent(wrapped["value"], "", "  ")

								if err != nil {
									panic(fmt.Sprintf("cannot unwrap example: %e", err))
								}

								newOperation.ExampleResponse = string(unwrapped)

							}

							if contentValue.Schema != nil {
								schema, err := contentValue.Schema.BuildSchema()
								if err != nil {
									panic(fmt.Sprintf("cannot build schema: %e", err))
								}

								schemaMarshalled, err := schema.MarshalJSONInline()
								if err != nil {
									panic(fmt.Sprintf("cannot marshal schema: %e", err))
								}

								var rawSchema map[string]any
								json.Unmarshal(schemaMarshalled, &rawSchema)

								parsed := parseResponseSchema(rawSchema)

								var indented bytes.Buffer
								json.Indent(&indented, []byte(parsed), "", "  ")
								newOperation.Schema = indented.String()

							}

						}
					}

				}

				newOperation.Responses = append(newOperation.Responses, newResponse)
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

		if parameter.Required != nil {
			param.Required = *parameter.Required
		}

		if len(schema.Type) > 0 {
			if schema.Type[0] == "array" {
				if schema.Items != nil && schema.Items.A != nil {
					t := schema.Items.A.Schema().Type
					if len(t) > 0 {
						param.Schema = t[0] + "[]"
					}
				}

			} else {
				param.Schema = schema.Type[0]
			}
		}

		var vals []string
		if schema.Minimum != nil {
			vals = append(vals, fmt.Sprintf("Minimum: %d", int(*schema.Minimum)))
		}

		if schema.Maximum != nil {
			vals = append(vals, fmt.Sprintf("Maximum: %d", int(*schema.Maximum)))
		}

		if schema.Default != nil {
			vals = append(vals, fmt.Sprintf("Default: %s", schema.Default.Value))
		}

		param.Extra = strings.Join(vals, ", ")

		if schema.Example != nil {
			if schema.Example.Value != "" {
				param.Example = schema.Example.Value
			} else {
				for i, c := range schema.Example.Content {
					param.Example += c.Value
					if i < len(schema.Example.Content)-1 {
						param.Example += ","
					}
				}
			}
		}

		parsed = append(parsed, param)
	}

	return parsed
}

func parseResponseSchema(t map[string]any) string {
	r, ok := t["type"]
	if !ok {
		panic("Error while parsing the schema. Key \"type\" does not exist.")
	}

	c, ok := r.(string)
	if !ok {
		panic("Error while parsing the schema. Value of key \"type\" is not a string.")
	}

	var res string
	switch c {
	case "array":
		i, ok := t["items"]
		if !ok {
			panic("Error while parsing the schema. Array is missing \"items\" sibling.")
		}

		n, ok := i.(map[string]any)
		if !ok {
			panic("Error while parsing the schema. Value of array item is not a map.")
		}

		res = fmt.Sprintf("[%s]", parseResponseSchema(n))

	case "object":
		p, ok := t["properties"]
		if !ok {
			panic("Error while parsing the schema. Object is missing \"properties\" sibling.")
		}

		dt, ok := p.(map[string]any)
		if !ok {
			panic("Error while parsing the schema. Value of object property is not a map.")
		}

		res = "{"
		for k, v := range dt {
			pv, ok := v.(map[string]any)
			if !ok {
				panic("Error while parsing the schema. Value of nested object property is not a map.")
			}

			res += fmt.Sprintf("\"%s\":%s,", k, parseResponseSchema(pv))
		}

		res = res[:len(res)-1]
		res += "}"

	default:
		res = fmt.Sprintf("\"%s\"", c)

	}

	return res

}

func defaultHandler(tmpl *template.Template, content *Content, w http.ResponseWriter, _ *http.Request) {
	tmpl.ExecuteTemplate(w, "app.html", content)
}
