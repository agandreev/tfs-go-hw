package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
)

const (
	timeout        = 10 * time.Second
	orderURL       = "https://demo-futures.kraken.com/derivatives/api/v3/sendorder"
	derivativePath = "/derivatives"
	ioc            = "ioc"
	success        = "success"
	placed         = "placed"
)

// KrakenAPI implements Rest API communication with stock market.
type KrakenAPI struct {
	client *http.Client
}

// NewKrakenAPI returns pointer to KrakenAPI.
func NewKrakenAPI() *KrakenAPI {
	return &KrakenAPI{
		client: &http.Client{Timeout: timeout},
	}
}

// getSign creates sign for stock market's private methods.
func (api *KrakenAPI) getSign(apiPath string, urlValues url.Values,
	privateKey string) (string, error) {
	sha := sha256.New()

	if _, err := sha.Write([]byte(urlValues.Encode() + apiPath)); err != nil {
		return "", fmt.Errorf("sha encoding error: <%w>", err)
	}
	decodedSecretKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("base64 decoding error: <%w>", err)
	}
	mac := hmac.New(sha512.New, decodedSecretKey)

	if _, err = mac.Write(sha.Sum(nil)); err != nil {
		return "", fmt.Errorf("mac encoding error: <%w>", err)
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// AddOrder adds order by StockMarketEvent.
func (api *KrakenAPI) AddOrder(event domain.StockMarketEvent, user *domain.User) (domain.OrderInfo, error) {
	if event.Volume <= 0 {
		event.Volume = 1
	}
	limitPrice := countLimitPrice(event.Signal, event.Close, user.GetLimit(event.Name))
	urlValues := url.Values{
		"symbol":     {strings.ToLower(event.Name)},
		"side":       {string(event.Signal)},
		"orderType":  {ioc},
		"size":       {strconv.FormatInt(event.Volume, 10)},
		"limitPrice": {fmt.Sprintf("%.1f", limitPrice)},
	}

	req, err := http.NewRequest("POST", orderURL, nil)
	if err != nil {
		return domain.OrderInfo{}, fmt.Errorf("can't send order request: <%w>", err)
	}
	parse, err := url.Parse(orderURL)
	if err != nil {
		return domain.OrderInfo{}, fmt.Errorf("can't parse order url: <%w>", err)
	}
	apiPath := parse.Path
	apiPath = strings.TrimPrefix(apiPath, derivativePath)
	signature, err := api.getSign(apiPath, urlValues, user.PrivateKey)
	if err != nil {
		return domain.OrderInfo{}, err
	}
	req.Header.Add("APIKey", user.PublicKey)
	req.Header.Add("Authent", signature)
	req.URL.RawQuery = urlValues.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return domain.OrderInfo{}, fmt.Errorf("can't send order request: <%w>", err)
	}
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusAccepted {
		return domain.OrderInfo{}, fmt.Errorf(
			"order request is sended, but status code: <%d>", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return domain.OrderInfo{}, fmt.Errorf("can't read order response: <%w>", err)
	}
	var orderResponse domain.OrderResponse
	err = json.Unmarshal(data, &orderResponse)
	if err != nil {
		return domain.OrderInfo{}, fmt.Errorf("can't unmarshal order response: <%w>", err)
	}
	if err = checkOrderResponse(orderResponse); err != nil {
		return domain.OrderInfo{}, err
	}

	message, err := orderResponse.Message(event.Name)
	if err != nil {
		return domain.OrderInfo{}, fmt.Errorf(
			"can't structurized order response: <%w>", err)
	}
	return message, nil
}

// checkOrderResponse checks order status.
func checkOrderResponse(response domain.OrderResponse) error {
	if response.Result != success {
		return fmt.Errorf("can't process order cause of stock market side problem")
	}
	if response.SendStatus.Status != placed {
		return fmt.Errorf("can't process order because of <%s>", response.SendStatus.Status)
	}
	return nil
}

func countLimitPrice(side domain.Signal, price float64, limit float64) float64 {
	var limitPrice float64
	switch side {
	case domain.Buy:
		limitPrice = price * (1 + limit)
	case domain.Sell:
		limitPrice = price * (1 - limit)
	default:
		limitPrice = price
	}
	return limitPrice
}
