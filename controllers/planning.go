package controllers

import (
	"backend/models"
	"backend/utils/token"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TravelPlanInput struct {
	ID          string          `json:"id" validate:"required"`          // プランID
	Title       string          `json:"title" validate:"required"`       // 例：「京都1日観光プラン」
	Description string          `json:"description" validate:"required"` // プランの説明
	Items       []PlanItemInput `json:"items" validate:"required,dive"`  // プランの各項目
	TotalCost   int             `json:"totalCost" validate:"required"`   // 合計費用
	CreatedAt   time.Time       `json:"createdAt" validate:"required"`   // 作成日時
	UpdatedAt   time.Time       `json:"updatedAt" validate:"required"`   // 更新日時
	CreatorID   uint            `json:"creatorId" validate:"required"`   // プラン作成者のユーザーID
	IsPublic    bool            `json:"isPublic" validate:"required"`    // プランの公開状態
}

type PlanItemInput struct {
	ID          string    `json:"id" validate:"required"`          // アクティビティID
	PlanID      string    `json:"planId" validate:"required"`      // プランID
	Type        string    `json:"type" validate:"required"`        // "visit"(訪問)、"transport"(移動)、"meal"(食事)など
	Title       string    `json:"title" validate:"required"`       // 例：「清水寺観光」
	Description string    `json:"description" validate:"required"` // 詳細説明
	Location    string    `json:"location" validate:"required"`    // 場所
	StartTime   time.Time `json:"startTime" validate:"required"`   // 開始時間
	EndTime     time.Time `json:"endTime" validate:"required"`     // 終了時間
	Duration    int       `json:"duration" validate:"required"`    // 所要時間（分）
	Cost        int       `json:"cost" validate:"required"`        // 費用（円）
	Notes       string    `json:"notes" validate:"required"`       // メモ
	Order       int       `json:"order" validate:"required"`       // 順序
}

// CreatePlan プランを作成する
func CreatePlan(c *gin.Context) {
	var input TravelPlanInput

	// リクエストのJSONデータをPlanInput構造体にバインドする
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// トークンからユーザーIDを取得
	userId, err := token.ExtractTokenId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証に失敗しました"})
		return
	}

	// プランオブジェクトを作成
	plan := models.TravelPlan{
		Title:       input.Title,
		Description: input.Description,
		TotalCost:   input.TotalCost,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatorID:   userId,
	}

	// データベースに保存
	err = models.DB.Create(&plan).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "プランの作成に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plan})
}

// CreatePlanItem プランにアイテムを追加する
func CreatePlanItem(c *gin.Context) {
	id := c.Param("id")
	var input PlanItemInput

	// リクエストのJSONデータをPlanItemInput構造体にバインドする
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// プランを取得
	var plan models.TravelPlan
	if err := models.DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "プランが見つかりません"})
		return
	}

	// トークンからユーザーIDを取得し、プランの作成者と一致するか確認
	userId, err := token.ExtractTokenId(c)
	if err != nil || userId != plan.CreatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "このプランにアクセスする権限がありません"})
		return
	}

	// プランアイテムオブジェクトを作成
	item := models.PlanItem{
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		Location:    input.Location,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Duration:    input.Duration,
		Cost:        input.Cost,
		Notes:       input.Notes,
		Order:       input.Order,
	}

	item.PlanID = id

	models.DB.Create(&item)
	c.JSON(http.StatusOK, gin.H{"data": item})

}

// GetPlan 指定されたIDのプランを取得する
func GetPlan(c *gin.Context) {
	id := c.Param("id")
	var plan models.TravelPlan

	// プランを取得 (リレーションを含む)
	if err := models.DB.Preload("PlanItem").First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "プランが見つかりません"})
		return
	}

	// 非公開プランの場合は作成者のみアクセス可能
	if !plan.IsPublic {
		userId, err := token.ExtractTokenId(c)
		if err != nil || userId != plan.CreatorID {
			c.JSON(http.StatusForbidden, gin.H{"error": "このプランにアクセスする権限がありません"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": plan})
}

// UpdatePlan 指定されたIDのプランを更新する
func UpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var plan models.TravelPlan

	// プランを取得
	if err := models.DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "プランが見つかりません"})
		return
	}

	// トークンからユーザーIDを取得し、プランの作成者と一致するか確認
	userId, err := token.ExtractTokenId(c)
	if err != nil || userId != plan.CreatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "このプランを更新する権限がありません"})
		return
	}

	// 入力を取得
	var input TravelPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// プランを更新（StatusとCreatorIDは変更しない）
	updatedPlan := models.TravelPlan{
		Title:       input.Title,
		Description: input.Description,
		TotalCost:   input.TotalCost,
		UpdatedAt:   time.Now(),
		IsPublic:    input.IsPublic,
	}

	models.DB.Model(&plan).Updates(updatedPlan)
	c.JSON(http.StatusOK, gin.H{"data": plan})
}

// UpdatePlanStatus プランのステータスを更新する
func UpdatePlanStatus(c *gin.Context) {
	id := c.Param("id")
	var plan models.TravelPlan

	// プランを取得
	if err := models.DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "プランが見つかりません"})
		return
	}

	// トークンからユーザーIDを取得し、プランの作成者と一致するか確認
	userId, err := token.ExtractTokenId(c)
	if err != nil || userId != plan.CreatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "このプランを更新する権限がありません"})
		return
	}

	// 入力を取得
	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ステータスのバリデーション
	validStatuses := map[string]bool{"draft": true, "confirmed": true, "completed": true, "cancelled": true}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なステータスです"})
		return
	}

	// ステータスのみ更新
	models.DB.Model(&plan).Update("status", input.Status)
	c.JSON(http.StatusOK, gin.H{"data": plan})
}

// DeletePlan 指定されたIDのプランを削除する
func DeletePlan(c *gin.Context) {
	id := c.Param("id")
	var plan models.TravelPlan

	// プランを取得
	if err := models.DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "プランが見つかりません"})
		return
	}

	// トークンからユーザーIDを取得し、プランの作成者と一致するか確認
	userId, err := token.ExtractTokenId(c)
	if err != nil || userId != plan.CreatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "このプランを削除する権限がありません"})
		return
	}

	// プランを削除
	models.DB.Delete(&plan)
	c.JSON(http.StatusOK, gin.H{"data": "プランが削除されました"})
}

// GetMyPlans 自分が作成したプランを取得する
func GetMyPlans(c *gin.Context) {
	// トークンからユーザーIDを取得
	userId, err := token.ExtractTokenId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証に失敗しました"})
		return
	}

	var plans []models.TravelPlan
	// ユーザーIDに基づいてプランを取得
	models.DB.Where("creator_id = ?", userId).Find(&plans)

	c.JSON(http.StatusOK, gin.H{"data": plans})
}

// GetPublicPlans 公開プランを取得する
func GetPublicPlans(c *gin.Context) {
	var plans []models.TravelPlan

	// カテゴリでフィルタリング（オプション）
	category := c.Query("category")
	query := models.DB.Where("is_public = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 日付でソート
	query.Order("start_date desc").Find(&plans)

	c.JSON(http.StatusOK, gin.H{"data": plans})
}
