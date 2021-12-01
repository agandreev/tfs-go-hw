package orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNotConnected = errors.New("there is no db connection")
)

type OrderStorage struct {
	pool   *pgxpool.Pool
	Config ConnectionConfig
}

type ConnectionConfig struct {
	Username string
	Password string
	NameDB   string
	Port     string
}

func (storage *OrderStorage) Connect() error {
	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s"+
		"?sslmode=disable", storage.Config.Username, storage.Config.Password,
		storage.Config.Port, storage.Config.NameDB)
	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("can't connect db <%w>", err)
	}
	storage.pool = pool
	return nil
}

func (storage OrderStorage) AddOrder(info domain.OrderInfo) error {
	if storage.pool == nil {
		return ErrNotConnected
	}
	_, err := storage.pool.Exec(context.Background(),
		"INSERT INTO orders(name, orderID, price, amount, side) VALUES($1, $2, $3, $4, $5)", info.Name,
		info.OrderID, info.Price, info.Amount, info.Side)
	if err != nil {
		return fmt.Errorf("can't add to db <%w>", err)
	}
	return nil
}

func (storage OrderStorage) GetOrders(offset int64) ([]domain.OrderInfo, error) {
	rows, err := storage.pool.Query(context.Background(), "SELECT * FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counter int64
	var tmp domain.OrderInfo
	var id int64
	infos := make([]domain.OrderInfo, 0)

	for rows.Next() && (counter < offset) {
		err = rows.Scan(&id, &tmp.Name, &tmp.OrderID, &tmp.Price, &tmp.Amount, &tmp.Side)
		if err != nil {
			return nil, fmt.Errorf("can't read from db <%w>", err)
		}
		infos = append(infos, tmp)
		counter++
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("buf db error <%w>", err)
	}
	return infos, nil
}

func (storage OrderStorage) Shutdown() {
	if storage.pool != nil {
		storage.pool.Close()
	}
}
