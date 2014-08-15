package reverseproxy

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	"github.com/mix3/go-tenco/storage"
)

type reverseProxy struct {
	storage     storage.Storager
	domain      string
	matcher     *regexp.Regexp
	rootHandler Handler
}

type Handler func(http.ResponseWriter, *http.Request, storage.Storager)

func split(host string) (string, string) {
	ret := strings.Split(host, ":")
	if len(ret) <= 1 {
		return ret[0], ""
	}
	return ret[0], ret[1]
}

func (rp *reverseProxy) parseSubDomain(host string) string {
	match := rp.matcher.FindStringSubmatch(host)
	if 1 < len(match) {
		return match[1]
	}
	return ""
}

func (rp *reverseProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	host, _ := split(req.Host)
	subdomain := rp.parseSubDomain(host)
	if subdomain == "" {
		if rp.rootHandler == nil {
			http.NotFound(w, req)
		} else {
			rp.rootHandler(w, req, rp.storage)
		}
		return
	}
	data, err := rp.storage.Get(subdomain)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			http.NotFound(w, req)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	u, err := url.Parse(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httputil.NewSingleHostReverseProxy(u).ServeHTTP(w, req)
}

func New(domain string, storage storage.Storager, handler Handler) *reverseProxy {
	return &reverseProxy{
		storage:     storage,
		domain:      domain,
		matcher:     regexp.MustCompile(fmt.Sprintf(`^(.+?)\.%s$`, domain)),
		rootHandler: handler,
	}
}
