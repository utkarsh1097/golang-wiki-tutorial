package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

/*
	This tutorial guides you in creating a wiki-like website.
*/

//Data structure meant for describing a wiki page
//Wiki is a website of interconnected pages with titles and content
type Page struct {
	Title string
	Body  []byte //[]byte, and not string because this is the type expected by the 'io/ioutil' module
}

func (p *Page) save() error { //error is Nil if everything goes smooth
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600) //from go 1.16, same as os.WriteFile
}

func loadpage(title string) (*Page, error) { //error handling for case where said file does not exist
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, err
}

/*
	template renderer for handlers
	template.ParseFiles reads the template files everytime which is heavy on io.
	For that we can instead cache the files at once during program initialization.
*/

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))	//template.Must handles panic situations, so no need to handle nil cases separately

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl + ".html", p)
	if err != nil {
		http.Error(w, "soemthing weird is happening: "+err.Error(), http.StatusInternalServerError)
	}
}

/*
	To prevent clients from passing arbitrary paths to the server, we can do a regex validation.
	The variable below stores the rules to be checked for in a variable 
*/

var validPath = regexp.MustCompile("^/(save|edit|view)/([a-zA-Z0-9]+)$")

/*
	Extract the title from the URL using path validation
*/
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid page title")
	}
	fmt.Println(m)
	return m[2], nil
}

/*
	Handler meant to be used when requests hit a specified path
	w: response object. response is written to the client
	r: request object. it is the request sent from client
*/
func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return 
	}
	p, err := loadpage(title)
	if err != nil {
		http.Error(w, "could not find page.", http.StatusNotFound)
	} else {
		renderTemplate(w, "view", p)
	}
}

/*
	Handler for displaying an "edit form" that takes in information
	for creating a new wiki page, if a page does not exist. If it does,
	display the pre-filled information in form?
*/
func editHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return 
	}
	p, err := loadpage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

/*
	Handler for saving a submitted page. In case the page already exists,
	the existing one is ovewritten.
*/
func saveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	p := &Page{Body: []byte(body), Title: title}
	err = p.save()
	if err == nil {
		http.Redirect(w, r, "/view/"+title, http.StatusFound)
	} else {
		http.Error(w, "soemthing weird is happening: "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	//handlers
	http.HandleFunc("/view/", viewHandler) //viewHandler assigned to "/view/" path
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	log.Panic(http.ListenAndServe("localhost:8080", nil)) //listen and raise panic if error thrown. error throws iff program exits\

}
