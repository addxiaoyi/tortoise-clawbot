package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ============ Gateway Cluster ============

// ClusterConfig 集群配置
type ClusterConfig struct {
	NodeID         string            // 节点 ID
	ClusterName    string            // 集群名称
	Peers          []string         // 对等节点地址
	AdvertiseAddr  string           // 广播地址
	AdvertisePort  int              // 广播端口
	JoinToken      string            // 加入令牌
	LeaderElection bool             // 启用领导者选举
	HeartbeatInterval time.Duration // 心跳间隔
	ElectionTimeout   time.Duration // 选举超时
	MaxPeers         int            // 最大对等节点数
}

// ClusterNode 集群节点
type ClusterNode struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Addr      string    `json:"addr"`
	Port      int       `json:"port"`
	Role      NodeRole  `json:"role"`
	State     NodeState `json:"state"`
	IsLeader  bool      `json:"is_leader"`
	StartTime time.Time `json:"start_time"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Metadata  map[string]string `json:"metadata"`
}

// NodeRole 节点角色
type NodeRole string

const (
	RoleLeader    NodeRole = "leader"
	RoleFollower  NodeRole = "follower"
	RoleCandidate NodeRole = "candidate"
	RoleObserver  NodeRole = "observer"
)

// NodeState 节点状态
type NodeState string

const (
	StateInit       NodeState = "init"
	StateJoining    NodeState = "joining"
	StateActive     NodeState = "active"
	StateLeaving    NodeState = "leaving"
	StateFailed     NodeState = "failed"
)

// Cluster 集群管理器
type Cluster struct {
	config   *ClusterConfig
	nodes    map[string]*ClusterNode
	mu       sync.RWMutex
	conns    map[string]*websocket.Conn
	ctx      context.Context
	cancel   context.CancelFunc
	handlers []ClusterEventHandler
	httpClient *http.Client
}

// ClusterEventHandler 集群事件处理器
type ClusterEventHandler func(event *ClusterEvent)

// ClusterEvent 集群事件
type ClusterEvent struct {
	Type      ClusterEventType `json:"type"`
	NodeID    string           `json:"node_id"`
	Node      *ClusterNode     `json:"node,omitempty"`
	Data      interface{}      `json:"data,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

// ClusterEventType 集群事件类型
type ClusterEventType string

const (
	EventNodeJoined   ClusterEventType = "node_joined"
	EventNodeLeft     ClusterEventType = "node_left"
	EventNodeFailed   ClusterEventType = "node_failed"
	EventLeaderChange ClusterEventType = "leader_change"
	EventSync         ClusterEventType = "sync"
)

// NewCluster 创建集群管理器
func NewCluster(config *ClusterConfig) *Cluster {
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 5 * time.Second
	}
	if config.ElectionTimeout == 0 {
		config.ElectionTimeout = 15 * time.Second
	}
	if config.MaxPeers == 0 {
		config.MaxPeers = 100
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	return &Cluster{
		config:      config,
		nodes:       make(map[string]*ClusterNode),
		conns:       make(map[string]*websocket.Conn),
		ctx:         ctx,
		cancel:      cancel,
		handlers:    make([]ClusterEventHandler, 0),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Start 启动集群
func (c *Cluster) Start() error {
	// 添加自己为节点
	c.addNode(&ClusterNode{
		ID:        c.config.NodeID,
		Name:      c.config.NodeID,
		Addr:      c.config.AdvertiseAddr,
		Port:      c.config.AdvertisePort,
		Role:      RoleCandidate,
		State:     StateJoining,
		StartTime: time.Now(),
		Metadata:  make(map[string]string),
	})
	
	// 加入集群
	if len(c.config.Peers) > 0 {
		for _, peer := range c.config.Peers {
			if err := c.joinPeer(peer); err != nil {
				log.Printf("⚠️ 加入对等节点失败 %s: %v", peer, err)
			}
		}
	}
	
	// 启动领导者选举
	if c.config.LeaderElection {
		go c.runLeaderElection()
	}
	
	// 启动心跳
	go c.runHeartbeat()
	
	// 启动状态同步
	go c.runStateSync()
	
	c.setNodeState(StateActive)
	
	log.Printf("✅ Gateway 集群已启动 (node: %s, peers: %d)", c.config.NodeID, len(c.config.Peers))
	return nil
}

// Stop 停止集群
func (c *Cluster) Stop() {
	c.setNodeState(StateLeaving)
	
	c.mu.Lock()
	for id, conn := range c.conns {
		conn.Close()
		delete(c.conns, id)
	}
	c.mu.Unlock()
	
	c.cancel()
	log.Printf("🛑 Gateway 集群已停止")
}

// joinPeer 加入对等节点
func (c *Cluster) joinPeer(addr string) error {
	url := fmt.Sprintf("http://%s/api/cluster/join", addr)
	
	reqData := map[string]interface{}{
		"node_id": c.config.NodeID,
		"addr":    c.config.AdvertiseAddr,
		"port":    c.config.AdvertisePort,
		"token":   c.config.JoinToken,
	}
	
	body, _ := json.Marshal(reqData)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("join failed: %d", resp.StatusCode)
	}
	
	return nil
}

// addNode 添加节点
func (c *Cluster) addNode(node *ClusterNode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	node.LastHeartbeat = time.Now()
	c.nodes[node.ID] = node
}

// removeNode 移除节点
func (c *Cluster) removeNode(nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if node, ok := c.nodes[nodeID]; ok {
		if conn, ok := c.conns[nodeID]; ok {
			conn.Close()
			delete(c.conns, nodeID)
		}
		delete(c.nodes, nodeID)
	}
}

// getNode 获取节点
func (c *Cluster) getNode(nodeID string) *ClusterNode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodes[nodeID]
}

