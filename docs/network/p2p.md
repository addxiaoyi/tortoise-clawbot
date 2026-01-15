# Tortoise P2P 去中心化网络

## 概述

Tortoise 网络层采用 P2P 架构，实现去中心化通信、分布式存储和抗审查能力。

## 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                    Tortoise Network Layer                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                      DHT Network                             │ │
│  │   ┌─────┐    ┌─────┐    ┌─────┐    ┌─────┐    ┌─────┐     │ │
│  │   │Node │────│Node │────│Node │────│Node │────│Node │     │ │
│  │   └─────┘    └─────┘    └─────┘    └─────┘    └─────┘     │ │
│  │       │           │           │           │                 │ │
│  │       └───────────┴───────────┴───────────┘                 │ │
│  │                     Kademlia DHT                             │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Transport Layer                           │ │
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │ │
│  │   │  TCP    │  │  UDP    │  │  QUIC   │  │ WebRTC  │       │ │
│  │   └─────────┘  └─────────┘  └─────────┘  └─────────┘       │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Protocol Layer                            │ │
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │ │
│  │   │Discovery│  │Routing  │  │Storage  │  │Messaging│       │ │
│  │   └─────────┘  └─────────┘  └─────────┘  └─────────┘       │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## 核心实现

### P2P 节点

