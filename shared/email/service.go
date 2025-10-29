package email

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	IsSSL    bool     `json:"is_ssl"`
}

// EmailService 邮件服务
type EmailService struct {
	config *EmailConfig
}

// NewEmailService 创建邮件服务
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{config: config}
}

// SendVerificationEmail 发送验证邮件
func (es *EmailService) SendVerificationEmail(to, verificationCode string) error {
	subject := "邮箱验证码"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">邮箱验证</h2>
			<p>您好！</p>
			<p>您的验证码是：<strong style="color: #007bff; font-size: 24px;">%s</strong></p>
			<p>验证码有效期为10分钟，请及时使用。</p>
			<p>如果这不是您的操作，请忽略此邮件。</p>
			<hr style="margin: 20px 0; border: none; border-top: 1px solid #eee;">
			<p style="color: #666; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
		</div>
	`, verificationCode)

	return es.sendEmail(to, subject, body)
}

// SendWelcomeEmail 发送欢迎邮件
func (es *EmailService) SendWelcomeEmail(to, username string) error {
	subject := "欢迎注册博客系统"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">欢迎 %s！</h2>
			<p>恭喜您成功注册我们的博客系统！</p>
			<p>现在您可以：</p>
			<ul>
				<li>发表评论和互动</li>
				<li>使用钱包功能</li>
				<li>享受更多服务</li>
			</ul>
			<p>感谢您的支持！</p>
			<hr style="margin: 20px 0; border: none; border-top: 1px solid #eee;">
			<p style="color: #666; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
		</div>
	`, username)

	return es.sendEmail(to, subject, body)
}

// sendEmail 发送邮件
func (es *EmailService) sendEmail(to, subject, body string) error {
	e := email.NewEmail()
	e.From = es.config.From
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(body)

	var auth smtp.Auth
	if es.config.IsSSL {
		auth = smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.Host)
	}

	addr := fmt.Sprintf("%s:%d", es.config.Host, es.config.Port)

	if es.config.IsSSL {
		err := e.SendWithTLS(addr, auth, nil)
		if err != nil {
			return fmt.Errorf("failed to send email with TLS: %v", err)
		}
	} else {
		err := e.Send(addr, auth)
		if err != nil {
			return fmt.Errorf("failed to send email: %v", err)
		}
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

// SendPaymentNotification 发送支付通知邮件
func (es *EmailService) SendPaymentNotification(to, amount, description string) error {
	subject := "支付通知"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">支付成功</h2>
			<p>您的支付已完成！</p>
			<p><strong>支付金额：</strong>¥%s</p>
			<p><strong>支付说明：</strong>%s</p>
			<p>感谢您的使用！</p>
			<hr style="margin: 20px 0; border: none; border-top: 1px solid #eee;">
			<p style="color: #666; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
		</div>
	`, amount, description)

	return es.sendEmail(to, subject, body)
}
