package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"quote-voting-backend/config"
	"quote-voting-backend/models"
	"quote-voting-backend/utils"

	"github.com/gin-gonic/gin"
)

// POST /quotes
func CreateQuote(c *gin.Context) {
	var quote models.Quote
	if err := c.ShouldBindJSON(&quote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, _ := utils.GetUserIDFromToken(c)
	quote.AuthorID = userId
	quote.CreatedAt = time.Now()
	quote.UpdatedAt = time.Now()

	if err := config.DB.Create(&quote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้าง quote ได้"})
		return
	}

	var createdQuote models.Quote
	if err := config.DB.Preload("Author").First(&createdQuote, quote.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดข้อมูลผู้สร้าง quote ได้"})
		return
	}

	resp := models.QuoteCreateResponse{
		ID:          createdQuote.ID,
		Content:     createdQuote.Content,
		AuthorID:    createdQuote.AuthorID,
		AuthorName:  createdQuote.Author.Username,
		IsAnonymous: createdQuote.IsAnonymous,
		CreatedAt:   createdQuote.CreatedAt,
		UpdatedAt:   createdQuote.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

// GET /quotes
func GetQuotes(c *gin.Context) {
	searchContent := c.Query("searchContent")        // search ใน content
	searchAuthor := c.Query("searchAuthor")          // search ใน username ของ author
	sortBy := c.DefaultQuery("sort", "votes")        // votes, name, updated_at
	direction := c.DefaultQuery("direction", "desc") // asc, desc
	dateFilter := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// pageLength, _ := strconv.Atoi(c.DefaultQuery("pageLength", "10"))
	// offset := (page - 1) * pageLength

	var totalItems int64
	var quotes []models.Quote
	joinedUsers := false

	// Base query
	query := config.DB.Model(&models.Quote{}).
		Preload("Author").
		Preload("Votes").
		Where("DATE(quotes.created_at) = ?", dateFilter)

	// Clone countQuery
	countQuery := config.DB.Model(&models.Quote{}).
		Where("DATE(quotes.created_at) = ?", dateFilter)

	// Search เฉพาะ content
	if searchContent != "" {
		query = query.Where("quotes.content ILIKE ?", "%"+searchContent+"%")
		countQuery = countQuery.Where("quotes.content ILIKE ?", "%"+searchContent+"%")
	}

	// Search เฉพาะ author name
	if searchAuthor != "" {
		if strings.ToLower(searchAuthor) == "anonymous" {
			// กรองเฉพาะ quote ที่เป็น anonymous
			query = query.Where("quotes.is_anonymous = ?", true)
			countQuery = countQuery.Where("quotes.is_anonymous = ?", true)
		} else {
			// กรองเฉพาะ quote ที่ไม่ใช่ anonymous และ username ตรง search
			if !joinedUsers {
				query = query.Joins("JOIN users ON users.id = quotes.author_id")
				countQuery = countQuery.Joins("JOIN users ON users.id = quotes.author_id")
				joinedUsers = true
			}
			query = query.Where("quotes.is_anonymous = ? AND users.username ILIKE ?", false, "%"+searchAuthor+"%")
			countQuery = countQuery.Where("quotes.is_anonymous = ? AND users.username ILIKE ?", false, "%"+searchAuthor+"%")
		}
	}

	// Sort
	switch sortBy {
	case "name":
		if !joinedUsers {
			query = query.Joins("JOIN users ON users.id = quotes.author_id")
			countQuery = countQuery.Joins("JOIN users ON users.id = quotes.author_id")
			joinedUsers = true
		}
		query = query.Order("users.username " + direction)
	case "updated_at":
		query = query.Order("quotes.updated_at " + direction)
	case "votes":
		fallthrough
	default:
		query = query.
			Select("quotes.*, COUNT(votes.id) as vote_count").
			Joins("LEFT JOIN votes ON votes.quote_id = quotes.id").
			Group("quotes.id").
			Order("vote_count " + direction)
	}

	// Count (แยกจาก query ที่มี GROUP BY เพื่อความแม่นยำ)
	countQuery.Count(&totalItems)

	// Pagination
	// query = query.Limit(pageLength).Offset(offset)

	// Fetch data
	if err := query.Find(&quotes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  500,
			"message": "ไม่สามารถดึง quotes ได้",
			"data":    []interface{}{},
		})
		return
	}

	responses := make([]models.QuoteSearchResponse, 0)
	for _, q := range quotes {
		authorName := q.Author.Username
		if q.IsAnonymous {
			authorName = "Anonymous"
		}

		responses = append(responses, models.QuoteSearchResponse{
			ID:          q.ID,
			Content:     q.Content,
			AuthorID:    q.AuthorID,
			AuthorName:  authorName,
			IsAnonymous: q.IsAnonymous,
			CreatedAt:   q.CreatedAt,
			UpdatedAt:   q.UpdatedAt,
			VoteCount:   len(q.Votes),
		})
	}

	// Success Response
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "Quotes fetched successfully",
		// "page":       page,
		// "pageLength": pageLength,
		"totalItems": totalItems,
		"data":       responses,
	})
}

