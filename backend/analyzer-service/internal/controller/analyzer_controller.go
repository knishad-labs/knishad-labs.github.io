package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"analyzer-service/internal/model"
	"analyzer-service/internal/repository"
	"analyzer-service/internal/services"
	"analyzer-service/internal/utils"
	"github.com/gin-gonic/gin"
)

type AnalyzeRequest struct {
	DatabaseConnectionID uint   `json:"database_connection_id" binding:"required"`
	AIProviderID         uint   `json:"ai_provider_id" binding:"required"`
	Query                string `json:"query" binding:"required"`
	Rules                string `json:"rules"` // optional custom rules
}

type ApplyFixRequest struct {
	SuggestionIndex int `json:"suggestion_index"` // which suggestion in the array to apply, defaults to 0
}

// AnalyzeQuery runs EXPLAIN on the target DB and gets AI advice
func AnalyzeQuery(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Fetch DB connection configuration
	var dbConn model.DatabaseConnection
	if err := repository.DB.First(&dbConn, req.DatabaseConnectionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database connection not found"})
		return
	}

	decryptedPassword, err := utils.Decrypt(dbConn.PasswordEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt database password"})
		return
	}

	// 2. Fetch AI provider configuration
	var aiProv model.AIProvider
	if err := repository.DB.First(&aiProv, req.AIProviderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI provider config not found"})
		return
	}

	decryptedApiKey, err := utils.Decrypt(aiProv.ApiKeyEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt AI API key"})
		return
	}

	// 3. Connect to Target DB and execute EXPLAIN query
	targetConn := repository.TargetDBConnection{
		DbType:       dbConn.DbType,
		Host:         dbConn.Host,
		Port:         dbConn.Port,
		Username:     dbConn.Username,
		Password:     decryptedPassword,
		DatabaseName: dbConn.DatabaseName,
	}

	explainPlan, err := repository.ExecuteExplain(targetConn, req.Query)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Target DB query execution plan check failed: " + err.Error()})
		return
	}

	// 4. Send plan to AI Provider
	aiResult, err := services.AnalyzeQueryWithAI(aiProv.ApiEndpoint, decryptedApiKey, aiProv.ModelName, dbConn.DbType, req.Query, explainPlan, req.Rules)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI analysis query failed: " + err.Error()})
		return
	}

	// 5. Serialize AI result back to JSON string for Postgres JSONB field
	aiResultBytes, err := json.Marshal(aiResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal AI results"})
		return
	}

	// 6. Save optimization report
	report := model.OptimizationReport{
		DatabaseConnectionID: dbConn.ID,
		QueryText:            req.Query,
		ExecutionPlan:        explainPlan,
		AnalysisResult:       string(aiResultBytes),
		Status:               "Pending Review",
	}

	if err := repository.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save optimization report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetReports retrieves all query optimization reports
func GetReports(c *gin.Context) {
	var reports []model.OptimizationReport
	if err := repository.DB.Order("created_at desc").Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}
	c.JSON(http.StatusOK, reports)
}

// GetReport retrieves a single query optimization report by ID
func GetReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var report model.OptimizationReport
	if err := repository.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}
	c.JSON(http.StatusOK, report)
}

// ApplyFix runs the AI recommended SQL change (e.g. CREATE INDEX) directly on target database
func ApplyFix(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var report model.OptimizationReport
	if err := repository.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.Status == "Applied" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This fix has already been applied"})
		return
	}

	var req ApplyFixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// default to index 0 if payload omitted
		req.SuggestionIndex = 0
	}

	// 1. Unmarshal the AI analysis result
	var analysis services.QueryAnalysisResult
	if err := json.Unmarshal([]byte(report.AnalysisResult), &analysis); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read report suggestions detail"})
		return
	}

	if len(analysis.Suggestions) <= req.SuggestionIndex {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion index"})
		return
	}

	suggestion := analysis.Suggestions[req.SuggestionIndex]
	if suggestion.Sql == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This suggestion does not contain an executable SQL fix"})
		return
	}

	// 2. Fetch target database connection credentials
	var dbConn model.DatabaseConnection
	if err := repository.DB.First(&dbConn, report.DatabaseConnectionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target database connection details not found"})
		return
	}

	decryptedPassword, err := utils.Decrypt(dbConn.PasswordEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt database credentials"})
		return
	}

	targetConn := repository.TargetDBConnection{
		DbType:       dbConn.DbType,
		Host:         dbConn.Host,
		Port:         dbConn.Port,
		Username:     dbConn.Username,
		Password:     decryptedPassword,
		DatabaseName: dbConn.DatabaseName,
	}

	// 3. Execute the SQL command
	err = repository.ExecuteQuery(targetConn, suggestion.Sql)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to apply SQL fix to target database: " + err.Error()})
		return
	}

	// 4. Update status in configuration database
	now := time.Now()
	report.Status = "Applied"
	report.AppliedFix = suggestion.Sql
	report.AppliedAt = &now

	if err := repository.DB.Save(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fix succeeded but failed to update status in reports database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "SQL fix executed successfully on target database!",
		"applied_sql": suggestion.Sql,
	})
}
