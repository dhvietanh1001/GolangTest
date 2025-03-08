# Giới Thiệu

Dự án này bao gồm ba phần chính:

1. **Gọi vào Groq API**: Ứng dụng web cho phép người dùng nhập prompt và nhận kết quả từ Groq API, hiển thị kết quả dưới dạng Markdown.
2. **Sinh file SSML từ hội thoại**: Trang web cho phép chọn voice và tạo file SSML từ hội thoại gốc.
3. **Tạo hội thoại từ prompt và trích xuất từ mới**: Ứng dụng web tự động hóa việc tạo hội thoại, trích xuất từ vựng, và lưu vào cơ sở dữ liệu PostgreSQL.

---

## Cài Đặt

Để cài đặt dự án, bạn cần có các công cụ sau:

- **Golang** 
- **PostgreSQL**

---

## Cách Chạy

### Đối với Bài 01 và 03

1. **Cấu hình cơ sở dữ liệu**:
   - Vào file `.env` và điền các thông tin kết nối PostgreSQL.
   - Sửa API key nếu không chạy được.

2. **Khởi chạy ứng dụng**:
   - Sử dụng IDE mở project lên và chạy.
   - Hoặc dùng lệnh:
     ```bash
     go run main.go
     ```

3. **Truy cập ứng dụng**:
   - Mở trình duyệt và truy cập: [http://localhost:8080](http://localhost:8080).

### Đối với Bài 02

- Mở thư mục chứa bài 02 và double-click vào file `index.html`.

---

## Các Bước Thực Hiện

### 01. Call Groq API

- Người dùng nhập prompt vào text area.
- Ứng dụng gọi Groq API và hiển thị kết quả.

![Ảnh minh họa](https://github.com/user-attachments/assets/d6c54ce1-160f-4413-8fd6-194c49db5b92)

---

### 02. Sinh file SSML từ hội thoại

- Chọn hai voice từ danh sách.
- Nhập hội thoại và tạo file SSML.

![Ảnh minh họa](https://github.com/user-attachments/assets/ffa4144a-6086-418f-81c9-0885024c96a5)

---

### 03. Tạo hội thoại từ prompt và trích xuất từ mới

- Chạy ứng dụng lên và nhấn vào nút **"Bắt đầu"**.
- Hệ thống sẽ tự động nhập các câu prompt và hiển thị kết quả lên màn hình.
- Kết quả sẽ được lưu vào cơ sở dữ liệu.



https://github.com/user-attachments/assets/b9f3ef5f-e038-4b75-b8f1-bf73dbc85a25


---