// PUT /quotes/:id
func UpdateQuote(c *gin.Context) {
	quoteID, _ := strconv.Atoi(c.Param("id"))

	var quote models.Quote
	if err := config.DB.First(&quote, quoteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ quote นี้"})
		return
	}

	userID, _ := utils.GetUserIDFromToken(c)
	if quote.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "คุณไม่ได้เป็นเจ้าของ quote นี้"})
		return
	}

	var input models.Quote
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quote.Content = input.Content
	quote.IsAnonymous = input.IsAnonymous
	quote.UpdatedAt = time.Now()

	if err := config.DB.Save(&quote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "แก้ไขไม่ได้"})
		return
	}

	// โหลดข้อมูล Author ใหม่เพื่อดึง username มาแสดง
	var updatedQuote models.Quote
	if err := config.DB.Preload("Author").First(&updatedQuote, quote.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดข้อมูลผู้สร้าง quote ได้"})
		return
	}

	resp := models.QuoteCreateResponse{
		ID:          updatedQuote.ID,
		Content:     updatedQuote.Content,
		AuthorID:    updatedQuote.AuthorID,
		AuthorName:  updatedQuote.Author.Username,
		IsAnonymous: updatedQuote.IsAnonymous,
		CreatedAt:   updatedQuote.CreatedAt,
		UpdatedAt:   updatedQuote.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

// POST /quotes/:id/vote
func VoteQuote(c *gin.Context) {
	quoteID, _ := strconv.Atoi(c.Param("quoteId"))
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// var input struct {
	// 	IsAnonymous bool `json:"is_anonymous"`
	// }
	// if err := c.ShouldBindJSON(&input); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
	// 	return
	// }

	today := time.Now().Format("2006-01-02")

	var existing models.Vote
	err = config.DB.
		Where("voter_id = ? AND vote_date = ?", userID, today).
		First(&existing).Error

	if err == nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "You have already voted a quote today"})
		return
	}

	// เพิ่ม vote
	vote := models.Vote{
		VoterID:     userID,
		QuoteID:     uint(quoteID),
		IsAnonymous: false,
		VoteDate:    time.Now(),
	}
	if err := config.DB.Create(&vote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to vote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Voted successfully"})
}

// GET /votes/today
func GetTodayVotes(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	today := time.Now().Format("2006-01-02")

	var votes []models.Vote
	err = config.DB.Preload("Quote").
		Where("voter_id = ? AND vote_date = ?", userID, today).
		Find(&votes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch today's votes"})
		return
	}

	if len(votes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": "You haven't voted today yet",
			"data":    []interface{}{},
		})
		return
	}

	type VoteResponse struct {
		ID           uint      `json:"id"`
		QuoteID      uint      `json:"quote_id"`
		QuoteContent string    `json:"quote_content"`
		IsAnonymous  bool      `json:"is_anonymous"`
		VoteDate     time.Time `json:"vote_date"`
	}

	resp := make([]VoteResponse, 0, len(votes))
	for _, v := range votes {
		resp = append(resp, VoteResponse{
			ID:           v.ID,
			QuoteID:      v.QuoteID,
			QuoteContent: v.Quote.Content,
			IsAnonymous:  v.IsAnonymous,
			VoteDate:     v.VoteDate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "Today's votes fetched successfully",
		"data":    resp,
	})
}

// GET /quotes/qotd
func GetQuoteOfTheDay(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	// Step 1: หา quote ที่ได้ vote เยอะที่สุดในวันนี้
	var maxVoteCount int
	err := config.DB.
		Table("votes").
		Select("COUNT(*)").
		Where("vote_date = ?", today).
		Group("quote_id").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&maxVoteCount).Error

	if err != nil || maxVoteCount == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": "No quotes have been voted today",
			"data":    []interface{}{},
		})
		return
	}

	// Step 2: ดึง quote ทั้งหมดที่มี vote เท่ากับ maxVoteCount
	var topQuoteIDs []uint
	err = config.DB.
		Table("votes").
		Select("quote_id").
		Where("vote_date = ?", today).
		Group("quote_id").
		Having("COUNT(*) = ?", maxVoteCount).
		Scan(&topQuoteIDs).Error

	if err != nil || len(topQuoteIDs) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  500,
			"message": "Failed to fetch quotes",
		})
		return
	}

	// Step 3: ดึงรายละเอียด quote เหล่านั้น
	var quotes []models.Quote
	err = config.DB.
		Preload("Author").
		Preload("Votes", "vote_date = ?", today).
		Where("id IN ?", topQuoteIDs).
		Find(&quotes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  500,
			"message": "Failed to fetch quotes",
		})
		return
	}

	// Step 4: เตรียม response
	responses := make([]models.QuoteSearchResponse, 0)
	for _, q := range quotes {
		name := q.Author.Username
		if q.IsAnonymous {
			name = "Anonymous"
		}
		responses = append(responses, models.QuoteSearchResponse{
			ID:          q.ID,
			Content:     q.Content,
			AuthorID:    q.AuthorID,
			AuthorName:  name,
			IsAnonymous: q.IsAnonymous,
			CreatedAt:   q.CreatedAt,
			UpdatedAt:   q.UpdatedAt,
			VoteCount:   len(q.Votes),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      200,
		"message":     "Quote(s) of the Day fetched successfully",
		"highestVote": maxVoteCount,
		"totalQuotes": len(responses),
		"data":        responses,
	})
}
