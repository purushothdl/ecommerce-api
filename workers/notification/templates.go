// workers/notification/templates.go
package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/purushothdl/ecommerce-api/events"
)

// TemplateService manages parsing and executing HTML email templates.
type TemplateService struct {
	templates *template.Template
}

// Custom template functions to format data nicely inside the HTML.
var templateFuncs = template.FuncMap{
	"formatAsMoney": func(amount float64) string {
		return fmt.Sprintf("₹%.2f", amount)
	},
	"formatAsDate": func(t time.Time) string {
		return t.Format("02 Jan 2006")
	},
}

func NewTemplateService(templatesDir string) (*TemplateService, error) {

	tmpls, err := template.New("").Funcs(templateFuncs).ParseGlob(filepath.Join(templatesDir, "*.gohtml"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates from %s: %w", templatesDir, err)
	}

	return &TemplateService{
		templates: tmpls,
	}, nil
}

// execute is a private helper to render a specific template with given data.
func (s *TemplateService) execute(templateName string, data any) (string, error) {
	var buf bytes.Buffer
	// Execute the specific template by name (e.g., "order_confirmed.gohtml").
	err := s.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}
	return buf.String(), nil
}

// --- Specific Email Generation Methods ---

func (s *TemplateService) GenerateOrderConfirmedEmail(payload events.OrderCreatedEvent) (subject string, body string, err error) {
	subject = fmt.Sprintf("Your GoKart Order #%s is Confirmed!", payload.OrderNumber)
	body, err = s.execute("order_confirmed.gohtml", payload)
	return
}

func (s *TemplateService) GenerateOrderPackedEmail(payload events.OrderPackedEvent) (subject string, body string, err error) {
	subject = fmt.Sprintf("Your GoKart Order #%d is Being Processed", payload.OrderID)
	body, err = s.execute("order_packed.gohtml", payload)
	return
}

func (s *TemplateService) GenerateOrderShippedEmail(payload events.OrderShippedEvent) (subject string, body string, err error) {
	subject = fmt.Sprintf("Your GoKart Order #%d Has Shipped!", payload.OrderID)
	body, err = s.execute("order_shipped.gohtml", payload)
	return
}

func (s *TemplateService) GenerateOrderDeliveredEmail(payload events.OrderDeliveredEvent) (subject string, body string, err error) {
	subject = fmt.Sprintf("Your GoKart Order #%d Has Been Delivered!", payload.OrderID)
	body, err = s.execute("order_delivered.gohtml", payload)
	return
}