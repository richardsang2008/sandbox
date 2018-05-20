package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"

	"github.com/gin-gonic/gin"
)


var log *logrus.Logger

func NewLogger(filename string, logLevel string) *logrus.Logger {
	if log != nil {
		return log
	}

	path := filename
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(86400)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(604800)*time.Second),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	logrus.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.InfoLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.DebugLevel: writer,
			logrus.WarnLevel:  writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		},
		&logrus.JSONFormatter{},
	))

	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  path,
		logrus.ErrorLevel: path,
		logrus.DebugLevel: path,
		logrus.WarnLevel:  path,
		logrus.FatalLevel: path,
		logrus.PanicLevel: path,
	}
	log = logrus.New()
	switch logLevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.PanicLevel)
	}
	log.Hooks.Add(lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{},
	))

	return log
}

func GetUsers(c *gin.Context) {
	log.Debug("I am here")
	c.JSON(200, "hello world")
}

type User struct {
	gorm.Model

	ID        uint `gorm:"primary_key`
	Uname     string `sql:"type:VARCHAR(255)"`
	CreatedAt time.Time
}

func main() {
	env := ""
	viper.SetConfigName("appconfig")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	} else {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println("Config file changed:", e.Name)
		})
		testingEnvEnable := viper.GetString("test.enable")
		devEnvEnable := viper.GetString("dev.enable")
		prodEnvEnable := viper.GetString("prod.enable")
		if testingEnvEnable == "true" {
			env = "test"
		}
		if devEnvEnable == "true" {
			env = "dev"
		}
		if prodEnvEnable == "true" {
			env = "prod"
		}
		envLogLevel := fmt.Sprintf("%s.log.level", env)
		envLogFile := fmt.Sprintf("%s.log.file", env)
		//envServerHost:=fmt.Sprintf("%s.server.host",env)
		envServerPort:=fmt.Sprintf("%s.server.port",env)
		logLevel := viper.GetString(envLogLevel)
		logFile := viper.GetString(envLogFile)
		//serverHost:=viper.GetString(envServerHost)
		serverPort:=viper.GetString(envServerPort)
		log := NewLogger(logFile, logLevel)

		r := gin.Default()
		v1 := r.Group("api/v1")
		//log.Debug("hello world")
		{
			v1.GET("/users", GetUsers)
		}
		db, err := gorm.Open("mysql", "mapuser:password@tcp(127.0.0.1:3306)/lisa_data2?charset=utf8&parseTime=True&loc=Local")
		db.CreateTable(&User{})
		defer db.Close()
		if err != nil {
			log.Panic("DB is not open ")
		}

		//user := User{Uname: "test1"}
		//fmt.Println(db.NewRecord(user)) // => returns `true` as primary key is blank
		//ret := db.Create(&user)

		//log.Warn("You should probably take a look at this.")
		//log.Error("Somthing failed but I am not quitting")
		/*log.WithFields(logrus.Fields{
			"animal": "walrus",
			"size":   10,
		}).Error("A group of walrus emerges from the ocean")*/
		ports := fmt.Sprintf(":%s",serverPort)
		log.Info("Server is running from port ",serverPort)
		r.Run(ports)
	}

}
