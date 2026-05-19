package ocrapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type OcrService struct {
	BaseURL string
	APIKey  string
}

func NewService(baseUrl, apiKey string) *OcrService {
	return &OcrService{
		BaseURL: baseUrl,
		APIKey:  apiKey,
	}
}

func (s *OcrService) pushFile(filename string) (*OCRResult, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	sniffLen := 512
	if len(data) < sniffLen {
		sniffLen = len(data)
	}
	mimeType := http.DetectContentType(data[:sniffLen])

	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowed[mimeType] {
		return nil, fmt.Errorf("unsupported file type: %s", mimeType)
	}

	req, err := http.NewRequest(
		"POST",
		s.BaseURL,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("Token", s.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result OCRResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *OcrService) getOutputText(uuid string) (*OCRTranscript, error) {
	req, _ := http.NewRequest(
		"GET",
		"https://api.chimege.com/v1.2/court/ocr-transcript",
		nil,
	)

	req.Header.Set("Token", s.APIKey)
	req.Header.Set("UUID", uuid)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("sorry unknown error occurred")
	}

	data := new(OCRTranscript)
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *OcrService) GetTextFromFile(filename string) (string, error) {
	result, err := s.pushFile(filename)
	if err != nil {
		return "", err
	}

	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for OCR result")
		case <-ticker.C:
			output, err := s.getOutputText(result.UUID)
			if err != nil {
				return "", err
			}

			if output.Done {
				var sb strings.Builder
				for _, ocrPage := range output.Result.OCR {
					sb.WriteString(ocrPage.Text)
					sb.WriteString("\n")
				}
				return sb.String(), nil
			}
		}
	}
}
