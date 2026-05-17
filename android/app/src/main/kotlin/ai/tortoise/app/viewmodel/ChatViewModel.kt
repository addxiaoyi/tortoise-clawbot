package ai.tortoise.app.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import ai.tortoise.app.data.model.ChatMessage
import ai.tortoise.app.data.model.MessageRole

data class ChatUiState(
    val messages: List<ChatMessage> = emptyList(),
    val isLoading: Boolean = false,
    val currentSession: String? = null
)

class ChatViewModel : ViewModel() {
    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private val messageIdCounter = java.util.concurrent.atomic.AtomicLong(0)

    init {
        // 添加欢迎消息
        addMessage(
            role = MessageRole.ASSISTANT,
            content = "你好！我是 Tortoise AI 助手。有什么可以帮助你的吗？"
        )
    }

    fun sendMessage(content: String) {
        if (content.isBlank()) return

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)

            // 添加用户消息
            addMessage(role = MessageRole.USER, content = content)

            // 模拟AI回复
            delay(1000)
            
            val response = generateResponse(content)
            addMessage(role = MessageRole.ASSISTANT, content = response)

            _uiState.value = _uiState.value.copy(isLoading = false)
        }
    }

    fun createSession() {
        _uiState.value = _uiState.value.copy(
            messages = emptyList(),
            currentSession = "session-${System.currentTimeMillis()}"
        )
        
        addMessage(
            role = MessageRole.ASSISTANT,
            content = "新会话已创建！有什么可以帮助你的吗？"
        )
    }

    private fun addMessage(role: MessageRole, content: String) {
        val message = ChatMessage(
            id = "msg-${messageIdCounter.incrementAndGet()}",
            role = role,
            content = content,
            createdAt = System.currentTimeMillis()
        )
        
        _uiState.value = _uiState.value.copy(
            messages = _uiState.value.messages + message
        )
    }

    private fun generateResponse(userMessage: String): String {
        return when {
            userMessage.contains("你好", ignoreCase = true) -> 
                "你好！很高兴见到你。有什么我可以帮助你的吗？"
            userMessage.contains("帮助", ignoreCase = true) -> 
                "我可以帮你完成很多任务，比如：\n• 回答问题\n• 编写代码\n• 分析数据\n• 写作助手\n请告诉我你需要什么帮助。"
            userMessage.contains("天气", ignoreCase = true) -> 
                "抱歉，我目前无法获取实时天气信息。但我可以帮你查询其他内容。"
            else -> 
                "收到你的消息：\"$userMessage\"\n\n作为 Tortoise AI，我会尽力帮助你。如果有具体问题请告诉我！"
        }
    }
}
