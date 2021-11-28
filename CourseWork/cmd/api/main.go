package main

import (
	"context"
	"fmt"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/controller"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/handlers"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/repository"
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
	srvPort               = "SRV_PORT"
	reconnectionsQuantity = "RECONNECTION_QUANTITY"
	ttlHours              = "TTL_HOURS"
	signKey               = "SIGN_KEY"
	//apiKey     = "API_KEY"
	//publicKey  = "vnjLbCnt4ReMVxepNxMqJ2JRh+Wg7Nqebi2YdUy6vhpviF0fnxEPNSjq"
	//privateKey = "z1JMXEjJXiJmkUUROrujpBzL2P53AixU3Vg3pMt7aFcnrfwpiLok/63BMAcvODFYQRHY4V/o7+i9agSdU4IqAxEu"
	//tgToken    = "2122664959:AAFQ8E2LCKOb2qfKWhX-qpw4uIvVTLIa0ro"
)

func main() {
	log := logrus.New()
	port, reconnections, secretKey, ttlHours, err := loadConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}

	users, err := repository.NewUserStorage(secretKey, ttlHours)
	if err != nil {
		log.Fatalf(err.Error())
	}
	trader := service.NewAlgoTrader(users, log, reconnections)
	trader.Run()

	cmd := msgwriters.ConsoleWriter{}
	trader.AddMessageWriter(cmd)

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

	//log := logrus.New()
	//
	//user := domain.NewUser("1")
	//user2 := domain.NewUser("2")
	//user2.PublicKey = "sf"
	//user2.PrivateKey = "sdf"
	//user.PublicKey = publicKey
	//user.PrivateKey = privateKey
	//
	//users := repository.NewUserStorage()
	//algoTrader := service.NewAlgoTrader(users, log)
	//algoTrader.Run()
	//err := algoTrader.AddUser(*user)
	//err = algoTrader.AddUser(*user2)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//
	//err = algoTrader.AddPair("2", "FI_ETHUSD_211126", domain.Candle5m)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//
	//err = algoTrader.AddPair("1", "FI_ETHUSD_211126", domain.Candle1m)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//
	//err = algoTrader.AddPair("1", "FI_ETHUSD_211126", domain.Candle5m)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//
	//<-time.Tick(time.Second * 5)
	//algoTrader.DeletePair("1", "FI_ETHUSD_211126", domain.Candle1m)
	//
	//<-time.Tick(time.Second * 5)
	//algoTrader.ShutDown()

	//api := service.NewKrakenAPI()
	//event := domain.StockMarketEvent{
	//	Signal: domain.Sell,
	//	Name:   "PI_BCHUSD",
	//	Volume: 5000,
	//	Close:  577,
	//}
	//message, err := api.AddOrder(event, &domain.User{
	//	Username:         1,
	//	PublicKey:  publicKey,
	//	PrivateKey: privateKey,
	//})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(message)
}

func loadConfig() (string, int64, string, int64, error) {
	viper.SetConfigFile("config.env")

	if err := viper.ReadInConfig(); err != nil {
		return "", 0, "", 0, fmt.Errorf("error while reading config file %s", err)
	}
	port, ok := viper.Get(srvPort).(string)
	if !ok {
		return "", 0, "", 0, fmt.Errorf("invalid %s type assertion", srvPort)
	}
	reconnections, err := getIntFromViper(reconnectionsQuantity)
	if err != nil {
		return "", 0, "", 0, err
	}
	sign, ok := viper.Get(signKey).(string)
	if !ok {
		return "", 0, "", 0, fmt.Errorf("invalid %s type assertion", signKey)
	}
	ttl, err := getIntFromViper(ttlHours)
	if err != nil {
		return "", 0, "", 0, err
	}
	return port, reconnections, sign, ttl, nil
}

func getIntFromViper(name string) (int64, error) {
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