// getLeader 获取领导者
func (c *Cluster) getLeader() *ClusterNode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	for _, node := range c.nodes {
		if node.IsLeader {
			return node
		}
	}
	return nil
}

// getAllNodes 获取所有节点
func (c *Cluster) getAllNodes() []*ClusterNode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	nodes := make([]*ClusterNode, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// setNodeState 设置节点状态
func (c *Cluster) setNodeState(state NodeState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if node, ok := c.nodes[c.config.NodeID]; ok {
		node.State = state
		node.LastHeartbeat = time.Now()
	}
}

// runLeaderElection 运行领导者选举
func (c *Cluster) runLeaderElection() {
	ticker := time.NewTicker(c.config.ElectionTimeout)
	defer ticker.Stop()
	
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.checkAndElectLeader()
		}
	}
}

// checkAndElectLeader 检查并选举领导者
func (c *Cluster) checkAndElectLeader() {
	leader := c.getLeader()
	
	// 检查领导者是否存活
	if leader != nil {
		c.mu.RLock()
		lastHB := leader.LastHeartbeat
		c.mu.RUnlock()
		
		if time.Since(lastHB) < c.config.ElectionTimeout {
			return // 领导者存活
		}
		log.Printf("⚠️ 领导者 %s 已失效", leader.ID)
	}
	
	// 开始选举
	c.startElection()
}

// startElection 开始选举
func (c *Cluster) startElection() {
	c.mu.Lock()
	
	// 将自己设为候选人
	if node, ok := c.nodes[c.config.NodeID]; ok {
		node.Role = RoleCandidate
	}
	
	// 计算任期
	term := time.Now().Unix()
	
	c.mu.Unlock()
	
	// 请求投票
	votes := 1 // 投自己一票
	
	nodes := c.getAllNodes()
	for _, node := range nodes {
		if node.ID == c.config.NodeID {
			continue
		}
		
		url := fmt.Sprintf("http://%s:%d/api/cluster/vote", node.Addr, node.Port)
		resp, err := c.httpClient.Post(url, "application/json", nil)
		if err != nil {
			continue
		}
		resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			votes++
		}
	}
	
	// 获得多数票成为领导者
	if votes > len(nodes)/2 {
		c.becomeLeader(term)
	}
}

// becomeLeader 成为领导者
func (c *Cluster) becomeLeader(term int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, node := range c.nodes {
		node.IsLeader = false
	}
	
	if node, ok := c.nodes[c.config.NodeID]; ok {
		node.Role = RoleLeader
		node.IsLeader = true
	}
	
	log.Printf("✅ 节点 %s 成为领导者 (term: %d)", c.config.NodeID, term)
	
	// 广播领导者变更
	c.emitEvent(&ClusterEvent{
		Type:   EventLeaderChange,
		NodeID: c.config.NodeID,
		Timestamp: time.Now(),
	})
}

// runHeartbeat 运行心跳
func (c *Cluster) runHeartbeat() {
	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat()
			c.checkPeers()
		}
	}
}

