// Cannibalized from:
// https://golang.org/doc/articles/wiki/
// https://www.freecodecamp.org/news/how-i-set-up-a-real-world-project-with-go-and-vue/
// https://github.com/neo4j-examples/movies-golang-bolt

package main

import (
	"fmt"
	// "encoding/json"
	// "io/ioutil"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	// "io"
	"log"
	"net/http"
	// "net/url"
	// "os"
	// "strconv"
	// "strings"

	"github.com/daaku/go.httpgzip"
)

// Handler Functions for routes / API endpoints 


// returns a function that returns the http response.
// You can test this with curl -v http://localhost:8080/search !!!! (and it returns! Nice!)
// ( or just type http://localhost:8080/search into browser but that won't work soon hehehe.)
func searchHandlerFunc(driver neo4j.Driver, database string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// w.Header().Set("Content-Type", "application/json")

		_, err := fmt.Fprintf(w, "search api called.")

		if err != nil {
			log.Println("error querying search:", err)
			return
		}
	}
}



// Main controller.

func main() {
	fmt.Println("You know.")

	port := "8080" // just set it to 8080. Ideally would get this from an ENV var ('LookupEnv' & setting it in the build vars)

	// defer unsafeClose(driver) // 'defer' is basically 'finally', executes on function return

	serveMux := http.NewServeMux()
	// serveMux.HandleFunc("/", defaultHandler)
	serveMux.HandleFunc("/search", searchHandlerFunc(nil, "blah"))
	// serveMux.HandleFunc("/adjacentnodes/", movieHandlerFunc(driver, configuration.Database))
	// serveMux.HandleFunc("/graphdata", graphHandler(driver, configuration.Database))

	fmt.Printf("Running on port %s, database is who knows where...", port)
	// fmt.Printf("Running on port %s, database is at %s\n", port, configuration.Url)

	// the handler below wants functions, so the handlers above should return functions ( functions that return functions! )
	panic(http.ListenAndServe(":"+port, httpgzip.NewHandler(serveMux)))

}
