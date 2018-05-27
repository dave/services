package deployer

import (
	"github.com/dave/services"
	"github.com/dave/services/session"
)

type Deployer struct {
	session *session.Session
	send    func(services.Message)
	config  Config
	index   map[string]map[bool]string
	prelude map[bool]string
}

func New(session *session.Session, send func(services.Message), index map[string]map[bool]string, prelude map[bool]string, config Config) *Deployer {
	c := &Deployer{}
	c.session = session
	c.send = send
	c.config = config
	c.index = index
	c.prelude = prelude
	return c
}

type Config struct {
	ConcurrentStorageUploads int
	IndexBucket              string
	PkgBucket                string
	Protocol                 string
	PkgHost                  string
}
