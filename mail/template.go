package mail

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path/filepath"
	texttemplate "text/template"

	"github.com/zenvisjr/building-scalable-microservices/logger"
)

func RenderTemplates(name string, data any) (string, string, error) {
	Logs := logger.GetGlobalLogger()
	ctx := context.Background()
	basePath, err := os.Getwd() // gets the working directory at runtime
	if err != nil {
		return "", "", fmt.Errorf("failed to get working dir: %w", err)
	}

	htmlPath := filepath.Join(basePath, "mail", "templates", name+".html")
	textPath := filepath.Join(basePath, "mail", "templates", name+".txt")

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		Logs.Error(ctx, "HTML template not found at path: "+htmlPath)
	}
	if _, err := os.Stat(textPath); os.IsNotExist(err) {
		Logs.Error(ctx, "Text template not found at path: "+textPath)
	}

	// Parse HTML template using html/template
	htmlTpl, err := htmltemplate.ParseFiles(htmlPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Parse plain text template using text/template
	textTpl, err := texttemplate.ParseFiles(textPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse text template: %w", err)
	}

	var htmlBuf, textBuf bytes.Buffer

	// Render HTML
	if err := htmlTpl.Execute(&htmlBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Render Text
	if err := textTpl.Execute(&textBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to render text template: %w", err)
	}

	return textBuf.String(), htmlBuf.String(), nil
}
