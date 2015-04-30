package db_object

import (
	"fmt"
)

type Product struct {
	ProductId    *int    `db:"product_id"`
	Title        string  `db:"title"`
	StoreId      *int    `db:"store_id"`
	OwnerId      *int    `db:"owner_id"`
	Price        float64 `db:"price"`
	Currency     string  `db:"currency"`
	Quantity     *int    `db:"quantity"`
	PhotoId      *int    `db:"photo_id"`
	Public       int     `db:"public"`
	Actived      int     `db:"actived"`
	CreationDate string  `db:"create_date"`
	Url          string  `db:"url"`
	ShopCategory string  `db:"shop_category"`
	UpdateDate   string  `db:"update_date"`
	Scraped      int     `db:"scraped"`
	Description  string  `db:"description"`
	ClaimID      int     `db:"claim_id"`
	BeforePrice  float64 `db:"before_price"`
	Deleted      int     `db:"deleted"`
}

func NewProduct() *Product {
	return new(Product)
}

func (product *Product) String() string {
	return fmt.Sprintf("%s %s", product.Title, product.Url)
}
