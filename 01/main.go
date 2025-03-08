package main

import (
	"encoding/json"
	"html/template"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/gomarkdown/markdown"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
)

// Định nghĩa struct để parse JSON response từ API Groq
type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Request struct {
	Prompt string `json:"prompt"`
}

func callGroqAPI(prompt string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	client := resty.New()

	// Gọi API Groq
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model": "gemma2-9b-it", //chon model
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
		}).
		Post("https://api.groq.com/openai/v1/chat/completions")

	if err != nil {
		return "", err
	}

	// Parse JSON để lấy phần "content"
	var groqResp GroqResponse
	err = json.Unmarshal(resp.Body(), &groqResp)
	if err != nil {
		return "", err
	}

	// Kiểm tra nếu có phản hồi hợp lệ
	if len(groqResp.Choices) > 0 {
		return groqResp.Choices[0].Message.Content, nil
	}

	return "No response from AI", nil
}

func main() {
	// Load API key từ .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := iris.New()
	app.Logger().SetLevel("debug")

	tmpl := iris.HTML("./views", ".html").Reload(true)
	app.RegisterView(tmpl)

	app.Use(iris.Compression)

	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	app.Post("/api/chat", func(ctx iris.Context) {
		var req Request
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "Invalid request"})
			return
		}

		response, err := callGroqAPI(req.Prompt)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "Failed to call Groq API"})
			return
		}

		// Convert Markdown thành HTML
		htmlResponse := template.HTML(markdown.ToHTML([]byte(response), nil, nil))
		ctx.JSON(iris.Map{"response": htmlResponse})
	})

	app.Listen(":8080")
}
