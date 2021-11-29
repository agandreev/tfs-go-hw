package orders

import "github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"

type OrderRepository interface {
	AddOrder(domain.OrderInfo) error
	GetOrders(int64) ([]domain.OrderInfo, error)
	Shutdown()
}
