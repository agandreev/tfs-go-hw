/*
	yszl9dv2@futures-demo.com
	6ns9e9u25xr2o0opmf2s

	F2+9p5LpLD1tfO4w8sGAB2mYaul62fMwrdFx/Aga1lI80STL5nP/j8EB
	8bDPoaTM0nhjNVVC/7UoWLzfWGujpcuJC8lMD+jg8UvzoSqbE99koiW0iCwJtXfkarQv2Uuawm8Z3OH95lv5d+rP
 */

package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

const(
	apiKey = "API_KEY"
)

func main() {
	fmt.Println("Hi!")
}

func loadAPiKey() string {
	viper.SetConfigFile("config.env")
	//viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	value, ok := viper.Get(apiKey).(string)
	if !ok {
		log.Fatalf("Invalid type assertion")
	}

	return value
}
