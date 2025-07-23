// monitor/main.go
// GitHub Actions Real-Time Monitor for GitLab CI - Branch-based Version 1.1.0
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitHub API Response Structures
type GitHubWorkflowRun struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	RunStartedAt string `json:"run_started_at"`
	HTMLURL      string `json:"html_url"`
	HeadSHA      string `json:"head_sha"`
	HeadBranch   string `json:"head_branch"`
	Event        string `json:"event"`
	RunNumber    int    `json:"run_number"`
	RunAttempt   int    `json:"run_attempt"`
}

type GitHubWorkflowRunsResponse struct {
	TotalCount   int                 `json:"total_count"`
	WorkflowRuns []GitHubWorkflowRun `json:"workflow_runs"`
}

type GitHubJob struct {
	ID          int             `json:"id"`
	RunID       int             `json:"run_id"`
	Name        string          `json:"name"`
	Status      string          `json:"status"`
	Conclusion  string          `json:"conclusion"`
	CreatedAt   string          `json:"created_at"`
	StartedAt   string          `json:"started_at"`
	CompletedAt string          `json:"completed_at"`
	HTMLURL     string          `json:"html_url"`
	Steps       []GitHubJobStep `json:"steps"`
}

type GitHubJobStep struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	Number      int    `json:"number"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

type GitHubJobsResponse struct {
	TotalCount int         `json:"total_count"`
	Jobs       []GitHubJob `json:"jobs"`
}

// GitHubActionsMonitor - Real-time monitor for GitHub Actions
type GitHubActionsMonitor struct {
	GitHubToken     string
	GitHubRepo      string
	GitLabToken     string
	GitLabProjectID string
	BranchName      string // Changed from CommitSHA to BranchName
	Environment     string // New: environment name (dev, qa, prod, next)
	LogFile         string
	APILogFile      string
	PollInterval    time.Duration
	HTTPClient      *http.Client
	LogsDir         string
}

// NewGitHubActionsMonitor creates a new monitor instance
func NewGitHubActionsMonitor() *GitHubActionsMonitor {
	pollInterval := 10 * time.Second
	if interval := os.Getenv("POLL_INTERVAL"); interval != "" {
		if parsed, err := time.ParseDuration(interval); err == nil {
			pollInterval = parsed
		}
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "unknown"
	}

	logsDir := "gitlab-logs"
	return &GitHubActionsMonitor{
		GitHubToken:     os.Getenv("GITHUB_TOKEN"),
		GitHubRepo:      os.Getenv("GITHUB_REPO"),
		GitLabToken:     os.Getenv("GITLAB_TOKEN"),
		GitLabProjectID: os.Getenv("GITLAB_PROJECT_ID"),
		BranchName:      os.Getenv("BRANCH_NAME"), // Read branch name from env
		Environment:     environment,
		LogsDir:         logsDir,
		LogFile:         filepath.Join(logsDir, fmt.Sprintf("github-deployment-%s.log", environment)),
		APILogFile:      filepath.Join(logsDir, fmt.Sprintf("github-api-responses-%s.log", environment)),
		PollInterval:    pollInterval,
		HTTPClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

// ensureLogDir creates the logs directory with proper permissions
func (gm *GitHubActionsMonitor) ensureLogDir() error {
	// Create directory with full permissions
	if err := os.MkdirAll(gm.LogsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Verify directory exists and is writable
	if info, err := os.Stat(gm.LogsDir); err != nil {
		return fmt.Errorf("logs directory not accessible: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("logs path exists but is not a directory")
	}

	// Test write permissions by creating a test file
	testFile := filepath.Join(gm.LogsDir, "test_write.tmp")
	if file, err := os.Create(testFile); err != nil {
		return fmt.Errorf("logs directory not writable: %w", err)
	} else {
		file.Close()
		os.Remove(testFile) // Clean up test file
	}

	return nil
}

// writeLog writes to both console and log file with better error handling
func (gm *GitHubActionsMonitor) writeLog(message string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// Always write to console first (visible in GitLab CI)
	fmt.Print(logEntry)

	// Ensure log directory exists
	if err := gm.ensureLogDir(); err != nil {
		fmt.Printf("ERROR: Cannot create log directory: %v\n", err)
		return err
	}

	// Write to log file with better error handling
	file, err := os.OpenFile(gm.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("ERROR: Cannot open log file %s: %v\n", gm.LogFile, err)
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(logEntry); err != nil {
		fmt.Printf("ERROR: Cannot write to log file: %v\n", err)
		return err
	}

	// Force flush to disk
	if err := file.Sync(); err != nil {
		fmt.Printf("WARNING: Cannot sync log file: %v\n", err)
	}

	return nil
}

// writeAPILog saves GitHub API responses for debugging
func (gm *GitHubActionsMonitor) writeAPILog(endpoint string, response interface{}) error {
	if err := gm.ensureLogDir(); err != nil {
		return err
	}

	file, err := os.OpenFile(gm.APILogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("ERROR: Cannot open API log file: %v\n", err)
		return err
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	jsonData, _ := json.MarshalIndent(response, "", "  ")

	logEntry := fmt.Sprintf("\n[%s] GET %s\n%s\n%s\n",
		timestamp, endpoint, string(jsonData), strings.Repeat("-", 80))

	_, err = file.WriteString(logEntry)
	if err != nil {
		fmt.Printf("ERROR: Cannot write to API log file: %v\n", err)
	}
	return err
}

// makeGitHubRequest makes authenticated requests to GitHub API
func (gm *GitHubActionsMonitor) makeGitHubRequest(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+gm.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "GitLab-GitHub-Monitor/1.1")

	return gm.HTTPClient.Do(req)
}

// getWorkflowRuns fetches all workflow runs
func (gm *GitHubActionsMonitor) getWorkflowRuns() (*GitHubWorkflowRunsResponse, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs", gm.GitHubRepo)

	resp, err := gm.makeGitHubRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	var response GitHubWorkflowRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	gm.writeAPILog(endpoint, response)
	return &response, nil
}

// getSpecificWorkflowRun fetches detailed workflow run information
func (gm *GitHubActionsMonitor) getSpecificWorkflowRun(runID int) (*GitHubWorkflowRun, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%d", gm.GitHubRepo, runID)

	resp, err := gm.makeGitHubRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	var run GitHubWorkflowRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	gm.writeAPILog(endpoint, run)
	return &run, nil
}

// getWorkflowJobs fetches all jobs for a workflow run
func (gm *GitHubActionsMonitor) getWorkflowJobs(runID int) (*GitHubJobsResponse, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%d/jobs", gm.GitHubRepo, runID)

	resp, err := gm.makeGitHubRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	var response GitHubJobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	gm.writeAPILog(endpoint, response)
	return &response, nil
}

// updateGitLabStatus updates GitLab external pipeline status
func (gm *GitHubActionsMonitor) updateGitLabStatus(state, description, targetURL, commitSHA string) error {
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/statuses/%s", gm.GitLabProjectID, commitSHA)

	payload := map[string]string{
		"state":       state,
		"context":     fmt.Sprintf("GitHub Actions - %s Environment", strings.Title(gm.Environment)),
		"description": description,
		"target_url":  targetURL,
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("PRIVATE-TOKEN", gm.GitLabToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := gm.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// getStatusSymbol returns emoji for status
func (gm *GitHubActionsMonitor) getStatusSymbol(status, conclusion string) string {
	switch status {
	case "queued":
		return "⏳"
	case "in_progress":
		return "🔄"
	case "completed":
		switch conclusion {
		case "success":
			return "✅"
		case "failure":
			return "❌"
		case "cancelled":
			return "⚠️"
		default:
			return "❓"
		}
	default:
		return "❓"
	}
}

// mapToGitLabState converts GitHub status to GitLab state
func (gm *GitHubActionsMonitor) mapToGitLabState(status, conclusion string) string {
	switch status {
	case "queued":
		return "pending"
	case "in_progress":
		return "running"
	case "completed":
		switch conclusion {
		case "success":
			return "success"
		case "failure":
			return "failed"
		case "cancelled":
			return "canceled"
		default:
			return "failed"
		}
	default:
		return "pending"
	}
}

// analyzeFailure provides detailed failure analysis
func (gm *GitHubActionsMonitor) analyzeFailure(run *GitHubWorkflowRun) {
	gm.writeLog("🔍 FAILURE ANALYSIS: Analyzing GitHub Actions deployment failure...")

	jobs, err := gm.getWorkflowJobs(run.ID)
	if err != nil {
		gm.writeLog(fmt.Sprintf("❌ Could not fetch job details: %v", err))
		return
	}

	gm.writeLog(fmt.Sprintf("📊 Total jobs in workflow: %d", jobs.TotalCount))

	for _, job := range jobs.Jobs {
		gm.writeLog(fmt.Sprintf("📋 Job: %s", job.Name))
		gm.writeLog(fmt.Sprintf("   Status: %s | Conclusion: %s", job.Status, job.Conclusion))

		if job.Conclusion == "failure" {
			gm.writeLog("   ❌ FAILED JOB - Step-by-step analysis:")
			for _, step := range job.Steps {
				symbol := gm.getStatusSymbol(step.Status, step.Conclusion)
				gm.writeLog(fmt.Sprintf("      %s Step %d: %s (%s)",
					symbol, step.Number, step.Name, step.Status))

				if step.Conclusion == "failure" {
					gm.writeLog(fmt.Sprintf("         ❌ FAILURE POINT: %s", step.Name))
					if step.StartedAt != "" && step.CompletedAt != "" {
						startTime, _ := time.Parse(time.RFC3339, step.StartedAt)
						endTime, _ := time.Parse(time.RFC3339, step.CompletedAt)
						duration := endTime.Sub(startTime)
						gm.writeLog(fmt.Sprintf("         ⏱️ Failed after: %v", duration.Round(time.Second)))
					}
				}
			}
		}

		gm.writeLog(fmt.Sprintf("   🔗 Job URL: %s", job.HTMLURL))
		gm.writeLog("   " + strings.Repeat("-", 50))
	}
}

// logDetailedStatus provides comprehensive status information
func (gm *GitHubActionsMonitor) logDetailedStatus(run *GitHubWorkflowRun) {
	jobs, err := gm.getWorkflowJobs(run.ID)
	if err != nil {
		gm.writeLog(fmt.Sprintf("⚠️ Could not fetch job details: %v", err))
		return
	}

	var queued, inProgress, completed, failed int
	var currentJobs []string

	for _, job := range jobs.Jobs {
		switch job.Status {
		case "queued":
			queued++
		case "in_progress":
			inProgress++
			currentJobs = append(currentJobs, job.Name)
		case "completed":
			completed++
			if job.Conclusion == "failure" {
				failed++
			}
		}
	}

	gm.writeLog(fmt.Sprintf("📊 Workflow: %s | Run #%d | Attempt #%d",
		run.Name, run.RunNumber, run.RunAttempt))

	if run.RunStartedAt != "" && run.UpdatedAt != "" {
		startTime, _ := time.Parse(time.RFC3339, run.RunStartedAt)
		updateTime, _ := time.Parse(time.RFC3339, run.UpdatedAt)
		duration := updateTime.Sub(startTime)
		gm.writeLog(fmt.Sprintf("⏱️ Duration: %v | Event: %s", duration.Round(time.Second), run.Event))
	}

	gm.writeLog(fmt.Sprintf("📈 Jobs Status: %d total | %d queued | %d running | %d completed | %d failed",
		jobs.TotalCount, queued, inProgress, completed, failed))

	if len(currentJobs) > 0 {
		gm.writeLog(fmt.Sprintf("🔄 Currently running: %s", strings.Join(currentJobs, ", ")))
	}
}

// validateConfig validates environment variables
func (gm *GitHubActionsMonitor) validateConfig() error {
	if gm.GitHubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN is required")
	}
	if gm.GitHubRepo == "" {
		return fmt.Errorf("GITHUB_REPO is required")
	}
	if gm.GitLabToken == "" {
		return fmt.Errorf("GITLAB_TOKEN is required")
	}
	if gm.GitLabProjectID == "" {
		return fmt.Errorf("GITLAB_PROJECT_ID is required")
	}
	if gm.BranchName == "" {
		return fmt.Errorf("BRANCH_NAME is required")
	}
	return nil
}

// startMonitoring starts the real-time monitoring process
func (gm *GitHubActionsMonitor) startMonitoring() error {
	// Test log file creation immediately
	gm.writeLog("🧪 Testing log file creation...")

	if err := gm.validateConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	gm.writeLog("🚀 GitLab CI: Real-Time GitHub Actions Monitor Started")
	gm.writeLog(fmt.Sprintf("📁 GitHub Repository: %s", gm.GitHubRepo))
	gm.writeLog(fmt.Sprintf("🌿 Monitoring Branch: %s", gm.BranchName))
	gm.writeLog(fmt.Sprintf("🏗️ Environment: %s", strings.ToUpper(gm.Environment)))
	gm.writeLog(fmt.Sprintf("⏰ Polling every: %v", gm.PollInterval))
	gm.writeLog(fmt.Sprintf("📝 Log file: %s", gm.LogFile))
	gm.writeLog(fmt.Sprintf("📊 API log file: %s", gm.APILogFile))
	gm.writeLog("👥 GitLab developers can see GitHub deployment status here")
	gm.writeLog(strings.Repeat("=", 60))

	var lastStatus, lastConclusion string
	var lastRunID int
	monitoringStart := time.Now()

	// Main monitoring loop with regular logging
	for {
		currentTime := time.Now()
		elapsed := currentTime.Sub(monitoringStart)

		// Log every 10 seconds regardless of status changes
		gm.writeLog(fmt.Sprintf("⏰ Monitoring tick - Elapsed: %v", elapsed.Round(time.Second)))

		runs, err := gm.getWorkflowRuns()
		if err != nil {
			gm.writeLog(fmt.Sprintf("❌ Error fetching workflow runs: %v", err))
			time.Sleep(gm.PollInterval)
			continue
		}

		gm.writeLog(fmt.Sprintf("📊 Found %d workflow runs", runs.TotalCount))

		var currentRun *GitHubWorkflowRun
		for _, run := range runs.WorkflowRuns {
			// Modified: Match by branch name instead of commit SHA
			if run.HeadBranch == gm.BranchName {
				detailed, err := gm.getSpecificWorkflowRun(run.ID)
				if err == nil {
					currentRun = detailed
					break
				}
			}
		}

		if currentRun == nil {
			gm.writeLog(fmt.Sprintf("⏳ No workflow found for branch %s yet...", gm.BranchName))
			time.Sleep(gm.PollInterval)
			continue
		}

		// Always log current status
		statusChanged := currentRun.Status != lastStatus ||
			currentRun.Conclusion != lastConclusion ||
			currentRun.ID != lastRunID

		symbol := gm.getStatusSymbol(currentRun.Status, currentRun.Conclusion)
		statusMsg := fmt.Sprintf("%s GitHub Actions (%s): %s", symbol, strings.ToUpper(gm.Environment), strings.ToUpper(currentRun.Status))

		if currentRun.Status == "completed" && currentRun.Conclusion != "" {
			statusMsg += fmt.Sprintf(" (%s)", strings.ToUpper(currentRun.Conclusion))
		}

		statusMsg += fmt.Sprintf(" | Run ID: %d | Branch: %s", currentRun.ID, currentRun.HeadBranch)
		gm.writeLog(statusMsg)

		if statusChanged {
			gm.writeLog(fmt.Sprintf("🔗 GitHub URL: %s", currentRun.HTMLURL))

			// Log detailed status
			gm.logDetailedStatus(currentRun)

			// Update GitLab external status
			gitlabState := gm.mapToGitLabState(currentRun.Status, currentRun.Conclusion)
			description := fmt.Sprintf("GitHub Actions (%s): %s", gm.Environment, currentRun.Status)
			if currentRun.Conclusion != "" {
				description += fmt.Sprintf(" (%s)", currentRun.Conclusion)
			}

			if err := gm.updateGitLabStatus(gitlabState, description, currentRun.HTMLURL, currentRun.HeadSHA); err != nil {
				gm.writeLog(fmt.Sprintf("⚠️ Warning: GitLab status update failed: %v", err))
			} else {
				gm.writeLog(fmt.Sprintf("✅ GitLab external status updated: %s", gitlabState))
			}

			lastStatus = currentRun.Status
			lastConclusion = currentRun.Conclusion
			lastRunID = currentRun.ID
		}

		// Handle completion
		if currentRun.Status == "completed" {
			totalDuration := time.Since(monitoringStart).Round(time.Second)
			symbol := gm.getStatusSymbol(currentRun.Status, currentRun.Conclusion)

			gm.writeLog(fmt.Sprintf("%s GitHub Actions deployment completed: %s",
				symbol, strings.ToUpper(currentRun.Conclusion)))

			if currentRun.Conclusion == "success" {
				gm.writeLog(fmt.Sprintf("🎉 GitHub Actions deployment to %s SUCCESSFUL!", strings.ToUpper(gm.Environment)))
			} else if currentRun.Conclusion == "failure" {
				gm.writeLog(fmt.Sprintf("💥 GitHub Actions deployment to %s FAILED!", strings.ToUpper(gm.Environment)))
				gm.analyzeFailure(currentRun)
			} else if currentRun.Conclusion == "cancelled" {
				gm.writeLog(fmt.Sprintf("⚠️ GitHub Actions deployment to %s was CANCELLED", strings.ToUpper(gm.Environment)))
			}

			gm.writeLog(fmt.Sprintf("⏱️ Total monitoring duration: %v", totalDuration))
			gm.writeLog(strings.Repeat("=", 60))
			gm.writeLog("🏁 Real-time monitoring completed!")
			gm.writeLog("📊 Complete GitHub API responses saved to: " + gm.APILogFile)
			gm.writeLog("📋 This log available as GitLab CI artifact")
			break
		}

		// Add timeout check
		if elapsed > 45*time.Minute {
			gm.writeLog("⏰ Monitoring timeout reached (45 minutes)")
			break
		}

		time.Sleep(gm.PollInterval)
	}

	return nil
}

func main() {
	monitor := NewGitHubActionsMonitor()

	if err := monitor.startMonitoring(); err != nil {
		log.Printf("❌ GitHub Actions monitoring failed: %v", err)
		os.Exit(1)
	}
}
