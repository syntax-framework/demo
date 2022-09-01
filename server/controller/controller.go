package controller

import (
	"github.com/julienschmidt/httprouter"
	"github.com/syntax-framework/shtml/sht"
	"net/http"
)

type Something struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Params         httprouter.Params
}

type Controller struct {
}

type LiveState struct {
}

// On p
func (l LiveState) On(event string, callback func(params map[string]interface{})) {

}

type ProcessFun func(scope *sht.Scope, ctx Something)
type LiveFun func(scope *sht.Scope, ctx Something, live *LiveState)

func NewController(controller ProcessFun) *LiveState {
	return nil
}

func NewLiveController(process ProcessFun, mount LiveFun) *Controller {
	return nil
}

func init() {
	controller := NewController(func(scope *sht.Scope, ctx Something) {
		// Controller simple, manipula o escopo e finaliza
		// Scope só possui os parametros recebido na tag html (param-name="value")
	})

	liveController := NewLiveController(
		func(scope *sht.Scope, ctx Something) {
			// Mesmo que uma controller simples, manipula escopo e finaliza

			scope.Set("nome", "beribecanta")

		},
		func(scope *sht.Scope, ctx Something, live *LiveState) {

			// Magica do Live Controller, mantém estado de vida longo
			// Método mount é invocado com os parametros recebidos na tag html original
			// sistema serializa e escreve no DOM os dados serializados e criptografados, salva também o fingerprint
			// se a conexão ocorrer usando o mesmo fingerprint, client não submete dados criptografados no mount
			// Sempre que atualizar o scope

			live.On("xpto", func(params map[string]interface{}) {
				// params["MouseX"]
				scope.Set("name", "Alex")
			})
		},
	)

	println(controller)
	println(liveController)
}
