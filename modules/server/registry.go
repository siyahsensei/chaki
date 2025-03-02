package server

import (
	"github.com/Trendyol/chaki/modules/server/controller"
	"github.com/Trendyol/chaki/modules/server/route"
	"github.com/Trendyol/chaki/modules/swagger"
	"github.com/Trendyol/chaki/util/slc"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

type registry struct {
	ct     any
	base   string
	name   string
	mws    []fiber.Handler
	routes []route.Route
}

func parseControllers(cts ...controller.Controller) []*registry {
	return slc.Map(cts, newRegistry)
}

func newRegistry(ctr controller.Controller) *registry {
	return &registry{
		ct:     ctr,
		base:   ctr.Prefix(),
		name:   ctr.Name(),
		mws:    ctr.Middlewares(),
		routes: ctr.Routes(),
	}
}

func (r *registry) parsePath(path string) string {
	path, err := url.JoinPath(r.base, path)
	if err != nil {
		panic(err.Error())
	}
	return path
}

func (r *registry) toMeta(h route.Route) route.Meta {
	m := h.Meta()
	if m.Name == "" {
		m.Name = r.parsePath(m.Path)
	}
	m.Path = r.parsePath(m.Path)
	return m
}

func (r *registry) SwaggerDefs() []swagger.EndpointDef {
	metas := slc.Map(r.routes, r.toMeta)
	return slc.Map(metas, r.toSwagDefinition)
}

func (r *registry) toSwagDefinition(m route.Meta) swagger.EndpointDef {
	return swagger.EndpointDef{
		RequestType:  m.Req,
		ResponseType: m.Res,
		Group:        r.name,
		Name:         m.Name,
		Endpoint:     m.Path,
		Method:       m.Method,
	}
}
