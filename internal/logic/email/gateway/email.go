package gateway

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"os"
	"unibee/internal/logic/email/sender"
	"unibee/utility"
)

type SmtpConfig struct {
	SmtpHost      string `json:"smtpHost"`
	SmtpPort      int    `json:"smtpPort"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	AuthType      string `json:"authType"`
	OauthToken    string `json:"oauthToken"`
	UseTLS        bool   `json:"useTLS"`
	SkipTLSVerify bool   `json:"skipTLSVerify"`
}

func SendEmailToUser(f *sender.Sender, emailGatewayKey string, mailTo string, subject string, body string) (result string, err error) {
	if f == nil {
		f = sender.GetDefaultSender()
	}
	from := mail.NewEmail(f.Name, f.Address)
	to := mail.NewEmail(mailTo, mailTo)
	plainTextContent := body
	htmlContent := "<div>" + body + " </div>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(emailGatewayKey)
	response, err := client.Send(message)
	if err != nil {
		fmt.Printf("SendEmailToUser error:%s\n", err.Error())
		return "", err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return utility.MarshalToJsonString(response), nil
}

func SendPdfAttachEmailToUser(f *sender.Sender, emailGatewayKey string, mailTo string, subject string, body string, pdfFilePath string, pdfFileName string) (result string, err error) {
	if f == nil {
		f = sender.GetDefaultSender()
	}
	from := mail.NewEmail(f.Name, f.Address)
	to := mail.NewEmail(mailTo, mailTo)
	plainTextContent := body
	htmlContent := "<div>" + body + " </div>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	attach := mail.NewAttachment()
	dat, err := os.ReadFile(pdfFilePath)
	if err != nil {
		fmt.Println(err)
	}
	encoded := base64.StdEncoding.EncodeToString(dat)
	attach.SetContent(encoded)
	attach.SetType("application/pdf")
	attach.SetFilename(pdfFileName)
	attach.SetDisposition("attachment")
	message.AddAttachment(attach)
	client := sendgrid.NewSendClient(emailGatewayKey)
	response, err := client.Send(message)
	if err != nil {
		fmt.Printf("SendPdfAttachEmailToUser error:%s\n", err.Error())
		return "", err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return utility.MarshalToJsonString(response), nil
}

func SendSmtpEmailToUser(f *sender.Sender, smtpConfigData string, mailTo string, subject string, body string) (result string, err error) {
	if f == nil {
		f = sender.GetDefaultSender()
	}
	cfg, err := parseSmtpConfig(smtpConfigData)
	if err != nil {
		return "", err
	}
	msg, err := buildSmtpMessage(f.Address, mailTo, subject, body, "", "")
	if err != nil {
		return "", err
	}
	return sendBySmtp(cfg, f.Address, mailTo, msg)
}

func SendSmtpAttachEmailToUser(f *sender.Sender, smtpConfigData string, mailTo string, subject string, body string, pdfFilePath string, pdfFileName string) (result string, err error) {
	if f == nil {
		f = sender.GetDefaultSender()
	}
	cfg, err := parseSmtpConfig(smtpConfigData)
	if err != nil {
		return "", err
	}
	msg, err := buildSmtpMessage(f.Address, mailTo, subject, body, pdfFilePath, pdfFileName)
	if err != nil {
		return "", err
	}
	return sendBySmtp(cfg, f.Address, mailTo, msg)
}

func parseSmtpConfig(data string) (*SmtpConfig, error) {
	cfg := &SmtpConfig{}
	err := utility.UnmarshalFromJsonString(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp config: %w", err)
	}
	if cfg.SmtpHost == "" {
		return nil, fmt.Errorf("invalid smtp config: smtpHost is required")
	}
	if cfg.SmtpPort <= 0 {
		if cfg.UseTLS {
			cfg.SmtpPort = 465
		} else {
			cfg.SmtpPort = 587
		}
	}
	if cfg.AuthType == "" {
		cfg.AuthType = "plain"
	}
	return cfg, nil
}

func buildSmtpMessage(fromAddress string, toAddress string, subject string, htmlBody string, attachPath string, attachName string) ([]byte, error) {
	headers := map[string]string{
		"From":         fromAddress,
		"To":           toAddress,
		"Subject":      mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version": "1.0",
	}

	if attachPath == "" {
		headers["Content-Type"] = `text/html; charset="UTF-8"`
		var sb strings.Builder
		for k, v := range headers {
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		sb.WriteString("\r\n")
		sb.WriteString(htmlBody)
		return []byte(sb.String()), nil
	}

	if attachName == "" {
		attachName = filepath.Base(attachPath)
	}
	fileBytes, err := os.ReadFile(attachPath)
	if err != nil {
		return nil, err
	}

	boundary := fmt.Sprintf("UNIBEE-%d", time.Now().UnixNano())
	headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%q", boundary)

	var sb strings.Builder
	for k, v := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	sb.WriteString("\r\n")

	writer := multipart.NewWriter(&sb)
	_ = writer.SetBoundary(boundary)

	bodyHeader := make(textproto.MIMEHeader)
	bodyHeader.Set("Content-Type", `text/html; charset="UTF-8"`)
	bodyPart, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return nil, err
	}
	_, err = bodyPart.Write([]byte(htmlBody))
	if err != nil {
		return nil, err
	}

	attachHeader := make(textproto.MIMEHeader)
	attachHeader.Set("Content-Type", "application/pdf")
	attachHeader.Set("Content-Transfer-Encoding", "base64")
	attachHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, attachName))
	attachPart, err := writer.CreatePart(attachHeader)
	if err != nil {
		return nil, err
	}
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(fileBytes)))
	base64.StdEncoding.Encode(encoded, fileBytes)
	_, err = attachPart.Write(encoded)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return []byte(sb.String()), nil
}

func sendBySmtp(cfg *SmtpConfig, fromAddress string, toAddress string, message []byte) (string, error) {
	address := fmt.Sprintf("%s:%d", cfg.SmtpHost, cfg.SmtpPort)
	tlsConfig := &tls.Config{ServerName: cfg.SmtpHost, InsecureSkipVerify: cfg.SkipTLSVerify}

	sendWithConn := func(conn net.Conn, forceStartTLS bool) error {
		client, err := smtp.NewClient(conn, cfg.SmtpHost)
		if err != nil {
			_ = conn.Close()
			return err
		}
		defer func() {
			_ = client.Quit()
			_ = client.Close()
		}()

		if forceStartTLS {
			if ok, _ := client.Extension("STARTTLS"); !ok {
				return fmt.Errorf("smtp server does not support STARTTLS")
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				return err
			}
		} else if !cfg.UseTLS {
			if ok, _ := client.Extension("STARTTLS"); ok {
				if err := client.StartTLS(tlsConfig); err != nil {
					return err
				}
			}
		}

		auth, err := buildSmtpAuth(cfg)
		if err != nil {
			return err
		}
		if auth != nil {
			if ok, _ := client.Extension("AUTH"); !ok {
				return fmt.Errorf("smtp server does not support AUTH")
			}
			if err := client.Auth(auth); err != nil {
				return err
			}
		}

		if err := client.Mail(fromAddress); err != nil {
			return err
		}
		if err := client.Rcpt(toAddress); err != nil {
			return err
		}

		wc, err := client.Data()
		if err != nil {
			return err
		}
		if _, err = wc.Write(message); err != nil {
			_ = wc.Close()
			return err
		}
		if err = wc.Close(); err != nil {
			return err
		}

		return nil
	}

	if cfg.UseTLS {
		conn, err := tls.Dial("tcp", address, tlsConfig)
		if err == nil {
			if err = sendWithConn(conn, false); err == nil {
				return `{"status":250,"message":"smtp accepted"}`, nil
			}
		}

		// Fallback for servers configured for STARTTLS-only ports (e.g., 587).
		plainConn, plainErr := net.Dial("tcp", address)
		if plainErr != nil {
			if err != nil {
				return "", fmt.Errorf("smtp implicit TLS failed: %v; STARTTLS fallback dial failed: %v", err, plainErr)
			}
			return "", plainErr
		}
		if startTLSErr := sendWithConn(plainConn, true); startTLSErr != nil {
			if err != nil {
				return "", fmt.Errorf("smtp implicit TLS failed: %v; STARTTLS fallback failed: %v", err, startTLSErr)
			}
			return "", startTLSErr
		}
		return `{"status":250,"message":"smtp accepted"}`, nil
	} else {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return "", err
		}
		if err = sendWithConn(conn, false); err != nil {
			return "", err
		}
		return `{"status":250,"message":"smtp accepted"}`, nil
	}
}

func buildSmtpAuth(cfg *SmtpConfig) (smtp.Auth, error) {
	switch strings.ToLower(cfg.AuthType) {
	case "none":
		return nil, nil
	case "plain", "":
		if cfg.Username == "" || cfg.Password == "" {
			return nil, fmt.Errorf("smtp username/password required for plain auth")
		}
		return smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SmtpHost), nil
	case "cram-md5":
		if cfg.Username == "" || cfg.Password == "" {
			return nil, fmt.Errorf("smtp username/password required for cram-md5 auth")
		}
		return smtp.CRAMMD5Auth(cfg.Username, cfg.Password), nil
	case "xoauth2":
		return nil, fmt.Errorf("smtp xoauth2 auth is not supported yet")
	default:
		return nil, fmt.Errorf("unsupported smtp auth type: %s", cfg.AuthType)
	}
}
