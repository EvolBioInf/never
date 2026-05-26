package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
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
	Parameters      []Parameter
	Responses       []Response
	Schema          string
	ExampleResponse string
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
	Example     string
}

type WrappedJson struct {
	value any
}

func main() {
	fmt.Println("reading spec...")
	spec, _ := os.ReadFile("../static/api_spec.json")

	fmt.Println("read spec")

	fmt.Println("creating document...")
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
		content.Tags = append(content.Tags, Tag{Name: tag.Name, Description: tag.Description})
	}

	fmt.Println("reading paths")
	for pathName, pathItem := range docModel.Model.Paths.PathItems.FromOldest() {
		newPath := Path{Name: pathName}

		pathParams := extractParameters(pathItem.Parameters)

		for verb, operation := range pathItem.GetOperations().FromOldest() {
			newOperation := Operation{
				Verb:        verb,
				Description: operation.Description,
				Tags:        operation.Tags,
			}

			queryParams := extractParameters(operation.Parameters)
			newOperation.Parameters = append(newOperation.Parameters, queryParams...)
			newOperation.Parameters = append(newOperation.Parameters, pathParams...)

			examplePath := newPath.Name
			for _, param := range pathParams {
				examplePath = strings.ReplaceAll(examplePath, param.Name, param.Example)
			}

			examplePath = strings.ReplaceAll(examplePath, "{", "")
			examplePath = strings.ReplaceAll(examplePath, "}", "")

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
					if newResponse.Code == 200 && contentValue != nil {

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
							var indented bytes.Buffer
							json.Indent(&indented, []byte(parseType(rawSchema)), "", "  ")

							newOperation.Schema = indented.String()
						}
					}

					newOperation.Responses = append(newOperation.Responses, newResponse)
				}
			}
			newPath.Operations = append(newPath.Operations, newOperation)
		}

		content.Paths = append(content.Paths, newPath)
	}
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

		schemaMarshalled, err := schema.MarshalJSONInline()

		if schema.Items != nil && schema.Items.A != nil {
			t := schema.Items.A.Schema().Type
			if len(t) > 0 {
				fmt.Println(t[0])
			}
		}

		var rawSchema map[string]any
		json.Unmarshal(schemaMarshalled, &rawSchema)
		var indented bytes.Buffer
		json.Indent(&indented, []byte(parseType(rawSchema)), "", "  ")

		param.Schema = indented.String()

		s, err := parameter.Schema.BuildSchema()
		if err != nil {
			panic(fmt.Sprintf("cannot build parameter schema: %e", err))
		}

		var vals []string
		if s.Minimum != nil {
			vals = append(vals, fmt.Sprintf("Minimum: %f", *s.Minimum))
		}

		if s.Maximum != nil {
			vals = append(vals, fmt.Sprintf("Maximum: %f", *s.Maximum))
		}

		if s.Default != nil {
			vals = append(vals, fmt.Sprintf("Default: %s", s.Default.Value))
		}

		param.Extra = strings.Join(vals, ", ")

		if parameter.Required != nil {
			param.Required = *parameter.Required
		}

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

func parseType(t map[string]any) string {
	r, ok := t["type"]
	if !ok {
		return ""
	}

	c, ok := r.(string)
	if !ok {
		return ""
	}

	var res string
	switch c {
	case "array":
		i, ok := t["items"]
		if !ok {
			return ""
		}

		n, ok := i.(map[string]any)
		if !ok {
			return ""
		}

		res = fmt.Sprintf("[%s]", parseType(n))
	case "object":
		p, ok := t["properties"]
		if !ok {
			return ""
		}

		dt, ok := p.(map[string]any)
		if !ok {
			return ""
		}

		res = "{"
		for k, v := range dt {
			pv, ok := v.(map[string]any)
			if !ok {
				return ""
			}
			res += fmt.Sprintf("\"%s\":%s,", k, parseType(pv))
		}
		res = res[:len(res)-1]
		res += "}"
	default:
		res = fmt.Sprintf("\"%s\"", c)
	}

	return res
}
