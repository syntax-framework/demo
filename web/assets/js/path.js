let HANDLER = Symbol() // ao inves de usar symbol, usar caractere não permitido na url, como "/h"
let MIDDLEWARES = Symbol()

let GET = {
  "/": {
    "user": {
      [HANDLER]: () => "/user",
      "*": {
        [MIDDLEWARES]: [
          [() => "/user/*filepath", ["filepath"]],
        ],
      },
      ":": {
        [HANDLER]: [() => "/user/:id", ["id"]],
        [MIDDLEWARES]: [
          [() => "/user/:userId", ["userId"]],
        ],
        "edit": {
          [HANDLER]: [() => "/user/:id/edit", ["id"]],
        },
        "update": {
          [HANDLER]: [() => "/user/:id/update", ["id"]],
          ":": {
            [HANDLER]: [() => "/user/:uid/update/:case", ["uid", "case"]],
          }
        }
      }
    }
  }
}


let POST = {
  paths: {
    "user": {
      handler: () => "/user", // 1
      paths: {
        "*": {
          middlewares: [
            [() => "/user/*filepath", ["filepath"]],
          ],
        },
        ":": {
          handler: [() => "/user/:id", ["id"]],
          middlewares: [
            [() => "/user/:userId", ["userId"]],
          ],
          paths: {
            ":": {
              handler: [() => "/user/:id/:action", ["id"]],
              paths: {
                "xpto": {
                  handler: [() => "/user/:uid/:action/xpto", ["uid", "case"]], // /user/33/update/xpto
                }
              }
            },
            "edit": {
              handler: [() => "/user/:id/edit", ["id"]],
            },
            "update": {
              handler: [() => "/user/:id/update", ["id"]],
              paths: {
                ":": {
                  handler: [() => "/user/:uid/update/:case", ["uid", "case"]], // /user/33/update/xpto
                }
              }
            }
          }
        }
      }
    }
  }
}

const Entry = {
  path: [], // string
  priority: 1,
  handler: {
    fn: () => 1,
    paramNames: []
  }
}

let HANDLERS_BY_PATH_LENGTH = [
  /* 0 */
  [
    ["user", [() => "/user", []]]
  ],
  /* 1 */
  [
    ["user", ":", [() => "/user/:id", ["id"]]]
  ],
  /* 2 */
  [
    ["user", ":", ":", [() => "/user/:id/:action", ["id"]]],
    ["user", ":", "edit", [() => "/user/:id/edit", ["id"]]],
    ["user", ":", "update", [() => "/user/:id/update", ["id"]]],
  ],
  /* 3 */
  [
    ["user", ":", ":", "xpto", [() => "/user/:uid/:action/xpto", ["uid", "case"]]], // /user/33/update/xpto
    ["user", ":", "update", ":", [() => "/user/:uid/update/:case", ["uid", "case"]]], // /user/33/update/xpto
    // formula da prioridade
  ],
]

  // calculo da prioridade
  // Partes da esquerda tem maiss prioridade do que a direita
  // Um aprametro estático tem peso 2
  // O parametro exato a esquerda tem mais prioridade do que
  (["user", ":", ":", "xpto"]).reduce((prev, part, index, array) => {
    let weight = 3
    if (part === ":") {
      weight = 2
    } else if (part === "*") {
      weight = 1
    }
    return prev + ((array.length - index) * weight)
  }, 0)


  // analise combinatório de termos e pesos
  (function (combinations, weightStatic, weightParameter, weightWildcard) {
    let result = combinations.map(item => {
      let priority = 0
      if (item !== "/") {
        priority = item.split('/').reduce((prev, part, index, array) => {
          let weight = weightStatic
          if (part.startsWith(":")) {
            weight = weightParameter
          } else if (part.startsWith("*")) {
            weight = weightWildcard
          }
          let height = array.length - index;
          return prev + (height * height * weight)
        }, 0)
      }
      return `${priority}  ${item}`
    })
    console.log(JSON.stringify(result, null, 2))
  })([
    "/blog/category/page/subpage",
    "/blog/category/page/:subpage",
    "/blog/category/page/*subpage",
    "/blog/category/:page/subpage",
    "/blog/category/:page/:subpage",
    "/blog/category/:page/*subpage",
    "/blog/:category/page/subpage",
    "/blog/:category/page/:subpage",
    "/blog/:category/page/*subpage",
    "/blog/:category/:page/subpage",
    "/blog/:category/:page/:subpage",
    "/blog/:category/:page/*subpage",
    "/:blog/category/page/subpage",
    "/:blog/category/page/:subpage",
    "/:blog/category/page/*subpage",
    "/:blog/category/:page/subpage",
    "/:blog/category/:page/:subpage",
    "/:blog/category/:page/*subpage",
    "/:blog/:category/page/subpage",
    "/:blog/:category/page/:subpage",
    "/:blog/:category/page/*subpage",
    "/:blog/:category/:page/subpage",
    "/:blog/:category/:page/:subpage",
    "/:blog/:category/:page/*subpage",
    "/blog/category/page",
    "/blog/category/:page",
    "/blog/category/*page",
    "/blog/:category/page",
    "/blog/:category/:page",
    "/blog/:category/*page",
    "/:blog/category/page",
    "/:blog/category/:page",
    "/:blog/category/*page",
    "/:blog/:category/page",
    "/:blog/:category/:page",
    "/:blog/:category/*page",
    "/blog/category",
    "/blog/:category",
    "/blog/*category",
    "/blog",
    "/:blog",
    "/*blog",
    "/",
  ], 3, 2, 1)

