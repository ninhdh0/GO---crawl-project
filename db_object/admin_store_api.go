package db_object

import (
	"fmt"
)

type AdminAPI struct {
	Id      int    `db:"id"`
	Title   string `db:"title"`
	Url     string `db:"url"`
	StoreId int    `db:"store_id"`
	OwnerId int    `db:"owner_id"`
}

func New() *AdminAPI {
	return new(AdminAPI)
}

func (api *AdminAPI) ToString() string {
	return fmt.Sprintln("%s %s", api.Title, api.Url)
}
