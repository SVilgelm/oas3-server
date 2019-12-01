package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"github.com/SVilgelm/oas3-server/pkg/config"
	"github.com/SVilgelm/oas3-server/pkg/server"
)

// Page is a structure to render templates
type Page struct {
	Title    string
	Body     string
	Articles []string
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: string(body)}, nil
}

func noescape(str string) template.HTML {
	return template.HTML(str)
}

var fn = template.FuncMap{
	"noescape": noescape,
}

var templates = map[string]*template.Template{
	"edit": template.Must(
		template.ParseFiles(
			"templates/base.html",
			"templates/edit.html",
		),
	).Funcs(fn),
	"view": template.Must(
		template.ParseFiles(
			"templates/base.html",
			"templates/view.html",
		),
	).Funcs(fn),
}

func renderTemplate(w http.ResponseWriter, name string, p *Page) {
	files, err := ioutil.ReadDir("data")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	for _, file := range files {
		if file.Name() == ".keep" {
			continue
		}
		title := strings.TrimSuffix(file.Name(), ".txt")
		p.Articles = append(p.Articles, title)
	}

	err = templates[name].ExecuteTemplate(w, name+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var body strings.Builder
	if body.Len() > 0 {
		body.WriteString("</ul>")
	}

	p := &Page{
		Title: "",
	}
	renderTemplate(w, "view", p)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]

	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: body}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		println("Config Validation Error: ", err.Error())
		os.Exit(1)
	}
	srv := server.NewServer(cfg)
	srv.HandleFunc("wiki.list", listHandler)
	srv.HandleFunc("wiki.create", createHandler)
	srv.HandleFunc("wiki.edit", editHandler)
	srv.HandleFunc("wiki.save", saveHandler)
	srv.HandleFunc("wiki.view", viewHandler)
	srv.Serve()
}