// sendHeartbeat 发送心跳
func (c *Cluster) sendHeartbeat() {
	c.mu.RLock()
	myNode := c.nodes[c.config.NodeID]
	c.mu.RUnlock()
	
	nodes := c.getAllNodes()
	for _, node := range nodes {
		if node.ID == c.config.NodeID {
			continue
		}
		
		url := fmt.Sprintf("http://%s:%d/api/cluster/heartbeat", node.Addr, node.Port)
		
		data := map[string]interface{}{
			"from_id": c.config.NodeID,
			"term":    time.Now().Unix(),
			"state":   myNode.State,
		}
		
		body, _ := json.Marshal(data)
		resp, err := c.httpClient.Post(url, "application/json", nil)
		if err != nil {
			continue
		}
		resp.Body.Close()
	}
}

// checkPeers 检查对等节点
func (c *Cluster) checkPeers() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for id, node := range c.nodes {
		if id == c.config.NodeID {
			continue
		}
		
		if time.Since(node.LastHeartbeat) > c.config.ElectionTimeout*2 {
			log.Printf("⚠️ 节点 %s 已失效", id)
			c.removeNode(id)
			
			c.emitEvent(&ClusterEvent{
				Type:   EventNodeFailed,
				NodeID: id,
				Node:   node,
				Timestamp: time.Now(),
			})
		}
	}
}

// runStateSync 运行状态同步
func (c *Cluster) runStateSync() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.syncState()
		}
	}
}

// syncState 同步状态
func (c *Cluster) syncState() {
	leader := c.getLeader()
	if leader == nil || leader.ID != c.config.NodeID {
		return // 只有领导者同步
	}
	
	c.emitEvent(&ClusterEvent{
		Type:   EventSync,
		Data:   c.getAllNodes(),
		Timestamp: time.Now(),
	})
}

// OnEvent 注册事件处理器
func (c *Cluster) OnEvent(handler ClusterEventHandler) {
	c.handlers = append(c.handlers, handler)
}

// emitEvent 发送事件
func (c *Cluster) emitEvent(event *ClusterEvent) {
	for _, handler := range c.handlers {
		go handler(event)
	}
}

// IsLeader 检查是否为领导者
func (c *Cluster) IsLeader() bool {
	node := c.getNode(c.config.NodeID)
	return node != nil && node.IsLeader
}

// GetNodeCount 获取节点数量
func (c *Cluster) GetNodeCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.nodes)
}

// GetClusterStatus 获取集群状态
type ClusterStatus struct {
	NodeID      string `json:"node_id"`
	ClusterName string `json:"cluster_name"`
	IsLeader    bool   `json:"is_leader"`
	NodeCount   int    `json:"node_count"`
	LeaderID    string `json:"leader_id,omitempty"`
	Peers       []*ClusterNode `json:"peers"`
}

func (c *Cluster) GetStatus() *ClusterStatus {
	leader := c.getLeader()
	
	return &ClusterStatus{
		NodeID:      c.config.NodeID,
		ClusterName: c.config.ClusterName,
		IsLeader:    c.IsLeader(),
		NodeCount:   c.GetNodeCount(),
		LeaderID:    func() string { if leader != nil { return leader.ID }; return "" }(),
		Peers:       c.getAllNodes(),
	}
}

// HandleHTTP HTTP 处理器
func (c *Cluster) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/cluster/join":
		c.handleJoin(w, r)
	case "/api/cluster/heartbeat":
		c.handleHeartbeat(w, r)
	case "/api/cluster/vote":
		c.handleVote(w, r)
	case "/api/cluster/status":
		c.handleStatus(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (c *Cluster) handleJoin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID string `json:"node_id"`
		Addr   string `json:"addr"`
		Port   int    `json:"port"`
		Token  string `json:"token"`
	}
	
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Token != c.config.JoinToken {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	
	c.addNode(&ClusterNode{
		ID:   req.NodeID,
		Addr: req.Addr,
		Port: req.Port,
		State: StateJoining,
		Metadata: make(map[string]string),
	})
	
	c.emitEvent(&ClusterEvent{
		Type:   EventNodeJoined,
		NodeID: req.NodeID,
		Timestamp: time.Now(),
	})
	
	w.WriteHeader(http.StatusOK)
}

func (c *Cluster) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromID string `json:"from_id"`
		Term   int64  `json:"term"`
		State  NodeState `json:"state"`
	}
	
	json.NewDecoder(r.Body).Decode(&req)
	
	if node := c.getNode(req.FromID); node != nil {
		c.mu.Lock()
		node.LastHeartbeat = time.Now()
		node.State = req.State
		c.mu.Unlock()
	}
	
	w.WriteHeader(http.StatusOK)
}

func (c *Cluster) handleVote(w http.ResponseWriter, r *http.Request) {
	// 简化投票逻辑
	w.WriteHeader(http.StatusOK)
}

func (c *Cluster) handleStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(c.GetStatus())
}
