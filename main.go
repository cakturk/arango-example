package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type EdgeObject struct {
	From string `json:"_from"`
	To   string `json:"_to"`
}

type Production struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
}

type CastType int

const (
	Actor CastType = iota + 1
	Director
	Producer
)

type Cast struct {
	Name string   `json:"name"`
	Type CastType `json:"type"`
}

func main() {
	var (
		name = flag.String("dbname", "example_db", "name of the db to be created")
		user = flag.String("dbuser", "root", "database user")
		pass = flag.String("dbpass", "", "database password")
		host = flag.String("dbhost", "http://localhost:8529", "db host name")
	)
	flag.Parse()
	ctx := context.Background()
	// Create an HTTP connection to the database
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{*host},
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})
	if err != nil {
		log.Fatalf("failed to create HTTP connection: %v", err)
	}
	// Create a client
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(*user, *pass),
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
		return
	}

	// Create database
	db, err := c.CreateDatabase(ctx, *name, nil)
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	edgeDef := driver.EdgeDefinition{
		Collection: "moviegraph",
		To:         []string{"cast"},
		From:       []string{"movies"},
	}
	options := driver.CreateGraphOptions{
		EdgeDefinitions: []driver.EdgeDefinition{edgeDef},
	}
	graph, err := db.CreateGraphV2(ctx, "myGraph", &options)
	if err != nil {
		log.Fatalf("Failed to create graph: %v", err)
	}
	_ = graph

	movieCol, err := graph.VertexCollection(ctx, "movies")
	if err != nil {
		log.Fatalf("Failed to get vertex collection: %v", err)
	}
	// Create document
	movie := Production{
		Title: "Apocalypse Now",
		Year:  1979,
	}
	metaMov, err := movieCol.CreateDocument(ctx, &movie)
	if err != nil {
		log.Fatalf("failed to create document: %v", err)
	}
	fmt.Printf("Created document in collection '%s' in database '%s'\n", movieCol.Name(), db.Name())

	// Read the document back
	var result Production
	if _, err := movieCol.ReadDocument(ctx, metaMov.Key, &result); err != nil {
		log.Fatalf("failed to read document: %v", err)
	}
	fmt.Printf("Read book '%+v'\n", result)

	castCol, err := graph.VertexCollection(ctx, "cast")
	if err != nil {
		log.Fatalf("Failed to get vertex collection: %v", err)
	}
	actor := Cast{
		Name: "Marlon Brando",
		Type: Actor,
	}
	metaCast, err := castCol.CreateDocument(ctx, &actor)
	if err != nil {
		log.Fatalf("failed to create document: %v", err)
	}
	fmt.Printf("Created document in collection '%s' in database '%s'\n", castCol.Name(), db.Name())

	// add edge
	edgeCol, _, err := graph.EdgeCollection(ctx, "moviegraph")
	if err != nil {
		log.Fatalf("Failed to select edge collection: %v", err)
	}

	fmt.Printf("from: %+v to: %+v\n", metaMov, metaCast)
	edge := EdgeObject{From: string(metaMov.ID), To: string(metaCast.ID)}
	_, err = edgeCol.CreateDocument(ctx, edge)
	if err != nil {
		log.Fatalf("failed to create edge document: %v", err)
	}
}
