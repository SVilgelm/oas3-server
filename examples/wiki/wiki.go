package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SVilgelm/oas3-server/pkg/utils"

	"github.com/gorilla/mux"

	"github.com/SVilgelm/oas3-server/pkg/config"
	"github.com/SVilgelm/oas3-server/pkg/server"
)

var dataFolder string = "data"

// Page is a structure to render templates
type Page struct {
	Title    string
	Body     string
	Articles []string
}

func (p *Page) save() error {
	filename := filepath.Join(dataFolder, p.Title+".txt")
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
	filename := filepath.Join(dataFolder, title+".txt")
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
	if utils.Contains(utils.GetContentTypes(r), "application/json") {
		var res []string
		files, err := ioutil.ReadDir(dataFolder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, file := range files {
			if file.Name() == ".keep" {
				continue
			}
			title := strings.TrimSuffix(file.Name(), ".txt")
			res = append(res, title)
		}
		data, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.Write(data)
		return
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

func initServer() *server.Server {
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
	return srv
}

func main() {
	err := initServer().Serve()
	if err != nil {
		log.Fatal(err)
	}
}
