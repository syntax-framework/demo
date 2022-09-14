// Copyright 2022 Alex Rodin. All rights reserved.
// Use of this source code is governed by MIT license that can be found
// in the LICENSE file.

// Package router is a trie based high performance and easy to use HTTP request router.
//
// A trivial example is:
//
//	package main
//
//	import (
//	    "fmt"
//	    "github.com/nidorx/router"
//	    "net/http"
//	    "log"
//	)
//
//	func Index(w http.ResponseWriter, r *http.Request, _ router.Params) {
//	    fmt.Fprint(w, "Welcome!\n")
//	}
//
//	func Hello(w http.ResponseWriter, r *http.Request, ps router.Params) {
//	    fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
//	}
//
//	func main() {
//	    router := router.New()
//	    router.GET("/", Index)
//	    router.GET("/hello/:name", Hello)
//
//	    log.Fatal(http.ListenAndServe(":8080", router))
//	}
//
// The router matches incoming requests by the request method and the path.
// If a handle is registered for this path and method, the router delegates the
// request to that function.
// For the methods GET, POST, PUT, PATCH, DELETE and OPTIONS shortcut functions exist to
// register handles, for all other methods router.Handle can be used.
//
// The registered path, against which the router matches incoming requests, can
// contain two types of parameters:
//
//	Syntax    Type
//	:name     named parameter
//	*name     catch-all parameter
//
// Named parameters are dynamic path segments. They match anything until the
// next '/' or the path end:
//
//	Path: /blog/:category/:post
//
//	Requests:
//	 /blog/go/request-routers            match: category="go", post="request-routers"
//	 /blog/go/request-routers/           no match, but the router would redirect
//	 /blog/go/                           no match
//	 /blog/go/request-routers/comments   no match
//
// Catch-all parameters match anything until the path end, including the
// directory index (the '/' before the catch-all). Since they match anything
// until the end, catch-all parameters must always be the final path element.
//
//	Path: /files/*filepath
//
//	Requests:
//	 /files/                             match: filepath="/"
//	 /files/LICENSE                      match: filepath="/LICENSE"
//	 /files/templates/article.html       match: filepath="/templates/article.html"
//	 /files                              no match, but the router would redirect
//
// The value of parameters is saved as a slice of the Param struct, consisting
// each of a key and a value. The slice is passed to the Handle func as a third
// parameter.
// There are two ways to retrieve the value of a parameter:
//
//	// by the name of the parameter
//	user := ps.ByName("user") // defined by :user or *user
//
//	// by the index of the parameter. This way you can also get the name (key)
//	thirdKey   := ps[2].Key   // the name of the 3rd parameter
//	thirdValue := ps[2].Value // the value of the 3rd parameter
package main

