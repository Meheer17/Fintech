package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESClient interface {
	SendEmail(ctx context.Context, to string, subject string, bodyHtml string, bodyText string) error
}

type RealSESClient struct {
	client *ses.Client
}

func NewRealSESClient(cfg aws.Config) *RealSESClient {
	return &RealSESClient{
		client: ses.NewFromConfig(cfg),
	}
}

func (s *RealSESClient) SendEmail(ctx context.Context, to string, subject string, bodyHtml string, bodyText string) error {
	log.Printf("AWS SES: Sending email to %s with subject: %s", to, subject)
	
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(bodyHtml),
				},
				Text: &types.Content{
					Data: aws.String(bodyText),
				},
			},
			Subject: &types.Content{
				Data: aws.String(subject),
			},
		},
		Source: aws.String("no-reply@revenueiqdynamics.com"),
	}

	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email via AWS SES: %w", err)
	}
	log.Printf("AWS SES: Email sent successfully to %s", to)
	return nil
}

type MockSESClient struct{}

func NewMockSESClient() *MockSESClient {
	return &MockSESClient{}
}

func (s *MockSESClient) SendEmail(ctx context.Context, to string, subject string, bodyHtml string, bodyText string) error {
	log.Printf("[MOCK SES] Mock-sending email to %s", to)
	log.Printf("[MOCK SES] Subject: %s", subject)
	log.Printf("[MOCK SES] Body (HTML): %s", bodyHtml)
	log.Printf("[MOCK SES] Body (Text): %s", bodyText)
	return nil
}