```rust
// src/network/p2p/node.rs

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::{RwLock, mpsc};
use libp2p::{Swarm, Transport, NetworkBehaviour, PeerId, Multiaddr};
use libp2p::kad::{KadProtocol, Kademlia, Record};
use libp2p::mdns::{Mdns, MdnsEvent};
use libp2p::floodsub::{Floodsub, FloodsubEvent};
use libp2p::ping::{Ping, PingEvent};

/// 节点配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeConfig {
    /// 节点 ID
    pub node_id: String,
    /// 监听地址列表
    pub listen_addresses: Vec<String>,
    /// 是否引导节点
    pub is_bootstrap: bool,
    /// 引导节点地址
    pub bootstrap_nodes: Vec<BootstrapNode>,
    /// DHT 配置
    pub dht_config: DhtConfig,
    /// 存储配置
    pub storage_config: StorageConfig,
}

/// DHT 配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DhtConfig {
    /// 存储半径
    pub storage_radius: u32,
    /// 查询并行数
    pub query_parallelism: u32,
    /// 查询超时 (ms)
    pub query_timeout_ms: u64,
    /// 记录 TTL (s)
    pub record_ttl_seconds: u64,
}

/// 存储配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StorageConfig {
    /// 最大存储大小 (bytes)
    pub max_storage_bytes: u64,
    /// 清理间隔 (s)
    pub gc_interval_seconds: u64,
    /// 备份因子
    pub replication_factor: u32,
}

/// 引导节点
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BootstrapNode {
    pub peer_id: String,
    pub addresses: Vec<String>,
}

/// 节点状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum NodeStatus {
    Initializing,
    Connecting,
    Connected,
    Disconnected,
    Error(String),
}

/// 网络事件
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum NetworkEvent {
    /// 新节点连接
    PeerConnected { peer_id: String, addresses: Vec<String> },
    /// 节点断开
    PeerDisconnected { peer_id: String },
    /// 收到消息
    Message { peer_id: String, message: P2PMessage },
    /// DHT 记录更新
    RecordUpdated { key: String },
    /// 发现新节点
    NodeDiscovered { peer_id: String, addresses: Vec<String> },
}

/// P2P 消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct P2PMessage {
    pub id: String,
    pub source: String,
    pub destination: Option<String>,
    pub message_type: MessageType,
    pub payload: Vec<u8>,
    pub timestamp: i64,
    pub ttl: u32,
}

/// 消息类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum MessageType {
    /// 直接消息
    Direct,
    /// 广播
    Broadcast,
    /// 路由消息
    Routed,
    /// DHT 查询
    DhtQuery,
    /// 存储请求
    StoreRequest,
}

/// P2P 节点
pub struct P2PNode {
    config: NodeConfig,
    peer_id: PeerId,
    swarm: Swarm<NetworkBehaviour>,
    peers: Arc<RwLock<HashMap<PeerId, PeerInfo>>>,
    message_tx: mpsc::Sender<NetworkEvent>,
    running: Arc<RwLock<bool>>,
}

impl P2PNode {
    /// 创建新节点
    pub async fn new(config: NodeConfig) -> Result<Self> {
        let peer_id = PeerId::from_public_key(
            config.node_id.as_bytes().to_vec().try_into()?
        );

        // 构建传输层
        let transport = Self::build_transport(&config).await?;

        // 构建网络行为
        let behaviour = NetworkBehaviour::new(peer_id, &config)?;

        // 构建 Swarm
        let mut swarm = Swarm::new(transport, behaviour, peer_id.clone());

        // 监听地址
        for addr in &config.listen_addresses {
            let multiaddr: Multiaddr = addr.parse()?;
            Swarm::listen_on(&mut swarm, multiaddr)?;
        }

        // 连接引导节点
        if !config.is_bootstrap && !config.bootstrap_nodes.is_empty() {
            Self::bootstrap(&mut swarm, &config.bootstrap_nodes).await?;
        }

        let (message_tx, message_rx) = mpsc::channel(1000);

        Ok(Self {
            config,
            peer_id,
            swarm,
            peers: Arc::new(RwLock::new(HashMap::new())),
            message_tx,
            running: Arc::new(RwLock::new(false)),
        })
    }

    /// 构建传输层
    async fn build_transport(config: &NodeConfig) -> Result<Box<dyn Transport + Send + Sync>> {
        use libp2p::tcp::TcpConfig;
        use libp2p::yamux::YamuxConfig;
        use libp2p::noise::NoiseConfig;
        use libp2p::quic::QuicConfig;

        let tcp = TcpConfig::new();
        let quic = QuicConfig::new();
        let yamux = YamuxConfig::default();
        
        // TLS 配置
        let noise = NoiseConfig::new(
            libp2p::noise::Keypair::<libp2p::noise::X25519Spec>::new()
                .into_authentic(&libp2p::noise::Keypair::<libp2p::noise::X25519Spec>::new())
                .0,
        );

        Ok(tcp.or_transport(quic)
            .upgrade(libp2p::core::upgrade::Version::V1)
            .authenticate(noise)
            .multiplex(yamux)
            .boxed())
    }

    /// 引导连接
    async fn bootstrap(
        swarm: &mut Swarm<NetworkBehaviour>,
        nodes: &[BootstrapNode],
    ) -> Result<()> {
        for node in nodes {
            let peer_id: PeerId = node.peer_id.parse()?;
            for addr in &node.addresses {
                let multiaddr: Multiaddr = addr.parse()?;
                Swarm::dial_addr(swarm, multiaddr)?;
            }
        }
        Ok(())
    }

    /// 启动节点
    pub async fn start(&self) -> Result<()> {
        *self.running.write().await = true;

        let running = self.running.clone();
        let mut swarm = Swarm::new(self.swarm.clone());
        let message_tx = self.message_tx.clone();

        // 事件循环
        tokio::spawn(async move {
            while *running.read().await {
                match swarm.next().await {
                    Some(event) => {
                        if let Some(network_event) = Self::process_event(event) {
                            let _ = message_tx.send(network_event).await;
                        }
                    }
                    None => break,
                }
            }
        });

        Ok(())
    }

    /// 停止节点
    pub async fn stop(&self) -> Result<()> {
        *self.running.write().await = false;
        Ok(())
    }

    /// 处理事件
    fn process_event(event: SwarmEvent) -> Option<NetworkEvent> {
        match event {
            SwarmEvent::PeerConnected { peer_id, .. } => {
                Some(NetworkEvent::PeerConnected {
                    peer_id: peer_id.to_base58(),
                    addresses: vec![],
                })
            }
            SwarmEvent::PeerDisconnected { peer_id, .. } => {
                Some(NetworkEvent::PeerDisconnected {
                    peer_id: peer_id.to_base58(),
                })
            }
            _ => None,
        }
    }

    /// 发送消息
    pub async fn send_message(&self, message: P2PMessage) -> Result<()> {
        let peers = self.peers.read().await;
        
        if let Some(target) = &message.destination {
            let peer_id: PeerId = target.parse()?;
            self.swarm.send_direct_message(peer_id, message)?;
        } else {
            // 广播
            self.swarm.floodsub_publish(message)?;
        }
        
        Ok(())
    }

    /// 存储数据到 DHT
    pub async fn store(&self, key: &str, value: Vec<u8>) -> Result<()> {
        let record = Record {
            key: key.as_bytes().to_vec(),
            value,
            publisher: Some(self.peer_id.clone()),
            expires: None,
        };
        
        self.swarm.kademlia().put_record(record)?;
        Ok(())
    }

    /// 从 DHT 获取数据
    pub async fn get(&self, key: &str) -> Result<Option<Vec<u8>>> {
        let (ch, _) = self.swarm.kademlia().get_record(key.as_bytes());
        
        // 等待结果
        match ch.await {
            Ok(Ok(record)) => Ok(Some(record.value)),
            _ => Ok(None),
        }
    }

    /// 发布到主题
    pub async fn publish(&self, topic: &str, message: Vec<u8>) -> Result<()> {
        self.swarm.floodsub_publish(topic, message)?;
        Ok(())
    }

    /// 订阅主题
    pub async fn subscribe(&self, topic: &str) -> Result<()> {
        self.swarm.floodsub_subscribe(topic)?;
        Ok(())
    }

    /// 获取节点信息
    pub async fn get_peers(&self) -> Vec<PeerInfo> {
        let peers = self.peers.read().await;
        peers.values().cloned().collect()
    }

    /// 获取节点状态
    pub async fn status(&self) -> NodeStatus {
        if *self.running.read().await {
            NodeStatus::Connected
        } else {
            NodeStatus::Disconnected
        }
    }
}

/// 节点信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PeerInfo {
    pub peer_id: String,
    pub addresses: Vec<String>,
    pub is_connected: bool,
    pub latency_ms: Option<u64>,
    pub last_seen: i64,
}

/// Swarm 事件
#[derive(Debug)]
pub enum SwarmEvent {
    PeerConnected { peer_id: PeerId, addr: Multiaddr },
    PeerDisconnected { peer_id: PeerId },
    NewListenAddr { addr: Multiaddr },
}
```

