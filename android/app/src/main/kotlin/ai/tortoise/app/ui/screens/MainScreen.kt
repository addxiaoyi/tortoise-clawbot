package ai.tortoise.app.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import ai.tortoise.app.viewmodel.ChatViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainScreen() {
    var selectedTab by remember { mutableIntStateOf(0) }
    val navItems = listOf("对话", "渠道", "记忆", "设置")
    
    Scaffold(
        bottomBar = {
            NavigationBar {
                navItems.forEachIndexed { index, title ->
                    NavigationBarItem(
                        icon = {
                            when(index) {
                                0 -> Icon(Icons.Default.Chat, contentDescription = title)
                                1 -> Icon(Icons.Default.Notifications, contentDescription = title)
                                2 -> Icon(Icons.Default.Memory, contentDescription = title)
                                3 -> Icon(Icons.Default.Settings, contentDescription = title)
                                else -> Icon(Icons.Default.Home, contentDescription = title)
                            }
                        },
                        label = { Text(title) },
                        selected = selectedTab == index,
                        onClick = { selectedTab = index }
                    )
                }
            }
        }
    ) { padding ->
        Box(modifier = Modifier.padding(padding)) {
            when(selectedTab) {
                0 -> ChatScreen()
                1 -> ChannelsScreen()
                2 -> MemoryScreen()
                3 -> SettingsScreen()
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatScreen() {
    val viewModel = remember { ChatViewModel() }
    val uiState by viewModel.uiState.collectAsState()
    var inputText by remember { mutableStateOf("") }
    
    Column(modifier = Modifier.fillMaxSize()) {
        // 顶部栏
        TopAppBar(
            title = { Text("Tortoise") },
            actions = {
                IconButton(onClick = { viewModel.createSession() }) {
                    Icon(Icons.Default.Add, "新建会话")
                }
            }
        )
        
        // 消息列表
        LazyColumn(
            modifier = Modifier.weight(1f).fillMaxWidth(),
            contentPadding = PaddingValues(16.dp)
        ) {
            items(uiState.messages) { message ->
                MessageBubble(message)
            }
        }
        
        // 输入框
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = inputText,
                onValueChange = { inputText = it },
                modifier = Modifier.weight(1f),
                placeholder = { Text("输入消息...") },
                maxLines = 3
            )
            
            Spacer(modifier = Modifier.width(8.dp))
            
            FilledIconButton(
                onClick = {
                    if (inputText.isNotBlank()) {
                        viewModel.sendMessage(inputText)
                        inputText = ""
                    }
                }
            ) {
                Icon(Icons.Default.Send, "发送")
            }
        }
    }
}

@Composable
fun MessageBubble(message: ai.tortoise.app.data.model.ChatMessage) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalArrangement = if (message.isUser) Arrangement.End else Arrangement.Start
    ) {
        Surface(
            shape = MaterialTheme.shapes.medium,
            color = if (message.isUser) 
                MaterialTheme.colorScheme.primary 
            else 
                MaterialTheme.colorScheme.surfaceVariant,
            modifier = Modifier.widthIn(max = 280.dp)
        ) {
            Text(
                text = message.content,
                modifier = Modifier.padding(12.dp),
                color = if (message.isUser) 
                    MaterialTheme.colorScheme.onPrimary 
                else 
                    MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
fun ChannelsScreen() {
    Column(modifier = Modifier.fillMaxSize()) {
        TopAppBar(title = { Text("消息渠道") })
        
        // 渠道列表
        LazyColumn(
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(5) { index ->
                ChannelCard(
                    name = listOf("Telegram", "Discord", "Slack", "WhatsApp", "Signal")[index],
                    status = "已连接",
                    onToggle = {}
                )
            }
        }
    }
}

@Composable
fun ChannelCard(name: String, status: String, onToggle: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column {
                Text(name, style = MaterialTheme.typography.titleMedium)
                Text(status, style = MaterialTheme.typography.bodySmall)
            }
            Switch(checked = true, onCheckedChange = { onToggle() })
        }
    }
}

@Composable
fun MemoryScreen() {
    Column(modifier = Modifier.fillMaxSize()) {
        TopAppBar(title = { Text("记忆") })
        Text("记忆功能开发中...", modifier = Modifier.padding(16.dp))
    }
}

@Composable
fun SettingsScreen() {
    Column(modifier = Modifier.fillMaxSize()) {
        TopAppBar(title = { Text("设置") })
        
        List {
            item { ListItem(headlineContent = { Text("服务器地址") }, supportingContent = { Text("http://localhost:8080") }) }
            item { ListItem(headlineContent = { Text("主题") }, supportingContent = { Text("跟随系统") }) }
            item { ListItem(headlineContent = { Text("关于") }, supportingContent = { Text("Tortoise v0.1.0") }) }
        }
    }
}
