package ai.tortoise.app.data.model

data class ChatMessage(
    val id: String,
    val role: MessageRole,
    val content: String,
    val createdAt: Long,
    val isStreaming: Boolean = false
)

enum class MessageRole {
    USER,
    ASSISTANT,
    SYSTEM
}
