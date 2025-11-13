package utils

import (
	"crypto/rand"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/resend/resend-go/v2"
)

// EmailService handles email operations
type EmailService struct {
	client *resend.Client
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	apiKey := config.GetMailAPIKey()
	return &EmailService{
		client: resend.NewClient(apiKey),
	}
}

// EmailType represents different types of emails
type EmailType int

const (
	VerificationEmail EmailType = iota
	PasswordResetEmail
)

// GenerateVerificationCode generates a 6-digit verification code
func GenerateVerificationCode() string {
	const codeLength = 6
	const charset = "0123456789"

	b := make([]byte, codeLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	for i := range codeLength {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

// SendEmail sends an email based on the email type
func (e *EmailService) SendEmail(toEmail, code string, emailType EmailType) error {
	var subject, html string

	switch emailType {
	case VerificationEmail:
		subject = "验证码 - SZTU Trip Planner"
		html = generateVerificationEmailHTML(code)
	case PasswordResetEmail:
		subject = "密码重置请求 - SZTU Trip Planner"
		html = generatePasswordResetEmailHTML(code)
	}

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{toEmail},
		Subject: subject,
		Html:    html,
	}

	_, err := e.client.Emails.Send(params)
	return err
}

// generateEmailHTML creates the base HTML template
func generateEmailHTML(code, title, description string) string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f6f9fc;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .verification-code {
            background: #f8f9fa;
            border: 2px dashed #dee2e6;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            margin: 25px 0;
            font-size: 32px;
            font-weight: bold;
            color: #495057;
            letter-spacing: 8px;
            font-family: 'Courier New', monospace;
        }
        .instructions {
            background: #e7f3ff;
            border-left: 4px solid #1890ff;
            padding: 15px 20px;
            margin: 25px 0;
            border-radius: 4px;
        }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #6c757d;
            font-size: 14px;
            border-top: 1px solid #dee2e6;
        }
        .expiry-notice {
            color: #e74c3c;
            font-weight: 600;
            text-align: center;
            margin: 20px 0;
        }
        @media only screen and (max-width: 600px) {
            .container {
                margin: 10px;
                border-radius: 8px;
            }
            .content {
                padding: 30px 20px;
            }
            .verification-code {
                font-size: 24px;
                letter-spacing: 6px;
                padding: 15px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>SZTU Trip Planner</h1>
        </div>
        <div class="content">
            <h2>` + title + `</h2>
            <p>您好，</p>
            <p>` + description + `</p>
            
            <div class="verification-code">` + code + `</div>
            
            <div class="expiry-notice">
                ⚠️ 此验证码将在15分钟后过期
            </div>
            
            <div class="instructions">
                <strong>使用说明：</strong>
                <ul style="margin: 10px 0; padding-left: 20px;">
                    <li>在验证字段中输入此验证码</li>
                    <li>不要与任何人分享此验证码</li>
                    <li>如果您没有请求此验证码，请忽略此邮件</li>
                </ul>
            </div>
            
            <p>如果您有任何问题，请联系客户支持团队。</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 SZTU Trip Planner. All rights reserved.</p>
            <p>这是一封自动发送的邮件，请勿回复。</p>
        </div>
    </div>
</body>
</html>`
}

// generateVerificationEmailHTML generates HTML for email verification
func generateVerificationEmailHTML(code string) string {
	return generateEmailHTML(code, "邮箱验证", "感谢您使用SZTU Trip Planner。请使用以下验证码完成您的注册：")
}

// generatePasswordResetEmailHTML generates HTML for password reset
func generatePasswordResetEmailHTML(code string) string {
	return generateEmailHTML(code, "密码重置请求", "我们收到了您的密码重置请求。请使用以下验证码来重置您的密码：")
}