/*
  "165  /blog/category/page/subpage",
  "164  /blog/category/page/:subpage",
  "163  /blog/category/page/*subpage",
  "161  /blog/category/:page/subpage",
  "160  /blog/category/:page/:subpage",
  "159  /blog/category/:page/*subpage",
  "156  /blog/:category/page/subpage",
  "155  /blog/:category/page/:subpage",
  "154  /blog/:category/page/*subpage",
  "152  /blog/:category/:page/subpage",
  "151  /blog/:category/:page/:subpage",
  "150  /blog/:category/:page/*subpage",
  "149  /:blog/category/page/subpage",
  "148  /:blog/category/page/:subpage",
  "147  /:blog/category/page/*subpage",
  "145  /:blog/category/:page/subpage",
  "144  /:blog/category/:page/:subpage",
  "143  /:blog/category/:page/*subpage",
  "140  /:blog/:category/page/subpage",
  "139  /:blog/:category/page/:subpage",
  "138  /:blog/:category/page/*subpage",
  "136  /:blog/:category/:page/subpage",
  "135  /:blog/:category/:page/:subpage",
  "134  /:blog/:category/:page/*subpage",
  "90  /blog/category/page",
  "89  /blog/category/:page",
  "88  /blog/category/*page",
  "86  /blog/:category/page",
  "85  /blog/:category/:page",
  "84  /blog/:category/*page",
  "81  /:blog/category/page",
  "80  /:blog/category/:page",
  "79  /:blog/category/*page",
  "77  /:blog/:category/page",
  "76  /:blog/:category/:page",
  "75  /:blog/:category/*page",
  "42  /blog/category",
  "41  /blog/:category",
  "40  /blog/*category",
  "15  /blog",
  "14  /:blog",
  "13  /*blog",
 */

//	"/blog/category/page"		  = [-, -, -]	3*3 + 2*3 + 1*3 = 9 + 6 + 3 = 18
//	"/blog/category/:page"		= [-, -, :]	3*3 + 2*3 + 1*2 = 9 + 6 + 2 = 17
//	"/blog/category/*page"		= [-, -, *]	3*3 + 2*3 + 1*1 = 9 + 6 + 1 = 16
//	"/blog/:category/page"		= [-, :, -]	3*3 + 2*2 + 1*3 = 9 + 4 + 3 = 16
//	"/blog/:category/:page"		= [-, :, :]	3*3 + 2*2 + 1*2 = 9 + 4 + 2 = 15
//	"/blog/:category/*page"		= [-, :, *]	3*3 + 2*2 + 1*1 = 9 + 4 + 1 = 14
//	"/:blog/category/page"		= [:, -, -]	3*2 + 2*3 + 1*3 = 6 + 6 + 3 = 15
//	"/:blog/category/:page"		= [:, -, :]	3*2 + 2*3 + 1*2 = 6 + 6 + 2 = 14
//	"/:blog/category/*page"		= [:, -, *]	3*2 + 2*3 + 1*1 = 6 + 6 + 1 = 13
//	"/:blog/:category/page"	  = [:, :, -]	3*3 + 2*3 + 1*3 = 9 + 6 + 3
//	"/:blog/:category/:page"	= [:, :, :]	3*3 + 2*3 + 1*3 = 9 + 6 + 3
//	"/:blog/:category/*page"	= [:, :, *]	3*3 + 2*3 + 1*3 = 9 + 6 + 3
//	"/blog/category"				  = [-, -]
//	"/blog/:category"			    = [-, :]
//	"/blog/*category"				  = [-, *]
//	"/blog"						        = [-, -]
//	"/:blog"					        = [-, :]
//	"/*blog"					        = [-, *]

let MIDDLEWARES_BY_PATH_LENGTH = [
  null,
  /* 1 */
  [
    ["user", "*", [[() => "/user/*filepath", ["filepath"]]]],
    ["user", ":", [[() => "/user/:userId", ["userId"]]]],
  ],
]
