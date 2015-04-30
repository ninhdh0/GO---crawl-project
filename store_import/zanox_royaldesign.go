package store_import

import (
	"encoding/xml"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"styl/common"
	db "styl/db_management"
	"styl/db_object"
	"time"
)

type XmlZanox struct {
	ProductItems []XmlZanoxProduct `xml:"product"`
}

type XmlZanoxProduct struct {
	Title     string  `xml:"name"`
	Price     float64 `xml:"price"`
	Thumbnail string  `xml:"largeImage"`
	Url       string  `xml:"deepLink"`
	Currency  string  `xml:"currencyCode"`
	Category  string  `xml:"merchantCategoryPath"`
}

type ZanoxAPI struct {
	FilePath     string
	StoreId      int
	OwnerId      int
	CategoryList []string
}

var zanoxAPI *ZanoxAPI = nil

func NewZanoxAPI(filePath string, storeId int, ownerId int, categoryList []string) *ZanoxAPI {
	if zanoxAPI == nil {
		zanoxAPI = new(ZanoxAPI)
	}
	zanoxAPI.FilePath = filePath
	zanoxAPI.StoreId = storeId
	zanoxAPI.OwnerId = ownerId
	zanoxAPI.CategoryList = categoryList

	return zanoxAPI
}

func (api *ZanoxAPI) ImportFromAPI() {
	defer log.Flush()

	xmlFile, err := os.Open(api.FilePath)
	if err != nil {
		log.Error("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	var data XmlZanox
	body, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Error(err)
	}
	err = xml.Unmarshal(body, &data)
	if err != nil {
		log.Error(err)
	}

	log.Info("Number of Product item: ", len(data.ProductItems))
	for _, product := range data.ProductItems {
		api.InsertRow(product)
	}

	db.GetInstance().MarkDeletedProducts(api.StoreId)
	db.GetInstance().RemoveDeletedProductInCoreSearch()

	log.Info("Total Products : ", len(data.ProductItems))

}

func (api *ZanoxAPI) InsertRow(xmlData XmlZanoxProduct) {
	product := db_object.NewProduct()
	product.ProductId = nil
	product.StoreId = &api.StoreId
	product.OwnerId = &api.OwnerId
	product.ShopCategory = xmlData.Category
	product.Title = xmlData.Title
	product.Price = xmlData.Price
	product.Currency = strings.Trim(xmlData.Currency, " ")
	product.UpdateDate = time.Now().Format(TIME_FORMAT)
	product.CreationDate = time.Now().Format(TIME_FORMAT)
	product.Url = xmlData.Url
	product.Scraped = 1
	product.Description = ""
	product.BeforePrice = xmlData.Price
	product.Public = 1
	product.Actived = 1
	product.Deleted = 0

	isCatExist := common.IsExistInArray(product.ShopCategory, api.CategoryList)
	if isCatExist == false {
		return
	}

	var results = db.GetInstance().GetProductByUrl(product.Url)
	if len(results) > 1 {
		// update information
		product.ProductId = db.GetInstance().UniqueProductByUrl(results)
		db.GetInstance().UpdateProduct(product)
		log.Info("update product " + strconv.Itoa(*product.ProductId))
	} else if len(results) == 1 {
		// update information
		product.ProductId = results[0].ProductId
		db.GetInstance().UpdateProduct(product)
		log.Info("uupdate product " + strconv.Itoa(*product.ProductId))
	} else {
		// insert image
		// insert to authorize table
		// insert to search table

		// insert new product
		db.GetInstance().InsertProduct(product)
		log.Info("insert new")
	}
}
