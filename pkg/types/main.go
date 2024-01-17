package types

import (
	"cloud.google.com/go/civil"
)

type Todo struct {
	ID   string
	Date civil.Date
}
