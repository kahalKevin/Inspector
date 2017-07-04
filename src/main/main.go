package main

import (
	"runtime"
	"log"
	"flag"

	"github.com/spf13/viper"

	"inspector"
)

var confLocation *string
var conf, fileDir, port, smtp, sender, publicdir string
var smtpPort int
var recipient []string

func init() {
	confLocation = flag.String("config", "../conf", "string | location of config file")
    runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()
	
	conf := viper.New()
	conf.SetConfigName("inspector_config")
	conf.SetConfigType("toml")
	conf.AddConfigPath(*confLocation)

	err := conf.ReadInConfig()
	if err != nil {
		log.Fatal("Fatal error config file: %s \n", err)
	}

    publicdir = conf.GetString("publicdirectory")
	fileDir = conf.GetString("filedirectory")
	port = conf.GetString("inspectorport")
	smtp = conf.GetString("smtphost")
	smtpPort = conf.GetInt("smtpport")
	sender = conf.GetString("emailfrom")
	recipient = conf.GetStringSlice("emailrecipient")
	subject := conf.GetStringMapString("subject")
	body := conf.GetStringMapString("body")

	inspector.Start(fileDir, port, smtp, sender, publicdir, smtpPort, recipient, subject, body)
}
