package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/ledongthuc/pdf"
	"github.com/tidwall/gjson"
)

type pdfContent struct {
	Text       string   `json:"text"`
	Pages      []string `json:"pages"`
	TotalPages int      `json:"total_pages"`
}

// NewPDF extracts the plain text from the PDF body and exposes it as a JSON
// document {"text": ..., "pages": [...], "total_pages": ...}, so models can
// address it with regular gjson paths like "text" or "pages.0".
func NewPDF(body []byte, logger logger.Logger) *engineParser[*gjson.Result] {
	content, err := extractPDF(body, logger)
	if err != nil {
		logger.Errorw("unable to extract pdf content", "error", err.Error())
		return NewJson(nil, logger)
	}

	jsonBody, err := json.Marshal(content)
	if err != nil {
		logger.Errorw("unable to marshal pdf content", "error", err.Error())
		return NewJson(nil, logger)
	}

	return NewJson(jsonBody, logger)
}

func extractPDF(body []byte, logger logger.Logger) (content *pdfContent, err error) {
	// the underlying reader panics on some malformed documents
	defer func() {
		if r := recover(); r != nil {
			content = nil
			err = fmt.Errorf("unable to parse pdf: %v", r)
		}
	}()

	reader, err := pdf.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}

	totalPages := reader.NumPage()
	pages := make([]string, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			pages = append(pages, "")
			continue
		}
		text, errPage := page.GetPlainText(nil)
		if errPage != nil {
			logger.Errorw("unable to extract text from pdf page", "page", fmt.Sprintf("%d", i), "error", errPage.Error())
			pages = append(pages, "")
			continue
		}
		pages = append(pages, text)
	}

	return &pdfContent{
		Text:       strings.Join(pages, "\n"),
		Pages:      pages,
		TotalPages: totalPages,
	}, nil
}
