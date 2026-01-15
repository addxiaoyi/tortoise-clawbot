package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ServiceType 服务类型
type ServiceType string

const (
	TypeGateway ServiceType = "_tortoise._tcp"
	TypeAgent   ServiceType = "_tortoise-agent._tcp"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID        string
	Name     string
	Type     ServiceType
	Host     string
	Port     int
	Metadata map[string]string
}

// Config 发现服务配置
type Config struct {
	EnableMDNS bool
	EnableUPnP bool
	EnableSSDP bool
	Domain    string
	Port      int
}

// Service 发现服务
type Service struct {
	config Config

	// 本地服务
	localServices map[string]*ServiceInfo

	// 发现的服务
	discovered map[string]*ServiceInfo

	// 监听器
	listeners map[string]chan *ServiceInfo

	// 控制
	ctx    context.Context
	cancel context.CancelFunc

	// 统计
	stats Stats

	mu sync.RWMutex
}

// Stats 发现服务统计
type Stats struct {
	ServicesPublished  atomic.Int64
	ServicesDiscovered atomic.Int64
	LookupLatencyUs   atomic.Int64
}

// NewService 创建发现服务
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Service{
		config:        cfg,
		localServices: make(map[string]*ServiceInfo),
		discovered:    make(map[string]*ServiceInfo),
		listeners:     make(map[string]chan *ServiceInfo),
		ctx:           ctx,
		cancel:        cancel,
	}

	// 启动后台服务
	if cfg.EnableMDNS {
		go s.runMDNS()
	}
	if cfg.EnableUPnP {
		go s.runUPnP()
	}
	if cfg.EnableSSDP {
		go s.runSSDP()
	}

	return s
}

// Publish 发布服务 (mDNS)
func (s *Service) Publish(info *ServiceInfo) error {
	if info.ID == "" {
		info.ID = uuid.New().String()
	}

	s.mu.Lock()
	s.localServices[info.ID] = info
	s.stats.ServicesPublished.Add(1)
	s.mu.Unlock()

	// 通知监听器
	s.notifyListeners(info)

	return nil
}

// Unpublish 取消发布服务
func (s *Service) Unpublish(serviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.localServices[serviceID]; ok {
		delete(s.localServices, serviceID)
		s.stats.ServicesPublished.Add(-1)
	}
}

// Discover 发现服务
func (s *Service) Discover(serviceType ServiceType, timeout time.Duration) ([]*ServiceInfo, error) {
	start := time.Now()
	defer func() {
		s.stats.LookupLatencyUs.Store(time.Since(start).Microseconds())
	}()

	s.mu.RLock()
	results := make([]*ServiceInfo, 0)
	for _, info := range s.discovered {
		if info.Type == serviceType {
			results = append(results, info)
		}
	}
	s.mu.RUnlock()

	return results, nil
}

// Lookup 查找单个服务
func (s *Service) Lookup(serviceType ServiceType, name string) (*ServiceInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, info := range s.discovered {
		if info.Type == serviceType && info.Name == name {
			return info, nil
		}
	}
	return nil, fmt.Errorf("service not found")
}

// Subscribe 订阅服务发现
func (s *Service) Subscribe(ch chan *ServiceInfo) string {
	id := uuid.New().String()
	s.mu.Lock()
	s.listeners[id] = ch
	s.mu.Unlock()
	return id
}

// Unsubscribe 取消订阅
func (s *Service) Unsubscribe(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.listeners[id]; ok {
		close(ch)
		delete(s.listeners, id)
	}
}

// notifyListeners 通知所有监听器
func (s *Service) notifyListeners(info *ServiceInfo) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.listeners {
		select {
		case ch <- info:
		default:
		}
	}
}

// runMDNS 运行mDNS服务
func (s *Service) runMDNS() {
	// 监听本地网络变化
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			time.Sleep(30 * time.Second)
		}
	}
}

// runUPnP 运行UPnP服务
func (s *Service) runUPnP() {
	// UPnP端口映射
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			time.Sleep(60 * time.Second)
		}
	}
}

// runSSDP 运行SSDP服务
func (s *Service) runSSDP() {
	// SSDP发现
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			time.Sleep(30 * time.Second)
		}
	}
}

// GetLocalIP 获取本地IP
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

// GetLocalServices 获取本地服务列表
func (s *Service) GetLocalServices() []*ServiceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(s.localServices))
	for _, info := range s.localServices {
		services = append(services, info)
	}
	return services
}

// GetDiscoveredServices 获取发现的服务列表
func (s *Service) GetDiscoveredServices() []*ServiceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(s.discovered))
	for _, info := range s.discovered {
		services = append(services, info)
	}
	return services
}

// Stats 获取统计
func (s *Service) Stats() Stats {
	return Stats{
		ServicesPublished:  s.stats.ServicesPublished,
		ServicesDiscovered: s.stats.ServicesDiscovered,
		LookupLatencyUs:   s.stats.LookupLatencyUs,
	}
}

// Stop 停止服务
func (s *Service) Stop() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ch := range s.listeners {
		close(ch)
	}
	s.listeners = make(map[string]chan *ServiceInfo)
}

// MarshalJSON 序列化服务信息
func (info *ServiceInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":      info.ID,
		"name":    info.Name,
		"type":    info.Type,
		"host":    info.Host,
		"port":    info.Port,
		"metadata": info.Metadata,
	})
}
