package otp

type Service struct {
	generator Generator
	hasher    Hasher
}

func NewService(generator Generator, hasher Hasher) *Service {
	return &Service{
		generator: generator,
		hasher:    hasher,
	}
}

func (s *Service) Generate() (string, error) {
	code, err := s.generator.Generate()

	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) Hash(email string, code string) string {
	return s.hasher.Hash(
		email,
		code,
	)
}

func (s *Service) Verify(email string, code string, hash string) bool {
	return s.hasher.Verify(
		email,
		code,
		hash,
	)
}
