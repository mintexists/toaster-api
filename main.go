package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	paths, max := loadPaths()

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Use(middleware.RedirectSlashes)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"*"},
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		type Route struct {
			Route       string `json:"route"`
			Method      string `json:"method"`
			Description string `json:"description"`
		}
		type Api struct {
			Info   string  `json:"info"`
			Routes []Route `json:"routes"`
			Max    int     `json:"max"`
		}
		api := Api{
			Info: "Toaster Image API",
			Routes: []Route{
				{
					Route:       "/",
					Method:      "GET",
					Description: "Get the API info",
				},
				{
					Route:       "/{id}",
					Method:      "GET",
					Description: "Get the image by id",
				},
				{
					Route:       "/random",
					Method:      "GET",
					Description: "Get a random image",
				},
				{
					Route:       "/embed/{id}",
					Method:      "GET",
					Description: "Get the image as an opengraph/twitter embed card",
				},
				{
					Route:       "/embed/random",
					Method:      "GET",
					Description: "Get a random image as an opengraph/twitter embed card",
				},
			},
			Max: max - 1,
		}
		json.NewEncoder(w).Encode(api)
	})

	r.Get("/{img:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		number, _ := strconv.Atoi(chi.URLParam(r, "img"))
		number = number % max

		fileBytes, err := ioutil.ReadFile(paths[number])
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(fileBytes)
	})

	r.Get("/embed/{img:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("embed.html")
		if err != nil {
			log.Println(err)
			return
		}

		number, _ := strconv.Atoi(chi.URLParam(r, "img"))
		number = number % max

		url := fmt.Sprintf("http://%s/%d", r.Host, number)

		type data struct {
			ID  int
			URL string
		}

		d := data{
			ID:  number,
			URL: url,
		}

		tmpl.Execute(w, d)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)

		r.Get("/random", func(w http.ResponseWriter, r *http.Request) {
			number := rand.Intn(len(paths)) % max

			fileBytes, err := ioutil.ReadFile(paths[number])
			if err != nil {
				panic(err)
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(fileBytes)
		})

		r.Get("/embed/random", func(w http.ResponseWriter, r *http.Request) {
			tmpl, err := template.ParseFiles("embed.html")
			if err != nil {
				log.Println(err)
				return
			}

			number := rand.Intn(len(paths)) % max

			url := fmt.Sprintf("http://%s/%d", r.Host, number)

			type data struct {
				ID  int
				URL string
			}

			d := data{
				ID:  number,
				URL: url,
			}

			tmpl.Execute(w, d)
		})
	})

	http.ListenAndServe(os.Getenv("PORT"), r)
}

func loadPaths() ([]string, int) {
	paths, err := os.ReadDir("./Toasters/Toasters/")
	if err != nil {
		panic(err)
	}

	regex, err := regexp.Compile("[0-9]+")
	if err != nil {
		fmt.Println(err)
	}

	max := len(paths)

	sort.Slice(paths, func(i, j int) bool {
		a, err := strconv.Atoi(regex.FindString(paths[i].Name()))
		if err != nil {
			panic(err)
		}
		b, err := strconv.Atoi(regex.FindString(paths[j].Name()))
		if err != nil {
			panic(err)
		}
		return a < b
	})

	pathString := make([]string, max)

	for i, path := range paths {
		pathString[i] = filepath.Join("./Toasters/Toasters/", path.Name())
	}

	return pathString, max
}
