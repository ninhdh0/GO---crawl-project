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

type XmlTradedoubler struct {
	Products []XmlTradedoublerProduct `xml:"products>product"`
}

type XmlTradedoublerProduct struct {
	Title       string         `xml:"name"`
	Description string         `xml:"description"`
	Thumbnail   string         `xml:"productImage"`
	Offer       XmlOfferTag    `xml:"offers>offer"`
	Category    XmlCategoryTag `xml:"categories>category"`
}

type XmlOfferTag struct {
	Url   string      `xml:"productUrl"`
	Price XmlPriceTag `xml:"priceHistory>price"`
}

type XmlPriceTag struct {
	Value    string `xml:",chardata"`
	Currency string `xml:"currency,attr"`
}

type XmlCategoryTag struct {
	Name string `xml:"name,attr"`
}

type NullId struct {
	Id *int `db:"store_category"`
}

type TradeDoublerAPI struct {
	Filepath     string
	StoreId      int
	OwnerId      int
	CategoryList []string
}

var tradedoublerAPI *TradeDoublerAPI = nil

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
)

func NewTradedoublerAPI(filePath string, storeId int, ownerId int, categoryList []string) *TradeDoublerAPI {
	if tradedoublerAPI == nil {
		tradedoublerAPI = new(TradeDoublerAPI)
	}
	tradedoublerAPI.Filepath = filePath
	tradedoublerAPI.StoreId = storeId
	tradedoublerAPI.OwnerId = ownerId
	tradedoublerAPI.CategoryList = categoryList
	return tradedoublerAPI
}

func (api *TradeDoublerAPI) ImportFromAPI() {
	xmlFile, err := os.Open(api.Filepath)
	if err != nil {
		log.Error("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	var data XmlTradedoubler
	body, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Error(err)
	}
	err = xml.Unmarshal(body, &data)
	if err != nil {
		log.Error("Error Unmarshal data:", err)
	}

	log.Info("Number of Product item: ", len(data.Products))
	for _, product := range data.Products {
		api.InsertRow(product)
	}
	db.GetInstance().MarkDeletedProducts(api.StoreId)
	db.GetInstance().RemoveDeletedProductInCoreSearch()

	log.Info("Total Products : ", len(data.Products))
}

func (api *TradeDoublerAPI) InsertRow(xmlData XmlTradedoublerProduct) {

	product := db_object.NewProduct()
	product.StoreId = &api.StoreId
	product.OwnerId = &api.OwnerId
	product.ShopCategory = xmlData.Category.Name
	product.Title = xmlData.Title
	product.Price, _ = strconv.ParseFloat(xmlData.Offer.Price.Value, 64)
	product.Currency = strings.Trim(xmlData.Offer.Price.Currency, " ")
	product.CreationDate = time.Now().Format(TIME_FORMAT)
	product.UpdateDate = time.Now().Format(TIME_FORMAT)
	product.Url = xmlData.Offer.Url
	product.Scraped = 1
	product.Description = xmlData.Description
	product.BeforePrice = product.Price
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
