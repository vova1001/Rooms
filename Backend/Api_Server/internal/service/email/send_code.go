package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type STMP struct {
	host     string
	from     string
	port     string
	username string
	password string
}

func NewSTMP(host, from, port, username, password string) *STMP {
	return &STMP{host: host, from: from, port: port, username: username, password: password}
}

func (s *STMP) SendCode(ctx context.Context, email, code string) error {
	subject := "Your verification code"

	body := fmt.Sprintf(
		`
	<html>
	<body>

	<h2>Verification code</h2>

	<p>Your code:</p>

	<h1>%s</h1>

	<p>This code expires in 5 minutes.</p>

	</body>
	</html>
	`,
		code,
	)

	message := []byte(
		"From: " + s.from + "\r\n" +
			"To: " + email + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	err := smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{email}, message)

	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
