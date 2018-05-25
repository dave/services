package gcsdatabase

import (
	"cloud.google.com/go/datastore"
)

func New(client *datastore.Client) *Database {
	return &Database{
		Client: client,
	}
}

type Database struct {
	*datastore.Client
}
