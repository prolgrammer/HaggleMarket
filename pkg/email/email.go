package email

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EM-Stawberry/Stawberry/config"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

//go:generate go.uber.org/mock/mockgen -source=$GOFILE -destination=mock_email/mock_email.go -package=mock_email

type MailerService interface {
	Registered(userName string, userMail string)
	StatusUpdate(offerID uint, status string, userMail string)
	OfferReceived(offerID uint, userMail string)
	Stop(ctx context.Context)
	SendGuestOfferNotification(email string, subject string, body string)
}

type SMTPMailer struct {
	enabled bool
	ctx     context.Context
	ctxCanc context.CancelFunc
	dialer  *gomail.Dialer
	queue   chan *gomail.Message
	wg      sync.WaitGroup
	mutex   sync.Mutex
	stopped bool
	log     *zap.Logger
}

func NewMailer(log *zap.Logger, emailCfg *config.EmailConfig) MailerService {
	ctx, cancel := context.WithCancel(context.Background())
	m := &SMTPMailer{
		ctx:     ctx,
		ctxCanc: cancel,
		log:     log,
	}

	if !emailCfg.Enabled {
		m.log.Info("email notifications are disabled")
		return m
	}

	m.enabled = true

	m.dialer = gomail.NewDialer(emailCfg.SMTPHost, emailCfg.SMTPPort, emailCfg.From, emailCfg.Password)

	m.log.Info("creating email queue", zap.Int("size", emailCfg.QueueSize))
	m.queue = make(chan *gomail.Message, emailCfg.QueueSize)

	m.log.Info("starting mailer workers", zap.Int("pool size", emailCfg.WorkerPool))
	m.wg.Add(emailCfg.WorkerPool)
	for range emailCfg.WorkerPool {
		go m.worker()
	}

	m.log.Info("email notifications are enabled")

	return m
}

func (m *SMTPMailer) Stop(ctx context.Context) {
	if !m.enabled {
		m.log.Info("mailer stop called, but email is disabled")
		return
	}
	m.log.Info("mailer is stopping")

	m.ctxCanc()
	m.mutex.Lock()
	m.stopped = true
	close(m.queue)
	m.mutex.Unlock()

	workersDone := make(chan struct{}, 1)
	go func() {
		m.wg.Wait()
		workersDone <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		m.log.Info("mailer workers forcefully stopped (timeout), messages that remained in queue are lost")
	case <-workersDone:
		m.log.Info("email queue emptied, mailer workers stopped")
	}
}

func (m *SMTPMailer) sendEmailWithRetry(msg *gomail.Message) error {
	const maxSendRetries = 3
	const sendRetryDelay = time.Second

	for i := 0; i < maxSendRetries; i++ {
		if err := m.dialer.DialAndSend(msg); err != nil {
			m.log.Error("failed to send email, retrying...",
				zap.String("subject", msg.GetHeader("Subject")[0]),
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", maxSendRetries),
				zap.Error(err))
			time.Sleep(sendRetryDelay)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to send email after %d attempts for subject: %s",
		maxSendRetries, msg.GetHeader("Subject")[0])
}

func (m *SMTPMailer) worker() {
	defer m.wg.Done()
	for {
		select {
		case <-m.ctx.Done():
			for msg := range m.queue {
				if err := m.sendEmailWithRetry(msg); err != nil {
					m.log.Error("failed to send email during shutdown after retries", zap.Error(err))
				}
			}
			return

		case msg, ok := <-m.queue:
			if !ok {
				return
			}
			if err := m.sendEmailWithRetry(msg); err != nil {
				m.log.Error("failed to send email after retries", zap.Error(err))
			}
		}
	}
}

func (m *SMTPMailer) enqueue(msg *gomail.Message) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.stopped {
		m.log.Info("failed to enqueue email, mailer is stopped",
			zap.String("subject", msg.GetHeader("Subject")[0]))
		return
	}

	const maxRetries = 5
	const retryDelay = 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		select {
		case m.queue <- msg:
			return
		default:
			if i < maxRetries-1 {
				m.log.Warn("email queue is full, retrying...",
					zap.String("subject", msg.GetHeader("Subject")[0]),
					zap.Int("attempt", i+1),
					zap.Int("max_attempts", maxRetries))
				time.Sleep(retryDelay)
			}
		}
	}
}

func (m *SMTPMailer) StatusUpdate(offerID uint, status string, userMail string) {
	if !m.enabled {
		return
	}

	subject := fmt.Sprintf("Stawberry: Offer Status Update (ID %d)", offerID)
	body := fmt.Sprintf("The status of your offer (%d) has been changed to: %s", offerID, status)
	msg := m.createMessage(userMail, subject, body)

	m.enqueue(msg)
}

func (m *SMTPMailer) OfferReceived(offerID uint, userMail string) {
	if !m.enabled {
		return
	}

	subject := fmt.Sprintf("Stawberry: New Offer Received (ID %d)", offerID)
	body := fmt.Sprintf("A new offer (%d) has been received", offerID)
	msg := m.createMessage(userMail, subject, body)

	m.enqueue(msg)
}

func (m *SMTPMailer) Registered(userName string, userMail string) {
	if !m.enabled {
		return
	}

	subject := "Welcome to Strawberry!"
	body := fmt.Sprintf("Thank you for registering, %s.", userName)
	msg := m.createMessage(userMail, subject, body)

	m.enqueue(msg)
}

func (m *SMTPMailer) createMessage(to, subject, body string) *gomail.Message {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.dialer.Username)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)
	return msg
}

func (m *SMTPMailer) SendGuestOfferNotification(email string, subject string, body string) {
	if !m.enabled {
		return
	}

	msg := m.createMessage(email, subject, body)
	m.enqueue(msg)
}
