package email

import (
	"bytes"
	"crypto/tls"
	"net/smtp"

	"github.com/zinic/forculus/config"
)

func tlsConfig(smtpServer config.SMTPServer) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpServer.Host,
	}
}

func tlsDial(smtpServer config.SMTPServer) (*tls.Conn, error) {
	return tls.Dial("tcp", smtpServer.FormatAddress(), tlsConfig(smtpServer))
}

func formatMessage(body string, alert config.EmailAlert, smtpServer config.SMTPServer) []byte {
	var (
		messageBuffer = &bytes.Buffer{}
		headers       = map[string]string{
			"From":    smtpServer.Sender,
			"To":      alert.FormatRecipients(),
			"Subject": alert.SubjectTemplate,
		}
	)

	for key, value := range headers {
		messageBuffer.WriteString(key)
		messageBuffer.WriteString(": ")
		messageBuffer.WriteString(value)
		messageBuffer.WriteString("\r\n")
	}

	messageBuffer.WriteString("\r\n")
	messageBuffer.WriteString(body)

	return messageBuffer.Bytes()
}

func Send(body string, alert config.EmailAlert, smtpServer config.SMTPServer) error {
	if conn, err := tlsDial(smtpServer); err != nil {
		return err
	} else {
		defer conn.Close()

		if smtpClient, err := smtp.NewClient(conn, smtpServer.Host); err != nil {
			return err
		} else {
			defer smtpClient.Quit()

			auth := smtp.PlainAuth("", smtpServer.Username, smtpServer.Password, smtpServer.Host)
			if err := smtpClient.Auth(auth); err != nil {
				return err
			}

			if err = smtpClient.Mail(smtpServer.Sender); err != nil {
				return err
			}

			for _, recipient := range alert.Recipients {
				if err = smtpClient.Rcpt(recipient); err != nil {
					return err
				}
			}

			if dataWriter, err := smtpClient.Data(); err != nil {
				return err
			} else {
				defer dataWriter.Close()

				if _, err := dataWriter.Write(formatMessage(body, alert, smtpServer)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
