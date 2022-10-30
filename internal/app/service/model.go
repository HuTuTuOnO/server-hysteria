package service

type Service interface {
	Start() error
	Close() error
}
