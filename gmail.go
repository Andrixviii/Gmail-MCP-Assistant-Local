package main

import (
	"encoding/base64"
	"strings"

	"google.golang.org/api/gmail/v1"
)

type Email struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Snippet string `json:"snippet"`
	Unread  bool   `json:"unread"`
}

func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, header := range headers {
		if strings.EqualFold(header.Name, name) {
			return header.Value
		}
	}

	return ""
}

func getRecentEmails(
	service *gmail.Service,
	query string,
	limit int64,
) ([]Email, error) {
	if query == "" {
		query = "in:inbox newer_than:1d"
	}

	if limit <= 0 || limit > 20 {
		limit = 10
	}

	result, err := service.Users.Messages.List("me").
		Q(query).
		MaxResults(limit).
		Do()

	if err != nil {
		return nil, err
	}

	var emails []Email

	for _, message := range result.Messages {
		detail, err := service.Users.Messages.Get("me", message.Id).
			Format("metadata").
			MetadataHeaders("From", "To", "Subject", "Date").
			Do()

		if err != nil {
			return nil, err
		}

		email := Email{
			ID:      detail.Id,
			From:    getHeader(detail.Payload.Headers, "From"),
			To:      getHeader(detail.Payload.Headers, "To"),
			Subject: getHeader(detail.Payload.Headers, "Subject"),
			Date:    getHeader(detail.Payload.Headers, "Date"),
			Snippet: detail.Snippet,
			Unread:  contains(detail.LabelIds, "UNREAD"),
		}

		emails = append(emails, email)
	}

	return emails, nil
}

func getEmailDetail(
	service *gmail.Service,
	messageID string,
) (*Email, string, error) {
	message, err := service.Users.Messages.Get("me", messageID).
		Format("full").
		Do()

	if err != nil {
		return nil, "", err
	}

	email := &Email{
		ID:      message.Id,
		From:    getHeader(message.Payload.Headers, "From"),
		To:      getHeader(message.Payload.Headers, "To"),
		Subject: getHeader(message.Payload.Headers, "Subject"),
		Date:    getHeader(message.Payload.Headers, "Date"),
		Snippet: message.Snippet,
		Unread:  contains(message.LabelIds, "UNREAD"),
	}

	body := extractBody(message.Payload)

	return email, body, nil
}

func extractBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}

	if payload.MimeType == "text/plain" && payload.Body != nil {
		data, err := base64.RawURLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	for _, part := range payload.Parts {
		body := extractBody(part)

		if body != "" {
			return body
		}
	}

	return ""
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
