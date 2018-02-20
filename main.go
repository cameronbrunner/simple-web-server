package main

import (
	"github.com/go-redis/redis"

	"html/template"
	"log"
	"net/http"
	"regexp"
)

// Redis section...

func NewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return client
}

type Page struct {
	Title string
	Body  []byte
	Pages []string
}

func (p *Page) save(client *redis.Client) error {
	return client.Set(p.Title, p.Body, 0).Err()
}

func loadPage(client *redis.Client, title string) (*Page, error) {
	body, err := client.Get(title).Result()
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: []byte(body)}, nil
}

func viewHandler(client *redis.Client, w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(client, title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(client *redis.Client, w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(client, title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

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

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "list.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

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

func main() {
	client := NewClient()
	http.HandleFunc("/view/", makeHandler(client, viewHandler))
	http.HandleFunc("/edit/", makeHandler(client, editHandler))
	http.HandleFunc("/save/", makeHandler(client, saveHandler))
	http.HandleFunc("/", listHandler(client))

	log.Fatal(http.ListenAndServe(":8085", nil))
}
