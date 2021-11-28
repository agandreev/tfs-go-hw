package main

import (
	"context"
	"fmt"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/controller"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/handlers"
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
	//apiKey     = "API_KEY"
	//publicKey  = "vnjLbCnt4ReMVxepNxMqJ2JRh+Wg7Nqebi2YdUy6vhpviF0fnxEPNSjq"
	//privateKey = "z1JMXEjJXiJmkUUROrujpBzL2P53AixU3Vg3pMt7aFcnrfwpiLok/63BMAcvODFYQRHY4V/o7+i9agSdU4IqAxEu"
	//tgToken = "2122664959:AAFQ8E2LCKOb2qfKWhX-qpw4uIvVTLIa0ro"
)

func main() {
	log := logrus.New()
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
	trader := service.NewAlgoTrader(userStorage, log, reconnections)
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
