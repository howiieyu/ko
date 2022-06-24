package ko

import (
	"log"
	"net/http"
	"strings"
)


type router struct {
	roots map[string]*node
	handlers map[string]HandlerFunc
}


func newRouter() *router {
	return &router{
		roots: make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)

	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)

			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}


func (r *router) addRoute(method string, pattern string, hf HandlerFunc) {
	log.Printf("[Route] %4s - %s", method, pattern)
	parts := parsePattern(pattern)
	key := method + "-" + pattern

	// r.handlers[key] = hf
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	log.Printf("[Route] addRoute: %s", key)
	r.handlers[key] = hf
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)

	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)
	log.Printf("[Route] getRoute: Path %v RouteNode: %v", searchParts, n)
	if n != nil {
		parts := parsePattern(n.pattern)

		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}

			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	log.Printf("[%4s] handle %s", c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		hf, ok := r.handlers[key]
		if !ok {
			log.Printf("%s have not handlerFunc", key)
		} else {
			c.handlers = append(c.handlers, hf)
		}
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s \n", c.Path)
		})
	}
	c.Next()
}