import (
	"bytes"
	"errors"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

// Handle is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (variables).
type Handle func(w http.ResponseWriter, r *http.Request, params Params)

type Middleware func(w http.ResponseWriter, r *http.Request, params Params, next func())

type handler struct {
	id       int
	priority int      //
	path     string   // debug purpose
	fn       Handle   // A função de execução da rota
	parts    []string // ":", "*" or "string"
	params   []string // Nomes dos parametros da rota (Ex. '/user/:id' => ["id"])
}

type middleware struct {
	sequence int        // sequencial de adição do middleware
	priority int        //
	path     string     // debug purpose
	fn       Middleware // A função de execução do middleware
	parts    []string   // ":", "*" or "string"
	params   []string   // Nomes dos parametros da rota (Ex. '/user/:id' => ["id"])
}

type mHandlers struct {
	sequence int
	common   map[int][]*handler
	catchAll map[int][]*handler
}

type mMiddlewares struct {
	common   map[int][]*middleware
	catchAll map[int][]*middleware
}

type Router struct {
	handlers    map[string]*mHandlers    // { [HTTP_METHOD] => Handlers }
	middlewares map[string]*mMiddlewares // { [HTTP_METHOD] => Middlewares }
}

// router.Map(path string, &MyController{});

func (r *Router) Use(method, route string, handle Middleware) {

}

// GET is a shortcut for router.Handle(http.MethodGet, route, handle)
func (r *Router) GET(route string, handle Handle) {
	r.Handle(http.MethodGet, route, handle)
}

// HEAD is a shortcut for router.Handle(http.MethodHead, route, handle)
func (r *Router) HEAD(route string, handle Handle) {
	r.Handle(http.MethodHead, route, handle)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, route, handle)
func (r *Router) OPTIONS(route string, handle Handle) {
	r.Handle(http.MethodOptions, route, handle)
}

// POST is a shortcut for router.Handle(http.MethodPost, route, handle)
func (r *Router) POST(route string, handle Handle) {
	r.Handle(http.MethodPost, route, handle)
}

// PUT is a shortcut for router.Handle(http.MethodPut, route, handle)
func (r *Router) PUT(route string, handle Handle) {
	r.Handle(http.MethodPut, route, handle)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, route, handle)
func (r *Router) PATCH(route string, handle Handle) {
	r.Handle(http.MethodPatch, route, handle)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, route, handle)
func (r *Router) DELETE(route string, handle Handle) {
	r.Handle(http.MethodDelete, route, handle)
}

func (r *Router) Handle(method, route string, handle Handle) {
	if err := r.handle(method, route, handle); err != nil {
		panic(any(err))
	}
}

// validParamNameReg the name of route parameters must be made up of “word characters” ([A-Za-z0-9_]).
var validParamNameReg = regexp.MustCompile(`[A-Za-z0-9_]+`)

func isValidParam(name string) bool {
	return validParamNameReg.MatchString(name)
}

// The name of route parameters must be made up of “word characters” ([A-Za-z0-9_]).
func (r *Router) handle(method, route string, fn Handle) error {
	route = path.Clean(route)

	var parts []string       // path parts [":", "*", "STRING"]
	var params []string      // param names (Ex. for path "/user/:id", will be ["id"])
	cpath := &bytes.Buffer{} // path clean

	segments := strings.Split(strings.Trim(route, "/"), "/")
	for i, segment := range segments {

		cpath.WriteRune('/')

		var prefix string
		if len(segment) > 0 {
			prefix = segment[0:1]
		}
		isLastSegment := i == (len(segments) - 1)

		switch prefix {
		case ":", "*":
			// filepath (Ex. "/assets/js/*filepath")
			// param name (Ex. "/user/:id", "/user/:id/edit", "/filter/:id/:subId")
			paramName := strings.TrimPrefix(segment, prefix)
			if strings.ContainsAny(paramName, ":*") {
				// the wildcard name must not contain ':' and '*'
				return errors.New("only one wildcard per path segment is allowed in '" + route + "'")
			}

			if prefix == "*" {
				// catch-all
				if !isLastSegment {
					return errors.New("catch-all routes are only allowed at the end of the path in path '" + route + "'")
				}
				if paramName == "" {
					paramName = "filepath"
				}
			}

			if !isValidParam(paramName) {
				return errors.New("Invalid param ('" + paramName + "') in path '" + route + "'")
			}

			parts = append(parts, prefix)
			params = append(params, paramName)
			cpath.WriteString(prefix)
			cpath.WriteString(paramName)
		default:
			// named path part (Ex. "/user", "/assets/js/file.js")
			parts = append(parts, segment)
			if strings.ContainsAny(segment, ":*") {
				// the wildcard name must not contain ':' and '*'
				return errors.New("only one wildcard per path segment is allowed in '" + route + "'")
			}
			cpath.WriteString(segment)
		}
	}

	numParams := len(parts)
	priority := 0

	// Calculating the priority of this handler
	//
	// a) Left parts have higher priority than right
	// b) For each part of the path
	//    1. ("*") catch all parameter has weight 1
	//    2. (":") named parameter has weight 2
	//    3. ("string") An exact match has weight 3
	for i := 0; i < numParams; i++ {
		weight := 3
		switch parts[i] {
		case ":":
			weight = 2
		case "*":
			weight = 1
		}
		priority = priority + ((numParams - i) * weight)
	}

	if r.handlers == nil {
		r.handlers = make(map[string]*mHandlers)
	}

	root := r.handlers[method]
	if root == nil {
		root = &mHandlers{
			common:   map[int][]*handler{},
			catchAll: map[int][]*handler{},
		}
		r.handlers[method] = root
	}

	handle := &handler{
		id:       root.sequence,
		priority: priority,
		path:     cpath.String(),
		fn:       fn,
		parts:    parts,
		params:   params,
	}
	root.sequence++

	var exist bool
	var handlers []*handler

	isCatchAll := parts[numParams-1] == "*"

	if isCatchAll {
		// catch all
		if handlers, exist = root.catchAll[numParams]; !exist {
			handlers = []*handler{}
			root.catchAll[numParams] = handlers
		}
	} else {
		if handlers, exist = root.common[numParams]; !exist {
			handlers = []*handler{}
			root.common[numParams] = handlers
		}
	}

	// check if exists handlers, or has conflict
	for _, h2 := range handlers {
		if h2.path == handle.path {
			return errors.New("A handle is already registered for path '" + handle.path + "'")
		}

		if h2.priority == handle.priority {
			if strings.ContainsAny(handle.path, "*:") && strings.ContainsAny(h2.path, "*:") {
				hasConflict := true
				for i, h1part := range handle.parts {
					h2part := h2.parts[i]
					h2IsWildcard := h2part == ":" || h2part == "*"
					h1IsWildcard := h1part == "*" || h1part == ":"
					if h2IsWildcard && h1IsWildcard {
						if h2part == h1part {
							// still conflicts
							continue
						}
					} else if h2part != h1part {
						// '/cmd/:tool' vs '/search/:query'
						hasConflict = false
						break
					}
				}
				if hasConflict {
					return errors.New("wildcard route '" + handle.path + "' conflicts with existing wildcard route in path '" + h2.path + "'")
				}
			}
		}
	}

	handlers = append(handlers, handle)
	sort.Slice(handlers, func(i, j int) bool {
		// high priority at the beginning'
		a := handlers[i]
		b := handlers[j]

		return a.priority > b.priority
	})

	if isCatchAll {
		root.catchAll[numParams] = handlers
	} else {
		root.common[numParams] = handlers
	}

	return nil
}

// Lookup allows the manual lookup of a method + route combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter values.
func (r *Router) Lookup(method, route string) (*handler, Params) {
	if r.handlers == nil {
		return nil, nil
	}

	root := r.handlers[method]
	if root == nil {
		return nil, nil
	}

	pathSplit := strings.Split(strings.Trim(route, "/"), "/")
	index := len(pathSplit)

	var match *handler

	if handlers, exists := root.common[index]; exists {
	out:
		for _, h := range handlers {
			for i, part := range h.parts {
				if part == ":" {
					// part is named parameter, ok
					continue
				}
				if part != pathSplit[i] {
					// static part dont match
					continue out
				}
			}
			match = h
			break
		}
	}

	if match == nil {
		// try catch all
		if strings.HasSuffix(route, "/") {
			index++
		}
		for i := index; i >= 0; i-- {
			if handlers, exists := root.catchAll[i]; exists {
			out2:
				for _, h := range handlers {
					for j, part := range h.parts {
						if part == ":" {
							// part is named parameter, ok
							continue
						}
						if part == "*" {
							// part is catch all
							break
						}
						if part != pathSplit[j] {
							// static part dont match
							continue out2
						}
					}
					match = h
					break
				}
			}
			if match != nil {
				break
			}
		}
	}

	if match != nil {
		// parse parameters
		qtdParams := len(match.params)
		if qtdParams == 0 {
			return match, nil
		}

		paramIndex := 0
		params := make(Params, qtdParams)
		for i, part := range match.parts {
			if part == ":" {
				// part is named parameter, ok
				params[paramIndex].Key = match.params[paramIndex]
				params[paramIndex].Value = pathSplit[i]
				paramIndex++
			} else if part == "*" {
				// part is catch all
				params[paramIndex].Key = match.params[paramIndex]
				params[paramIndex].Value = "/" + strings.Join(pathSplit[i:], "/")
				break
			}
		}
		return match, params
	}

	return nil, nil
}
