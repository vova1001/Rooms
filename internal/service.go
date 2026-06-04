package internal

type partService struct {
	repo *PartRepo
}

func NewService(repo *PartRepo) *partService {
	return &partService{repo: repo}
}
