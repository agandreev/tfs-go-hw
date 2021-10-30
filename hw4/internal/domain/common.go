package domain

type GeneralMessage struct {
	ID       uint64 `json:"id,omitempty"`
	Content  string `json:"content"`
	SenderID uint64 `json:"sender_id"`
}

type DirectMessage struct {
	GeneralMessage
	ReceiverID uint64 `json:"receiver_id"`
}

type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
