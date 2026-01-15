//! 工具函数模块

use chrono::{DateTime, Utc};

/// 时间戳转日期时间
pub fn timestamp_to_datetime(ts: i64) -> DateTime<Utc> {
    DateTime::from_timestamp(ts, 0).unwrap_or_else(Utc::now)
}

/// 获取当前时间戳
pub fn now_timestamp() -> i64 {
    Utc::now().timestamp()
}

/// 格式化字节大小
pub fn format_bytes(bytes: u64) -> String {
    const UNITS: &[&str] = &["B", "KB", "MB", "GB", "TB"];
    let mut size = bytes as f64;
    let mut unit_idx = 0;
    
    while size >= 1024.0 && unit_idx < UNITS.len() - 1 {
        size /= 1024.0;
        unit_idx += 1;
    }
    
    format!("{:.2} {}", size, UNITS[unit_idx])
}

/// 生成随机 ID
pub fn generate_id() -> String {
    uuid::Uuid::new_v4().to_string()
}

/// 计算 token 数量 (粗略估算)
pub fn estimate_tokens(text: &str) -> usize {
    // 简单估算: 中文字符约 2 tokens, 英文约 0.75 tokens
    let chinese_chars = text.chars().filter(|c| c.len_utf8() > 1).count();
    let english_chars = text.len() - chinese_chars * 2;
    
    (chinese_chars * 2) + (english_chars * 3 / 4)
}
