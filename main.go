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

	go addStats()
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

	go addStats()
	tmpl.Execute(w, page)
}

func api(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	paths, err := filepath.Glob("./Toasters/Toasters/*")
	if err != nil {
		panic(err)
	}

	stats := getStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{'info': 'Toaster Image API. Routes are /img/[number], /img/random, /embed/[number], /img/random. Add /[anything] to the end of the URL to avoid cache issues', 'max': %d, 'hits': '%d'}", len(paths)-1, stats.Hits)
	go addStats()
}

func getStats() Stats {
	jsonFile, err := os.Open("stats.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var stats Stats
	json.Unmarshal(byteValue, &stats)
	return stats
}

func addStats() {
	jsonFile, err := os.Open("stats.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var stats Stats
	json.Unmarshal(byteValue, &stats)
	stats.Hits++
	file, _ := json.MarshalIndent(stats, "", " ")
	ioutil.WriteFile("stats.json", file, 0644)
}

type Stats struct {
	Hits int `json:"hits"`
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
