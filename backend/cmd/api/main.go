package main

import (
	"log"

	"envie-backend/internal/auth"
	"envie-backend/internal/database"
	"envie-backend/internal/handlers"
	"envie-backend/internal/middleware"
	"envie-backend/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env vars")
	}

	database.Connect()
	auth.InitOAuth()

	if err := storage.InitS3(); err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}
	log.Println("S3 storage initialized successfully")

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Master-Key-Version")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes
	r.GET("/auth/login", handlers.AuthLogin)
	r.GET("/auth/callback", handlers.AuthCallback)
	r.POST("/auth/exchange", handlers.AuthExchange)
	r.POST("/auth/refresh", handlers.AuthRefresh)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.GET("/me", handlers.GetMe)
		authorized.PUT("/me/public-key", handlers.SetPublicKey)
		authorized.POST("/me/rotate-master-key", handlers.RotateMasterKey)
		authorized.POST("/auth/logout", handlers.AuthLogout)

		// Identity
		authorized.POST("/devices", handlers.RegisterDevice)
		authorized.GET("/devices", handlers.GetDevices)
		authorized.DELETE("/devices", handlers.DeleteAllDevices)
		authorized.DELETE("/devices/:id", handlers.DeleteDevice)
		authorized.PUT("/devices/:id", handlers.UpdateDevice)

		// Project Routes
		authorized.POST("/projects", handlers.CreateProject)
		authorized.GET("/projects", handlers.GetProjects)
		authorized.GET("/projects/:id", handlers.GetProject)
		authorized.PUT("/projects/:id", handlers.UpdateProject)
		// Config Items
		authorized.GET("/projects/:id/config", handlers.GetConfigItems)
		authorized.PUT("/projects/:id/config", handlers.SyncConfigItems)
		authorized.DELETE("/projects/:id", handlers.DeleteProject)

		// Secret Manager Configs
		authorized.GET("/projects/:id/secret-managers", handlers.GetSecretManagerConfigs)
		authorized.POST("/projects/:id/secret-managers", handlers.CreateSecretManagerConfig)
		authorized.PUT("/projects/:id/secret-managers/:configId", handlers.UpdateSecretManagerConfig)
		authorized.DELETE("/projects/:id/secret-managers/:configId", handlers.DeleteSecretManagerConfig)

		// Project Access (Teams)
		authorized.GET("/projects/:id/teams", handlers.GetProjectTeams)
		authorized.POST("/projects/:id/teams", handlers.AddTeamToProject)

		// Key Rotation
		authorized.GET("/projects/:id/rotation", handlers.GetPendingRotation)
		authorized.POST("/projects/:id/rotation", handlers.InitiateKeyRotation)
		authorized.POST("/projects/:id/rotation/:rotationId/approve", handlers.ApproveKeyRotation)
		authorized.POST("/projects/:id/rotation/:rotationId/reject", handlers.RejectKeyRotation)
		authorized.DELETE("/projects/:id/rotation/:rotationId", handlers.CancelKeyRotation)
		authorized.GET("/pending-rotations", handlers.GetUserPendingRotations)

		// Project Files
		authorized.GET("/projects/:id/files", handlers.ListProjectFiles)
		authorized.POST("/projects/:id/files", handlers.UploadProjectFile)
		authorized.GET("/projects/:id/files/:fileId", handlers.DownloadProjectFile)
		authorized.DELETE("/projects/:id/files/:fileId", handlers.DeleteProjectFile)
		authorized.GET("/projects/:id/files-feks", handlers.GetProjectFilesForRotation)
		authorized.PUT("/projects/:id/files-feks", handlers.UpdateFileFEKs)

		// Organizations
		authorized.POST("/organizations", handlers.CreateOrganization)
		authorized.GET("/organizations", handlers.GetOrganizations)
		authorized.GET("/organizations/:id", handlers.GetOrganization)
		authorized.PUT("/organizations/:id", handlers.UpdateOrganization)
		authorized.GET("/organizations/:id/users", handlers.GetOrganizationUsers)

		// Teams
		authorized.POST("/teams", handlers.CreateTeam)
		authorized.GET("/teams", handlers.GetTeams)
		authorized.GET("/teams/my", handlers.GetMyTeams)
		authorized.PUT("/teams/:id/my-key", handlers.UpdateMyTeamKey)
		authorized.GET("/teams/:id/members", handlers.GetTeamMembers)
		authorized.POST("/teams/:id/members", handlers.AddTeamMember)
		authorized.PUT("/teams/:id/members/:userId", handlers.UpdateTeamMember)
		authorized.DELETE("/teams/:id/members/:userId", handlers.RemoveTeamMember)
	}

	r.Run(":8080")
}
