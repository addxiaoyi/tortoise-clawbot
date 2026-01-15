package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TailscaleService Tailscale 网络发现服务
type TailscaleService struct {
	config     *TailscaleConfig
	httpClient *http.Client
	nodes      map[string]*TailscaleNode
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// TailscaleConfig Tailscale 配置
type TailscaleConfig struct {
	APIKey          string        // Tailscale API Key
	Tailnet         string        // Tailnet 名称 (如 "example.com")
	UseTailscaled   bool          // 是否使用本地 tailscaled
	SocketPath      string        // tailscaled socket 路径
	PollInterval    time.Duration // 节点发现间隔
	FilterACL       []string      // ACL 过滤规则
	EnableDERP      bool          // 启用 DERP 中继
}

// TailscaleNode Tailscale 节点
type TailscaleNode struct {
	NodeID       string            `json:"nodeId"`
	Name         string            `json:"name"`
	Hostname     string            `json:"hostname"`
	IPv4         string            `json:"ipv4"`
	IPv6         string            `json:"ipv6"`
	TailnetAddr  string            `json:"tailnetAddr"`
	Online       bool              `json:"online"`
	LastSeen     time.Time         `json:"lastSeen"`
	OS           string            `json:"os"`
	Version      string            `json:"version"`
	ACLTags      []string          `json:"aclTags"`
	Roles        []string          `json:"roles"`
	Metadata     map[string]string `json:"metadata"`
}

// NewTailscaleService 创建 Tailscale 服务
func NewTailscaleService(config *TailscaleConfig) *TailscaleService {
	if config.PollInterval == 0 {
		config.PollInterval = 60 * time.Second
	}
	if config.SocketPath == "" {
		config.SocketPath = "/var/run/tailscale/tailscaled.sock"
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	return &TailscaleService{
		config:     config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		nodes:      make(map[string]*TailscaleNode),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start 启动服务
func (s *TailscaleService) Start() error {
	// 尝试本地 tailscaled
	if s.config.UseTailscaled {
		if err := s.checkLocalDaemon(); err != nil {
			log.Printf("⚠️ 本地 tailscaled 未运行: %v", err)
		}
	}
	
	// 启动节点发现
	go s.discoverNodes()
	
	log.Printf("✅ Tailscale 服务已启动 (tailnet: %s)", s.config.Tailnet)
	return nil
}

// Stop 停止服务
func (s *TailscaleService) Stop() {
	s.cancel()
	log.Printf("🛑 Tailscale 服务已停止")
}

// checkLocalDaemon 检查本地 tailscaled
func (s *TailscaleService) checkLocalDaemon() error {
	// 检查 tailscaled 进程
	cmd := exec.Command("tailscale", "status")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	if strings.Contains(string(output), "Logged out") {
		return fmt.Errorf("tailscale 未登录")
	}
	
	return nil
}

// discoverNodes 发现节点
func (s *TailscaleService) discoverNodes() {
	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.refreshNodes()
		}
	}
}

// refreshNodes 刷新节点列表
func (s *TailscaleService) refreshNodes() {
	if s.config.APIKey != "" {
		s.refreshViaAPI()
	} else if s.config.UseTailscaled {
		s.refreshViaLocal()
	}
}

// refreshViaAPI 通过 API 刷新
func (s *TailscaleService) refreshViaAPI() {
	url := fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/-/devices")
	
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("❌ Tailscale API 请求失败: %v", err)
		return
	}
	defer resp.Body.Close()
	
	var devices []TailscaleNode
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		log.Printf("❌ 解析 Tailscale 响应失败: %v", err)
		return
	}
	
	s.mu.Lock()
	for _, d := range devices {
		s.nodes[d.NodeID] = &d
	}
	s.mu.Unlock()
	
	log.Printf("💓 Tailscale 节点刷新: %d 节点", len(devices))
}

// refreshViaLocal 通过本地命令刷新
func (s *TailscaleService) refreshViaLocal() {
	cmd := exec.Command("tailscale", "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("❌ tailscale status 失败: %v", err)
		return
	}
	
	var status struct {
		Peer map[string]struct {
			ID         string `json:"ID"`
			HostName   string `json:"HostName"`
			DNSName    string `json:"DNSName"`
			IPv4       string `json:"IPv4"`
			IPv6       string `json:"IPv6"`
			Online     bool   `json:"Online"`
			LastSeen   int64  `json:"LastSeen"`
			OS         string `json:"OS"`
			Version    string `json:"Version"`
		} `json:"Peer"`
	}
	
	if err := json.Unmarshal(output, &status); err != nil {
		log.Printf("❌ 解析 tailscale 状态失败: %v", err)
		return
	}
	
	s.mu.Lock()
	for id, p := range status.Peer {
		s.nodes[id] = &TailscaleNode{
			NodeID:       id,
			Hostname:     p.HostName,
			TailnetAddr:  p.DNSName,
			IPv4:         p.IPv4,
			IPv6:         p.IPv6,
			Online:       p.Online,
			LastSeen:     time.Unix(p.LastSeen/1000, 0),
			OS:           p.OS,
			Version:      p.Version,
		}
	}
	s.mu.Unlock()
}

