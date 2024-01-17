package svc

type Service struct {
	// ...
}

func New() *Service {
	return &Service{}
}

func (*Service) Run() {
	// ...
}
