import SwiftUI

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
    }
}

// MARK: - App State
class AppState: ObservableObject {
    @Published var isLoggedIn = false
    @Published var serverURL = "http://localhost:8080"
    @Published var apiKey = ""
    @Published var currentUser: User?
    
    struct User: Codable, Identifiable {
        let id: String
        let email: String
        let name: String
    }
}

// MARK: - Chat Manager
class ChatManager: ObservableObject {
    @Published var sessions: [ChatSession] = []
    @Published var currentSession: ChatSession?
    @Published var messages: [ChatMessage] = []
    @Published var isLoading = false
    
    struct ChatSession: Identifiable, Codable {
        let id: String
        var title: String
        var model: String
        var updatedAt: Date
    }
    
    struct ChatMessage: Identifiable, Codable {
        let id: String
        let role: MessageRole
        let content: String
        let createdAt: Date
        var isStreaming: Bool = false
        
        enum MessageRole: String, Codable {
            case user, assistant, system
        }
    }
    
    func createSession() {
        let session = ChatSession(
            id: UUID().uuidString,
            title: "新会话",
            model: "gpt-4",
            updatedAt: Date()
        )
        sessions.insert(session, at: 0)
        currentSession = session
        messages = []
    }
    
    func sendMessage(_ content: String) {
        let userMessage = ChatMessage(
            id: UUID().uuidString,
            role: .user,
            content: content,
            createdAt: Date()
        )
        messages.append(userMessage)
        
        isLoading = true
        
        // TODO: 调用 API
    }
}

// MARK: - Channel Manager
class ChannelManager: ObservableObject {
    @Published var channels: [Channel] = []
    
    struct Channel: Identifiable, Codable {
        let id: String
        let type: ChannelType
        var name: String
        var status: ConnectionStatus
        
        enum ChannelType: String, Codable {
            case telegram, discord, slack, whatsapp, signal
        }
        
        enum ConnectionStatus: String, Codable {
            case connected, disconnected, connecting, error
        }
    }
    
    func connect(channel: Channel.ChannelType) {
        // TODO: 连接渠道
    }
}
