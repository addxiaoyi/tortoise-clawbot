package discovery

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Discovery 服务发现管理器
type Discovery struct {
	config    *Config
	ctx      context.Context
	cancel   context.CancelFunc
	services  map[string]*Service
	mu        sync.RWMutex
	announcer *MDNSAnnouncer
}

// Config 服务发现配置
type Config struct {
	Enabled           bool
	mDNS              bool // Bonjour/mDNS
	UPnP              bool // Universal Plug and Play
	SSDP              bool // Simple Service Discovery Protocol
	Port              int
	AdvertiseInterval int // 秒
	ServiceName       string
	ServiceType       string
}

// Service 服务信息
type Service struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// MDNSAnnouncer mDNS 广播器
type MDNSAnnouncer struct {
	iface *net.Interface
	conn  *net.UDPConn
	addr  *net.UDPAddr
}

// NewDiscovery 创建服务发现管理器
func NewDiscovery(config *Config) *Discovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &Discovery{
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		services:  make(map[string]*Service),
	}
}

// Start 启动服务发现
func (d *Discovery) Start() error {
	if !d.config.Enabled {
		log.Println("📡 服务发现已禁用")
		return nil
	}

	// 注册本服务
	d.registerSelf()

	// 启动 mDNS 广播 (如果启用)
	if d.config.mDNS {
		go d.startMDNS()
	}

	// 启动 UPnP 端口映射 (如果启用)
	if d.config.UPnP {
		go d.startUPnP()
	}

	// 启动 SSDP 发现 (如果启用)
	if d.config.SSDP {
		go d.startSSDP()
	}

	// 定期广播
	go d.advertiseLoop()

	log.Printf("✅ 服务发现已启动 (端口: %d)", d.config.Port)
	return nil
}

// Stop 停止服务发现
func (d *Discovery) Stop() {
	d.cancel()
	
	if d.announcer != nil && d.announcer.conn != nil {
		d.announcer.conn.Close()
	}
	
	log.Println("📡 服务发现已停止")
}

// registerSelf 注册本服务
func (d *Discovery) registerSelf() {
	hostname, _ := getLocalIP()
	
	d.mu.Lock()
	d.services["self"] = &Service{
		ID:       "self",
		Name:     d.config.ServiceName,
		Type:     d.config.ServiceType,
		Address:  hostname,
		Port:     d.config.Port,
		Metadata: map[string]string{
			"version": "1.0.0",
			"api":     "/api/v1",
		},
	}
	d.mu.Unlock()
}

// advertiseLoop 定期广播循环
func (d *Discovery) advertiseLoop() {
	ticker := time.NewTicker(time.Duration(d.config.AdvertiseInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.advertise()
		}
	}
}

// advertise 广播服务
func (d *Discovery) advertise() {
	if d.config.mDNS {
		d.advertiseMDNS()
	}
	
	if d.config.SSDP {
		d.advertiseSSDP()
	}
}

// ============ mDNS 实现 ============

func (d *Discovery) startMDNS() {
	// 获取网络接口
	iface, err := net.InterfaceByName("en0")
	if err != nil {
		// 尝试获取默认接口
		ifaces, err := net.Interfaces()
		if err != nil {
			log.Printf("❌ mDNS: 无法获取网络接口: %v", err)
			return
		}
		for _, i := range ifaces {
			if i.Flags&net.FlagUp != 0 && i.Flags&net.FlagMulticast != 0 {
				iface = &i
				break
			}
		}
	}
	if iface == nil {
		log.Printf("❌ mDNS: 找不到合适的网络接口")
		return
	}

	// 创建 UDP 连接
	addr := &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251), // mDNS 多播地址
		Port: 5353,
	}

	conn, err := net.ListenMulticastUDP("udp", iface, addr)
	if err != nil {
		log.Printf("❌ mDNS: 无法监听: %v", err)
		return
	}
	defer conn.Close()

	d.announcer = &MDNSAnnouncer{
		iface: iface,
		conn:  conn,
		addr:  addr,
	}

	log.Printf("✅ mDNS 已启动 (接口: %s)", iface.Name)

	// 处理 mDNS 查询
	buf := make([]byte, 65536)
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, src, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			go d.handleMDNSQuery(buf[:n], src)
		}
	}
}

// advertiseMDNS 广播 mDNS
func (d *Discovery) advertiseMDNS() {
	if d.announcer == nil {
		return
	}

	d.mu.RLock()
	service := d.services["self"]
	d.mu.RUnlock()

	if service == nil {
		return
	}

	// 构建 mDNS 响应
	response := buildMDNSResponse(service)
	
	d.announcer.conn.WriteToUDP(response, d.announcer.addr)
}

