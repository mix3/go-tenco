package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/mix3/go-tenco/reverseproxy"
	"github.com/mix3/go-tenco/storage"
	"github.com/mix3/go-tenco/storage/sqlite"
	"github.com/unrolled/render"
)

var myRender = render.New(render.Options{})

func parseHost(host string) (string, string) {
	ret := strings.Split(host, ":")
	if len(ret) <= 1 {
		return ret[0], ""
	}
	return ret[0], ret[1]
}

func checkAllowedHost(host string) bool {
	for _, v := range opts.AllowIP {
		if v == host {
			return true
		}
	}

	for _, v := range opts.AllowIPNet {
		if v.Contains(net.ParseIP(host)) {
			return true
		}
	}

	for _, v := range opts.AllowHost {
		if v == host {
			return true
		}
	}

	return false
}

func start(host string, handler http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(host)
}

func renderJson(w http.ResponseWriter, code int, result interface{}) {
	if result != nil {
		myRender.JSON(w, code, map[string]interface{}{
			"code":    fmt.Sprintf("%d", code),
			"message": http.StatusText(code),
			"result":  result,
		})
	} else {
		myRender.JSON(w, code, map[string]interface{}{
			"code":    fmt.Sprintf("%d", code),
			"message": http.StatusText(code),
		})
	}
}

func renderErrJson(w http.ResponseWriter, code int, err error) {
	renderJson(w, code, map[string]string{
		"error": err.Error(),
	})
}

func api(w http.ResponseWriter, req *http.Request, storage storage.Storager) {
	switch req.Method {
	case "GET":
		apiGet(w, req, storage)
	case "POST":
		apiSet(w, req, storage)
	case "DELETE":
		apiDel(w, req, storage)
	default:
		renderJson(w, http.StatusNotFound, nil)
	}
}

func apiGet(w http.ResponseWriter, req *http.Request, storage storage.Storager) {
	subdomain := req.FormValue("subdomain")
	if subdomain != "" {
		val, err := storage.Get(subdomain)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				renderErrJson(w, http.StatusNotFound, err)
			default:
				renderErrJson(w, http.StatusInternalServerError, err)
			}
			return
		}
		renderJson(w, http.StatusOK, map[string]string{
			subdomain: val,
		})
		return
	}

	// all
	val, err := storage.Map()
	if err != nil {
		renderErrJson(w, http.StatusInternalServerError, err)
		return
	}

	renderJson(w, http.StatusOK, val)
}

func apiSet(w http.ResponseWriter, req *http.Request, storage storage.Storager) {
	subdomain := req.FormValue("subdomain")

	if subdomain == "" {
		renderErrJson(w, http.StatusForbidden, fmt.Errorf("required: subdomain"))
		return
	}

	backend := req.FormValue("backend")
	if backend == "" {
		renderErrJson(w, http.StatusForbidden, fmt.Errorf("required: backend"))
		return
	}

	u, err := url.Parse(backend)
	if err != nil {
		renderErrJson(w, http.StatusInternalServerError, err)
		return
	}

	host, _ := parseHost(u.Host)
	if !checkAllowedHost(host) {
		renderErrJson(w, http.StatusBadRequest, fmt.Errorf("not allowed: %s", host))
		return
	}

	err = storage.Set(subdomain, backend)
	if err != nil {
		renderErrJson(w, http.StatusInternalServerError, err)
		return
	}

	renderJson(w, http.StatusOK, nil)
}

func apiDel(w http.ResponseWriter, req *http.Request, storage storage.Storager) {
	subdomain := req.FormValue("subdomain")

	if subdomain != "" {
		err := storage.Delete(subdomain)
		if err != nil {
			renderErrJson(w, http.StatusInternalServerError, err)
			return
		}
		renderJson(w, http.StatusOK, nil)
		return
	}

	// all
	all, err := storage.Map()
	if err != nil {
		renderErrJson(w, http.StatusInternalServerError, err)
		return
	}
	for k, _ := range all {
		err := storage.Delete(k)
		if err != nil {
			renderErrJson(w, http.StatusInternalServerError, err)
			return
		}
	}

	renderJson(w, http.StatusOK, nil)
}

func apiHandler(w http.ResponseWriter, req *http.Request, s storage.Storager) {
	switch req.URL.Path {
	case "/api":
		api(w, req, s)
	default:
		renderJson(w, http.StatusNotFound, nil)
	}
}

func main() {
	storage, err := sqlite.New(opts.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	proxyHandler := reverseproxy.New(
		opts.Domain,
		storage,
		apiHandler,
	)

	start(
		fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		proxyHandler,
	)
}
