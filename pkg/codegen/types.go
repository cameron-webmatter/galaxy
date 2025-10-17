package codegen

import (
	"github.com/cameron-webmatter/galaxy/pkg/parser"
	"github.com/cameron-webmatter/galaxy/pkg/router"
)

type HandlerGenerator struct {
	Component  *parser.Component
	Route      *router.Route
	ModuleName string
	BaseDir    string
}

type GeneratedHandler struct {
	PackageName  string
	Imports      []string
	FunctionName string
	Code         string
}

type MainGenerator struct {
	Handlers      []*GeneratedHandler
	Routes        []*router.Route
	ModuleName    string
	ManifestPath  string
	HasMiddleware bool
}
