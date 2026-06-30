package cron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"scheduler-service/internal/model"
	"scheduler-service/internal/repository"
	"github.com/robfig/cron/v3"
)

type AnalyzeRequest struct {
	DatabaseConnectionID uint   `json:"database_connection_id"`
	AIProviderID         uint   `json:"ai_provider_id"`
	Query                string `json:"query"`
	Rules                string `json:"rules"`
}

type OptimizationReportResponse struct {
	ID             uint   `json:"id"`
	AnalysisResult string `json:"analysis_result"`
}

type EmailReportRequest struct {
	ReportID   uint   `json:"report_id"`
	QueryText  string `json:"query_text"`
	ReportJSON string `json:"report_json"`
}

var CronManager *cron.Cron
var cronEntryMap map[uint]cron.EntryID

// StartCronEngine compiles all active scheduler jobs and boots up the robfig/cron runtime
func StartCronEngine() {
	CronManager = cron.New()
	cronEntryMap = make(map[uint]cron.EntryID)

	// Fetch active jobs
	var tasks []model.ScheduledTask
	if err := repository.DB.Where("is_active = ?", true).Find(&tasks).Error; err != nil {
		log.Printf("Error fetching scheduled tasks: %v", err)
		return
	}

	for _, task := range tasks {
		ScheduleTask(task)
	}

	CronManager.Start()
	log.Println("Scheduler cron engine started successfully.")
}

// ScheduleTask adds a single query scan task to the cron engine
func ScheduleTask(task model.ScheduledTask) {
	analyzerServiceURL := os.Getenv("ANALYZER_SERVICE_URL")
	if analyzerServiceURL == "" {
		analyzerServiceURL = "http://localhost:8081"
	}
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:8082"
	}

	jobFunc := func() {
		log.Printf("[Scheduler] Executing cron task '%s' (ID: %d)...", task.Name, task.ID)

		for _, query := range task.TargetQueries {
			log.Printf("[Scheduler] Initiating explain and query analysis for query: %s", query)

			// 1. Call analyzer REST endpoint
			payload := AnalyzeRequest{
				DatabaseConnectionID: task.DatabaseConnectionID,
				AIProviderID:         task.AIProviderID,
				Query:                query,
				Rules:                task.SkillRules,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Failed to marshal query analyzer request: %v", err)
				continue
			}

			resp, err := http.Post(
				fmt.Sprintf("%s/api/analyze", analyzerServiceURL),
				"application/json",
				bytes.NewBuffer(payloadBytes),
			)
			if err != nil {
				log.Printf("Failed to trigger query analyzer: %v", err)
				continue
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read query analysis output: %v", err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("Analyzer returned status %d: %s", resp.StatusCode, string(bodyBytes))
				continue
			}

			var reportResponse OptimizationReportResponse
			if err := json.Unmarshal(bodyBytes, &reportResponse); err != nil {
				log.Printf("Failed to parse report details: %v", err)
				continue
			}

			log.Printf("[Scheduler] Query analysis report %d generated.", reportResponse.ID)

			// 2. Call SMTP email notification worker REST endpoint
			emailPayload := EmailReportRequest{
				ReportID:   reportResponse.ID,
				QueryText:  query,
				ReportJSON: reportResponse.AnalysisResult,
			}
			emailPayloadBytes, err := json.Marshal(emailPayload)
			if err != nil {
				log.Printf("Failed to marshal email notification payload: %v", err)
				continue
			}

			mailResp, err := http.Post(
				fmt.Sprintf("%s/api/send-report", notificationServiceURL),
				"application/json",
				bytes.NewBuffer(emailPayloadBytes),
			)
			if err != nil {
				log.Printf("Failed to dispatch report to notification service: %v", err)
				continue
			}
			mailResp.Body.Close()
			log.Printf("[Scheduler] Notification dispatch triggered.")
		}
	}

	entryID, err := CronManager.AddFunc(task.CronExpression, jobFunc)
	if err != nil {
		log.Printf("Failed to queue job '%s' with cron '%s': %v", task.Name, task.CronExpression, err)
		return
	}

	cronEntryMap[task.ID] = entryID
	log.Printf("Cron successfully scheduled task '%s' (ID: %d) with pattern: %s", task.Name, task.ID, task.CronExpression)
}

// UnscheduleTask removes a specific job from active cron runners
func UnscheduleTask(taskID uint) {
	if entryID, exists := cronEntryMap[taskID]; exists {
		CronManager.Remove(entryID)
		delete(cronEntryMap, taskID)
		log.Printf("Unscheduled task ID %d from cron loop.", taskID)
	}
}
