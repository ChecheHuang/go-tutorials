---
name: go-chapter-quiz
description: Go 語言教學章節互動測驗工具。針對 go-tutorials 專案的 42 堂課，透過問答形式測試使用者對特定章節的掌握程度，答錯時立即講解直到完全理解。使用時機：使用者說「測驗」、「quiz」、「考考我」、「我想測試第 X 章」、「幫我出題」，或任何要測試 Go 教學章節知識的場景。
---

# Go Chapter Quiz

針對 `/c/work/alphacore/go-tutorials/tutorials/` 下的 42 堂 Go 課程進行互動式知識測驗。

## 工具準備

使用前先載入 AskUserQuestion：
```
ToolSearch: select:AskUserQuestion
```

## 測驗流程

### 第一步：選擇章節

用 AskUserQuestion 呈現章節清單，讓使用者選擇要測驗的章節。詳細章節列表見 `references/chapters.md`。

```
AskUserQuestion({
  question: "請選擇要測驗的章節：\n\n[章節列表...]\n\n請輸入章節編號（例如：01）或章節名稱",
  options: ["01", "02", ..., "42", "隨機一章"]
})
```

### 第二步：讀取章節內容

選定章節後，讀取該章節的 README：
```
/c/work/alphacore/go-tutorials/tutorials/{XX-chapter-name}/README.md
```
同時也可讀取 main.go 理解程式碼範例。

### 第三步：設計並出題

根據 README 內容，設計 **3–5 道題目**，涵蓋：
- 核心概念理解（是非題或選擇題）
- 實務應用（程式碼理解或填空）
- 常見陷阱（容易搞混的地方）

每次只問**一題**，使用 AskUserQuestion 呈現：

```
AskUserQuestion({
  question: "【第 X 題 / 共 Y 題】\n\n{題目內容}\n\nA. ...\nB. ...\nC. ...\nD. ...",
  options: ["A", "B", "C", "D"]  // 選擇題用選項；問答題不設 options
})
```

### 第四步：判斷答案

**答對**：
- 讚美 + 簡短說明為何正確（1–2 句）
- 繼續下一題

**答錯**：
- 明確告知答錯
- 詳細解釋正確答案的原因，包含相關程式碼範例
- 用 AskUserQuestion 再問一次相同或等效題目，確認理解：
  ```
  AskUserQuestion({
    question: "讓我們確認你已理解。{同概念換個角度的題目}",
    options: [...]
  })
  ```
- 直到答對才繼續

### 第五步：章節總結

所有題目結束後：
1. 統計得分（X / Y 題答對）
2. 列出答錯的概念重點
3. 用 AskUserQuestion 詢問是否繼續：
   ```
   AskUserQuestion({
     question: "本章測驗完成！你的得分：{X}/{Y}\n\n{弱點摘要}\n\n要繼續測驗其他章節嗎？",
     options: ["是，繼續選章節", "不了，結束測驗"]
   })
   ```

## 題目設計原則

- 優先測驗該章節**最核心的概念**，而非邊緣案例
- 選擇題設計 4 個選項，確保干擾選項是常見誤解
- 問答題用於測試「為什麼」或「什麼情況下使用」
- 程式碼題目要貼近 tutorials 的實際範例風格
- 難度由易到難：第 1–2 題概念題，第 3–4 題應用題，第 5 題進階或陷阱題

## 章節資訊

詳細的章節主題與核心考點，見 `references/chapters.md`。
