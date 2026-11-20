package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RecentEmailsInput struct {
	Query string `json:"query" jsonschema:"Gmail search query, contoh in:inbox is:unread"`
	Limit int64  `json:"limit" jsonschema:"Jumlah email maksimal"`
}

type RecentEmailsOutput struct {
	Emails []Email `json:"emails"`
	Count  int     `json:"count"`
}

type SearchEmailsInput struct {
	Query string `json:"query" jsonschema:"Query pencarian Gmail"`
	Limit int64  `json:"limit" jsonschema:"Jumlah maksimal hasil"`
}

type ReadEmailInput struct {
	MessageID string `json:"message_id" jsonschema:"ID email Gmail yang ingin dibaca"`
}

type ReadEmailOutput struct {
	Email *Email `json:"email"`
	Body  string `json:"body"`
}

func main() {
	service, err := getGmailService()
	if err != nil {
		log.Fatal(err)
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "gmail-assistant",
			Version: "1.0.0",
		},
		nil,
	)

	// Tool 1: Email terbaru
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_recent_emails",
			Description: "Mengambil email terbaru dari Gmail. " +
				"Gunakan untuk menjawab pertanyaan tentang email baru, " +
				"email belum dibaca, atau email terbaru.",
		},
		func(
			ctx context.Context,
			req *mcp.CallToolRequest,
			input RecentEmailsInput,
		) (*mcp.CallToolResult, RecentEmailsOutput, error) {
			emails, err := getRecentEmails(
				service,
				input.Query,
				input.Limit,
			)

			if err != nil {
				return nil, RecentEmailsOutput{}, err
			}

			return nil, RecentEmailsOutput{
				Emails: emails,
				Count:  len(emails),
			}, nil
		},
	)

	// Tool 2: Search email
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "search_emails",
			Description: "Mencari email Gmail berdasarkan query. " +
				"Contoh from:email@gmail.com, subject:bimbingan, " +
				"has:attachment, newer_than:7d.",
		},
		func(
			ctx context.Context,
			req *mcp.CallToolRequest,
			input SearchEmailsInput,
		) (*mcp.CallToolResult, RecentEmailsOutput, error) {
			emails, err := getRecentEmails(
				service,
				input.Query,
				input.Limit,
			)

			if err != nil {
				return nil, RecentEmailsOutput{}, err
			}

			return nil, RecentEmailsOutput{
				Emails: emails,
				Count:  len(emails),
			}, nil
		},
	)

	// Tool 3: Read email
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "read_email",
			Description: "Membaca detail dan isi email berdasarkan " +
				"message ID. Gunakan setelah menemukan email yang relevan.",
		},
		func(
			ctx context.Context,
			req *mcp.CallToolRequest,
			input ReadEmailInput,
		) (*mcp.CallToolResult, ReadEmailOutput, error) {
			email, body, err := getEmailDetail(
				service,
				input.MessageID,
			)

			if err != nil {
				return nil, ReadEmailOutput{}, err
			}

			return nil, ReadEmailOutput{
				Email: email,
				Body:  body,
			}, nil
		},
	)

	log.Println("Gmail MCP Server berjalan")

	if err := server.Run(
		context.Background(),
		&mcp.StdioTransport{},
	); err != nil {
		log.Fatal(err)
	}
}
