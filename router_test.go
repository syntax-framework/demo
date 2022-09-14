// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func Test_add_and_get(t *testing.T) {
	router := &Router{}

	routes := [...]string{
		"/hi",
		"/contact",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/α",
		"/β",
	}
	for _, route := range routes {
		router.GET(route, fakeHandler(route))
	}

	requests := []tRequest{
		{"/a", false, "/a", nil},
		{"/", true, "", nil},
		{"/hi", false, "/hi", nil},
		{"/contact", false, "/contact", nil},
		{"/co", false, "/co", nil},
		{"/con", true, "", nil},  // key mismatch
		{"/cona", true, "", nil}, // key mismatch
		{"/no", true, "", nil},   // no matching child
		{"/ab", false, "/ab", nil},
		{"/α", false, "/α", nil},
		{"/β", false, "/β", nil},
	}
	for _, tt := range requests {
		t.Run(tt.path, func(t *testing.T) {
			checkRequests(t, router, tt)
		})
	}
}

func Test_wildcard(t *testing.T) {
	router := &Router{}

	routes := [...]string{
		"/",
		"/cmd/:tool/:sub",
		"/cmd/:tool/",
		"/src/*filepath",
		"/src/js/:folder/:name/:file",
		"/src/:type/vendors/:name/index",
		"/src/css/:folder/:name/:file",
		"/src/:type/c/:name/index",
		"/search/",
		"/search/:query",
		"/user/:name",
		"/user/:name/about",
		"/files/:dir/*filepath",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/info/:user/public",
		"/info/:user/project/:project",
		"/src/a/:folder/:name/:file",
		"/src/:type/b/:name/index",
		"/src/b/:folder/:name/:file",
		"/src/c/:folder/:name/:file",
		"/src/:type/a/:name/index",
		"/src/d/:folder/:name/*file",
	}
	for _, route := range routes {
		router.GET(route, fakeHandler(route))
	}

	requests := []tRequest{
		{"/", false, "/", nil},
		{"/cmd/test/", false, "/cmd/:tool/", Params{Param{"tool", "test"}}},
		{"/cmd/test", false, "/cmd/:tool/", Params{Param{"tool", "test"}}},
		{"/cmd/test/3", false, "/cmd/:tool/:sub", Params{Param{"tool", "test"}, Param{"sub", "3"}}},
		{"/src/", false, "/src/*filepath", Params{Param{"filepath", "/"}}},
		{"/src/some/file.png", false, "/src/*filepath", Params{Param{"filepath", "/some/file.png"}}},
		{"/search/", false, "/search/", nil},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/search/someth!ng+in+ünìcodé/", false, "/search/:query", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/user/gopher", false, "/user/:name", Params{Param{"name", "gopher"}}},
		{"/user/gopher/about", false, "/user/:name/about", Params{Param{"name", "gopher"}}},
		{"/files/js/inc/framework.js", false, "/files/:dir/*filepath", Params{Param{"dir", "js"}, Param{"filepath", "/inc/framework.js"}}},
		{"/info/gordon/public", false, "/info/:user/public", Params{Param{"user", "gordon"}}},
		{"/info/gordon/project/go", false, "/info/:user/project/:project", Params{Param{"user", "gordon"}, Param{"project", "go"}}},
		{"/src/js/vendors/jquery/main.js", false, "/src/js/:folder/:name/:file",
			Params{
				Param{"folder", "vendors"},
				Param{"name", "jquery"},
				Param{"file", "main.js"},
			},
		},
		{"/src/css/vendors/jquery/main.css", false, "/src/css/:folder/:name/:file",
			Params{
				Param{"folder", "vendors"},
				Param{"name", "jquery"},
				Param{"file", "main.css"},
			},
		},
		{"/src/tpl/vendors/jquery/index", false, "/src/:type/vendors/:name/index",
			Params{
				Param{"type", "tpl"},
				Param{"name", "jquery"},
			},
		},
	}
	for _, tt := range requests {
		t.Run(tt.path, func(t *testing.T) {
			checkRequests(t, router, tt)
		})
	}
}

func Test_wildcard_conflict(t *testing.T) {
	routes := []tRoute{
		{"/src/*filepath", false},
		{"/src/*filepathx", true},
		{"/src/*", true},
		{"/src/", false},
		{"/src1/", false},
		{"/src1/*filepath", false},
		{"/src2/*filepath", false},
		{"/search/:query", false},
		{"/search/invalid", false},
		{"/user/:name", false},
		{"/user/x", false},
		{"/user/:id", true},
		{"/id/:id", false},
		{"/id/:uuid", true},
		{"/user/:id/:action", false},
		{"/user/:id/update", false},
	}
	testRoutes(t, routes)
}

