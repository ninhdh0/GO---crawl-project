package settings

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strings"
)

type DbItem struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	UserName string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Protocol string `json:"protocol"`
	Arg      string `json:"arg"`
}

type DbConf struct {
	Env  string   `json:"ENV"`
	Conf []DbItem `json:"CONF"`
}

type Settings struct {
	DBConf           *DbConf
	ConnectionString string
	DbConnection     *gorp.DbMap
	ConfPath         string
}

var setting *Settings = nil

func GetInstance(env string, filepath string) (*Settings, error) {
	var err error = nil
	if setting == nil {
		setting = new(Settings)
		err := setting.InitSettings(env, filepath)
		if err != nil {
			return nil, err
		}
	}

	return setting, err
}

func (setting *Settings) GetDBConnection() *gorp.DbMap {
	return setting.DbConnection
}

func (setting *Settings) GetConnectionString() string {
	return setting.ConnectionString
}

func (setting *Settings) GetConfFilePath() string {
	return setting.ConfPath
}

func (setting *Settings) GetEnvValue() string {
	return setting.DBConf.Env
}

func (setting *Settings) InitSettings(env string, filepath string) error {
	file, err := os.Open(filepath)
	var conf *DbConf = new(DbConf)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&conf)
	if err != nil {
		return err
	}

	var params DbItem
	if env == "" {
		env = conf.Env
	}

	for _, block := range conf.Conf {
		if block.Name == env {
			params = block
			break
		}
	}

	host := params.Host
	port := params.Port
	user := params.UserName
	pass := params.Password
	dbname := params.Database
	protocol := params.Protocol
	dbargs := params.Arg

	if strings.Trim(dbargs, " ") != "" {
		dbargs = "?" + dbargs
	} else {
		dbargs = ""
	}

	connectionStr := fmt.Sprintf("%s:%s@%s([%s]:%s)/%s%s", user, pass, protocol, host, port, dbname, dbargs)

	db, err := sql.Open("mysql", connectionStr)
	if err != nil {
		return err
	}

	Dbm := &gorp.DbMap{
		Db:      db,
		Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"},
	}

	setting.ConnectionString = connectionStr
	setting.DBConf = conf
	setting.ConfPath = filepath
	setting.DbConnection = Dbm

	return nil
}
