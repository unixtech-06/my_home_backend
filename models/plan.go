package models

import "time"

type TravelPlan struct {
	ID          string     `json:"id" validate:"required"`          // プランID
	Title       string     `json:"title" validate:"required"`       // 例：「京都1日観光プラン」
	Description string     `json:"description" validate:"required"` // プランの説明
	Items       []PlanItem `json:"items" validate:"required,dive"`  // プランの各項目
	TotalCost   int        `json:"totalCost" validate:"required"`   // 合計費用
	CreatedAt   time.Time  `json:"createdAt" validate:"required"`   // 作成日時
	UpdatedAt   time.Time  `json:"updatedAt" validate:"required"`   // 更新日時
	CreatorID   uint       `json:"creatorId" validate:"required"`   // プラン作成者のユーザーID
	IsPublic    bool       `json:"isPublic" validate:"required"`    // プランの公開状態
}

type PlanItem struct {
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
