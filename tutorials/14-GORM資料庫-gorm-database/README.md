# 第十三課：GORM 資料庫操作

## 學習目標

- 了解 ORM 是什麼、為什麼要使用它
- 學會 GORM 的 CRUD 操作
- 掌握 Preload（預載入關聯資料）
- 學會分頁、搜尋、排序等進階查詢

## 執行方式

```bash
cd tutorials/13-gorm-database
go mod init gorm-demo && go mod tidy
go run main.go
```

## 重點筆記

### 什麼是 ORM？

ORM（Object-Relational Mapping）讓你用程式語言的物件來操作資料庫，不需要寫 SQL：

```
Go 程式碼                    ↔    SQL
db.Create(&user)            →    INSERT INTO users (name, email) VALUES (?, ?)
db.First(&user, 1)          →    SELECT * FROM users WHERE id = 1
db.Save(&user)              →    UPDATE users SET ... WHERE id = ?
db.Delete(&user, 1)         →    DELETE FROM users WHERE id = 1
db.Where("age > ?", 25)     →    ... WHERE age > 25
```

### GORM CRUD 速查表

| 操作 | 方法 | SQL |
|------|------|-----|
| 建立 | `db.Create(&user)` | INSERT |
| 單筆查詢 | `db.First(&user, id)` | SELECT ... LIMIT 1 |
| 條件查詢 | `db.Where("x = ?", v).First(&u)` | SELECT ... WHERE |
| 查詢全部 | `db.Find(&users)` | SELECT * |
| 更新全部 | `db.Save(&user)` | UPDATE (所有欄位) |
| 更新部分 | `db.Model(&u).Update("col", val)` | UPDATE (單一欄位) |
| 刪除 | `db.Delete(&user, id)` | DELETE |

### Preload（最重要的概念之一）

```go
// 沒有 Preload → article.User 是空的
db.First(&article, 1)

// 有 Preload → article.User 有資料
db.Preload("User").First(&article, 1)

// 多層 Preload
db.Preload("User").Preload("Comments").Preload("Comments.User").First(&article, 1)
```

### 在專案中的對應

`internal/repository/article_repository.go`：
```go
func (r *articleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
    db := r.db.Model(&domain.Article{})

    if query.Search != "" {
        db = db.Where("title LIKE ?", "%"+query.Search+"%")
    }

    db.Count(&total)

    db.Preload("User").
       Order("created_at DESC").
       Offset(offset).
       Limit(query.PageSize).
       Find(&articles)
}
```

## 練習

1. 新增一個 `Comment` 模型，建立與 Article 的一對多關係
2. 用 Preload 同時載入文章的作者和留言
3. 實作分頁查詢：第 2 頁、每頁 5 筆