// GetNodes 获取所有节点
func (s *TailscaleService) GetNodes() []*TailscaleNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	nodes := make([]*TailscaleNode, 0, len(s.nodes))
	for _, n := range s.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

// GetOnlineNodes 获取在线节点
func (s *TailscaleService) GetOnlineNodes() []*TailscaleNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	nodes := make([]*TailscaleNode, 0)
	for _, n := range s.nodes {
		if n.Online {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

// GetNodeByHostname 通过主机名获取节点
func (s *TailscaleService) GetNodeByHostname(hostname string) *TailscaleNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, n := range s.nodes {
		if n.Hostname == hostname || n.Name == hostname {
			return n
		}
	}
	return nil
}

// GetNodeByRole 通过角色获取节点
func (s *TailscaleService) GetNodeByRole(role string) []*TailscaleNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	nodes := make([]*TailscaleNode, 0)
	for _, n := range s.nodes {
		for _, r := range n.Roles {
			if r == role {
				nodes = append(nodes, n)
				break
			}
		}
	}
	return nodes
}

// ConnectToNode 连接到节点
func (s *TailscaleService) ConnectToNode(hostname string) (net.Conn, error) {
	node := s.GetNodeByHostname(hostname)
	if node == nil {
		return nil, fmt.Errorf("节点未找到: %s", hostname)
	}
	
	if !node.Online {
		return nil, fmt.Errorf("节点不在线: %s", hostname)
	}
	
	// 使用 tailnet 地址连接
	addr := strings.TrimSuffix(node.TailnetAddr, ".")
	conn, err := net.DialTimeout("tcp", addr+":22", 10*time.Second)
	if err != nil {
		return nil, err
	}
	
	return conn, nil
}

// RunCommand 在远程节点执行命令
func (s *TailscaleService) RunCommand(hostname string, cmd string) (string, error) {
	node := s.GetNodeByHostname(hostname)
	if node == nil {
		return "", fmt.Errorf("节点未找到: %s", hostname)
	}
	
	// 使用 tailscale ssh 执行
	sshCmd := exec.Command("tailscale", "ssh", "--exit-node", node.Hostname, "hostname")
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	
	return string(output), nil
}

// PingNode Ping 节点
func (s *TailscaleService) PingNode(hostname string) (time.Duration, error) {
	node := s.GetNodeByHostname(hostname)
	if node == nil {
		return 0, fmt.Errorf("节点未找到: %s", hostname)
	}
	
	addr := node.TailnetAddr
	if node.IPv4 != "" {
		addr = node.IPv4
	}
	
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr+":80", 5*time.Second)
	if err != nil {
		return 0, err
	}
	conn.Close()
	
	return time.Since(start), nil
}

// FilterNodes 过滤节点
func (s *TailscaleService) FilterNodes(predicate func(*TailscaleNode) bool) []*TailscaleNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	nodes := make([]*TailscaleNode, 0)
	for _, n := range s.nodes {
		if predicate(n) {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

// GetNetworkMap 获取网络拓扑
type NetworkMap struct {
	Nodes     []*TailscaleNode `json:"nodes"`
	Edges     []NetworkEdge    `json:"edges"`
	Timestamp time.Time        `json:"timestamp"`
}

type NetworkEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Latency int64 `json:"latency"`
}

func (s *TailscaleService) GetNetworkMap() *NetworkMap {
	nodes := s.GetNodes()
	edges := make([]NetworkEdge, 0)
	
	// 构建网络边
	for _, from := range nodes {
		for _, to := range nodes {
			if from.NodeID != to.NodeID && from.Online && to.Online {
				edges = append(edges, NetworkEdge{
					From: from.Hostname,
					To:   to.Hostname,
				})
			}
		}
	}
	
	return &NetworkMap{
		Nodes:     nodes,
		Edges:     edges,
		Timestamp: time.Now(),
	}
}
