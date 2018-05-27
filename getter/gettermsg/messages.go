package gettermsg

import "encoding/gob"

func RegisterTypes() {
	gob.Register(Downloading{})
}

type Downloading struct {
	Starting bool
	Message  string
	Done     bool
}
