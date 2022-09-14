package controllers

import (
	"github.com/syntax-framework/shtml/sht"
	"github.com/syntax-framework/syntax/syntax"
)

func myLiveControllerSetup(scope *sht.Scope, params map[string]interface{}) {
	// Mesmo que uma controller simples, manipula escopo e finaliza

	scope.Set("value", "Valor da Live Controller")
	scope.Set("method", func() string {
		return "Método da Live Controller"
	})
}

func myLiveController(scope *sht.Scope, params map[string]interface{}, live *syntax.LiveState) {

	// Magica do Live Controller, mantém estado de vida longo
	// Método mount é invocado com os parametros recebidos na tag html original
	// sistema serializa e escreve no DOM os dados serializados e criptografados, salva também o fingerprint
	// se a conexão ocorrer usando o mesmo fingerprint, client não submete dados criptografados no mount

	// Sempre que atualizar o scope

	// ouvindo redi
	live.On("change", func(params map[string]interface{}) {
		// params["MouseX"]
		scope.Set("name", "Alex")
		scope.Set("value", "Clicou aqui")
	})

}

func RegisterMyLiveController(app *syntax.Syntax) {
	app.RegisterController("MyLiveController", myLiveControllerSetup, myLiveController)
}
