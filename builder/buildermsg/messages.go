package buildermsg

import (
	"encoding/gob"
)

func RegisterTypes() {
	gob.Register(Building{})
}

type Building struct {
	Starting bool
	Message  string
	Done     bool
}
