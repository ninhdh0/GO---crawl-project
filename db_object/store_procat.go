package db_object

import (
	"fmt"
)

type ProCat struct {
	ProcatId   *int `db:"procat_id"`
	ProductId  *int `db:"product_id"`
	CategoryId *int `db:"category_id"`
}

func NewProcat() *ProCat {
	return new(ProCat)
}

func (procat *ProCat) ToString() string {
	return fmt.Sprintln("%i %i", *procat.CategoryId, *procat.ProductId)
}
