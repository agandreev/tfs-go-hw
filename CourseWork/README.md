# Preparation

----
## Install

    git clone https://github.com/agandreev/tfs-go-hw.git

## Fill config.env file

    SRV_PORT=8000
    RECONNECTION_QUANTITY=5
    SIGN_KEY=
    TTL_HOURS=1
    TG_TOKEN=
    DB_USER=user
    DB_PSWD=passwd
    db_Name=fintech
    DB_Port=5442

## Up database

    docker-compose up

## Run the app

    go run cmd/app/main.go

----
# Rest API

----
Below you can read the descriptions of the endpoints calls

----
**Add user**
----
This option allow you to add your user to the service specifying stock market's public and private keys. Also, you can add your telegram id if you want to be informed about all operations.

* **URL**

  /auth/register

* **Method:**

  `POST`

*  **URL Params**

   None

* **Data Params**

  **Required:**
  ```
  {
  "username": "",
  "public_key": "",
  "private_key": ""
  }
  ```
  **Optional:**
  ```
  {
  "username": "",
  "public_key": "",
  "private_key": "",
  "telegram_id": 0
  }
  ```

* **Success Response:**

  If successful, then you should receive only status code.

    * **Code:** `201 CREATED`

* **Error Response:**

  In case of failure, you should receive status code and error message.

    * **Code:** `400 BAD REQUEST`
      **Content:** `{ error : "the user's id, public key or private key is empty" }`

* **Sample Call:**

  ```
  curl --data '{"username": "1","public_key": "1","private_key": "1"}' \
  http://localhost:<port>/auth/register
  ```
  ----
**Login**
----
This option allows you to get JWT token and use it for authorization in future steps.
Authorization is carried out using `username`, `public key` and `private key`

* **URL**

  /auth/login

* **Method:**

  `POST`

*  **URL Params**

   None

* **Data Params**

  **Required:**
  ```
  {
  "username": "",
  "public_key": "",
  "private_key": ""
  }
  ```

* **Success Response:**

  If successful, then you should receive only status code.

    * **Code:** `200 OK`
      **Content:** `{"token":"xxx"}`

* **Error Response:**

  In case of failure, you should receive status code and error message.

    * **Code:** `400 BAD REQUEST`
      **Content:** `{ error : "incorrect public or private key" }`
    * **Code:** `501 INTERNAL SERVER ERROR`
      **Content:** `{ error : "json: unsupported value: NaN" }`

* **Sample Call:**

  ```
  curl --data '{"username": "1","public_key": "1","private_key": "1"}' \
  http://localhost:<port>/auth/login
  ```
  ----
**Set keys**
----
This option requires authorization by JWT token stored as header. 
It allows you to switch `public key` or `private key`

* **URL**

  /users/set_keys

* **Method:**

  `POST`

*  **URL Params**

   None

* **Data Params**

  **Required:**
  ```
  {
  "public_key": "",
  "private_key": ""
  }
  ```

* **Success Response:**

  If successful, then you should receive only status code.

    * **Code:** `201 CREATED`

* **Error Response:**

  In case of failure, you should receive status code and error message.

    * **Code:** `400 BAD REQUEST`
      **Content:** `{ error : "incorrect public or private key" }`
    * **Code:** `401 UNAUTHORIZED`
      **Content:** `{ error : "token contains an invalid number of segments" }`
    * **Code:** `501 INTERNAL SERVER ERROR`
      **Content:** `{ error : "json: unsupported value: NaN" }`

* **Sample Call:**

  ```
  curl -H 'Authorization: Bearer xxx' \
  --data '{"public_key": "1","private_key": "1"}' \
  http://localhost:<port>/auth/login
  ```
  ----
**Start pair**
----
This option requires authorization by JWT token stored as header.
It allows you to subscribe on trading pair that you're interested in by candle interval.
Also, it works only for 1, 2, 5 and 10 minutes candles.

* **URL**

  /pair/start

* **Method:**

  `POST`

*  **URL Params**

   None

* **Data Params**

  **Required:**
  ```
  {
    "pair_name": "PI_BCHUSD",
    "pair_interval": "candles_trade_1m",
    "indicator_name": "Donchian"
  }
  ```
  **Optional:**
  ```
  {
    "pair_name": "PI_BCHUSD",
    "pair_interval": "candles_trade_1m",
    "indicator_name": "Donchian",
    "limit": 0.05
  }
  ```

* **Success Response:**

  If successful, then you should receive only status code.

    * **Code:** `201 CREATED`

* **Error Response:**

  In case of failure, you should receive status code and error message.

    * **Code:** `400 BAD REQUEST`
      **Content:** `{ error : "unsupported candle type" }`
    * **Code:** `401 UNAUTHORIZED`
      **Content:** `{ error : "token contains an invalid number of segments" }`

* **Sample Call:**

  ```
  curl -H 'Authorization: Bearer xxx' \
  --data '{"pair_name": "PI_BCHUSD","pair_interval": "candles_trade_1m","indicator_name": "Donchian"}' \
  http://localhost:<port>/pair/start
  ```

  ----
**Stop pair**
----
This option requires authorization by JWT token stored as header.
It allows you to unsubscribe trading pair. And it will stopped if users' quantity is zero.

* **URL**

  /pair/stop

* **Method:**

  `POST`

*  **URL Params**

   None

* **Data Params**

  **Required:**
  ```
  {
    "pair_name": "PI_BCHUSD",
    "pair_interval": "candles_trade_1m"
  }
  ```

* **Success Response:**

  If successful, then you should receive only status code.

  * **Code:** `200 OK`

* **Error Response:**

  In case of failure, you should receive status code and error message.

  * **Code:** `400 BAD REQUEST`
    **Content:** `{ error : "unsupported candle type" }`
  * **Code:** `401 UNAUTHORIZED`
    **Content:** `{ error : "token contains an invalid number of segments" }`

* **Sample Call:**

  ```
  curl -H 'Authorization: Bearer xxx' \
  --data '{"pair_name": "PI_BCHUSD","pair_interval": "candles_trade_1m"}' \
  http://localhost:<port>/pair/stop
  ```