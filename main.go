package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

var noCacheHeaders = map[string]string{
	"Expires":         "0",
	"Cache-Control":   "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

func imageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	regex, err := regexp.Compile("[0-9]+")
	if err != nil {
		fmt.Println(err)
	}

	paths, err := filepath.Glob("./Toasters/Toasters/*")
	if err != nil {
		panic(err)
	}

	number, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		if strings.Split(ps.ByName("id"), "/")[0] == "random" {
			for k, v := range noCacheHeaders {
				w.Header().Set(k, v)
			}
			number = rand.Intn(len(paths) - 1)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 Not Found")
			return
		}
	}

	number = number % len(paths)

	sort.Slice(paths, func(i, j int) bool {
		a, err := strconv.Atoi(regex.FindString(paths[i]))
		if err != nil {
			panic(err)
		}
		b, err := strconv.Atoi(regex.FindString(paths[j]))
		if err != nil {
			panic(err)
		}
		return a < b
	})

	fmt.Println(number, paths[number])

	http.ServeFile(w, r, paths[number])
}

type EmbedData struct {
	PagePath string
	ID       int
	URL      string
}

func embedImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println(r.URL.Scheme)
	tmpl, err := template.ParseFiles("embed.html")
	if err != nil {
		log.Println(err)
		return
	}

	paths, err := filepath.Glob("./Toasters/Toasters/*")
	if err != nil {
		panic(err)
	}

	number, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		fmt.Println("Nya")
		fmt.Println(strings.Split(ps.ByName("id"), "/")[0])
		if strings.Split(ps.ByName("id"), "/")[0] == "random" {
			for k, v := range noCacheHeaders {
				w.Header().Set(k, v)
			}
			number = rand.Intn(len(paths) - 1)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 Not Found")
			return
		}
	}

	number = number % len(paths)
	pagePath := fmt.Sprintf("http://%s/img/%d", r.Host, number)

	page := EmbedData{
		PagePath: pagePath,
		ID:       number,
		URL:      pagePath,
	}

	tmpl.Execute(w, page)
}

func api(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	paths, err := filepath.Glob("./Toasters/Toasters/*")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{'info': 'Toaster Image API', 'max': %d}", len(paths)-1)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	router := httprouter.New()
	router.GET("/embed/:id", embedImage)
	router.GET("/embed/:id/*_", embedImage)
	router.GET("/img/:id/*_", imageHandler)
	router.GET("/img/:id", imageHandler)
	router.GET("/", api)
	fmt.Println("Listening on port", os.Getenv("PORT"))
	http.ListenAndServe(os.Getenv("PORT"), router)
}
