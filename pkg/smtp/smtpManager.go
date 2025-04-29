package smtp

import (
	"crypto/tls"
	"fmt"
	"github.com/medods/auth-service/internal/config"
	"net/smtp"
	"strings"
)

type EmailSender struct {
	config *config.SMTPConfig
}

func NewEmailSender(config *config.SMTPConfig) *EmailSender {
	return &EmailSender{config: config}
}

func (s *EmailSender) SendAlert(subject, body string) error {
	if s.config.Username == "" || s.config.Password == "" {
		fmt.Printf("\n=== Email Alert ===\nSubject: %s\nBody: %s\n==================\n", subject, body)
		return nil
	}

	headers := make([]string, 0)
	headers = append(headers, fmt.Sprintf("From: %s", s.config.From))
	headers = append(headers, fmt.Sprintf("Subject: %s", subject))
	headers = append(headers, "MIME-Version: 1.0")
	headers = append(headers, "Content-Type: text/plain; charset=utf-8")

	// Собираем сообщение
	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	if s.config.UseTLS {
		// Настраиваем TLS конфигурацию
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("ошибка при установке TLS соединения: %w", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			return fmt.Errorf("ошибка при создании SMTP клиента: %w", err)
		}
		defer c.Close()

		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("ошибка аутентификации: %w", err)
		}

		if err = c.Mail(s.config.From); err != nil {
			return fmt.Errorf("ошибка при указании отправителя: %w", err)
		}

		if err = c.Rcpt(s.config.From); err != nil {
			return fmt.Errorf("ошибка при указании получателя: %w", err)
		}

		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("ошибка при подготовке данных: %w", err)
		}
		defer w.Close()

		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("ошибка при отправке данных: %w", err)
		}
	} else {
		// Отправляем без TLS
		err := smtp.SendMail(addr, auth, s.config.From, []string{s.config.From}, []byte(message))
		if err != nil {
			return fmt.Errorf("ошибка при отправке письма: %w", err)
		}
	}

	return nil
}
