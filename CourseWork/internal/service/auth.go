package service

type Auth struct {
	apiKey string
}

func NewAuth(key string) *Auth {
	return &Auth{
		apiKey: key,
	}
}