### Kademlia DHT

```rust
// src/network/dht/kademlia.rs

use anyhow::Result;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use sha2::{Sha256, Digest};

/// Kademlia DHT 实现
pub struct KademliaDHT {
    /// 节点 ID
    node_id: [u8; 32],
    /// 路由表
    routing_table: Arc<RwLock<RoutingTable>>,
    /// 存储
    storage: Arc<RwLock<HashMap<Vec<u8>, DHTRecord>>>,
    /// 配置
    config: KademliaConfig,
}

/// Kademlia 配置
#[derive(Debug, Clone)]
pub struct KademliaConfig {
    /// K 值 (每个桶的最大节点数)
    pub k: u32,
    /// Alpha 值 (并发查询数)
    pub alpha: u32,
    /// 节点 ID 长度 (bits)
    pub id_length: u32,
}

impl Default for KademliaConfig {
    fn default() -> Self {
        Self {
            k: 20,
            alpha: 3,
            id_length: 256,
        }
    }
}

/// 路由表
#[derive(Debug)]
pub struct RoutingTable {
    /// K-buckets
    buckets: Vec<KBucket>,
}

impl RoutingTable {
    pub fn new(k: u32) -> Self {
        Self {
            buckets: (0..256).map(|_| KBucket::new(k)).collect(),
        }
    }

    /// 添加节点
    pub fn add(&mut self, node_id: [u8; 32], info: NodeContact) {
        let bucket_index = Self::distance_index(&self.buckets[0].nodes.first().map(|n| n.id), &node_id);
        self.buckets[bucket_index].add(node_id, info);
    }

    /// 查找最近节点
    pub fn find_closest(&self, target: &[u8; 32], count: u32) -> Vec<NodeContact> {
        let mut all_nodes: Vec<_> = self.buckets.iter()
            .flat_map(|b| b.nodes.clone())
            .collect();
        
        all_nodes.sort_by(|a, b| {
            let dist_a = Self::xor_distance(&a.id, target);
            let dist_b = Self::xor_distance(&b.id, target);
            dist_a.cmp(&dist_b)
        });
        
        all_nodes.into_iter().take(count as usize).collect()
    }

    fn distance_index(a: &Option<[u8; 32]>, b: &[u8; 32]) -> usize {
        match a {
            Some(id) => Self::xor_distance(id, b) as usize,
            None => 255,
        }
    }

    fn xor_distance(a: &[u8; 32], b: &[u8; 32]) -> u32 {
        let mut dist = 0u32;
        for i in 0..32 {
            let x = a[i] ^ b[i];
            if x != 0 {
                dist = (32 - i) as u32 + x.leading_zeros();
                break;
            }
        }
        dist
    }
}

/// K-Bucket
#[derive(Debug)]
pub struct KBucket {
    nodes: Vec<NodeContact>,
    k: u32,
}

impl KBucket {
    pub fn new(k: u32) -> Self {
        Self { nodes: vec![], k }
    }

    pub fn add(&mut self, node_id: [u8; 32], info: NodeContact) {
        if let Some(pos) = self.nodes.iter().position(|n| n.id == node_id) {
            self.nodes[pos] = info;
        } else if self.nodes.len() < self.k as usize {
            self.nodes.push(info);
        }
    }
}

/// 节点联系信息
#[derive(Debug, Clone)]
pub struct NodeContact {
    pub id: [u8; 32],
    pub addresses: Vec<String>,
    pub last_seen: i64,
}

/// DHT 记录
#[derive(Debug, Clone)]
pub struct DHTRecord {
    pub key: Vec<u8>,
    pub value: Vec<u8>,
    pub publisher: Option<[u8; 32]>,
    pub expires: Option<i64>,
    pub sequence_number: u64,
}

impl KademliaDHT {
    /// 创建新 DHT
    pub fn new(node_id: &[u8], config: KademliaConfig) -> Self {
        let mut hasher = Sha256::new();
        hasher.update(node_id);
        let result = hasher.finalize();
        let node_id = result.into();

        Self {
            node_id,
            routing_table: Arc::new(RwLock::new(RoutingTable::new(config.k))),
            storage: Arc::new(RwLock::new(HashMap::new())),
            config,
        }
    }

    /// 生成节点 ID
    pub fn generate_node_id(peer_id: &str) -> [u8; 32] {
        let mut hasher = Sha256::new();
        hasher.update(peer_id.as_bytes());
        hasher.update(b"tortoise-dht-salt");
        hasher.finalize().into()
    }

    /// 生成记录键
    pub fn generate_key(data: &[u8]) -> [u8; 32] {
        let mut hasher = Sha256::new();
        hasher.update(data);
        hasher.finalize().into()
    }

    /// 存储记录
    pub async fn put(&self, key: &[u8], value: Vec<u8>) -> Result<()> {
        let record = DHTRecord {
            key: key.to_vec(),
            value,
            publisher: Some(self.node_id),
            expires: None,
            sequence_number: 0,
        };

        self.storage.write().await.insert(key.to_vec(), record);

        // 发布到网络
        self.replicate_to_network(key).await?;

        Ok(())
    }

    /// 获取记录
    pub async fn get(&self, key: &[u8]) -> Result<Option<DHTRecord>> {
        let storage = self.storage.read().await;
        Ok(storage.get(key).cloned())
    }

    /// 查找节点
    pub async fn find_node(&self, target: &[u8; 32]) -> Vec<NodeContact> {
        let table = self.routing_table.read().await;
        table.find_closest(target, self.config.alpha)
    }

    /// 迭代查询
    pub async fn iterative_find(&self, key: &[u8; 32]) -> Result<Vec<DHTRecord>> {
        let mut closest_nodes = self.find_node(key).await;
        let mut queried = vec![];
        let mut results = vec![];

        while !closest_nodes.is_empty() && queried.len() < self.config.k as usize {
            let node = closest_nodes.remove(0);
            if queried.contains(&node.id) {
                continue;
            }

            // 查询节点
            match self.query_node(&node, key).await {
                Ok(found) => {
                    queried.push(node.id);
                    results.extend(found);
                }
                Err(_) => continue,
            }

            // 更新最近节点
            let new_nodes = self.find_node(key).await;
            for n in new_nodes {
                if !closest_nodes.contains(&n) && !queried.contains(&n.id) {
                    closest_nodes.push(n);
                }
            }
        }

        Ok(results)
    }

    async fn query_node(&self, node: &NodeContact, key: &[u8; 32]) -> Result<Vec<DHTRecord>> {
        // 实际实现会通过网络请求
        Ok(vec![])
    }

    async fn replicate_to_network(&self, key: &[u8]) -> Result<()> {
        // 复制到最近的 K 个节点
        let nodes = self.find_node(&Self::generate_key(key)).await;
        for node in nodes.into_iter().take(self.config.k as usize) {
            self.replicate_record(&node, key).await?;
        }
        Ok(())
    }

    async fn replicate_record(&self, node: &NodeContact, key: &[u8]) -> Result<()> {
        Ok(())
    }

    /// 添加到路由表
    pub async fn add_peer(&self, peer_id: &str, addresses: Vec<String>) {
        let node_id = Self::generate_node_id(peer_id);
        let contact = NodeContact {
            id: node_id,
            addresses,
            last_seen: chrono::Utc::now().timestamp(),
        };

        let mut table = self.routing_table.write().await;
        table.add(node_id, contact);
    }
}
```

