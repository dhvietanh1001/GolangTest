package main

import (
	"03/db"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
)

var chatHistory []map[string]string
var dialogContent string
var translatedWordsJSON string
var promot1st string = "Tạo một hội thoại bằng tiếng Việt, gồm 6 câu, ngắn gọn, đơn giản,\nhỏi đường đi đến hồ Hoàn Kiếm ở Hà nội giữa một Mỹ tên James và\nngười Việt nam tên Lan. Chỉ xuất ra hội thoại không cần giải thích."
var promot2nd string = "Từ hội thoại trên hãy lọc ra danh sách các từ quan trọng,\nbỏ qua danh từ tên riêng cần học. Không cần giải thích xuất\nkết quả ra dạng JSON ."
var promot3rd string = "Dịch từng từ trong danh sách dưới sang tiếng Anh rồi trả JSON\ngồm mảng trong đó mỗi phần tử sẽ gồm từ tiếng Việt và từ\ntiếng Anh tương đương. Không cần giải thích."

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

// Gọi API Groq và xử lý kết quả
func callGroqAPI(history []map[string]string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	client := resty.New()

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model":    "deepseek-r1-distill-llama-70b",
			"messages": history,
		}).
		Post("https://api.groq.com/openai/v1/chat/completions")

	if err != nil {
		return "", err
	}

	var groqResp GroqResponse
	err = json.Unmarshal(resp.Body(), &groqResp)
	if err != nil {
		return "", err
	}

	if len(groqResp.Choices) > 0 {
		return groqResp.Choices[0].Message.Content, nil
	}

	return "No response from AI", nil
}

// Quản lý quy trình gửi tuần tự 3 promot
func executeSequentialPrompts(initialPrompt string) ([]string, error) {
	var allResults []string

	// Step 1: Gửi promot đầu tiên
	chatHistory = append(chatHistory, map[string]string{"role": "user", "content": promot1st})
	result1, err := callGroqAPI(chatHistory)
	if err != nil {
		return nil, err
	}
	cleanResult1 := RemoveThinkTags(result1)
	allResults = append(allResults, cleanResult1)

	// Cập nhật lịch sử chat với kết quả từ promot1
	chatHistory = append(chatHistory, map[string]string{"role": "assistant", "content": cleanResult1})

	// Step 2: Gửi promot tiếp theo với lịch sử
	chatHistory = append(chatHistory, map[string]string{"role": "user", "content": promot2nd})
	result2, err := callGroqAPI(chatHistory)
	if err != nil {
		return nil, err
	}
	cleanResult2 := RemoveThinkTags(result2)
	cleanResult2 = ExtractBracketContent(cleanResult2)
	allResults = append(allResults, cleanResult2)

	// Cập nhật lịch sử chat với kết quả từ promot2
	chatHistory = append(chatHistory, map[string]string{"role": "assistant", "content": cleanResult2})

	// Step 3: Gửi promot cuối cùng với lịch sử
	chatHistory = append(chatHistory, map[string]string{"role": "user", "content": promot3rd})
	result3, err := callGroqAPI(chatHistory)
	if err != nil {
		return nil, err
	}
	cleanResult3 := RemoveThinkTags(result3)
	cleanResult3 = ExtractBracketContent(cleanResult3)
	allResults = append(allResults, cleanResult3)

	// Cập nhật lịch sử chat với kết quả từ promot3
	chatHistory = append(chatHistory, map[string]string{"role": "assistant", "content": cleanResult3})

	return allResults, nil
}

