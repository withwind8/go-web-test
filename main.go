//go:generate go-bindata  -prefix "tmpl" tmpl

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/withwind8/go-web-test/logger"
	"github.com/withwind8/middleware"
)

type Page struct {
	Title string `json:"title"`
	Body  string `json:"content"`
}

var datePath = "data/"

func (p *Page) save() error {
	filename := datePath + p.Title + ".txt"
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
	filename := datePath + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: string(body)}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/欢迎", http.StatusFound)
}

var templates *template.Template

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var pageLink = regexp.MustCompile(`\[(.+)\]`)

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// title, err := getTitle(w, r)
		// if err != nil {
		// 	return
		// }
		title := mux.Vars(r)["title"]
		fn(w, r, title)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// for json
func renderPage(w http.ResponseWriter, r *http.Request, tmpl string, p *Page) {
	if r.Header.Get("accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(p)
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		renderTemplate(w, tmpl, p)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/"+title+"/edit", http.StatusFound)
		return
	}

	renderPage(w, r, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: body}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/"+title, http.StatusFound)
}

func basicAuth(h http.HandlerFunc, requiredUser, requiredPassword string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, hasAuth := r.BasicAuth()
		log.Print(user, password, hasAuth)
		if hasAuth && user == requiredUser && password == requiredPassword {
			h(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func initTmpl() {
	templates = template.New("t").Funcs(template.FuncMap{
		"parselink": func(s string) template.HTML {
			return template.HTML(string(pageLink.ReplaceAllFunc([]byte(s), func(str []byte) []byte {
				matched := pageLink.FindStringSubmatch(string(str))[1]
				return []byte(fmt.Sprintf("<a href=\"/%s\">%s</a>", matched, matched))
			})))
		}})
	for _, tmpl := range AssetNames() {
		templates = template.Must(templates.New(tmpl).Parse(string(MustAsset(tmpl))))
	}
}

func main() {
	initTmpl()

	app := middleware.New()
	app.Use(logger.New())

	router := mux.NewRouter()
	router.HandleFunc("/", handler)
	router.Handle("/favicon.ico", http.NotFoundHandler())
	router.HandleFunc("/{title}", makeHandler(viewHandler)).Methods("GET")
	router.HandleFunc("/{title}", makeHandler(saveHandler)).Methods("POST")
	router.HandleFunc("/{title}/edit", basicAuth(makeHandler(editHandler), "admin", "123456")).Methods("GET")
	app.UseHandler(router)

	log.Fatal(app.Listen(":8080"))
}
