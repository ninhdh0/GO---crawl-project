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

type XmlTradeTracker struct {
	ProductItems []XmlTradeTrackerProduct `xml:"product"`
}

type XmlTradeTrackerProduct struct {
	Title       string                  `xml:"name"`
	Price       XmlTradeTrackerPriceTag `xml:"price"`
	Thumbnail   string                  `xml:"images>image"`
	Url         string                  `xml:"URL"`
	Category    string                  `xml:"categories>category"`
	Description string                  `xml:"description"`
}

type XmlTradeTrackerPriceTag struct {
	Value    float64 `xml:",chardata"`
	Currency string  `xml:"currency,attr"`
}

type TradeTrackerAPI struct {
	Filepath     string
	StoreId      int
	OwnerId      int
	CategoryList []string
}

var tradetrackerAPI *TradeTrackerAPI = nil

func NewTradetrackerAPI(filePath string, storeId int, ownerId int, categoryList []string) *TradeTrackerAPI {
	if tradetrackerAPI == nil {
		tradetrackerAPI = new(TradeTrackerAPI)
	}
	tradetrackerAPI.Filepath = filePath
	tradetrackerAPI.StoreId = storeId
	tradetrackerAPI.OwnerId = ownerId
	tradetrackerAPI.CategoryList = categoryList
	return tradetrackerAPI
}

func (api *TradeTrackerAPI) ImportFromAPI() {
	xmlFile, err := os.Open(api.Filepath)
	if err != nil {
		log.Error("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	var data XmlTradeTracker
	body, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Error(err)
	}
	err = xml.Unmarshal(body, &data)
	if err != nil {
		log.Error("Error Unmarshal data:", err)
	}

	log.Info("Number of Product item: ", len(data.ProductItems))

	for _, product := range data.ProductItems {
		api.InsertRow(product)
	}
	db.GetInstance().MarkDeletedProducts(api.StoreId)
	db.GetInstance().RemoveDeletedProductInCoreSearch()

	log.Info("Total Products : ", len(data.ProductItems))
}

func (api *TradeTrackerAPI) InsertRow(xmlData XmlTradeTrackerProduct) {

	product := db_object.NewProduct()
	product.StoreId = &api.StoreId
	product.OwnerId = &api.OwnerId
	product.ShopCategory = xmlData.Category
	product.Title = xmlData.Title
	product.Price = xmlData.Price.Value
	product.Currency = strings.Trim(xmlData.Price.Currency, " ")
	product.CreationDate = time.Now().Format(TIME_FORMAT)
	product.UpdateDate = time.Now().Format(TIME_FORMAT)
	product.Url = xmlData.Url
	product.Scraped = 1
	product.Description = xmlData.Description
	product.BeforePrice = xmlData.Price.Value
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
