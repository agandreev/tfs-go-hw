package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"
)

const (
	timeout = 10 * time.Second
	accountURL = "https://demo-futures.kraken.com/derivatives/api/v3/accounts"
)

type KrakenApi struct {
	client *http.Client
	secretKey string
	publicKey string
}

func (api KrakenApi) AccountBalance() {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", accountURL, nil)
	if err != nil {
		return err
	}

	//q := url2.Values{}
	//q.Add("start", "1")
	//q.Add("limit", "5000")
	//q.Add("convert", "USD")

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", key)
	//req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respData, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	fmt.Println(string(respData))

	body := resp.Body
	defer body.Close()
	var mapResponse domain.MapResponseJSON
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &mapResponse)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", mapResponse)
	fmt.Println(mapResponse.Len())
	return nil
}
