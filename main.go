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

	"github.com/joho/godotenv"
)

var noCacheHeaders = map[string]string{
	"Expires":         "0",
	"Cache-Control":   "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

func imageHandler(w http.ResponseWriter, r *http.Request) {

	regex, err := regexp.Compile("[0-9]+")
	if err != nil {
		fmt.Println(err)
	}

	paths, err := filepath.Glob("./Toasters/Toasters/*")
	if err != nil {
		panic(err)
	}

	number, err := strconv.Atoi(regex.FindString(r.URL.Path))
	if err != nil {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			// json.NewEncoder(w).Encode({'info': 'Toaster Image API', 'max': %d})
			fmt.Fprintf(w, "{'info': 'Toaster Image API', 'max': %d}", len(paths)-1)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 Not Found")
		}
		return
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

func randomImage(w http.ResponseWriter, r *http.Request) {

	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}

	paths, err := filepath.Glob("./Toasters/Toasters/*")
	if err != nil {
		panic(err)
	}

	regex, err := regexp.Compile("[0-9]+")
	if err != nil {
		fmt.Println(err)
	}

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

	number := rand.Intn(len(paths) - 1)

	fmt.Println(number, paths[number])

	http.ServeFile(w, r, paths[number])
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	http.HandleFunc("/", imageHandler)
	http.HandleFunc("/random", randomImage)
	fmt.Println("Listening on port", os.Getenv("PORT"))
	http.ListenAndServe(os.Getenv("PORT"), nil)
}
