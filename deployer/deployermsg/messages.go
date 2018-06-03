package deployermsg

import (
	"encoding/gob"
)

func RegisterTypes() {
	gob.Register(ArchiveIndex{})
	gob.Register(Archive{})
}

// ArchiveIndex is a list of dependencies.
type ArchiveIndex map[string]ArchiveIndexItem

// ArchiveIndexItem is an item in ArchiveIndex. Unchanged is true if the client already has cached as
// specified by Cache in the Update message. Unchanged dependencies are not sent as Archive messages.
type ArchiveIndexItem struct {
	Hash      string // Hash of the js file
	Unchanged bool   // Unchanged is true if the package already exists in the client cache.
}

// Archive contains information about the JS and the stripped GopherJS archive file.
type Archive struct {
	Path     string
	Hash     string // Hash of the resultant js
	Standard bool
}
