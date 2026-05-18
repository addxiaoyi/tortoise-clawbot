/**
 * Voice STT - 语音转文字
 */

export interface STTConfig {
  language: string;
  continuous: boolean;
  interimResults: boolean;
}

export interface STTResult {
  transcript: string;
  confidence: number;
  isFinal: boolean;
}

/**
 * Speech-to-Text 服务
 */
export class STTService {
  private config: STTConfig;
  private isListening: boolean = false;
  private recognition: any = null;

  constructor(config: Partial<STTConfig> = {}) {
    this.config = {
      language: config.language || 'zh-CN',
      continuous: config.continuous ?? false,
      interimResults: config.interimResults ?? true,
    };
  }

  /**
   * 检查浏览器支持
   */
  isSupported(): boolean {
    return typeof window !== 'undefined' && ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window);
  }

  /**
   * 开始监听
   */
  start(onResult: (result: STTResult) => void, onError?: (error: string) => void): void {
    if (!this.isSupported()) {
      onError?.('Speech recognition not supported');
      return;
    }

    if (this.isListening) return;

    const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
    this.recognition = new SpeechRecognition();
    
    this.recognition.lang = this.config.language;
    this.recognition.continuous = this.config.continuous;
    this.recognition.interimResults = this.config.interimResults;
    
    this.recognition.onresult = (event: any) => {
      for (let i = event.resultIndex; i < event.results.length; i++) {
        const transcript = event.results[i][0].transcript;
        const confidence = event.results[i][0].confidence;
        const isFinal = event.results[i].isFinal;
        
        onResult({ transcript, confidence, isFinal });
      }
    };

    this.recognition.onerror = (event: any) => {
      onError?.(event.error);
      if (event.error !== 'no-speech') {
        this.isListening = false;
      }
    };

    this.recognition.onend = () => {
      this.isListening = false;
    };

    this.recognition.start();
    this.isListening = true;
  }

  /**
   * 停止监听
   */
  stop(): void {
    if (this.recognition) {
      this.recognition.stop();
      this.recognition = null;
    }
    this.isListening = false;
  }

  /**
   * 获取状态
   */
  getStatus(): { isListening: boolean; language: string } {
    return {
      isListening: this.isListening,
      language: this.config.language,
    };
  }
}

export const sttService = new STTService();
