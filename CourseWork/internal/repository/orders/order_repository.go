package orders

import "github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"

type OrderRepository interface {
	AddOrder(info domain.OrderInfo) error
	GetOrders(info domain.OrderInfo) ([]domain.OrderInfo, error)
}
