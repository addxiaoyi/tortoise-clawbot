import SwiftUI

/// Tortoise macOS App - SwiftUI 实现
@main
struct TortoiseApp: App {
    @StateObject private var appState = AppState()
    @StateObject private var chatManager = ChatManager()
    @StateObject private var channelManager = ChannelManager()
    
    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(appState)
                .environmentObject(chatManager)
                .environmentObject(channelManager)
        }
        .windowStyle(.hiddenTitleBar)
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("新会话") {
                    chatManager.createSession()
                }
                .keyboardShortcut("n", modifiers: [.command])
            }
            
            CommandMenu("渠道") {
                Button("连接 Telegram") {
                    channelManager.connect(channel: .telegram)
                }
                Button("连接 Discord") {
                    channelManager.connect(channel: .discord)
                }
                Button("连接 Slack") {
                    channelManager.connect(channel: .slack)
                }
                Divider()
                Button("断开所有") {
                    channelManager.disconnectAll()
                }
            }
            
            CommandMenu("视图") {
                Toggle("显示侧边栏", isOn: $appState.showSidebar)
                Toggle("显示工具栏", isOn: $appState.showToolbar)
                Divider()
                Button("全屏") {
                    NSApp.mainMenu?.windows.first?.toggleFullScreen(nil)
                }
                .keyboardShortcut("f", modifiers: [.command, .control])
            }
        }
        
        // 设置窗口
        Settings {
            SettingsView()
                .environmentObject(appState)
        }
    }
}

/// 应用状态
class AppState: ObservableObject {
    @Published var showSidebar = true
    @Published var showToolbar = true
    @Published var currentView: AppView = .chat
    @Published var theme: AppTheme = .system
    
    var serverURL: String = "http://localhost:8080"
    var apiKey: String = ""
    
    enum AppView: String, CaseIterable {
        case chat = "对话"
        case channels = "渠道"
        case memory = "记忆"
        case plugins = "插件"
        case settings = "设置"
    }
    
    enum AppTheme: String, CaseIterable {
        case light = "浅色"
        case dark = "深色"
        case system = "跟随系统"
    }
}

/// 内容视图
struct ContentView: View {
    @EnvironmentObject var appState: AppState
    @EnvironmentObject var chatManager: ChatManager
    
    var body: some View {
        NavigationSplitView {
            SidebarView()
        } detail: {
            if appState.currentView == .chat {
                ChatDetailView()
            } else if appState.currentView == .channels {
                ChannelsView()
            } else if appState.currentView == .memory {
                MemoryView()
            } else if appState.currentView == .plugins {
                PluginsView()
            } else if appState.currentView == .settings {
                SettingsView()
            }
        }
    }
}

/// 侧边栏视图
struct SidebarView: View {
    @EnvironmentObject var appState: AppState
    @EnvironmentObject var chatManager: ChatManager
    
    var body: some View {
        List(selection: $appState.currentView) {
            Section("导航") {
                NavigationLink(value: AppState.AppView.chat) {
                    Label("对话", systemImage: "bubble.left.and.bubble.right")
                }
                NavigationLink(value: AppState.AppView.channels) {
                    Label("渠道", systemImage: "cable.connector")
                }
                NavigationLink(value: AppState.AppView.memory) {
                    Label("记忆", systemImage: "brain")
                }
                NavigationLink(value: AppState.AppView.plugins) {
                    Label("插件", systemImage: "puzzlepiece.extension")
                }
            }
            
            Section("会话") {
                ForEach(chatManager.sessions) { session in
                    SessionRow(session: session)
                }
            }
            
            Section("设置") {
                NavigationLink(value: AppState.AppView.settings) {
                    Label("设置", systemImage: "gear")
                }
            }
        }
        .listStyle(.sidebar)
        .navigationDestination(for: AppState.AppView.self) { view in
            Text(view.rawValue)
        }
    }
}

/// 会话行
struct SessionRow: View {
    let session: ChatSession
    @EnvironmentObject var chatManager: ChatManager
    
    var body: some View {
        HStack {
            VStack(alignment: .leading) {
                Text(session.title)
                    .font(.headline)
                Text(session.updatedAt, style: .relative)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            Spacer()
            if session.unreadCount > 0 {
                Text("\(session.unreadCount)")
                    .font(.caption2)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.accentColor)
                    .foregroundColor(.white)
                    .clipShape(Capsule())
            }
        }
        .padding(.vertical, 4)
        .contentShape(Rectangle())
        .onTapGesture {
            chatManager.selectSession(session)
        }
    }
}

/// 聊天详情视图
struct ChatDetailView: View {
    @EnvironmentObject var chatManager: ChatManager
    @State private var inputText = ""
    @FocusState private var isInputFocused: Bool
    
    var body: some View {
        VStack(spacing: 0) {
            // 消息列表
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(chatManager.currentMessages) { message in
                            MessageBubble(message: message)
                        }
                    }
                    .padding()
                }
                .onChange(of: chatManager.currentMessages.count) { _ in
                    withAnimation {
                        proxy.scrollTo(chatManager.currentMessages.last?.id)
                    }
                }
            }
            
            Divider()
            
            // 输入区域
            HStack(spacing: 12) {
                Button(action: {}) {
                    Image(systemName: "plus.circle")
                        .font(.title2)
                }
                .buttonStyle(.plain)
                
                TextField("输入消息...", text: $inputText, axis: .vertical)
                    .textFieldStyle(.plain)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(Color(nsColor: .textBackgroundColor))
                    .cornerRadius(18)
                    .focused($isInputFocused)
                    .lineLimit(1...5)
                
                Button(action: sendMessage) {
                    Image(systemName: "arrow.up.circle.fill")
                        .font(.title)
                        .foregroundColor(inputText.isEmpty ? .gray : .accentColor)
                }
                .buttonStyle(.plain)
                .disabled(inputText.isEmpty)
            }
            .padding()
        }
        .navigationTitle(chatManager.currentSession?.title ?? "新会话")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                Menu {
                    Button("复制会话") {}
                    Button("导出会话") {}
                    Divider()
                    Button("删除会话", role: .destructive) {
                        chatManager.deleteCurrentSession()
                    }
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
            }
        }
    }
    
    private func sendMessage() {
        guard !inputText.isEmpty else { return }
        chatManager.sendMessage(inputText)
        inputText = ""
    }
}

/// 消息气泡
struct MessageBubble: View {
    let message: ChatMessage
    
    var body: some View {
        HStack(alignment: .top, spacing: 8) {
            if message.role == .user { Spacer() }
            
            VStack(alignment: message.role == .user ? .trailing : .leading, spacing: 4) {
                if message.role != .user {
                    HStack(spacing: 4) {
                        Image(systemName: "smartface")
                            .font(.caption)
                        Text("Tortoise")
                            .font(.caption.bold())
                    }
                    .foregroundColor(.secondary)
                }
                
                Text(message.content)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 10)
                    .background(message.role == .user ? Color.accentColor : Color(nsColor: .textBackgroundColor))
                    .foregroundColor(message.role == .user ? .white : .primary)
                    .cornerRadius(18)
                
                if message.isStreaming {
                    ProgressView()
                        .scaleEffect(0.5)
                }
            }
            
            if message.role != .user { Spacer() }
        }
    }
}

#Preview {
    ContentView()
        .environmentObject(AppState())
        .environmentObject(ChatManager())
        .environmentObject(ChannelManager())
}
