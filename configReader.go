package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var dirPath string

type DB struct {
	Type     string `json:"Type"`
	User     string `json:"User"`
	Password string `json:"Password"`
	Port     string `json:"Port"`
	Host     string `json:"Host"`
	Database string `json:"Database"`
}

type Host struct {
	Domain      string `json:"Domain"`
	Port        string `json:"Port"`
	Version     string `json:"Version"`
	Hostpath    string `json:"Hostpath"`
	Logfilepath string `json:"Logfilepath"`
}

type Security struct {
	Ip          string `json:"Ip"`
	Port        string `json:"Port"`
	AccessToken string `json:"AccessToken"`
}

type Configuration struct {
	DB       DB       `json:"DB"`
	Host     Host     `json:"Host"`
	Security Security `json:"Security"`
}

func GetDirPath() string {
	envPath := os.Getenv("GO_CONFIG_DIR")
	if envPath == "" {
		envPath = "./"
	}
	fmt.Println(envPath)
	return envPath
}

func LoadDefaultConfig() Configuration {
	confPath := filepath.Join("E:\\DuoProject\\Service\\GO-Projects\\src\\DVP-CampaignNumberUploader", "conf.json")
	fmt.Println("GetDefaultConfig config path: ", confPath)
	content, operr := ioutil.ReadFile(confPath)
	if operr != nil {
		fmt.Println(operr)
	}

	defconfiguration := Configuration{}
	deferr := json.Unmarshal(content, &defconfiguration)
	if deferr != nil {
		log.Panic(deferr)
	}
	return defconfiguration
}

func LoadConfiguration() Configuration {
	dirPath = GetDirPath()
	confPath := filepath.Join(dirPath, "custom-environment-variables.json")
	fmt.Println("InitiateRedis config path: ", confPath)

	content, operr := ioutil.ReadFile(confPath)
	if operr != nil {
		fmt.Println(operr)
	}
	envConfig := Configuration{}
	envconfiguration := Configuration{}
	enverr := json.Unmarshal(content, &envconfiguration)
	if enverr != nil {
		fmt.Println("Fail to Load Environment Settings and Load Default Settings :", enverr)
		envConfig = LoadDefaultConfig()
	} else {

		envConfig.DB.Database = os.Getenv(envconfiguration.DB.Database)
		envConfig.DB.Database = os.Getenv(envconfiguration.DB.Database)
		envConfig.DB.Host = os.Getenv(envconfiguration.DB.Host)
		envConfig.DB.Password = os.Getenv(envconfiguration.DB.Password)
		envConfig.DB.Port = os.Getenv(envconfiguration.DB.Port)
		envConfig.DB.Type = os.Getenv(envconfiguration.DB.Type)
		envConfig.DB.User = os.Getenv(envconfiguration.DB.User)

		envConfig.Host.Domain = os.Getenv(envconfiguration.Host.Domain)
		envConfig.Host.Hostpath = os.Getenv(envconfiguration.Host.Hostpath)
		envConfig.Host.Logfilepath = os.Getenv(envconfiguration.Host.Logfilepath)
		envConfig.Host.Port = os.Getenv(envconfiguration.Host.Port)
		envConfig.Host.Version = os.Getenv(envconfiguration.Host.Version)

		envConfig.Security.AccessToken = os.Getenv(envconfiguration.Security.AccessToken)
		envConfig.Security.Ip = os.Getenv(envconfiguration.Security.Ip)
		envConfig.Security.Port = os.Getenv(envconfiguration.Security.Port)
	}
	fmt.Println("Configurations :", envConfig)
	return envConfig
}
