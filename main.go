package main

import (
	"github.com/go-redis/redis"

	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
)

// Build a new redis client against a server
func NewClient(server string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     server + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return client
}

// Structure for holding page data
type Page struct {
	Title   string
	Body    []byte
	Pages   []string
	Version string
}

// Write a page's state to redis
func (p *Page) save(client *redis.Client) error {
	return client.Set(p.Title, p.Body, 0).Err()
}

// Load a page's state from redis
func loadPage(client *redis.Client, title string) (*Page, error) {
	body, err := client.Get(title).Result()
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: []byte(body)}, nil
}

// Handle view requests
func viewHandler(client *redis.Client, w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(client, title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// Handle edit requests
func editHandler(client *redis.Client, w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(client, title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// Handle save requests
func saveHandler(client *redis.Client, w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// Initialize our tempaltes
var templates = template.Must(template.ParseFiles("edit.html", "view.html", "list.html"))

// Render a template using a given page state
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	p.Version = runtime.Version()
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Valid paths for our pages
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// Make a handler for our page operations
func makeHandler(client *redis.Client, fn func(*redis.Client, http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(client, w, r, m[2])
	}
}

// Handle our index page
func listHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pages, err := client.Keys("*").Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderTemplate(w, "list", &Page{Pages: pages})
	}
}

// Main entry point for applicaiton
func main() {
	server := "localhost"
	if len(os.Args) > 1 {
		server = os.Args[1]
	}
	client := NewClient(server)
	http.HandleFunc("/view/", makeHandler(client, viewHandler))
	http.HandleFunc("/edit/", makeHandler(client, editHandler))
	http.HandleFunc("/save/", makeHandler(client, saveHandler))
	http.HandleFunc("/", listHandler(client))

	log.Fatal(http.ListenAndServe(":8085", nil))
}
