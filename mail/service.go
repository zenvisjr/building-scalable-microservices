package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Service interface {
	SendEmail(ctx context.Context, to, subject, templateName string, data any) error
	StartEmailSubscriber() error
}

type MailConfig struct {
	FromEmail string `envconfig:"FROM_EMAIL"`
	ApiKey    string `envconfig:"SENDGRID_API_KEY"`
	FromName  string `envconfig:"FROM_NAME"`
}

type MailService struct {
	config MailConfig
	client *sendgrid.Client
	nats   *nats.Conn
}

func NewMailService(cfg MailConfig, nc *nats.Conn) Service {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Creating new MailService instance")
	return &MailService{
		config: cfg,
		client: sendgrid.NewSendClient(cfg.ApiKey),
		nats:   nc,
	}
}

func (m *MailService) SendEmail(ctx context.Context, to, subject, templateName string, data any) error {
	Logs := logger.GetGlobalLogger()
	// Logs.Info(ctx, "Sending email to "+to)
	// Logs.Info(ctx, "api key: "+m.config.ApiKey)
	// Logs.Info(ctx, "from email: "+m.config.FromEmail)
	// Logs.Info(ctx, "from name: "+m.config.FromName)

	if to == "" || subject == "" || templateName == "" {
		Logs.Warn(ctx, "SendEmail called with empty fields")
		return fmt.Errorf("email: missing required fields")
	}

	textBody, htmlBody, err := RenderTemplates(templateName, data)
	if err != nil {
		Logs.Error(ctx, "Template rendering failed: "+err.Error())
		return err
	}

	from := mail.NewEmail(m.config.FromName, m.config.FromEmail)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, textBody, htmlBody)

	Logs.LocalOnlyInfo("Sending email to " + to)

	response, err := m.client.Send(message)
	if err != nil {
		Logs.Error(ctx, "Failed to send email: "+err.Error())
		return err
	}

	if response.StatusCode >= 400 {
		Logs.Error(ctx, "SendGrid error: status "+strconv.Itoa(response.StatusCode)+", body: "+response.Body)
		return err
	}

	Logs.Info(ctx, "Email sent successfully to "+to)
	Logs.LocalOnlyInfo("Email sent successfully to " + to)
	return nil
}

//we will subscribe to nats server

func (m *MailService) StartEmailSubscriber() error {
	Logs := logger.GetGlobalLogger()

	Logs.LocalOnlyInfo("Starting email subscriber")

	_, err := m.nats.Subscribe("emails.send", func(msg *nats.Msg) {
		Logs.LocalOnlyInfo("📩 Received email job from NATS")

		var payload struct {
			To           string            `json:"to"`
			Subject      string            `json:"subject"`
			TemplateName string            `json:"templateName"`
			TemplateData map[string]string `json:"templateData"`
		}

		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			Logs.Error(context.Background(), "Invalid email job payload: "+err.Error())
			return
		}

		ctx := context.Background()
		Logs.Info(ctx, "Received email job from NATS")
		Logs.Info(ctx, "Sending email to "+payload.To)
		err := m.SendEmail(ctx, payload.To, payload.Subject, payload.TemplateName, payload.TemplateData)
		if err != nil {
			Logs.Error(ctx, "Failed to send email from queue: "+err.Error())
		}
	})
	return err
}
