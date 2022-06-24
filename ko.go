package ko

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

func Default() *Engine {
	e := New()
	e.Use(Logger(), Recovery())
	return e
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

func (g *RouterGroup) Group(prefix string, middlewares ...HandlerFunc) *RouterGroup {
	e := g.engine
	newGroup := RouterGroup{
		prefix:      g.prefix + prefix,
		parent:      g,
		engine:      e,
		middlewares: middlewares,
	}
	e.groups = append(e.groups, &newGroup)
	return &newGroup
}

func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := g.prefix + relativePath
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		file := c.Param("filepath")

		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (g *RouterGroup) Static(relativePath string, root string) {
	h := g.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	g.GET(urlPattern, h)
}

func (g *RouterGroup) addRoute(method string, comp string, h HandlerFunc) {
	pattern := g.prefix + comp
	log.Printf("[Route] %4s - %s", method, pattern)
	g.engine.router.addRoute(method, pattern, h)
}

func (g *RouterGroup) GET(comp string, h HandlerFunc) {
	g.addRoute("GET", comp, h)
}

func (g *RouterGroup) POST(comp string, h HandlerFunc) {
	g.addRoute("POST", comp, h)
}

func (g *RouterGroup) PUT(comp string, h HandlerFunc) {
	g.addRoute("PUT", comp, h)
}

func (g *RouterGroup) DELETE(comp string, h HandlerFunc) {
	g.addRoute("DELETE", comp, h)
}

func (g *RouterGroup) OPTION(comp string, h HandlerFunc) {
	g.addRoute("OPTION", comp, h)
}

func (g *RouterGroup) PATCH(comp string, h HandlerFunc) {
	g.addRoute("PATCH", comp, h)
}

type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}

type HandlerFunc func(c *Context)

func New() *Engine {
	e := Engine{
		router: newRouter(),
	}
	e.RouterGroup = &RouterGroup{engine: &e}
	e.groups = []*RouterGroup{e.RouterGroup}
	return &e
}

func (e *Engine) SetFuncMap(fm template.FuncMap) {
	e.funcMap = fm
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = e
	e.router.handle(c)
}

func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}
