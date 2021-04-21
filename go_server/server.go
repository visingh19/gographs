// Cannibalized from:
// https://golang.org/doc/articles/wiki/
// https://www.freecodecamp.org/news/how-i-set-up-a-real-world-project-with-go-and-vue/
// https://github.com/neo4j-examples/movies-golang-bolt

package main

import (
	"fmt"
	"math/rand"
	"time"
	"encoding/json"
	// "io/ioutil"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"io"
	"log"
	"net/http"
	// "net/url"
	// "os"
	"strconv"
	"strings"

	"github.com/daaku/go.httpgzip"
)

// Handler Functions for routes / API endpoints 

///////////////////////////////////////////////////////
//////////    DATA STRUCTS         ////////////////////
///////////////////////////////////////////////////////

// the json:"identifier" means on encoding, we get identifier:value

type Person struct {
	Name string `json:"name"`
}

type D3Data struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

type Node struct {
	Title string `json:"title"`
	Label string `json:"label"`
}

type Link struct {
	Source int `json:"source"`
	Target int `json:"target"`
	Relationship string `json:"relationship"` //extra, express relation ( e.g. A BUYS_FROM B)
}

type Neo4jConfiguration struct {
	Url      string
	Username string
	Password string
	Database string
}

///////////////////////////////////////////////////////
/////////// API ENDPOINTS       ///////////////////////
///////////////////////////////////////////////////////


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


// some helpers ripped from https://github.com/neo4j-examples/movies-golang-bolt demo code.
// now connects.
func graphHandler(driver neo4j.Driver, database string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		session := driver.NewSession(neo4j.SessionConfig{
			AccessMode:   neo4j.AccessModeRead,
			DatabaseName: database,
		})
		defer unsafeClose(session)

		limit := parseLimit(req)
		query := `MATCH (m:Movie)<-[:ACTED_IN]-(a:Person)
				  RETURN m.title as movie, collect(a.name) as cast
				  LIMIT $limit `

		d3Resp, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			records, err := tx.Run(query, map[string]interface{}{"limit": limit})
			if err != nil {
				return nil, err
			}
			result := D3Data{}
			for records.Next() {
				record := records.Record()
				title, _ := record.Get("movie")
				actors, _ := record.Get("cast")
				result.Nodes = append(result.Nodes, Node{Title: title.(string), Label: "movie"})
				movIdx := len(result.Nodes) - 1
				for _, actor := range actors.([]interface{}) {
					idx := -1
					for i, node := range result.Nodes {
						if actor == node.Title && node.Label == "actor" {
							idx = i
							break
						}
					}
					if idx == -1 {
						result.Nodes = append(result.Nodes, Node{Title: actor.(string), Label: "actor"})
						result.Links = append(result.Links, Link{Source: len(result.Nodes) - 1, Target: movIdx})
					} else {
						result.Links = append(result.Links, Link{Source: idx, Target: movIdx})
					}
				}
			}
			return result, nil
		})
		if err != nil {
			log.Println("error querying graph:", err)
			return
		}
		err = json.NewEncoder(w).Encode(d3Resp)
		if err != nil {
			log.Println("error writing graph response:", err)
		}
	}
}

///////////////////////////////////////////////////////
////////// CONFIGS & HELPERS //////////////////////////
///////////////////////////////////////////////////////

// neo4j driver configuration. Ideally set by env vars ( 'LookupEnv', etc )
// Everything is pass-by-value so you gotta pointer your structs or else?

func setNeo4jConfigs() *Neo4jConfiguration {
	return &Neo4jConfiguration {
		Url:		"bolt://localhost",
		Username:	"neo4j",
		Password:	"root", // You really want this as an env var so it's not just whoosh, available.
		Database:	"", // on neo4j v4, this should just be empty string.
	}
}

// func to empty out DB ( todo )
func fillNeo4jDB(driver neo4j.Driver, database string) (string, error) {

	session := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: database,
	})
	defer unsafeClose(session)


	var fNames = []string{"Alice", "Bob", "Carol", "Danny", "Eve", "Finn", "Gina", "Hank", "Irene", "John" } // 10
	var lNames = []string{"Andrews", "Brick", "Carter", "Daves", "Erickson"} // 5 -> max unique AWS 'persons' = 50 ( 10 x 5 )
	var relations = []string{"MANAGES", "REPORTS_TO", "PAYS", "MONITORS", "AUTHORIZES"} // 5 -> 5 random relationships ( and yes, a person can manage themselves haha. )

	// Now we create 200 relationships - somewhat dense --> (person)-(relation)->(person) 200 times.
	// At worst ( pigeonhole principle ) each of the 50 possible nodes has 4 relations, fairly connected!

	// Because there are few nodes ( 50 or less ) and many relations ( 200 ), deletion should be easy.
	// Does send 600 lines to the transaction, so that might cause problems?

	var query strings.Builder
	rand.Seed(time.Now().Unix())
	for i := 0; i < 200; i++ {
		fullNameOne := fNames[rand.Intn(len(fNames))] + " " + lNames[rand.Intn(len(lNames))]
		fullNameTwo := fNames[rand.Intn(len(fNames))] + " " + lNames[rand.Intn(len(lNames))]
		relation := relations[rand.Intn(len(relations))]
		iStr := strconv.Itoa(i);


		subQuery :=	"MERGE (A"+iStr+":Person {name:'"+fullNameOne+"'})\n" +
					"MERGE (B"+iStr+":Person {name:'"+fullNameTwo+"'})\n" +
					"MERGE (A"+iStr+")-[:"+relation+"]->(B"+iStr+")\n"

		query.WriteString(subQuery)
	}

	// Query String exists.

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			query.String(), map[string]interface{}{})
		if err != nil {
			return nil, err
		}

		return nil, nil
	}) // end write.

	if err != nil { // if write had errors, return them.
		return "", err
	}

	return "Worked", nil

}


