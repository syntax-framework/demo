package controllers

import (
	"github.com/syntax-framework/shtml/sht"
	"github.com/syntax-framework/syntax/syntax"
)

func myControllerSetup(scope *sht.Scope, params map[string]interface{}) {
	// Controller simple, manipula o escopo e finaliza
	// Scope só possui os parametros recebido na tag html (param-name="value")

	scope.Set("value", "Valor da Controller Thawan")
	scope.Set("method", func() string {
		return "Método da Controller"
	})
}

func RegisterMyController(app *syntax.Syntax) {
	app.RegisterController("MyController", myControllerSetup, nil)
}
