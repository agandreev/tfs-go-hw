package main

import (
	"context"
	"fmt"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/controller"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/handlers"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/repository/orders"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/repository/users"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/service"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/service/msgwriters"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	configPath            = "config.env"
	srvPort               = "SRV_PORT"
	reconnectionsQuantity = "RECONNECTION_QUANTITY"
	ttlHours              = "TTL_HOURS"
	signKey               = "SIGN_KEY"
	tgToken               = "TG_TOKEN"
	dbUser                = "DB_USER"
	dbPSWD                = "DB_PSWD"
	dbName                = "db_Name"
	dbPort                = "DB_Port"
	//apiKey     = "API_KEY"
	//publicKey  = "vnjLbCnt4ReMVxepNxMqJ2JRh+Wg7Nqebi2YdUy6vhpviF0fnxEPNSjq"
	//privateKey = "z1JMXEjJXiJmkUUROrujpBzL2P53AixU3Vg3pMt7aFcnrfwpiLok/63BMAcvODFYQRHY4V/o7+i9agSdU4IqAxEu"
	//tgToken = "2122664959:AAFQ8E2LCKOb2qfKWhX-qpw4uIvVTLIa0ro"
)

func main() {
	log := logrus.New()

	orderStorage, err := connectDB()
	if err != nil {
		log.Fatalf(err.Error())
	}
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf(err.Error())
	}
	port, err := loadString(srvPort)
	if err != nil {
		log.Fatalf(err.Error())
	}
	key, err := loadString(signKey)
	if err != nil {
		log.Fatalf(err.Error())
	}
	tg, err := loadString(tgToken)
	if err != nil {
		log.Fatalf(err.Error())
	}
	reconnections, err := loadInt(reconnectionsQuantity)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hours, err := loadInt(ttlHours)
	if err != nil {
		log.Fatalf(err.Error())
	}

	userStorage, err := users.NewUserStorage(key, hours)
	if err != nil {
		log.Fatalf(err.Error())
	}
	trader := service.NewAlgoTrader(userStorage, orderStorage, log, reconnections)
	trader.Run()

	tgBot, err := msgwriters.NewTelegramBot(tg)
	if err != nil {
		log.Fatalf(err.Error())
	}
	trader.AddMessageWriter(tgBot)

	handler := handlers.Handler{Trader: trader}

	srv := controller.NewServer(handler)
	go func() {
		if err = srv.Run(port); err != nil {
			log.Fatalf("ERROR: running server is failed <%s>", err)
		}
	}()

	log.Print("Server is running")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Print("Server is shutting down")

	if err = srv.Shutdown(context.Background()); err != nil {
		log.Errorf("ERROR: graceful shutdown is broken <%s>", err.Error())
	}
}

func loadInt(name string) (int64, error) {
	strValue, ok := viper.Get(name).(string)
	if !ok {
		return 0, fmt.Errorf("invalid %s type assertion", name)
	}
	value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s type assertion", name)
	}
	return value, nil
}

func loadString(name string) (string, error) {
	value, ok := viper.Get(name).(string)
	if !ok {
		return "", fmt.Errorf("invalid %s type assertion", name)
	}
	return value, nil
}

func connectDB() (*orders.OrderStorage, error) {
	user, err := loadString(dbUser)
	if err != nil {
		return nil, err
	}
	password, err := loadString(dbPSWD)
	name, err := loadString(dbName)
	port, err := loadString(dbPort)
	orderStorage := orders.OrderStorage{}
	err = orderStorage.Connect(orders.ConnectionConfig{
		Username: user,
		Password: password,
		NameDB:   name,
		Port:     port,
	})
	if err != nil {
		return nil, err
	}
	return &orderStorage, nil
}
