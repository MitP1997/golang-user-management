package clients

import (
	"crypto/tls"
	"os"
	"strconv"

	"github.com/MitP1997/golang-user-management/internal/errors"
	"github.com/MitP1997/golang-user-management/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	gomail "gopkg.in/mail.v2"
)

type MailerClient struct {
	logger             *zap.Logger
	from               string
	password           string
	smtpHost           string
	smtpPort           int
	insecureSkipVerify bool
}

func NewMailerClient(logger *zap.Logger) *MailerClient {
	smtpPort, _ := strconv.Atoi(os.Getenv("EMAIL_SMTP_PORT"))
	skipVerify, _ := strconv.ParseBool(os.Getenv("EMAIL_TLS_INSECURE_SKIP_VERIFY"))
	return &MailerClient{
		// can't use context logger as this will be processed in a goroutine and the context might end before the goroutine finishes
		logger:             logger,
		from:               os.Getenv("EMAIL_FROM_ADDRESS"),
		password:           os.Getenv("EMAIL_FROM_PASSWORD"),
		smtpHost:           os.Getenv("EMAIL_SMTP_HOST"),
		smtpPort:           smtpPort,
		insecureSkipVerify: skipVerify,
	}
}

func (m *MailerClient) SendMail(ctx *gin.Context, to string, subject string, body string) *errors.Error {
	logger := utils.GetContextLogger(ctx)
	logger.Info("Sending email", zap.String("to", to), zap.String("subject", subject))
	mail := gomail.NewMessage()
	mail.SetHeader("From", m.from)
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/plain", body)
	d := gomail.NewDialer(m.smtpHost, m.smtpPort, m.from, m.password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: m.insecureSkipVerify}
	if e := d.DialAndSend(mail); e != nil {
		logger.Error("Error sending email", zap.Error(e))
		// add retry logic in case of network glitches
		// ideally use a queue to store the email and retry sending it later
		return errors.SendMailError(e)
	}
	m.logger.Info("Email sent successfully", zap.String("to", to), zap.String("subject", subject))
	return nil
}
