package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"tinder-for-clubs-backend/common"
)

//DBCredential struct
type DBCredential struct {
	DBAddress string `yaml:"db-address"`
	DBUser    string `yaml:"db-user"`
	DBPass    string `yaml:"db-pass"`
	DBPort    string `yaml:"db-port"`
	DBName    string `yaml:"db-name"`
}

type General struct {
	PictureStoragePath string `yaml:"static-storage-path"`
}

//GlobalConfiguration struct
type GlobalConfiguration struct {
	DBCredential DBCredential `yaml:"db-config"`
	General      General      `yaml:"general"`
}

//GetConnectionString Build a database connection
func (c *DBCredential) GetConnectionString() string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?multiStatements=TRUE&parseTime=true&charset=utf8mb4,utf8", c.DBUser, c.DBPass, c.DBAddress, c.DBPort, c.DBName)
}

//ConnectionCredentialLogString Get database connection information
func (c *DBCredential) ConnectionCredentialLogString() string {
	return fmt.Sprintf("Username: %v  Address: %v  DBName: %v\n", c.DBUser, c.DBAddress, c.DBName)
}

//Package profile information
func readConfig(path string) (GlobalConfiguration, error) {
	log.Println("Starting to load configuration file ...")
	dat, err := ioutil.ReadFile(path)
	common.ErrFatalLog(err)
	t := GlobalConfiguration{}
	err = yaml.Unmarshal(dat, &t)

	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("file %s does not exist", path)
		} else {
			log.Fatalf("unknown cacheerror: %v", err)
		}
	}
	return t, nil

}

//ReadConfig Get configuration file information from yml
func ReadConfig() GlobalConfiguration {
	configFilePath := flag.String("config", "./config.yml", "The path to the configuration file")
	flag.Parse()
	log.Printf("Using configuration file %s", *configFilePath)
	globalConfig, err := readConfig(*configFilePath)
	common.ErrFatalLog(err)
	return globalConfig

}