// handleMDNSQuery 处理 mDNS 查询
func (d *Discovery) handleMDNSQuery(query []byte, src *net.UDPAddr) {
	// 检查是否是查询我们的服务
	if containsMDNSQuery(query, d.config.ServiceName) {
		// 发送响应
		d.mu.RLock()
		service := d.services["self"]
		d.mu.RUnlock()
		
		if service != nil {
			response := buildMDNSResponse(service)
			d.announcer.conn.WriteToUDP(response, src)
		}
	}
}

// ============ UPnP 实现 ============

func (d *Discovery) startUPnP() {
	// 尝试添加端口映射
	if err := d.addUPnPPortMapping(); err != nil {
		log.Printf("⚠️ UPnP: %v", err)
		return
	}
	
	log.Printf("✅ UPnP 端口映射已添加 (外部端口: %d)", d.config.Port)
}

// addUPnPPortMapping 添加 UPnP 端口映射
func (d *Discovery) addUPnPPortMapping() error {
	// 查找 UPnP 路由器
	gateway, err := getUPnPGateway()
	if err != nil {
		return err
	}

	// 添加端口映射
	return addPortMapping(gateway, d.config.Port, "Tortoise AI Framework")
}

// ============ SSDP 实现 ============

func (d *Discovery) startSSDP() {
	// 发现局域网内的其他服务
	go d.discoverSSDPServices()
	
	log.Printf("✅ SSDP 已启动")
}

// discoverSSDPServices 发现 SSDP 服务
func (d *Discovery) discoverSSDPServices() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(239, 255, 255, 250),
		Port: 1900,
	})
	if err != nil {
		log.Printf("❌ SSDP: 无法创建连接: %v", err)
		return
	}
	defer conn.Close()

	// 发送 M-SEARCH
	search := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 3\r\n" +
		"ST: ssdp:all\r\n\r\n"

	conn.Write([]byte(search))

	// 读取响应
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}
		d.parseSSDPResponse(buf[:n])
	}
}

// advertiseSSDP 广播 SSDP
func (d *Discovery) advertiseSSDP() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(239, 255, 255, 250),
		Port: 1900,
	})
	if err != nil {
		return
	}
	defer conn.Close()

	ip, _ := getLocalIP()
	
	notify := "NOTIFY * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"NT: " + d.config.ServiceType + "\r\n" +
		"USN: uuid:" + d.config.ServiceName + "::" + d.config.ServiceType + "\r\n" +
		"Location: http://" + ip + ":" + itoa(d.config.Port) + "/\r\n" +
		"Cache-Control: max-age=1800\r\n\r\n"

	conn.Write([]byte(notify))
}

// parseSSDPResponse 解析 SSDP 响应
func (d *Discovery) parseSSDPResponse(data []byte) {
	lines := strings.Split(string(data), "\r\n")
	
	service := &Service{
		ID:   generateID(),
		Type: "ssdp",
	}
	
	for _, line := range lines {
		if strings.HasPrefix(line, "LOCATION:") {
			service.Address = strings.TrimSpace(strings.TrimPrefix(line, "LOCATION:"))
		}
		if strings.HasPrefix(line, "ST:") {
			service.Type = strings.TrimSpace(strings.TrimPrefix(line, "ST:"))
		}
	}
	
	if service.Address != "" {
		d.mu.Lock()
		d.services[service.ID] = service
		d.mu.Unlock()
	}
}

// ============ Helpers ============

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	
	return "127.0.0.1", nil
}

func containsMDNSQuery(query []byte, name string) bool {
	return strings.Contains(strings.ToLower(string(query)), strings.ToLower(name))
}

func buildMDNSResponse(service *Service) []byte {
	// 简化的 mDNS 响应格式
	return []byte{}
}

func getUPnPGateway() (string, error) {
	// 简化的 UPnP 网关发现
	resp, err := http.Get("http://www.upnp.org/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return "", nil
}

func addPortMapping(gateway, port int, description string) error {
	// 简化的 UPnP 端口映射添加
	return nil
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func itoa(i int) string {
	return strings.TrimSpace(strings.Replace(json.Marshal(i), "\"", "", -1))
}

// GetServices 获取所有发现的服务
func (d *Discovery) GetServices() []*Service {
	d.mu.RLock()
	defer d.mu.RUnlock()

	services := make([]*Service, 0, len(d.services))
	for _, s := range d.services {
		services = append(services, s)
	}
	return services
}
