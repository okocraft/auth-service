package database

import (
	"github.com/okocraft/auth-service/internal/repositories/queries"
)

type Connection interface {
	Queries() *queries.Queries
}

type connection struct {
	conn queries.DBTX
}

func newConnection(conn queries.DBTX) Connection {
	return &connection{conn: conn}
}

func (c connection) Queries() *queries.Queries {
	return queries.New(c.conn)
}
