---
name: go-chapter-quiz
description: Go 語言教學章節互動測驗工具。針對 go-tutorials 專案的 42 堂課，透過問答形式測試使用者對特定章節的掌握程度，答錯時立即講解直到完全理解。使用時機：使用者說「測驗」、「quiz」、「考考我」、「我想測試第 X 章」、「幫我出題」，或任何要測試 Go 教學章節知識的場景。
---

# Go Chapter Quiz

針對 `tutorials/` 下的 42 堂 Go 課程進行互動式知識測驗。

## 工具準備

使用前先載入 AskUserQuestion：
```
ToolSearch: select:AskUserQuestion
```

## 測驗流程

### 第一步：選擇模式

```
AskUserQuestion({
  questions: [
    {
      question: "請選擇測驗模式：",
      header: "測驗模式",
      multiSelect: false,
      options: [
        { label: "學習模式（推薦）", description: "答錯立即講解，確認理解後才繼續下一題，共 5 題" },
        { label: "考試模式", description: "全部 5 題做完後再統一講解錯誤，模擬實際考試" },
        { label: "快速測驗", description: "只出 3 題，快速確認核心概念" }
      ]
    }
  ]
})
```

### 第二步：選擇章節

先以大類別縮小範圍（使用者也可直接透過 Other 輸入章節編號如「07」跳過此步）：

```
AskUserQuestion({
  questions: [
    {
      question: "要測驗哪個類別的章節？也可直接輸入章節編號（如「07」）：",
      header: "選擇類別",
      multiSelect: false,
      options: [
        { label: "基礎語法 (01–09)", description: "變數、流程控制、函式、struct、指標、介面、錯誤、套件、slice/map" },
        { label: "Web 開發 (10–18)", description: "整潔架構、HTTP、Gin、JSON、GORM、PostgreSQL、錯誤包裝、Middleware、JWT" },
        { label: "工程品質與進階 (19–29)", description: "測試、設定、日誌、benchmark、Docker、泛型、goroutine、Redis、WebSocket、進階 DB、進階架構" },
        { label: "基礎設施與高可用 (30–42)", description: "gRPC、Wire、訊息佇列、CQRS、CI/CD、Prometheus、pprof、OpenTelemetry、K8s、熔斷器、高可用、分散式一致性" }
      ]
    }
  ]
})
```

若使用者選擇類別，再顯示該類別下的具體章節（每次最多 4 個，剩餘讓使用者用 Other 輸入編號）：

```
AskUserQuestion({
  questions: [
    {
      question: "基礎語法系列，請選擇章節（或 Other 輸入其他編號）：",
      header: "選擇章節",
      multiSelect: false,
      options: [
        { label: "01 變數與型別", description: "var vs :=、zero value、型別轉換" },
        { label: "02 流程控制", description: "if/else、for、switch、defer" },
        { label: "03 函式", description: "多回傳值、variadic、first-class function" },
        { label: "04–09 其他章節", description: "請用 Other 輸入章節編號，如「05」" }
      ]
    }
  ]
})
```

詳細章節目錄與核心考點見 `references/chapters.md`。

### 第三步：讀取章節內容

選定章節（如 07）後讀取：
```
tutorials/07-error-handling/README.md
tutorials/07-error-handling/main.go   （有程式碼範例時一併讀取）
```

路徑格式：`tutorials/{XX-目錄名稱}/README.md`，目錄名稱對照見 `references/chapters.md`。

### 第四步：設計並出題

根據模式決定題數：學習 / 考試模式 5 題，快速測驗 3 題。

難度排序：概念題（第 1–2 題）→ 應用題（第 3–4 題）→ 陷阱或進階題（第 5 題）

**一般選擇題**（options 必須是 `{ label, description }` 物件）：

```
AskUserQuestion({
  questions: [
    {
      question: "【第 1 題 / 共 5 題】\n\n在 Go 中，以下哪個關於 error interface 的描述正確？",
      header: "第 1 題",
      multiSelect: false,
      options: [
        { label: "A", description: "error 是具有 Error() string 方法的 interface" },
        { label: "B", description: "error 是內建的 struct 型別" },
        { label: "C", description: "必須用 errors.New 才能建立 error" },
        { label: "D", description: "error 只能作為第一個回傳值" }
      ]
    }
  ]
})
```

**程式碼閱讀題**（使用 `preview` 讓使用者邊看程式碼邊作答，只有單選題支援 preview）：

```
AskUserQuestion({
  questions: [
    {
      question: "【第 3 題 / 共 5 題】\n\n以下程式碼執行後，err 的值為何？",
      header: "第 3 題",
      multiSelect: false,
      options: [
        {
          label: "A",
          description: "nil（沒有錯誤）",
          preview: "func divide(a, b int) (int, error) {\n    if b == 0 {\n        return 0, errors.New(\"除數不能為零\")\n    }\n    return a / b, nil\n}\n\n_, err := divide(10, 2)"
        },
        { label: "B", description: "\"除數不能為零\"" },
        { label: "C", description: "runtime panic" },
        { label: "D", description: "編譯錯誤" }
      ]
    }
  ]
})
```

### 第五步：判斷答案與回饋

**答對**：
- 讚美 + 1–2 句說明為何正確
- 學習模式：立即繼續下一題；考試模式：記錄，全部結束後統整

**答錯（學習模式）**：
- 說明答錯，給出正確答案與詳細解釋（含程式碼範例）
- 出等效題確認理解：
  ```
  AskUserQuestion({
    questions: [
      {
        question: "確認理解：{同概念換角度的題目}",
        header: "再確認",
        multiSelect: false,
        options: [{ label: "A", description: "..." }, ...]
      }
    ]
  })
  ```
- 再次答錯時：再解釋一次後直接繼續下一題（避免無限迴圈）

**答錯（考試模式）**：記錄錯誤，不立即說明，全部結束後統整講解。

### 第六步：章節總結

```
AskUserQuestion({
  questions: [
    {
      question: "本章測驗完成！\n得分：{X}/{Y}（學習模式以第一次答對計）\n\n{答錯概念列表（如有）}\n\n要繼續嗎？",
      header: "繼續測驗",
      multiSelect: false,
      options: [
        { label: "繼續選其他章節", description: "回到章節選擇" },
        { label: "重測本章", description: "換一批題目再測一次" },
        { label: "結束", description: "離開測驗" }
      ]
    }
  ]
})
```

考試模式在此處逐一講解錯題後再顯示上述選項。

## 題目設計原則

- 優先測該章節**最核心概念**，非邊緣案例
- 4 個選項的干擾選項要選**常見誤解**，提升鑑別度
- 程式碼題目貼近 `tutorials/` 實際程式碼風格
- 有程式碼展示時，使用 `preview` 欄位讓使用者邊看邊選
