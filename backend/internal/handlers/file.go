package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"envie-backend/internal/database"
	"envie-backend/internal/models"
	"envie-backend/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const MaxFileSize = 1 * 1024 * 1024 // 1MB limit

func checkStorageConfigured(c *gin.Context) bool {
	if !storage.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "File storage is not configured"})
		return false
	}
	return true
}

func ListProjectFiles(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil || access == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var files []models.ProjectFile
	if err := database.DB.
		Preload("UploadedUser").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	type FileResponse struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		SizeBytes    int64     `json:"sizeBytes"`
		MimeType     string    `json:"mimeType"`
		EncryptedFEK string    `json:"encryptedFek"`
		Checksum     string    `json:"checksum"`
		UploadedBy   struct {
			ID    uuid.UUID `json:"id"`
			Name  string    `json:"name"`
			Email string    `json:"email"`
		} `json:"uploadedBy"`
		CreatedAt time.Time `json:"createdAt"`
	}

	response := make([]FileResponse, len(files))
	for i, f := range files {
		response[i] = FileResponse{
			ID:           f.ID,
			Name:         f.Name,
			SizeBytes:    f.SizeBytes,
			MimeType:     f.MimeType,
			EncryptedFEK: f.EncryptedFEK,
			Checksum:     f.Checksum,
			CreatedAt:    f.CreatedAt,
		}
		response[i].UploadedBy.ID = f.UploadedUser.ID
		response[i].UploadedBy.Name = f.UploadedUser.Name
		response[i].UploadedBy.Email = f.UploadedUser.Email
	}

	c.JSON(http.StatusOK, response)
}

func UploadProjectFile(c *gin.Context) {
	if !checkStorageConfigured(c) {
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil || access == nil || !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := c.Request.ParseMultipartForm(MaxFileSize + 1024*1024); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form: " + err.Error()})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	encryptedData, err := io.ReadAll(io.LimitReader(file, MaxFileSize+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}

	if int64(len(encryptedData)) > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File too large. Max size is %d bytes", MaxFileSize)})
		return
	}

	fileName := c.PostForm("name")
	if fileName == "" {
		fileName = header.Filename
	}

	encryptedFEK := c.PostForm("encryptedFek")
	if encryptedFEK == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing encryptedFek"})
		return
	}

	checksum := c.PostForm("checksum")
	mimeType := c.PostForm("mimeType")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	originalSize := c.PostForm("originalSize")
	var sizeBytes int64
	if originalSize != "" {
		fmt.Sscanf(originalSize, "%d", &sizeBytes)
	} else {
		sizeBytes = int64(len(encryptedData))
	}

	fileID := uuid.New()
	s3Key := fmt.Sprintf("projects/%s/files/%s", projectID.String(), fileID.String())

	ctx := context.Background()
	if err := storage.UploadFile(ctx, s3Key, encryptedData, "application/octet-stream"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	projectFile := models.ProjectFile{
		ID:           fileID,
		ProjectID:    projectID,
		Name:         fileName,
		SizeBytes:    sizeBytes,
		MimeType:     mimeType,
		S3Key:        s3Key,
		EncryptedFEK: encryptedFEK,
		Checksum:     checksum,
		UploadedBy:   uid,
	}

	if err := database.DB.Create(&projectFile).Error; err != nil {
		storage.DeleteFile(ctx, s3Key)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        fileID,
		"name":      fileName,
		"sizeBytes": sizeBytes,
	})
}

func DownloadProjectFile(c *gin.Context) {
	if !checkStorageConfigured(c) {
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")
	fileIDStr := c.Param("fileId")

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil || access == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var file models.ProjectFile
	if err := database.DB.Where("id = ? AND project_id = ?", fileID, projectID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	ctx := context.Background()
	data, err := storage.DownloadFile(ctx, file.S3Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         base64.StdEncoding.EncodeToString(data),
		"encryptedFek": file.EncryptedFEK,
		"checksum":     file.Checksum,
		"name":         file.Name,
		"mimeType":     file.MimeType,
	})
}

func DeleteProjectFile(c *gin.Context) {
	if !checkStorageConfigured(c) {
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")
	fileIDStr := c.Param("fileId")

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil || access == nil || !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var file models.ProjectFile
	if err := database.DB.Where("id = ? AND project_id = ?", fileID, projectID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	ctx := context.Background()
	if err := storage.DeleteFile(ctx, file.S3Key); err != nil {
		// Log but continue - we still want to delete the DB record
		fmt.Printf("Warning: Failed to delete file from S3: %v\n", err)
	}

	if err := database.DB.Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func GetProjectFilesForRotation(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil || access == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	type FileFEK struct {
		ID           uuid.UUID `json:"id"`
		EncryptedFEK string    `json:"encryptedFek"`
	}

	var files []FileFEK
	database.DB.Model(&models.ProjectFile{}).
		Select("id, encrypted_fek").
		Where("project_id = ?", projectID).
		Scan(&files)

	c.JSON(http.StatusOK, files)
}

type UpdateFileFEKsRequest struct {
	Files []struct {
		ID           uuid.UUID `json:"id"`
		EncryptedFEK string    `json:"encryptedFek"`
	} `json:"files"`
}

func UpdateFileFEKs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil || access == nil || !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req UpdateFileFEKsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()
	for _, f := range req.Files {
		result := tx.Model(&models.ProjectFile{}).
			Where("id = ? AND project_id = ?", f.ID, projectID).
			Update("encrypted_fek", f.EncryptedFEK)

		if result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file FEK"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit updates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File FEKs updated successfully", "count": len(req.Files)})
}
