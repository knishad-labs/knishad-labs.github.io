package controller

import (
	"net/http"
	"strconv"

	"scheduler-service/internal/cron"
	"scheduler-service/internal/model"
	"scheduler-service/internal/repository"
	"github.com/gin-gonic/gin"
)

type ScheduledTaskInput struct {
	Name                 string   `json:"name" binding:"required"`
	CronExpression       string   `json:"cron_expression" binding:"required"`
	DatabaseConnectionID uint     `json:"database_connection_id" binding:"required"`
	AIProviderID         uint     `json:"ai_provider_id" binding:"required"`
	TargetQueries        []string `json:"target_queries" binding:"required"`
	SkillRules           string   `json:"skill_rules"`
	IsActive             *bool    `json:"is_active" binding:"required"`
}

// GetTasks returns all scheduled optimization tasks
func GetTasks(c *gin.Context) {
	var tasks []model.ScheduledTask
	if err := repository.DB.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scheduled tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTask returns a single scheduled optimization task by ID
func GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task model.ScheduledTask
	if err := repository.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// CreateTask adds a new scheduled task and queues it in the cron engine
func CreateTask(c *gin.Context) {
	var input ScheduledTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := model.ScheduledTask{
		Name:                 input.Name,
		CronExpression:       input.CronExpression,
		DatabaseConnectionID: input.DatabaseConnectionID,
		AIProviderID:         input.AIProviderID,
		TargetQueries:        model.StringArray(input.TargetQueries),
		SkillRules:           input.SkillRules,
		IsActive:             *input.IsActive,
	}

	if err := repository.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduled task"})
		return
	}

	// Schedule in Cron engine if active
	if task.IsActive {
		cron.ScheduleTask(task)
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask updates an existing task and updates the cron engine queue
func UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task model.ScheduledTask
	if err := repository.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var input ScheduledTaskInput // reuse validator
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Unschedule current from cron
	cron.UnscheduleTask(task.ID)

	task.Name = input.Name
	task.CronExpression = input.CronExpression
	task.DatabaseConnectionID = input.DatabaseConnectionID
	task.AIProviderID = input.AIProviderID
	task.TargetQueries = model.StringArray(input.TargetQueries)
	task.SkillRules = input.SkillRules
	task.IsActive = *input.IsActive

	if err := repository.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scheduled task"})
		return
	}

	// Reschedule in Cron if still active
	if task.IsActive {
		cron.ScheduleTask(task)
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask removes a task and stops its cron execution
func DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task model.ScheduledTask
	if err := repository.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Unschedule
	cron.UnscheduleTask(task.ID)

	if err := repository.DB.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled task deleted successfully"})
}
