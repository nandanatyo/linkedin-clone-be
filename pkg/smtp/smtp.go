package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type EmailService interface {
	SendVerificationEmail(to, fullName, code string) error
	SendPasswordResetEmail(to, fullName, code string) error
}

type emailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailService(host string, port int, username, password string) EmailService {
	return &emailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     username,
	}
}

func (s *emailService) SendVerificationEmail(to, fullName, code string) error {
	subject := "Verify Your Email - LinkedIn Clone"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Email Verification</h2>
			<p>Hi %s,</p>
			<p>Thank you for signing up! Please use the following code to verify your email address:</p>
			<h3 style="color: #0073b1; font-size: 24px; letter-spacing: 2px;">%s</h3>
			<p>This code will expire in 15 minutes.</p>
			<p>If you didn't create an account with us, you can safely ignore this email.</p>
			<br>
			<p>Best regards,<br>LinkedIn Clone Team</p>
		</body>
		</html>
	`, fullName, code)

	return s.sendEmail(to, subject, body)
}

func (s *emailService) SendPasswordResetEmail(to, fullName, code string) error {
	subject := "Reset Your Password - LinkedIn Clone"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset</h2>
			<p>Hi %s,</p>
			<p>You requested to reset your password. Please use the following code:</p>
			<h3 style="color: #0073b1; font-size: 24px; letter-spacing: 2px;">%s</h3>
			<p>This code will expire in 15 minutes.</p>
			<p>If you didn't request a password reset, you can safely ignore this email.</p>
			<br>
			<p>Best regards,<br>LinkedIn Clone Team</p>
		</body>
		</html>
	`, fullName, code)

	return s.sendEmail(to, subject, body)
}

func (s *emailService) sendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n"+
		"MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body)

	client, err := smtp.Dial(fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(s.from); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write([]byte(msg))
	return err
}
