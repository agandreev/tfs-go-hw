package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/controller"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/handlers"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/repository/orders"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/repository/users"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	logFile               = "logs.txt"
)

func main() {
	log := logrus.New()
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer file.Close()
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	port, key, tg, reconnections, hours, orderConfig, err := loadConfig()
	if err != nil {
		return
	}
	orderStorage := orders.OrderStorage{Config: *orderConfig}

	userStorage, err := users.NewUserStorage(key, hours)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tgBot, err := service.NewTelegramBot(tg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	trader := service.NewAlgoTrader(userStorage, &orderStorage, log, reconnections)
	trader.AddMessageWriter(tgBot)
	if err = trader.Run(); err != nil {
		log.Fatalf(err.Error())
	}

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

func loadConfig() (string, string, string, int64, int64, *orders.ConnectionConfig, error) {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return "", "", "", 0, 0, nil, fmt.Errorf("can't load config: %w", err)
	}
	port, err := loadString(srvPort)
	if err != nil {
		return "", "", "", 0, 0, nil, err
	}
	key, err := loadString(signKey)
	if err != nil {
		return "", "", "", 0, 0, nil, err
	}
	tg, err := loadString(tgToken)
	if err != nil {
		return "", "", "", 0, 0, nil, err
	}
	reconnections, err := loadInt(reconnectionsQuantity)
	if err != nil {
		return "", "", "", 0, 0, nil, err
	}
	hours, err := loadInt(ttlHours)
	if err != nil {
		return "", "", "", 0, 0, nil, err
	}
	connectionConfig, err := loadDBVars()
	if err != nil {
		return "", "", "", 0, 0, nil, err
	}
	return port, key, tg, reconnections, hours, connectionConfig, nil
}

func loadDBVars() (*orders.ConnectionConfig, error) {
	user, err := loadString(dbUser)
	if err != nil {
		return nil, err
	}
	password, err := loadString(dbPSWD)
	if err != nil {
		return nil, err
	}
	name, err := loadString(dbName)
	if err != nil {
		return nil, err
	}
	port, err := loadString(dbPort)
	if err != nil {
		return nil, err
	}
	orderConfig := &orders.ConnectionConfig{
		Username: user,
		Password: password,
		NameDB:   name,
		Port:     port,
	}
	return orderConfig, nil
}