### 加密通信

```rust
// src/network/crypto.rs

use anyhow::Result;
use serde::{Deserialize, Serialize};
use sodiumoxide::crypto::{box_, secretbox, hash};

/// 密钥对
#[derive(Debug, Clone)]
pub struct KeyPair {
    pub public_key: [u8; 32],
    secret_key: [u8; 32],
}

/// 临时密钥
#[derive(Debug, Clone)]
pub struct EphemeralKey {
    pub public_key: [u8; 32],
    secret_key: [u8; 32],
}

/// 加密消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncryptedMessage {
    pub nonce: [u8; 24],
    pub ciphertext: Vec<u8>,
    pub sender_pubkey: [u8; 32],
}

impl KeyPair {
    /// 生成新密钥对
    pub fn generate() -> Self {
        let (public_key, secret_key) = box_::gen_keypair();
        Self {
            public_key: public_key.0,
            secret_key: secret_key.0,
        }
    }

    /// 从种子生成
    pub fn from_seed(seed: &[u8]) -> Self {
        let mut hasher = hash::sha256::State::init();
        hasher.update(seed);
        let hash = hasher.finalize();
        
        let secret_key = box_::ScalarMult::from_slice(&hash.0).unwrap();
        let public_key = box_::curve25519xsalsa20poly1305::scalar_mult_base(&secret_key);
        
        Self {
            public_key: public_key.0,
            secret_key: secret_key.0,
        }
    }

    /// 加密消息 (给指定接收者)
    pub fn encrypt(&self, recipient_pk: &[u8; 32], message: &[u8]) -> EncryptedMessage {
        let nonce = box_::Nonce::gen();
        let (public_key, secret_key) = (
            box_::PublicKey(*recipient_pk),
            box_::SecretKey(box_::ScalarMult::from_slice(&self.secret_key).unwrap()),
        );

        let ciphertext = box_::seal(message, &nonce, &public_key, &secret_key);

        EncryptedMessage {
            nonce: nonce.0,
            ciphertext,
            sender_pubkey: self.public_key,
        }
    }

    /// 解密消息
    pub fn decrypt(&self, encrypted: &EncryptedMessage) -> Result<Vec<u8>> {
        let nonce = box_::Nonce(encrypted.nonce);
        let sender_pk = box_::PublicKey(encrypted.sender_pubkey);
        let secret_key = box_::SecretKey(box_::ScalarMult::from_slice(&self.secret_key).unwrap());

        let plaintext = box_::open(&encrypted.ciphertext, &nonce, &sender_pk, &secret_key)
            .map_err(|_| anyhow::anyhow!("Decryption failed"))?;

        Ok(plaintext)
    }
}

/// 对称加密 (用于会话密钥)
impl EphemeralKey {
    pub fn generate() -> Self {
        let key = secretbox::gen_key();
        Self {
            public_key: key.0,
            secret_key: key.0,
        }
    }

    pub fn encrypt(&self, message: &[u8]) -> EncryptedMessage {
        let nonce = secretbox::Nonce::gen();
        let key = secretbox::Key(self.secret_key);
        
        let ciphertext = secretbox::seal(message, &nonce, &key);

        EncryptedMessage {
            nonce: nonce.0,
            ciphertext,
            sender_pubkey: self.public_key,
        }
    }

    pub fn decrypt(&self, encrypted: &EncryptedMessage) -> Result<Vec<u8>> {
        let nonce = secretbox::Nonce(encrypted.nonce);
        let key = secretbox::Key(self.secret_key);

        secretbox::open(&encrypted.ciphertext, &nonce, &key)
            .map_err(|_| anyhow::anyhow!("Decryption failed"))
    }
}
```
