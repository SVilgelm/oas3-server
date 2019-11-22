package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/SVilgelm/oas3-server/pkg/config"
	"github.com/SVilgelm/oas3-server/pkg/server"
)

type Page struct {
	Title string
	Body  string
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

var templates = template.Must(
	template.New("").
		Funcs(fn).
		ParseFiles(
			"templates/edit.html",
			"templates/view.html",
			"templates/list.html",
		),
)

func renderTemplate(w http.ResponseWriter, name string, p *Page) {
	err := templates.ExecuteTemplate(w, name+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var body strings.Builder
	files, err := ioutil.ReadDir("data")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	for _, file := range files {
		if file.Name() == ".keep" {
			continue
		}
		if body.Len() == 0 {
			body.WriteString("<ul>")
		}
		name := strings.TrimSuffix(file.Name(), ".txt")
		body.WriteString("<li><a href=\"/view/")
		body.WriteString(name)
		body.WriteString("\">")
		body.WriteString(name)
		body.WriteString("</a></li>")
	}
	if body.Len() > 0 {
		body.WriteString("</ul>")
	}

	p := &Page{
		Title: "Pages",
		Body:  body.String(),
	}
	renderTemplate(w, "list", p)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.RequestURI, "/edit/")
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.RequestURI, "/edit/")
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
	title := strings.TrimPrefix(r.RequestURI, "/view/")
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func main() {
	cfg := config.SafeLoad("config.yaml")
	srv := server.NewServer(cfg)
	srv.HandleFunc("wiki.list", listHandler)
	srv.HandleFunc("wiki.create", createHandler)
	srv.HandleFunc("wiki.edit", editHandler)
	srv.HandleFunc("wiki.save", saveHandler)
	srv.HandleFunc("wiki.view", viewHandler)
	srv.Serve()
}