func InsertFullData(dialogContent string, translatedWordsJSON string) error {
	dialogContent = strings.TrimSpace(dialogContent)
	translatedWordsJSON = strings.TrimSpace(translatedWordsJSON)
	// Kết nối cơ sở dữ liệu
	err := db.Connect()
	if err != nil {
		return fmt.Errorf("kết nối thất bại: %v", err)
	}
	defer db.Close()

	// 1. Chèn đoạn hội thoại vào bảng dialog
	dialogID, err := db.InsertDialog("vi", dialogContent) // Mặc định ngôn ngữ là "vi" (tiếng Việt)
	if err != nil {
		return fmt.Errorf("lỗi chèn dữ liệu vào bảng dialog: %v", err)
	}
	fmt.Printf("Inserted dialog with ID: %d\n", dialogID)

	// 2. Giải mã danh sách từ vựng từ JSON
	type WordPair struct {
		Vi string `json:"vi"`
		En string `json:"en"`
	}

	var temp []map[string]string
	err = json.Unmarshal([]byte(translatedWordsJSON), &temp)
	if err != nil {
		return fmt.Errorf("lỗi giải mã JSON từ vựng: %v", err)
	}

	var wordPairs []WordPair
	for _, word := range temp {
		i := 0
		var vi, en string
		for _, value := range word {
			if i == 0 {
				vi = value
			} else if i == 1 {
				en = value
			}
			i++
		}

		if vi == "" || en == "" {
			return fmt.Errorf("lỗi dữ liệu từ vựng không đầy đủ: %+v", word)
		}

		wordPairs = append(wordPairs, WordPair{Vi: vi, En: en})
	}

	fmt.Printf("Parsed word pairs: %+v\n", wordPairs)

	// 3. Chèn từ vựng và tạo liên kết word-dialog
	for _, pair := range wordPairs {
		wordID, err := db.InsertWord("vi", pair.Vi, pair.En) // Mặc định từ tiếng Việt trước, tiếng Anh sau
		if err != nil {
			return fmt.Errorf("lỗi chèn từ '%s' vào bảng word: %v", pair.Vi, err)
		}
		fmt.Printf("Inserted word '%s' with ID: %d\n", pair.Vi, wordID)

		// Tạo liên kết giữa dialog và word trong bảng word_dialog
		err = db.InsertWordDialog(dialogID, wordID)
		if err != nil {
			return fmt.Errorf("lỗi tạo liên kết word-dialog giữa dialogID %d và wordID %d: %v", dialogID, wordID, err)
		}
		fmt.Printf("Linked dialogID %d with wordID %d\n", dialogID, wordID)
	}
	return nil
}

// Loại bỏ nội dung giữa <think> và </think>
func RemoveThinkTags(input string) string {
	re := regexp.MustCompile(`(?s)<think>.*?</think>\n?`)
	return re.ReplaceAllString(input, "")
}

// Lấy nội dung json
func ExtractBracketContent(input string) string {
	re := regexp.MustCompile(`\[[^\]]*\]`)
	match := re.FindString(input)
	if match == "" {
		return input // Không thay đổi nếu không tìm thấy []
	}
	return match
}

func main() {
	// Load API key từ .env
	if err := godotenv.Load(); err != nil {
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

	// Endpoint nhận từng bước của "promot" và gọi AI trả kết quả
	app.Post("/api/chat-step", func(ctx iris.Context) {
		var req map[string]string
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "Invalid request"})
			return
		}

		// Lấy thông tin "step" từ FE
		step := req["step"]

		var prompt string
		switch step {
		case "1":
			prompt = promot1st
		case "2":
			prompt = promot2nd
		case "3":
			prompt = promot3rd
		default:
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "Unhandled step"})
			return
		}

		// Cập nhật lịch sử
		chatHistory = append(chatHistory, map[string]string{"role": "user", "content": prompt})

		// Gửi yêu cầu đến AI
		result, err := callGroqAPI(chatHistory)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "Failed to call Groq API"})
			return
		}

		// Nhận phản hồi từ AI và loại bỏ các thẻ không mong muốn
		cleanResult := RemoveThinkTags(result)
		cleanResult = ExtractBracketContent(cleanResult)
		// Cập nhật lịch sử với phản hồi của AI
		chatHistory = append(chatHistory, map[string]string{"role": "assistant", "content": cleanResult})

		switch step {
		case "1":
			dialogContent = cleanResult
		case "2":
		case "3":
			translatedWordsJSON = cleanResult
			// Gọi hàm chèn dữ liệu với dialogContent là cleanResult1 và translatedWordsJSON là cleanResult3
			err = InsertFullData(dialogContent, translatedWordsJSON)
			if err != nil {
				log.Printf("Lỗi khi chèn dữ liệu: %v\n", err)
			}

			// Xóa toàn bộ dữ liệu
			chatHistory = nil
			dialogContent = ""
			translatedWordsJSON = ""

		default:
			return
		}
		// Trả về cả prompt và kết quả phản hồi từ AI
		ctx.JSON(iris.Map{"prompt": prompt, "response": cleanResult})
	})

	app.Listen(":8080")
}
