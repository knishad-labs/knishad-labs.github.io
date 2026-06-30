package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type QuerySuggestion struct {
	Type        string `json:"type"`
	Sql         string `json:"sql"`
	Description string `json:"description"`
}

type QueryAnalysisResult struct {
	Issue       string            `json:"issue"`
	Rationale   string            `json:"rationale"`
	Suggestions []QuerySuggestion `json:"suggestions"`
}

type SendReportRequest struct {
	ReportID   uint   `json:"report_id" binding:"required"`
	QueryText  string `json:"query_text" binding:"required"`
	ReportJSON string `json:"report_json" binding:"required"`
}

// SendReportEmail formats the report and sends it via SMTP
func SendReportEmail(c *gin.Context) {
	var req SendReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Parse report suggestions
	var analysis QueryAnalysisResult
	if err := json.Unmarshal([]byte(req.ReportJSON), &analysis); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse report JSON content: " + err.Error()})
		return
	}

	// 2. Fetch SMTP configurations from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpSender := os.Getenv("SMTP_SENDER")
	smtpRecipient := os.Getenv("SMTP_RECIPIENT")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || smtpRecipient == "" {
		log.Println("[SMTP] Missing SMTP configurations. Email skipped. (Check SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, SMTP_RECIPIENT env vars)")
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Report generated, but email skipped due to missing SMTP environment variables.",
		})
		return
	}

	if smtpSender == "" {
		smtpSender = smtpUser
	}

	// 3. Build HTML Body
	subject := fmt.Sprintf("Subject: [Query Optimizer Alert] Issue Detected on Report #%d\n", req.ReportID)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	
	var suggestionsHtml strings.Builder
	for i, sug := range analysis.Suggestions {
		suggestionsHtml.WriteString(fmt.Sprintf(`
			<div style="margin-bottom: 20px; padding: 15px; border-left: 4px solid #4F46E5; background-color: #F9FAFB;">
				<h4 style="margin: 0 0 10px 0; color: #1F2937;">Suggestion %d: %s</h4>
				<p style="margin: 0 0 10px 0; font-size: 14px; color: #4B5563;">%s</p>
				%s
			</div>
		`, i+1, strings.ToUpper(sug.Type), sug.Description, func() string {
			if sug.Sql != "" {
				return fmt.Sprintf(`
					<pre style="background-color: #111827; color: #F9FAFB; padding: 10px; border-radius: 6px; overflow-x: auto; font-family: monospace; font-size: 13px;">%s</pre>
				`, sug.Sql)
			}
			return ""
		}()))
	}

	htmlBody := fmt.Sprintf(`
		<html>
		<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; color: #374151; line-height: 1.5; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; border: 1px solid #E5E7EB; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1);">
				<div style="background-color: #4F46E5; color: white; padding: 20px; text-align: center;">
					<h2 style="margin: 0; font-weight: 600;">AI Query Optimizer Report</h2>
					<p style="margin: 5px 0 0 0; opacity: 0.9; font-size: 14px;">Report ID: #%d</p>
				</div>
				<div style="padding: 24px;">
					<h3 style="margin-top: 0; color: #111827;">Performance Bottleneck Found</h3>
					<p style="background-color: #FEF2F2; color: #991B1B; padding: 12px; border-radius: 6px; font-weight: 500; font-size: 15px;">
						%s
					</p>
					
					<h4 style="color: #374151;">Rationale</h4>
					<p style="font-size: 14px; color: #4B5563;">%s</p>
					
					<h4 style="color: #374151;">Target Query</h4>
					<pre style="background-color: #F3F4F6; padding: 12px; border-radius: 6px; font-family: monospace; font-size: 13px; overflow-x: auto; color: #1F2937;">%s</pre>
					
					<h3 style="color: #111827; border-bottom: 1px solid #E5E7EB; padding-bottom: 8px; margin-top: 30px;">Recommended Fixes</h3>
					%s
					
					<div style="margin-top: 30px; text-align: center;">
						<a href="http://localhost:3000/reports/%d" style="background-color: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 600; display: inline-block;">
							View in Dashboard
						</a>
					</div>
				</div>
				<div style="background-color: #F9FAFB; padding: 15px; border-top: 1px solid #E5E7EB; text-align: center; font-size: 12px; color: #9CA3AF;">
					Sent automatically by AI Query Optimizer Scheduler.
				</div>
			</div>
		</body>
		</html>
	`, req.ReportID, analysis.Issue, analysis.Rationale, req.QueryText, suggestionsHtml.String(), req.ReportID)

	msg := []byte(subject + mime + htmlBody)

	// 4. Authenticate and Send
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := smtpHost + ":" + smtpPort
	
	err = smtp.SendMail(addr, auth, smtpSender, []string{smtpRecipient}, msg)
	if err != nil {
		log.Printf("[SMTP] Failed to send email: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to send SMTP email: " + err.Error()})
		return
	}

	log.Printf("[SMTP] Report email sent successfully to %s", smtpRecipient)
	c.JSON(http.StatusOK, gin.H{"message": "Email report sent successfully!"})
}
