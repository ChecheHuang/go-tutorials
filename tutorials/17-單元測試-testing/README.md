# 第十六課：單元測試

## 學習目標

- 學會 Go 內建的測試框架 `testing`
- 掌握表格驅動測試（Table-Driven Tests）
- 學會用 Mock 測試依賴介面的程式碼
- 了解 `t.Run`、`t.Errorf`、`t.Fatalf` 的差異

## 執行方式

```bash
cd tutorials/16-testing
go test -v                   # 詳細輸出
go test -cover               # 顯示覆蓋率
go test -run TestAdd         # 只跑特定測試
go test -run TestAdd/正數相加  # 只跑特定子測試
```

## 重點筆記

### Go 測試的命名規則

| 規則 | 範例 |
|------|------|
| 檔名以 `_test.go` 結尾 | `user_test.go` |
| 函式以 `Test` 開頭 | `TestCreateUser` |
| 參數是 `*testing.T` | `func TestXxx(t *testing.T)` |
| 與被測試檔案同 package | `package main` |

### t.Errorf vs t.Fatalf

```go
t.Errorf("...")  // 記錄錯誤，繼續執行後面的測試
t.Fatalf("...")  // 記錄錯誤，立即停止當前測試
```

### 表格驅動測試的模板

```go
func TestXxx(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {"案例1", input1, expected1, false},
        {"案例2", input2, expected2, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr { ... }
            if result != tt.expected { ... }
        })
    }
}
```

### Mock 的原理

```
正式環境：
  UserService → UserRepository → 真實資料庫

測試環境：
  UserService → mockUserRepository → map（記憶體）
```

因為 UserService 依賴的是**介面**而非具體實作，所以可以輕鬆替換。

### 在專案中的對應

`internal/usecase/user_usecase_test.go` 和 `internal/handler/user_handler_test.go` 就是用這些技巧寫成的。

## 練習

1. 為 `Divide` 函式新增更多表格驅動測試案例
2. 讓 `mockUserRepository.Create` 模擬失敗的情況，測試 Service 的錯誤處理
3. 執行 `go test -cover` 查看覆蓋率，嘗試提高到 90% 以上
