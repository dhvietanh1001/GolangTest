package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"os"
)

var Conn *pgx.Conn

func Connect() error {
	// Load biến môi trường từ file .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Không thể load file .env, kiểm tra lại.")
	}

	// Lấy thông tin kết nối từ biến môi trường
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL không được thiết lập")
	}

	// Kết nối đến PostgreSQL
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("không thể kết nối đến PostgreSQL: %w", err)
	}

	Conn = conn
	fmt.Println("Kết nối PostgreSQL thành công!")
	return nil
}

func Close() {
	if Conn != nil {
		if err := Conn.Close(context.Background()); err != nil {
			fmt.Println("Lỗi khi đóng kết nối PostgreSQL:", err)
		} else {
			fmt.Println("Đã đóng kết nối PostgreSQL")
		}
	}
}

// InsertDialog lưu hội thoại vào bảng dialog và trả về ID của hội thoại vừa chèn
func InsertDialog(lang string, content string) (int64, error) {
	if Conn == nil {
		return 0, fmt.Errorf("kết nối chưa được thiết lập")
	}

	query := `INSERT INTO dialog (lang, content) VALUES ($1, $2) RETURNING id`
	var id int64
	err := Conn.QueryRow(context.Background(), query, lang, content).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("lỗi khi chèn dữ liệu: %w", err)
	}
	return id, nil
}

// InsertWord lưu từ vựng vào bảng word và trả về ID của từ vừa chèn
func InsertWord(lang string, content string, translate string) (int64, error) {
	if Conn == nil {
		return 0, fmt.Errorf("kết nối chưa được thiết lập")
	}

	query := `INSERT INTO word (lang, content, translate) VALUES ($1, $2, $3) RETURNING id`
	var id int64
	err := Conn.QueryRow(context.Background(), query, lang, content, translate).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("lỗi khi chèn từ vựng: %w", err)
	}
	return id, nil
}

type TranslatedWord struct {
	Vi string `json:"vi"`
	En string `json:"en"`
}

type TranslatedWords struct {
	Words []TranslatedWord `json:"translated_words"`
}

// InsertWordDialog liên kết từ vựng với hội thoại trong bảng word_dialog
func InsertWordDialog(dialogID int64, wordID int64) error {
	query := `INSERT INTO word_dialog (dialog_id, word_id) VALUES ($1, $2)`
	_, err := Conn.Exec(context.Background(), query, dialogID, wordID)
	if err != nil {
		return fmt.Errorf("lỗi khi tạo liên kết word-dialog: %w", err)
	}
	return nil
}
