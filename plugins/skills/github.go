// GitHub 集成技能
package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// GitHubSkill GitHub 集成技能
type GitHubSkill struct {
	token string
}

// NewGitHubSkill 创建 GitHub 技能
func NewGitHubSkill(token string) *GitHubSkill {
	return &GitHubSkill{token: token}
}

// Execute 执行技能
func (s *GitHubSkill) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Action   string `json:"action"`
		Owner    string `json:"owner,omitempty"`
		Repo     string `json:"repo,omitempty"`
		IssueNum int    `json:"issue_number,omitempty"`
		Title    string `json:"title,omitempty"`
		Body     string `json:"body,omitempty"`
		PRNumber int    `json:"pr_number,omitempty"`
		Comment  string `json:"comment,omitempty"`
		Branch   string `json:"branch,omitempty"`
		Message  string `json:"message,omitempty"`
		Path     string `json:"path,omitempty"`
		Content  string `json:"content,omitempty"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	switch params.Action {
	case "list_repos":
		return s.listRepos()
	case "get_repo":
		return s.getRepo(params.Owner, params.Repo)
	case "create_issue":
		return s.createIssue(params.Owner, params.Repo, params.Title, params.Body)
	case "list_issues":
		return s.listIssues(params.Owner, params.Repo)
	case "get_issue":
		return s.getIssue(params.Owner, params.Repo, params.IssueNum)
	case "close_issue":
		return s.closeIssue(params.Owner, params.Repo, params.IssueNum)
	case "create_comment":
		return s.createComment(params.Owner, params.Repo, params.IssueNum, params.Comment)
	case "create_pr":
		return s.createPR(params.Owner, params.Repo, params.Title, params.Body, params.Branch)
	case "list_prs":
		return s.listPRs(params.Owner, params.Repo)
	case "get_file":
		return s.getFile(params.Owner, params.Repo, params.Path)
	case "update_file":
		return s.updateFile(params.Owner, params.Repo, params.Path, params.Message, params.Content, params.Branch)
	default:
		return json.Marshal(map[string]interface{}{
			"error": "Unknown action: " + params.Action,
		})
	}
}

func (s *GitHubSkill) apiCall(method, url string, body interface{}) ([]byte, error) {
	client := &http.Client{}
	
	var reqBody *strings.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = strings.NewReader(string(jsonBody))
	} else {
		reqBody = strings.NewReader("")
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (s *GitHubSkill) listRepos() (json.RawMessage, error) {
	data, err := s.apiCall("GET", "https://api.github.com/user/repos", nil)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	var repos []map[string]interface{}
	json.Unmarshal(data, &repos)

	var names []string
	for _, repo := range repos {
		names = append(names, repo["full_name"].(string))
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"repos":   names,
	})
}

func (s *GitHubSkill) getRepo(owner, repo string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	data, err := s.apiCall("GET", url, nil)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

func (s *GitHubSkill) createIssue(owner, repo, title, body string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	data, err := s.apiCall("POST", url, map[string]interface{}{
		"title": title,
		"body":  body,
	})
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

func (s *GitHubSkill) listIssues(owner, repo string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	data, err := s.apiCall("GET", url, nil)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"issues":  string(data),
	})
}

func (s *GitHubSkill) getIssue(owner, repo string, num int) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, num)
	data, err := s.apiCall("GET", url, nil)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

func (s *GitHubSkill) closeIssue(owner, repo string, num int) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, num)
	data, err := s.apiCall("PATCH", url, map[string]interface{}{
		"state": "closed",
	})
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

func (s *GitHubSkill) createComment(owner, repo string, num int, comment string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, num)
	data, err := s.apiCall("POST", url, map[string]interface{}{
		"body": comment,
	})
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

func (s *GitHubSkill) createPR(owner, repo, title, body, branch string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	data, err := s.apiCall("POST", url, map[string]interface{}{
		"title": title,
		"body":  body,
		"head":  branch,
		"base":  "main",
	})
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

func (s *GitHubSkill) listPRs(owner, repo string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	data, err := s.apiCall("GET", url, nil)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"pulls":   string(data),
	})
}

func (s *GitHubSkill) getFile(owner, repo, path string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	data, err := s.apiCall("GET", url, nil)
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

func (s *GitHubSkill) updateFile(owner, repo, path, msg, content, branch string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	data, err := s.apiCall("PUT", url, map[string]interface{}{
		"message": msg,
		"content": content,
		"branch":  branch,
	})
	if err != nil {
		return json.Marshal(map[string]interface{}{"error": err.Error()})
	}

	return data, nil
}

// Metadata 返回技能元数据
func (s *GitHubSkill) Metadata() *SkillMetadata {
	return &SkillMetadata{
		Name:        "github",
		Description: "Interact with GitHub repositories",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to perform",
					"enum": []string{
						"list_repos", "get_repo", "create_issue", "list_issues",
						"get_issue", "close_issue", "create_comment", "create_pr",
						"list_prs", "get_file", "update_file",
					},
				},
				"owner":   map[string]interface{}{"type": "string"},
				"repo":    map[string]interface{}{"type": "string"},
				"title":   map[string]interface{}{"type": "string"},
				"body":    map[string]interface{}{"type": "string"},
				"comment": map[string]interface{}{"type": "string"},
				"path":    map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
				"branch":  map[string]interface{}{"type": "string"},
				"message": map[string]interface{}{"type": "string"},
			},
			"required": []string{"action"},
		},
	}
}
