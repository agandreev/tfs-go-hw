package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	timeout        = 10 * time.Second
	orderURL       = "https://demo-futures.kraken.com/derivatives/api/v3/sendorder"
	derivativePath = "/derivatives"
	ioc            = "ioc"
	success        = "success"
	placed         = "placed"
)

// KrakenApi implements Rest API communication with stock market.
type KrakenApi struct {
	client *http.Client
}

// NewKrakenAPI returns pointer to KrakenApi.
func NewKrakenAPI() *KrakenApi {
	return &KrakenApi{
		client: &http.Client{Timeout: timeout},
	}
}

// getSign creates sign for stock market's private methods.
func (api *KrakenApi) getSign(apiPath string, urlValues url.Values,
	privateKey string) (string, error) {
	sha := sha256.New()

	if _, err := sha.Write([]byte(urlValues.Encode() + apiPath)); err != nil {
		return "", err
	}
	decodedSecretKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha512.New, decodedSecretKey)

	if _, err = mac.Write(sha.Sum(nil)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil

}

// AddOrder adds order by StockMarketEvent.
func (api *KrakenApi) AddOrder(event domain.StockMarketEvent, user *domain.User) (domain.Message, error) {
	urlValues := url.Values{
		"symbol":     {strings.ToLower(event.Name)},
		"side":       {string(event.Signal)},
		"orderType":  {ioc},
		"size":       {strconv.FormatInt(event.Volume, 10)},
		"limitPrice": {fmt.Sprintf("%f", event.Close)},
	}

	req, err := http.NewRequest("POST", orderURL, nil)
	parse, err := url.Parse(orderURL)
	if err != nil {
		return domain.Message{}, err
	}
	apiPath := parse.Path
	if strings.HasPrefix(apiPath, derivativePath) {
		apiPath = apiPath[len(derivativePath):]
	}
	signature, err := api.getSign(apiPath, urlValues, user.PrivateKey)
	if err != nil {
		return domain.Message{}, err
	}
	req.Header.Add("APIKey", user.PublicKey)
	req.Header.Add("Authent", signature)
	req.URL.RawQuery = urlValues.Encode()

	//todo: add response code checking
	resp, err := api.client.Do(req)
	if err != nil {
		return domain.Message{}, err
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return domain.Message{}, err
	}
	var orderResponse domain.OrderResponse
	err = json.Unmarshal(data, &orderResponse)
	if err != nil {
		return domain.Message{}, err
	}
	if err = checkOrderResponse(orderResponse); err != nil {
		return domain.Message{}, err
	}

	message, err := orderResponse.Message()
	if err != nil {
		return domain.Message{}, err
	}
	message.Name = event.Name
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
