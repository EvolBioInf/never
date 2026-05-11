package main

import (
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
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
	Description string
	Parameters  []Parameter
	Responses   []Response
	Example     string
}

type Parameter struct {
	ParamType   string
	Name        string
	Description string
	Required    bool
	Example     any
	Schema      any
}

type Response struct {
	Code        int
	Description string
	Mime        string
	Schema      any
	Example     any
}

func main() {
	fmt.Println("reading spec...")
	spec, _ := os.ReadFile("api_spec.json")

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
		content.Paths = append(content.Paths, Path{
			Name: pathName,
		})

		var pathParams []Parameter
		for _, parameter := range pathItem.Parameters {
			schema, err := parameter.Schema.BuildSchema()

			if err != nil {
				panic(fmt.Sprintf("cannot build schema: %e", err))
			}

			pathParams = append(pathParams, Parameter{
				ParamType:   "Path",
				Name:        parameter.Name,
				Description: parameter.Description,
				Example:     schema.Example.Value,
				Required:    *parameter.Required,
				Schema:      schema.Type[0],
			})

			fmt.Println(parameter.Name, schema)
		}
	}
}
