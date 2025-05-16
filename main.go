package main

import (
	"backend/controllers"
	"backend/middlewares"
	"backend/models"
	"github.com/gin-gonic/gin"
)

func main() {
	models.ConnectDataBase()

	router := gin.Default()

	err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	if err != nil {
		return
	}

	// 認証不要の公開API
	public := router.Group("/api")
	{
		// 既存のルート
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)

		// 新規：公開プラン取得用エンドポイント
		public.GET("/public-plans", controllers.GetPublicPlans)
	}

	// JWT認証が必要なAPI
	protected := router.Group("/api/admin")
	protected.Use(middlewares.JwtAuthMiddleware())
	{
		// 既存のユーザー情報取得
		protected.GET("/user", controllers.CurrentUser)
	}

	// 新規：プランニング関連の認証が必要なAPI
	plans := router.Group("/api/plans")
	plans.Use(middlewares.JwtAuthMiddleware())
	{
		// プラン管理
		plans.POST("", controllers.CreatePlan)
		plans.GET("/:id", controllers.CreatePlanItem)
		plans.GET("/:id", controllers.GetPlan)
		plans.PUT("/:id", controllers.UpdatePlan)
		plans.PATCH("/:id/status", controllers.UpdatePlanStatus)
		plans.DELETE("/:id", controllers.DeletePlan)
		plans.GET("", controllers.GetMyPlans)

		// アクティビティ管理 - :id に統一
		// アクティビティ管理はコメントアウトします
		// plans.POST("/:id/activities", controllers.AddActivityToPlan)
		// plans.PUT("/:id/activities/:activity_id", controllers.UpdateActivity)
		// plans.DELETE("/:id/activities/:activity_id", controllers.RemoveActivity)
	}

	err = router.Run(":8080")
	if err != nil {
		return
	}
}
