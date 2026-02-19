package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/javiergonzalez/adusa-tui/internal/types"
)

type Credentials struct {
	Email  string
	Token  string
	Domain string
}

type JiraResponse struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description any    `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Assignee struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
	} `json:"fields"`
}

func LoadCredentials() (*Credentials, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	scratchDir := filepath.Join(homeDir, ".scratch")

	emailPath := filepath.Join(scratchDir, "jira.email")
	tokenPath := filepath.Join(scratchDir, "jira.token")

	emailBytes, err := os.ReadFile(emailPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read jira.email: %w", err)
	}
	email := strings.TrimSpace(string(emailBytes))

	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read jira.token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))

	domainPath := filepath.Join(scratchDir, "jira.domain")
	domainBytes, err := os.ReadFile(domainPath)
	var domain string
	if err != nil {
		domain = "atlassian.com"
	} else {
		domain = strings.TrimSpace(string(domainBytes))
	}

	return &Credentials{
		Email:  email,
		Token:  token,
		Domain: domain,
	}, nil
}

func FetchTicket(key string) (*types.TicketInfo, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/issue/%s", creds.Domain, key)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(creds.Email, creds.Token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var jiraResp JiraResponse
	if err := json.NewDecoder(resp.Body).Decode(&jiraResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	description := ""
	if jiraResp.Fields.Description != nil {
		switch v := jiraResp.Fields.Description.(type) {
		case string:
			description = v
		case map[string]any:
			if content, ok := v["content"].([]any); ok {
				var sb strings.Builder
				for _, block := range content {
					if blockMap, ok := block.(map[string]any); ok {
						if content2, ok := blockMap["content"].([]any); ok {
							for _, item := range content2 {
								if itemMap, ok := item.(map[string]any); ok {
									if text, ok := itemMap["text"].(string); ok {
										sb.WriteString(text)
									}
								}
							}
						}
					}
				}
				description = sb.String()
			}
		}
	}

	return &types.TicketInfo{
		Key:         jiraResp.Key,
		Summary:     jiraResp.Fields.Summary,
		Description: description,
		Status:      jiraResp.Fields.Status.Name,
		Assignee:    jiraResp.Fields.Assignee.DisplayName,
		Priority:    jiraResp.Fields.Priority.Name,
	}, nil
}

func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, ".scratch", "tickets")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cacheDir, nil
}

func SaveTicketCache(ticket *types.TicketInfo) error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s.md", ticket.Key))

	content := fmt.Sprintf("# %s\n\n**Summary:** %s\n\n**Status:** %s\n**Assignee:** %s\n**Priority:** %s\n\n## Description\n\n%s\n",
		ticket.Key,
		ticket.Summary,
		ticket.Status,
		ticket.Assignee,
		ticket.Priority,
		ticket.Description,
	)

	if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

func LoadTicketCache(key string) (*types.TicketInfo, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s.md", key))

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	content := string(data)

	ticket := &types.TicketInfo{
		Key: key,
	}

	lines := strings.Split(content, "\n")
	inDescription := false
	var descLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "**Summary:**") {
			ticket.Summary = strings.TrimSpace(strings.TrimPrefix(line, "**Summary:**"))
		} else if strings.HasPrefix(line, "**Status:**") {
			ticket.Status = strings.TrimSpace(strings.TrimPrefix(line, "**Status:**"))
		} else if strings.HasPrefix(line, "**Assignee:**") {
			ticket.Assignee = strings.TrimSpace(strings.TrimPrefix(line, "**Assignee:**"))
		} else if strings.HasPrefix(line, "**Priority:**") {
			ticket.Priority = strings.TrimSpace(strings.TrimPrefix(line, "**Priority:**"))
		} else if strings.HasPrefix(line, "## Description") {
			inDescription = true
		} else if inDescription {
			descLines = append(descLines, line)
		}
	}

	ticket.Description = strings.TrimSpace(strings.Join(descLines, "\n"))

	return ticket, nil
}
