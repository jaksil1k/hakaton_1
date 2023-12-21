package service

import "github.com/tunx321/task5/internal/db"


type Repository interface{
	GetAllTransactions(string)([]db.FrontData, error)
}


type Service struct{
	Repo Repository
}


func NewService(repo Repository) *Service {
	return &Service{
			Repo: repo,
	}
}


func (s *Service) GetAllTransactions(address string)([]db.FrontData, error){
	data, err := s.Repo.GetAllTransactions(address)
	if err !=nil{
		return nil, err
	}
	return data, nil
}