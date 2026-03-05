package engine

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/2comjie/mcpflow/internal/model"
)

func executeEmail(cfg *model.EmailConfig, ctx *WorkflowContext) (any, error) {
	if cfg == nil {
		return nil, fmt.Errorf("email config is nil")
	}

	contentType := cfg.ContentType
	if contentType == "" {
		contentType = "text/plain"
	}

	to := resolveTemplate(cfg.To, ctx)
	subject := resolveTemplate(cfg.Subject, ctx)
	body := resolveTemplate(cfg.Body, ctx)

	toList := strings.Split(to, ",")
	for i := range toList {
		toList[i] = strings.TrimSpace(toList[i])
	}

	header := fmt.Sprintf("From: %s\r\nTo: %s\r\n", cfg.From, to)
	if cfg.Cc != "" {
		header += fmt.Sprintf("Cc: %s\r\n", cfg.Cc)
	}
	header += fmt.Sprintf("Subject: %s\r\n", subject)
	header += fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType)
	header += "\r\n"

	msg := []byte(header + body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	// Port 465 requires implicit TLS; smtp.SendMail only supports STARTTLS (port 587)
	if cfg.SMTPPort == 465 {
		if err := sendMailTLS(addr, cfg.SMTPHost, cfg.Username, cfg.Password, cfg.From, toList, msg); err != nil {
			return nil, fmt.Errorf("send email (tls): %w", err)
		}
	} else {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
		if err := smtp.SendMail(addr, auth, cfg.From, toList, msg); err != nil {
			return nil, fmt.Errorf("send email: %w", err)
		}
	}

	return map[string]any{
		"status": "sent",
		"to":     cfg.To,
	}, nil
}

// sendMailTLS sends email over implicit TLS (port 465).
func sendMailTLS(addr, host, username, password string, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	auth := smtp.PlainAuth("", username, password, host)
	if err = c.Auth(auth); err != nil {
		// Some servers reject PlainAuth over already-encrypted connections; try LOGIN
		if loginErr := c.Auth(&loginAuth{username, password}); loginErr != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err = c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	for _, rcpt := range to {
		if err = c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt to %s: %w", rcpt, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return c.Quit()
}

// loginAuth implements smtp.Auth for the LOGIN mechanism (required by some SMTP servers).
type loginAuth struct{ username, password string }

func (a *loginAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch strings.ToLower(string(fromServer)) {
		case "username:":
			return []byte(a.username), nil
		case "password:":
			return []byte(a.password), nil
		}
	}
	return nil, nil
}

