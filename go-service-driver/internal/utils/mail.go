package utils

import (
	"crypto/rand"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/resend/resend-go/v2"
)

func GenerateCode() string {
	const codeLength = 6
	const charset = "0123456789"

	b := make([]byte, codeLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	for i := 0; i < codeLength; i++ {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

func GenerateVerificationCodeExpiry() time.Time {
	return time.Now().Add(15 * time.Minute)
}

func GenerateVerificationEmailHTML(code string) string {
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
            <h2>Email Verification</h2>
            <p>Hello,</p>
            <p>Thank you for using SZTU Trip Planner. Please use the following verification code to complete your registration:</p>
            
            <div class="verification-code">` + code + `</div>
            
            <div class="expiry-notice">
                ⚠️ This code will expire in 15 minutes
            </div>
            
            <div class="instructions">
                <strong>Instructions:</strong>
                <ul style="margin: 10px 0; padding-left: 20px;">
                    <li>Enter this code in the verification field</li>
                    <li>Do not share this code with anyone</li>
                    <li>If you didn't request this code, please ignore this email</li>
                </ul>
            </div>
            
            <p>If you have any questions, please contact our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 SZTU Trip Planner. All rights reserved.</p>
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`
}

func SendVerificationEmail(toEmail, code string) error {
	apiKey := config.GetMailAPIKey()

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{toEmail},
		Subject: "Your SZTU Trip Planner Verification Code",
		Html:    GenerateVerificationEmailHTML(code),
	}

	_, err := client.Emails.Send(params)
	return err
}
