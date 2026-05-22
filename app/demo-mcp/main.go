package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type greetArgs struct {
	Name string `json:"name"`
}

type echoArgs struct {
	Message string `json:"message"`
	Upper   bool   `json:"upper"`
}

func main() {
	addr := env("ADDR", ":3003")

	s := server.NewMCPServer(
		"mcpflow-go-demo",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	s.AddTool(
		mcp.NewTool("greet",
			mcp.WithDescription("Generate a short greeting for a person."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the person to greet."),
			),
		),
		mcp.NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args greetArgs) (*mcp.CallToolResult, error) {
			name := strings.TrimSpace(args.Name)
			if name == "" {
				return mcp.NewToolResultError("name is required"), nil
			}
			return mcp.NewToolResultText("Hi " + name + ", this response came from a Go MCP server."), nil
		}),
	)

	s.AddTool(
		mcp.NewTool("echo",
			mcp.WithDescription("Echo a message, optionally converting it to uppercase."),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Message to echo."),
			),
			mcp.WithBoolean("upper",
				mcp.Description("Whether to convert the message to uppercase."),
				mcp.DefaultBool(false),
			),
		),
		mcp.NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args echoArgs) (*mcp.CallToolResult, error) {
			message := args.Message
			if args.Upper {
				message = strings.ToUpper(message)
			}
			return mcp.NewToolResultText(message), nil
		}),
	)

	s.AddTool(
		mcp.NewTool("now",
			mcp.WithDescription("Return the current server time in RFC3339 format."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result := map[string]string{
				"time":     time.Now().Format(time.RFC3339),
				"timezone": time.Now().Location().String(),
			}
			return mcp.NewToolResultJSON(result)
		},
	)

	httpServer := server.NewStreamableHTTPServer(s)

	log.Printf("Go MCP demo server listening on %s", addr)
	log.Printf("MCP endpoint: http://localhost:%s/mcp", strings.TrimPrefix(addr, ":"))
	log.Printf("Available tools: greet, echo, now")

	if err := httpServer.Start(addr); err != nil {
		log.Fatal(fmt.Errorf("start mcp server: %w", err))
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
