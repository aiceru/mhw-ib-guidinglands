package main

import (
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"net/http"
)

var renderer *render.Render

func init() {
	renderer = render.New()
}

func mainpage(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	renderer.HTML(w, http.StatusOK, "index", map[string]string{"title": "MHW IB Guiding Land Monster List"})
}

func main() {
	router := httprouter.New()

	router.GET("/", mainpage)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}

