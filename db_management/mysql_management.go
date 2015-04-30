package db_management

import (
	log "github.com/cihub/seelog"
	"github.com/coopernurse/gorp"

	"strconv"
	"strings"
	"styl/db_object"
	"time"
)

const (
	PRODUCT_TABLE_NAME = "your_product_table"
	CORE_SEARCH_TABLE  = "your_search_table"
	ADMIN_STORE_API    = "your_admin_table"
)

type DbManagement struct {
	DbConnection *gorp.DbMap
}

var instance *DbManagement = nil

func GetInstance() *DbManagement {
	if instance == nil {
		instance = new(DbManagement)
	}
	return instance
}

func (dbManager *DbManagement) SetDbConnection(dbMap *gorp.DbMap) {
	dbManager.DbConnection = dbMap
}

func (dbManager *DbManagement) GetDbConnection() *gorp.DbMap {
	return dbManager.DbConnection
}

func (db *DbManagement) InsertProcat(data db_object.ProCat) {

}

func (db *DbManagement) InsertProduct(data *db_object.Product) {
	db.DbConnection.AddTableWithName(db_object.Product{}, PRODUCT_TABLE_NAME)
	err := db.DbConnection.Insert(data)

	if err != nil {
		log.Error("InsertProduct " + err.Error())
	}
}

func (db *DbManagement) UpdateProduct(data *db_object.Product) {
	db.DbConnection.AddTableWithName(db_object.Product{}, PRODUCT_TABLE_NAME).SetKeys(false, "ProductId")
	_, err := db.DbConnection.Update(data)
	if err != nil {
		log.Error(err)
	}
}

func (db *DbManagement) GetProductByUrl(url string) []db_object.Product {
	defer log.Flush()
	var product = []db_object.Product{}
	var query string = ""

	query += "select product_id,title,store_id, owner_id, price, currency, quantity, photo_id, public, actived,create_date,"
	query += "url, shop_category,update_date, scraped,description,claim_id,before_price,deleted"
	query += " from " + PRODUCT_TABLE_NAME + " where url = \"" + strings.Replace(url, "\"", "'", -1) + "\""

	_, err := db.DbConnection.Select(&product, query)
	if err != nil {
		log.Error("GetProductByUrl" + err.Error())
	}
	return product
}

func (db *DbManagement) UniqueProductByUrl(products []db_object.Product) *int {
	var productId *int
	for i, record := range products {
		if i != 0 {
			db.DeleteProduct(record.ProductId)
		} else {
			productId = record.ProductId
		}
	}
	return productId
}

func (db *DbManagement) DeleteProduct(id *int) {
	defer log.Flush()
	var query string = "DELETE FROM " + PRODUCT_TABLE_NAME + " WHERE product_id = " + string(*id)
	result, err := db.DbConnection.Exec(query)

	if err != nil {
		log.Error(err)
	}

	number, err := result.RowsAffected()
	log.Info("Delete product : " + string(*id) + "; number of effected rows : " + string(number))
}

func (db *DbManagement) InsertImageToStorage() {

}

func (db *DbManagement) RemoveDeletedProductInCoreSearch() {
	defer log.Flush()
	var query, sub_query, month, year, day string
	month = strconv.Itoa(int(time.Now().Month()))
	year = strconv.Itoa(time.Now().Year())
	day = strconv.Itoa(time.Now().Day())

	sub_query = "SELECT product_id FROM " + PRODUCT_TABLE_NAME
	sub_query += " WHERE deleted = 1 and YEAR(update_date) = " + year
	sub_query += " AND MONTH(update_date) = " + month
	sub_query += " AND DAY(update_date) = " + day
	query = "DELETE FROM " + CORE_SEARCH_TABLE + " WHERE type='store_product' AND id IN (" + sub_query + ")"

	result, err := db.DbConnection.Exec(query)

	if err != nil {
		log.Error("RemoveDeletedProductInCoreSearch" + err.Error())
		return
	}

	number, err := result.RowsAffected()
	log.Info("Number of row effected : " + string(number))
}

func (db *DbManagement) MarkDeletedProducts(storeId int) {
	defer log.Flush()
	var query, month, year, day string
	month = strconv.Itoa(int(time.Now().Month()))
	year = strconv.Itoa(time.Now().Year())
	day = strconv.Itoa(time.Now().Day())

	query = "UPDATE " + PRODUCT_TABLE_NAME + " SET deleted=1"
	query += " WHERE YEAR(update_date) <> " + year
	query += " AND MONTH(update_date) <> " + month
	query += " AND DAY(update_date) <> " + day
	query += " AND store_id = " + strconv.Itoa(int(storeId))

	result, err := db.DbConnection.Exec(query)
	if err != nil {
		log.Error("MarkDeletedProducts" + err.Error())
	}

	number, _ := result.RowsAffected()
	log.Info("Number of row effected : " + string(number))
}

func (db *DbManagement) GetListOfAPI() []db_object.AdminAPI {
	defer log.Flush()

	var api = []db_object.AdminAPI{}
	_, err := db.DbConnection.Select(&api, "select * from "+ADMIN_STORE_API)

	if err != nil {
		log.Error("GetListOfAPI" + err.Error())
	}

	return api
}
