package deployermsg

import (
	"encoding/gob"
)

func RegisterTypes() {
	gob.Register(Index{})
	gob.Register(Archive{})
}

// Index is an ordered list of dependencies.
type Index map[string]IndexItem

// IndexItem is an item in Index. Unchanged is true if the client already has cached as specified by
// Cache in the Update message. Unchanged dependencies are not sent as Archive messages.
type IndexItem struct {
	Hash      string // Hash of the js file
	Unchanged bool   // Unchanged is true if the package already exists in the client cache.
}

// Archive contains information about the JS and the stripped GopherJS archive file.
type Archive struct {
	Path     string
	Hash     string // Hash of the resultant js
	Standard bool
}
