package session

const (
	authSessionPrefix         = "session:auth:"
	registrationSessionPrefix = "session:registration:"
)

type Service struct {
	repo      RepositoryS
	generator Generator
	haser     Hasher
}

func NewService(repo RepositoryS, generator Generator, hasher Hasher) *Service {
	return &Service{
		repo:      repo,
		generator: generator,
		haser:     hasher,
	}
}
