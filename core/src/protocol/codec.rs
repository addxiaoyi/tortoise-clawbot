//! Protocol codec for binary encoding/decoding

use bytes::{Bytes, Buf, BufMut, BytesMut};
use std::io::{Read, Write};
use super::{MessageFrame, Flags, PROTOCOL_MAGIC};

/// Codec error type
#[derive(Debug, thiserror::Error)]
pub enum CodecError {
    #[error("Invalid magic bytes")]
    InvalidMagic,
    #[error("Invalid version: {0}")]
    InvalidVersion(u16),
    #[error("Payload too large: {0} bytes")]
    PayloadTooLarge(usize),
    #[error("Compression error: {0}")]
    Compression(String),
    #[error("Serialization error: {0}")]
    Serialization(String),
}

/// Protocol codec for framing and encoding
pub struct ProtocolCodec {
    max_payload_size: usize,
    enable_compression: bool,
}

impl Default for ProtocolCodec {
    fn default() -> Self {
        Self::new()
    }
}

impl ProtocolCodec {
    pub fn new() -> Self {
        Self {
            max_payload_size: 10 * 1024 * 1024, // 10MB
            enable_compression: true,
        }
    }
    
    pub fn with_max_payload(mut self, size: usize) -> Self {
        self.max_payload_size = size;
        self
    }
    
    pub fn with_compression(mut self, enabled: bool) -> Self {
        self.enable_compression = enabled;
        self
    }
    
    /// Encode a message frame to bytes
    pub fn encode(&self, frame: &MessageFrame) -> Result<Bytes, CodecError> {
        let mut buf = BytesMut::with_capacity(12 + frame.payload.len());
        
        // Magic bytes
        buf.put_slice(PROTOCOL_MAGIC);
        
        // Version (2 bytes)
        buf.put_u16(frame.version);
        
        // Message type (2 bytes)
        buf.put_u16(frame.msg_type);
        
        // Flags (2 bytes)
        buf.put_u16(frame.flags.0);
        
        // Length (4 bytes, payload only)
        let len = frame.payload.len() as u32;
        buf.put_u32(len);
        
        // Payload
        buf.put_slice(&frame.payload);
        
        Ok(buf.freeze())
    }
    
    /// Decode a message frame from bytes
    pub fn decode(&self, mut bytes: Bytes) -> Result<MessageFrame, CodecError> {
        let mut buf = bytes.copy_to_bytes(12).as_ref();
        
        // Check magic
        let magic = [buf[0], buf[1], buf[2], buf[3]];
        if &magic != PROTOCOL_MAGIC {
            return Err(CodecError::InvalidMagic);
        }
        
        // Version
        let version = (&buf[4..6]).read_u16::<BigEndian>();
        if version != 1 {
            return Err(CodecError::InvalidVersion(version));
        }
        
        // Message type
        let msg_type = (&buf[6..8]).read_u16::<BigEndian>();
        
        // Flags
        let flags = Flags((&buf[8..10]).read_u16::<BigEndian>());
        
        // Length
        let len = (&buf[10..14]).read_u32::<BigEndian>() as usize;
        
        if len > self.max_payload_size {
            return Err(CodecError::PayloadTooLarge(len));
        }
        
        // Payload
        let payload = bytes.copy_to_bytes(len).to_vec();
        
        Ok(MessageFrame {
            version,
            msg_type,
            flags,
            payload,
        })
    }
    
    /// Compress payload using zstd
    pub fn compress(&self, data: &[u8]) -> Result<Vec<u8>, CodecError> {
        zstd::encode_all(data, 3)
            .map_err(|e| CodecError::Compression(e.to_string()))
    }
    
    /// Decompress payload using zstd
    pub fn decompress(&self, data: &[u8]) -> Result<Vec<u8>, CodecError> {
        zstd::decode_all(data)
            .map_err(|e| CodecError::Compression(e.to_string()))
    }
}

use byteorder::{BigEndian, ReadBytesExt};

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_encode_decode_roundtrip() {
        let codec = ProtocolCodec::new();
        
        let frame = MessageFrame::new(
            0x0003, // Request
            b"Hello, Tortoise!".to_vec(),
        );
        
        let encoded = codec.encode(&frame).unwrap();
        let decoded = codec.decode(encoded).unwrap();
        
        assert_eq!(frame.msg_type, decoded.msg_type);
        assert_eq!(frame.payload, decoded.payload);
    }
    
    #[test]
    fn test_compression() {
        let codec = ProtocolCodec::new();
        
        let data = vec![0u8; 1000]; // Highly compressible
        
        let compressed = codec.compress(&data).unwrap();
        let decompressed = codec.decompress(&compressed).unwrap();
        
        assert_eq!(data, decompressed);
        assert!(compressed.len() < data.len());
    }
}