func Test_duplicate_path(t *testing.T) {
	router := &Router{}

	routes := [...]string{
		"/",
		"/doc/",
		"/src/*filepath",
		"/search/:query",
		"/user/:name",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			router.GET(route, fakeHandler(route))
		})
		if recv != nil {
			t.Fatalf("panic inserting route '%s': %v", route, recv)
		}

		// Add again
		recv = catchPanic(func() {
			router.GET(route, nil)
		})
		if recv == nil {
			t.Fatalf("no panic while inserting duplicate route '%s", route)
		}
	}

	requests := []tRequest{
		{"/", false, "/", nil},
		{"/doc/", false, "/doc/", nil},
		{"/src/some/file.png", false, "/src/*filepath", Params{Param{"filepath", "/some/file.png"}}},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/user/gopher", false, "/user/:name", Params{Param{"name", "gopher"}}},
	}
	for _, tt := range requests {
		t.Run(tt.path, func(t *testing.T) {
			checkRequests(t, router, tt)
		})
	}
}

func Test_empty_wildcard_name(t *testing.T) {
	router := &Router{}

	routes := [...]string{
		"/user:",
		"/user:/",
		"/cmd/:/",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			router.GET(route, nil)
		})
		if recv == nil {
			t.Fatalf("no panic while inserting route with empty wildcard name '%s", route)
		}
	}
}

func Test_catch_all_conflict(t *testing.T) {
	routes := []tRoute{
		{"/src/*filepath/x", true},
		{"/src2/", false},
		{"/src2/*filepath/x", true},
		{"/src3/*filepath", false},
		{"/src3/*filepath/x", true},
		{"/*", false},
		{"/*filepath", true},
	}
	testRoutes(t, routes)
}

func Test_catch_max_params(t *testing.T) {
	router := &Router{}
	var route = "/cmd/*filepath"
	router.GET(route, fakeHandler(route))
}

func Test_double_wildcard(t *testing.T) {
	const panicMsg = "only one wildcard per path segment is allowed in"

	routes := [...]string{
		"/:foo:bar",
		"/:foo:bar/",
		"/:foo*bar",
	}

	for _, route := range routes {
		router := &Router{}
		recv := catchPanic(func() {
			router.GET(route, nil)
		})

		rs := fmt.Sprintf("%v", recv)
		if !strings.HasPrefix(rs, panicMsg) {
			t.Fatalf(`"Expected panic "%s" for route '%s', got "%v"`, panicMsg, route, recv)
		}
	}
}

// Used as a workaround since we can't compare functions or their addresses
var fakeHandlerValue string

func fakeHandler(val string) Handle {
	return func(http.ResponseWriter, *http.Request, Params) {
		fakeHandlerValue = val
	}
}

type tRequest struct {
	path       string
	nilHandler bool
	route      string
	ps         Params
}

type tRoute struct {
	path     string
	conflict bool
}

func checkRequests(t *testing.T, router *Router, request tRequest) {
	h, ps := router.Lookup(http.MethodGet, request.path)

	if h == nil {
		if !request.nilHandler {
			t.Errorf("handle mismatch for route '%s': Expected non-nil handle", request.path)
		}
	} else if request.nilHandler {
		t.Errorf("handle mismatch for route '%s': Expected nil handle", request.path)
	} else {
		h.fn(nil, nil, nil)
		if fakeHandlerValue != request.route {
			t.Errorf("handle mismatch for route '%s': Wrong handle (%s != %s)", request.path, fakeHandlerValue, request.route)
		}
	}

	if !reflect.DeepEqual(ps, request.ps) {
		t.Errorf("Params mismatch for route '%s'", request.path)
	}
}

func testRoutes(t *testing.T, routes []tRoute) {
	router := &Router{}

	for _, route := range routes {
		recv := catchPanic(func() {
			router.GET(route.path, nil)
		})

		if route.conflict {
			if recv == nil {
				t.Errorf("no panic for conflicting route '%s'", route.path)
			}
		} else if recv != nil {
			t.Errorf("unexpected panic for route '%s': %v", route.path, recv)
		}
	}
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	testFunc()
	return
}
