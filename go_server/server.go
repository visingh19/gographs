// Cannibalized from:
// https://golang.org/doc/articles/wiki/
// https://www.freecodecamp.org/news/how-i-set-up-a-real-world-project-with-go-and-vue/
// https://github.com/neo4j-examples/movies-golang-bolt

package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/daaku/go.httpgzip"
)

