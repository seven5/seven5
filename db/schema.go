package db

import (
	"github.com/eaigner/hood"
)

type Article struct {
	Id      hood.Id `sql:"pk" validate:"presence"`
	Author  string  `sql: "size(255)" validate:"presence"`
	Content string  `validate:"presence"`
	Created hood.Created
	Updated hood.Updated
}
