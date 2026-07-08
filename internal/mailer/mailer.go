package mailer

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/lancer/log/internal/config"
)

func SendPasswordResetCode(cfg config.SMTPConfig, to, brand, code string) error {
	if !cfg.Ready() {
		return fmt.Errorf("smtp is not configured")
	}
	brand = strings.TrimSpace(brand)
	if brand == "" {
		brand = "lancer.log"
	}

	subject := fmt.Sprintf("%s 管理端密码重置验证码", brand)
	body := fmt.Sprintf("你的 %s 管理端密码重置验证码是：%s\n\n验证码 10 分钟内有效。如果不是你本人操作，请忽略这封邮件。\n", brand, code)

	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\n", cfg.From)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: %s\r\n", encodeSubject(subject))
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&msg, "\r\n%s", body)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	// 465 用隐式 TLS；587/25 用 STARTTLS 或明文
	if cfg.Port == 465 {
		return sendMailImplicitTLS(addr, cfg.Host, auth, cfg.From, []string{to}, msg.Bytes())
	}
	return smtp.SendMail(addr, auth, cfg.From, []string{to}, msg.Bytes())
}

func sendMailImplicitTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{ServerName: host})
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("tls handshake: %w", err)
	}
	defer tlsConn.Close()

	c, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}
	defer c.Quit()

	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt to: %w", err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return nil
}

func encodeSubject(s string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}