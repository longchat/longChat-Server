package graph

import (
	_ "github.com/go-cq/cq"
	"github.com/jmoiron/sqlx"
)

var graphDb *sqlx.DB

func NewDb(neo4jURL string) error {
	var err error
	graphDb, err = sqlx.Connect("neo4j-cypher", neo4jURL)
	if err != nil {
		return err
	}
	return nil
}

func FinDb() {
	graphDb.Close()
}