// func to add random data into the DB ( todo ) - mocking 'getting aws resource'
// adds 50 random people and 100 random relationships between them based on a set of relationship names.


// func to just print every 'Person' ( to console, not an API! )
func actorPrinter(driver neo4j.Driver, database string) {

	log.Println("Hi")

	// https://neo4j.com/developer/go/
	session := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: database,
	})
	defer unsafeClose(session)

	// cypher query works in neo4j desktop...
	limit := 10
	query := `MATCH (p:Person)
			  RETURN p.name as name 
			  LIMIT $limit `

	log.Println(query)

	actorNames, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		log.Println("Inside read transaction.");
		records, err := tx.Run(query, map[string]interface{}{"limit": limit})
		if err != nil {
			return nil, err
		}
		var result []string;
		for records.Next() {
			record := records.Record()
			name, _ := record.Get("name");
			result = append(result, name.(string));
		}
		return result, nil
	})
	if err != nil {
		log.Println("error querying graph:", err)
		return
	}

	log.Println("Actor names are %s", actorNames);
}

// func to add random message to db... 
// https://neo4j.com/developer/go/
func helloWorld(driver neo4j.Driver, database string) (string, error) {
	// already done by main's 'NewDriver' method on neo4jConfig
	// driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	// if err != nil {
	// 	return "", err
	// }
	// defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: database,
	})
	defer session.Close()

	greeting, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
			map[string]interface{}{"message": "hello, world"})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return "", err
	}

	return greeting.(string), nil
}

// some helpers ripped from https://github.com/neo4j-examples/movies-golang-bolt demo code.
func unsafeClose(closeable io.Closer) {
	if err := closeable.Close(); err != nil {
		log.Fatal(fmt.Errorf("could not close resource: %w", err))
	}
}

// some helpers ripped from https://github.com/neo4j-examples/movies-golang-bolt demo code.
func (nc *Neo4jConfiguration) newDriver() (neo4j.Driver, error) {
	return neo4j.NewDriver(nc.Url, neo4j.BasicAuth(nc.Username, nc.Password, ""))
}


// some helpers ripped from https://github.com/neo4j-examples/movies-golang-bolt demo code.
func parseLimit(req *http.Request) int {
	limits := req.URL.Query()["limit"]
	limit := 50
	if len(limits) > 0 {
		var err error
		if limit, err = strconv.Atoi(limits[0]); err != nil {
			limit = 50
		}
	}
	return limit
}


///////////////////////////////////////////////////////
/////////// MAIN CONTROLLER ///////////////////////////
///////////////////////////////////////////////////////

// Main controller.

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	port := "8080" // just set it to 8080. Ideally would get this from an ENV var ('LookupEnv' & setting it in the build vars)

	// neo4jConfig := parseConfiguration()
	neo4jConfig := setNeo4jConfigs()
	driver, err := neo4jConfig.newDriver()
	if err != nil {
		log.Fatal(err)
	}
	defer unsafeClose(driver) // 'defer' is basically 'finally', executes on function return

	serveMux := http.NewServeMux()
	// serveMux.HandleFunc("/", defaultHandler)
	// htmlData := http.FileServer(http.Dir("my/vue/frontend/files"));
	// serveMux.handle("/", htmlData);
	// serveMux.HandleFunc("/api/search", searchHandlerFunc(driver, neo4jConfig.Database))
	// serveMux.HandleFunc("/adjacentnodes/", movieHandlerFunc(driver, configuration.Database))
	serveMux.HandleFunc("/graph", graphHandler(driver, neo4jConfig.Database))

	fmt.Printf("Running on port %s, database is at %s...", port, neo4jConfig.Url)
	fmt.Printf("%+v\n", neo4jConfig)
	// fmt.Printf("Running on port %s, database is at %s\n", port, configuration.Url)

	// basic check from DB
	// fmt.Printf("Running 'get Actors' function...")
	// actorPrinter(driver, neo4jConfig.Database)

	// basic add to DB
	// fmt.Printf("Running 'hello world' function...")
	// helloWorld(driver, neo4jConfig.Database)

	fillNeo4jDB(driver, neo4jConfig.Database)

	// the handler below wants functions, so the handlers above should return functions ( functions that return functions! )
	panic(http.ListenAndServe(":"+port, httpgzip.NewHandler(serveMux)))

}