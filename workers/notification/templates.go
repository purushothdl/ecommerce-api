// workers/notification/templates.go
package notification

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/purushothdl/ecommerce-api/events"
)

//go:embed all:templates/*.gohtml
var templateFS embed.FS

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

func NewTemplateService() (*TemplateService, error) {
	// Create a new template and register the custom functions first.
	// Then, parse the embedded files into this template.
	tmpls, err := template.New("").Funcs(templateFuncs).ParseFS(templateFS, "templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates from embedded FS: %w", err)
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
	subject = fmt.Sprintf("Your GoKart Order #%s is Being Processed", payload.OrderNumber)
	body, err = s.execute("order_packed.gohtml", payload)
	return
}

func (s *TemplateService) GenerateOrderShippedEmail(payload events.OrderShippedEvent) (subject string, body string, err error) {
	subject = fmt.Sprintf("Your GoKart Order #%s Has Shipped!", payload.OrderNumber)
	body, err = s.execute("order_shipped.gohtml", payload)
	return
}

func (s *TemplateService) GenerateOrderDeliveredEmail(payload events.OrderDeliveredEvent) (subject string, body string, err error) {
	subject = fmt.Sprintf("Your GoKart Order #%s Has Been Delivered!", payload.OrderNumber)
	body, err = s.execute("order_delivered.gohtml", payload)
	return
}