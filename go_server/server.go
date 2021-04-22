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


// some helpers ripped/edited from https://github.com/neo4j-examples/movies-golang-bolt demo code.
// now connects.
func graphHandlerFunc(driver neo4j.Driver, database string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		session := driver.NewSession(neo4j.SessionConfig{
			AccessMode:   neo4j.AccessModeRead,
			DatabaseName: database,
		})
		defer unsafeClose(session)

		// set query.
		limit := parseLimit(req)
		query := `MATCH (a:Person)-[rel]->(b:Person)
				  RETURN a.name as aName, b.name as bName, type(rel) as relName, id(a) as fromNode, id(b) as toNode
				  LIMIT $limit`

		// run query as read transaction.
		d3Resp, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			records, err := tx.Run(query, map[string]interface{}{"limit": limit})
			if err != nil {
				return nil, err
			}
			log.Println("read transaction in graphHandler ran")
			result := D3Data{}
			idMap := make([]int, 50) // these are 0'd, so we set them to -1, assuming 50 to start. At worst O(n) mem, n = max # of nodes
			for idx, _ := range idMap {
				idMap[idx] = -1
			}

			for records.Next() {
				record := records.Record()

				aName, _ := record.Get("aName") // string
				bName, _ := record.Get("bName") // string
				relName, _ := record.Get("relName") // string
				fromNode64, _ := record.Get("fromNode") // identity as per GraphDB of a, int64
				toNode64, _ := record.Get("toNode") // identity as per GraphDB of b, int64
				// convert from int64 to int, we don't need more than 2^32 bits of space. using type assertion.
				fromNode := int(fromNode64.(int64)) 
				toNode := int(toNode64.(int64))


				// if a 'fromNode' or a 'toNode' is out of the range of the index ( e.g. the 51st node has id 50 )
				if ( fromNode >= len(idMap) ) {
					newSlice := make([]int, fromNode - len(idMap) + 1)
					for idx, _ := range newSlice { newSlice[idx] = -1 }
					idMap = append(idMap, newSlice...) // spread operator on new slice, create enough space for new nodes
				}

				if ( toNode >= len(idMap) ) {
					newSlice := make([]int, toNode - len(idMap) + 1)
					for idx, _ := range newSlice { newSlice[idx] = -1 }
					idMap = append(idMap, newSlice...) // spread operator on new slice, create enough space for new nodes
				}

				// then we can set the mappings to connect links together. 
				// we are indirectly parsing a set of nodes from a set of link relations.
				// e.g. Tom of Graph id 5 (0-index'd) results in idMap[5] -> 0'th node in our slice
				// and then we can always see going forward that Tom is either placed or not placed in our node
				// and what his id is in the node list!
				if ( idMap[fromNode] == -1 ) {
					idMap[fromNode] = len(result.Nodes) 
					result.Nodes = append(result.Nodes, Node{Title: aName.(string), Label: "Person"})
				}
				if ( idMap[toNode] == -1 ) {
					idMap[toNode] = len(result.Nodes) 
					result.Nodes = append(result.Nodes, Node{Title: bName.(string), Label: "Person"})
				}

				// once you've added in any nodes for people that did not exist, we now add the relation/link.
				result.Links = append(result.Links, Link{Source: idMap[fromNode], Target: idMap[toNode], Relationship: relName.(string)})
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


func resetGraphHandlerFunc(driver neo4j.Driver, database string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		_, err := emptyNeo4jDB(driver, database)
		if err != nil {
			log.Println("error emptying graph:", err)
			return
		} else {
			_, err = fillNeo4jDB(driver, database, 50)
			if err != nil {
				log.Println("error filling graph:", err)
				return
			} else {
				_, err = fmt.Fprintf(w, "Success")
			}
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
func emptyNeo4jDB(driver neo4j.Driver, database string) (string, error) {
	log.Println("Entered emptyNeo4jDB.")
	defer func() {
		log.Println("Exited emptyNeo4jDB.")
	}()

	session := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: database,
	})
	defer unsafeClose(session)

	var query strings.Builder
	subQuery := "MATCH (n) DETACH DELETE n"
	query.WriteString(subQuery)

	// Query String exists.

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			query.String(), map[string]interface{}{})
		log.Println("Write transaction finshed in emptyNeo4jDB.")
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

// func to add random data into the DB - mocking 'getting aws resource'
// adds up to 50 random people and numRelations random relationships between them based on a set of relationship names.
// Personal experience: ~numRelations/10 = seconds to completion, so if you choose >300, your network call will timeout ( but the process will still run)
// Why: Many browsers will automatically timeout a request after 30 seconds.
func fillNeo4jDB(driver neo4j.Driver, database string, numRelations int) (string, error) {
	log.Println("Entered fillNeo4jDB.")
	defer func() {
		log.Println("Exited fillNeo4jDB.")
	}()

	session := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: database,
	})
	defer unsafeClose(session)


	var fNames = []string{"Alice", "Bob", "Carol", "Danny", "Eve", "Finn", "Gina", "Hank", "Irene", "John" } // 10
	var lNames = []string{"Andrews", "Brick", "Carter", "Daves", "Erickson"} // 5 -> max unique AWS 'persons' = 50 ( 10 x 5 )
	var relations = []string{"MANAGES", "REPORTS_TO", "PAYS", "MONITORS", "AUTHORIZES"} // 5 -> 5 random relationships ( and yes, a person can manage themselves haha. )

	// Now we create numRelations relationships - somewhat dense --> (person)-(relation)->(person) 200 times.
	// At worst ( pigeonhole principle ) each of the 50 possible nodes has 4 relations, fairly connected!

	// Because there are few nodes ( 50 or less always ), deletion should be easy.
	// Sends 3*numRelations lines to the transaction.

	var query strings.Builder
	rand.Seed(time.Now().Unix())
	for i := 0; i < numRelations; i++ {
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
		log.Println("Write transaction finshed in fillNeo4jDB.")
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

//////////////////////////////////////////
////// HELPFUL TEST CODE /////////////////
//////////////////////////////////////////

// func to just print every 'Person' ( to console, not an API! )
// useful for the neo4j demo movies data set, simple test!
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

	// Serve webpage

	htmlData := http.FileServer(http.Dir("./frontend/dist"));
	serveMux.Handle("/", htmlData);

	// APIs

	serveMux.HandleFunc("/api/search", searchHandlerFunc(driver, neo4jConfig.Database))
	// serveMux.HandleFunc("/adjacentnodes/", movieHandlerFunc(driver, configuration.Database))

	serveMux.HandleFunc("/api/resetgraph", resetGraphHandlerFunc(driver, neo4jConfig.Database)) // empties graph & fills it with a new fake data set
	serveMux.HandleFunc("/api/graph", graphHandlerFunc(driver, neo4jConfig.Database)) // simply returns a set of nodes and links.

	// Self loggers & testing functions ( logs to console, not site / api )

	fmt.Printf("Running on port %s, database is at %s...", port, neo4jConfig.Url)
	fmt.Printf("%+v\n", neo4jConfig)

	// basic check from DB
	// fmt.Printf("Running 'get Actors' function...")
	// actorPrinter(driver, neo4jConfig.Database)

	// basic add to DB
	// fmt.Printf("Running 'hello world' function...")
	// helloWorld(driver, neo4jConfig.Database)

	// fills db with n relations
	// fillNeo4jDB(driver, neo4jConfig.Database, 20)

	// empties db of all nodes & relations
	// emptyNeo4jDB(driver, neo4jConfig.Database)

	// listen & serve call ( serves website )

	// the handler below wants functions, so the handlers above should return functions ( functions that return functions! )
	panic(http.ListenAndServe(":"+port, httpgzip.NewHandler(serveMux)))

}