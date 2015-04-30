package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
	"io"
	"net/http"
	"os"
	"strconv"
	db "styl/db_management"
	model "styl/db_object"
	settings "styl/settings"
	giant "styl/store_import"
)

const (
	SETTING_PATH                        = "settings.json"
	ZANOX_ROYAL_CATEGORIES_PATH         = "category-list/zanox-royal.txt"
	ZANOX_SCANDINAVIAN_CATEGORIES_PATH  = "category-list/zanox-scandinavian.txt"
	TRADEDOUBLER_ELLOS_CATEGORIES_PATH  = "category-list/tradedoubler-ellos.txt"
	TRADEDOUBLER_ROOM21_CATEGORIES_PATH = "category-list/tradedoubler-room21.txt"
	TRADETRACKER_CATEGORIES_PATH        = "category-list/tradetracker.txt"
	ZANOX_FILE_PATH                     = "zanox_product_api.xml"
	TRADETRACKER_FILE_PATH              = "tradetracker.xml"
	ZANOX_URL_DOWNLOAD                  = "http://productdata.zanox.com/exportservice/v1/rest/27462547C66644784.xml?ticket=A4A582A2F5366A0EDF8BBF73A447525D&productIndustryId=1&gZipCompress=null"
	TRADOUBLE_ELLOS_FILE_PATH           = "tradedoubler_ellos_product_api.xml"
	TRADOUBLE_ELLOS_URL_PATH            = "http://api.tradedoubler.com/1.0/productsUnlimited.xml;fid=9259?token=your-token"
	TRADOUBLE_ROOM21_FILE_PATH          = "tradedoubler_room21_product_api.xml"
	TRADOUBLE_ROOM21_URL_PATH           = "http://api.tradedoubler.com/1.0/productsUnlimited.xml;fid=17710?token=your-token"
	TRADETRACKER_URL_DOWNLOAD           = "http://pf.tradetracker.net/?aid=173883&encoding=utf-8&type=xml-v2&fid=680353&categoryType=2&additionalType=2"
)

var (
	listOfAPI    []model.AdminAPI   = nil
	globalConfig *settings.Settings = nil
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Error("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func init() {
	logger, err := log.LoggerFromConfigAsFile("log-conf.xml")

	if err != nil {
		log.Info(err)
	}

	log.ReplaceLogger(logger)

	globalConfig, err = settings.GetInstance("", SETTING_PATH)
	dbMange := db.GetInstance()
	dbMange.SetDbConnection(globalConfig.DbConnection)

	log.Info(globalConfig.GetConnectionString())

}

func main() {

	defer log.Flush()
	connectionStr := getConnectionString()

	conn, err := amqp.Dial(connectionStr)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		3,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(

		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			data, err := base64.StdEncoding.DecodeString(string(d.Body))
			if err != nil {
				log.Error("error:", err)
			}

			res := map[string]string{}
			json.Unmarshal([]byte(data), &res)
			log.Info("Received a message: %v", res)
			run(res)
			d.Ack(false)
		}
	}()

	log.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func getConnectionString() string {
	var env string = globalConfig.GetEnvValue()
	var connectionStr string = ""
	if env == "dev" {
		connectionStr = "amqp://guest:guest@localhost:45672/"
	} else {
		connectionStr = "amqp://guest:guest@localhost:5672/"
	}
	log.Info(connectionStr)
	return connectionStr
}

func getListCategoryConf(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func downloadFromUrl(url string, fileName string) {
	fmt.Println("Downloading", url, "to", fileName)
	log.Info("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

	fmt.Println(n, "bytes downloaded.")
}

func run(params map[string]string) {

	var data model.AdminAPI
	data.OwnerId, _ = strconv.Atoi(params["owner_id"])
	data.StoreId, _ = strconv.Atoi(params["store_id"])
	data.Title = params["title"]

	switch data.Title {
	case "test":
		log.Info("this is the testing")
	case "gallerimarkveien":

		//gallerimarkveien.Run(data.Url, data.StoreId, data.OwnerId)

	case "tradedoubler_room21":
		categories, _ := getListCategoryConf(TRADEDOUBLER_ROOM21_CATEGORIES_PATH)
		log.Info("Category list size : ", len(categories))
		//check file is exist
		//if not download it
		if !fileExists(TRADOUBLE_ROOM21_FILE_PATH) {
			downloadFromUrl(TRADOUBLE_ROOM21_URL_PATH, TRADOUBLE_ROOM21_FILE_PATH)
		}

		tradedoubleAPI := giant.NewTradedoublerAPI(TRADOUBLE_ROOM21_FILE_PATH, data.StoreId, data.OwnerId, categories)
		tradedoubleAPI.ImportFromAPI()

	case "tradedoubler_ellos":
		categories, _ := getListCategoryConf(TRADEDOUBLER_ELLOS_CATEGORIES_PATH)
		log.Info("Category list size : ", len(categories))
		//check file is exist
		//if not download it
		if !fileExists(TRADOUBLE_ELLOS_FILE_PATH) {
			downloadFromUrl(TRADOUBLE_ELLOS_URL_PATH, TRADOUBLE_ELLOS_FILE_PATH)
		}
		tradedoubleAPI := giant.NewTradedoublerAPI(TRADOUBLE_ELLOS_FILE_PATH, data.StoreId, data.OwnerId, categories)
		tradedoubleAPI.ImportFromAPI()

	case "zanox_royaldesign":
		categories, _ := getListCategoryConf(ZANOX_ROYAL_CATEGORIES_PATH)
		log.Info("Category list size : ", len(categories))

		//check file is exist
		//if not download it
		if !fileExists(ZANOX_FILE_PATH) {
			downloadFromUrl(ZANOX_URL_DOWNLOAD, ZANOX_FILE_PATH)
		}

		zanoxAPI := giant.NewZanoxAPI(ZANOX_FILE_PATH, data.StoreId, data.OwnerId, categories)
		zanoxAPI.ImportFromAPI()

	case "zanox_scandinavian":
		categories, _ := getListCategoryConf(ZANOX_SCANDINAVIAN_CATEGORIES_PATH)
		log.Info("Category list size : ", len(categories))
		if !fileExists(ZANOX_FILE_PATH) {
			downloadFromUrl(ZANOX_URL_DOWNLOAD, ZANOX_FILE_PATH)
		}
		zanoxAPI := giant.NewZanoxAPI(ZANOX_FILE_PATH, data.StoreId, data.OwnerId, categories)
		zanoxAPI.ImportFromAPI()

	case "tradetracker":
		categories, _ := getListCategoryConf(TRADETRACKER_CATEGORIES_PATH)
		log.Info("Category list size : ", len(categories))
		if !fileExists(TRADETRACKER_FILE_PATH) {
			downloadFromUrl(TRADETRACKER_URL_DOWNLOAD, TRADETRACKER_FILE_PATH)
		}
		tradetrackerAPI := giant.NewTradetrackerAPI(TRADETRACKER_FILE_PATH, data.StoreId, data.OwnerId, categories)
		tradetrackerAPI.ImportFromAPI()
	case "cahetu":
		//cahetu_product.Run(data.Url, data.StoreId, data.OwnerId)

	case "cahetu-category":
		//cahetu_category.Run(data.Url)

	}

}
