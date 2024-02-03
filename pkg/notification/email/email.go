package email

type EmailAccount struct {
	Name    string
	Address string
}

type Content struct {
	PlainText string
	Html      string
}

// TODO: implement the notifier interface
type Agent interface {
	Sender() EmailAccount
	Send(to EmailAccount, subject string, content Content) error
}

// func Send() error {
// 	from := mail.NewEmail("Example User", "flightdeals852@gmail.com")
// 	subject := "Sending with SendGrid is Fun"
// 	to := mail.NewEmail("Example User", "flightdeals852@gmail.com")
// 	plainTextContent := "and easy to do anywhere, even with Go"
// 	htmlContent := "<strong>and easy to do anywhere, even with Go</strong>"
// 	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

// 	log.Println(config.Cfg.Email.SendGridAPIKey)

// 	client := sendgrid.NewSendClient(config.Cfg.Email.SendGridAPIKey)
// 	response, err := client.Send(message)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	} else {
// 		fmt.Println(response.StatusCode)
// 		fmt.Println(response.Body)
// 		fmt.Println(response.Headers)
// 		return nil
// 	}
// }
