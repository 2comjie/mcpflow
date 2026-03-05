package engine

import (
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

	// 构建收件人列表
	toList := strings.Split(to, ",")
	for i := range toList {
		toList[i] = strings.TrimSpace(toList[i])
	}

	// 构建邮件内容
	header := fmt.Sprintf("From: %s\r\nTo: %s\r\n", cfg.From, to)
	if cfg.Cc != "" {
		header += fmt.Sprintf("Cc: %s\r\n", cfg.Cc)
	}
	header += fmt.Sprintf("Subject: %s\r\n", subject)
	header += fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType)
	header += "\r\n"

	msg := []byte(header + body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)

	if err := smtp.SendMail(addr, auth, cfg.From, toList, msg); err != nil {
		return nil, fmt.Errorf("send email: %w", err)
	}

	return map[string]any{
		"status": "sent",
		"to":     cfg.To,
	}, nil
}
