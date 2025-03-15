package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Gettable interface {
	Get(name string) string
	GetDef(name, def string) string
}

type Accessibly interface {
	Middleware(access, url string) func(ctx *fiber.Ctx) error
}

type RouteFunc func(ctx *fiber.Ctx) error
type Route struct {
	Role    string
	Path    string
	Method  string
	Handler RouteFunc
}

var routes = make([]Route, 0, 32)

func AddRoute(method, path, role string, route RouteFunc) {
	routes = append(routes, Route{role, path, method, route})
}

func Listen(cfg Gettable, access Accessibly) (chan bool, *fiber.App) {
	srv := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             4 * 1024 * 1024,
	})
	srv.Use(cors.New())
	srv.Use(compress.New())
	srv.Static("/", "./static")
	addr := cfg.GetDef("server.addr", ":8080")
	for _, route := range routes {
		switch route.Method {
		case "GET":
			srv.Get(route.Path, access.Middleware(route.Role, route.Path), route.Handler)
		case "POST":
			srv.Post(route.Path, access.Middleware(route.Role, route.Path), route.Handler)
		case "PUT":
			srv.Put(route.Path, access.Middleware(route.Role, route.Path), route.Handler)
		case "DELETE":
			srv.Delete(route.Path, access.Middleware(route.Role, route.Path), route.Handler)
		}
	}
	sigs := make(chan bool, 1)
	go func() {
		if err := srv.Listen(addr); err != nil {
			log.Error(err)
		}
		log.Info("server stopped")
		sigs <- true
	}()
	return sigs, srv
}
