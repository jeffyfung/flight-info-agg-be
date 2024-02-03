package sendgrid

import (
	"fmt"
	"log"

	"github.com/go-errors/errors"
	"github.com/jeffyfung/flight-info-agg/config"
	"github.com/jeffyfung/flight-info-agg/pkg/email"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridAgent struct {
	sender email.EmailAccount
}

func New() SendGridAgent {
	sender := email.EmailAccount{
		Name:    "852 Flight Deals",
		Address: config.Cfg.Email.FromEmail,
	}
	return SendGridAgent{sender}
}

func (sg SendGridAgent) Sender() email.EmailAccount {
	return sg.sender
}

func (sg SendGridAgent) Send(to email.EmailAccount, subject string, content email.Content) error {
	sender := sg.Sender()
	from := mail.NewEmail(sender.Name, sender.Address)
	target := mail.NewEmail(to.Name, to.Address)
	message := mail.NewSingleEmail(from, subject, target, content.PlainText, content.Html)

	client := sendgrid.NewSendClient(config.Cfg.Email.SendGridAPIKey)
	res, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return errors.New("Error sending email (SendGrid)" + err.Error())
	}
	fmt.Printf("Email sent to %s. Response: %v\n", to.Address, res.StatusCode)
	return nil

}
