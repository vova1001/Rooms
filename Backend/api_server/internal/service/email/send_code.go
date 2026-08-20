package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
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
	return &STMP{
		host:     host,
		from:     from,
		port:     port,
		username: username,
		password: password,
	}
}

func (s *STMP) SendCode(
	ctx context.Context,
	email string,
	code string,
) error {
	addr := net.JoinHostPort(s.host, s.port)

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}

	startTLSSupported, _ := client.Extension("STARTTLS")
	if !startTLSSupported {
		return fmt.Errorf("smtp server does not support STARTTLS")
	}

	if err := client.StartTLS(&tls.Config{
		ServerName: s.host,
		MinVersion: tls.VersionTLS12,
	}); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}

	auth := smtp.PlainAuth(
		"",
		s.username,
		s.password,
		s.host,
	)

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := client.Rcpt(email); err != nil {
		return fmt.Errorf("smtp recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	subject := "Your verification code"

	body := fmt.Sprintf(`
<html>
<body>
	<h2>Verification code</h2>
	<p>Your code:</p>
	<h1>%s</h1>
	<p>This code expires in 5 minutes.</p>
</body>
</html>
`, code)

	message := []byte(
		"From: " + s.from + "\r\n" +
			"To: " + email + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp finish message: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}

	return nil
}
