/**
 * Voice TTS - 文字转语音
 */

export interface TTSConfig {
  voice?: string;
  rate: number;
  pitch: number;
  volume: number;
}

/**
 * Text-to-Speech 服务
 */
export class TTSService {
  private config: TTSConfig;
  private isSpeaking: boolean = false;
  private synth: SpeechSynthesis;

  constructor(config: Partial<TTSConfig> = {}) {
    this.config = {
      voice: config.voice,
      rate: config.rate ?? 1.0,
      pitch: config.pitch ?? 1.0,
      volume: config.volume ?? 1.0,
    };
    this.synth = window.speechSynthesis;
  }

  /**
   * 检查浏览器支持
   */
  isSupported(): boolean {
    return typeof window !== 'undefined' && 'speechSynthesis' in window;
  }

  /**
   * 获取可用声音
   */
  getVoices(): SpeechSynthesisVoice[] {
    return this.synth.getVoices();
  }

  /**
   * 获取中文声音
   */
  getChineseVoice(): SpeechSynthesisVoice | null {
    const voices = this.getVoices();
    return voices.find(v => v.lang.includes('zh')) || null;
  }

  /**
   * 朗读文本
   */
  speak(text: string, onEnd?: () => void): void {
    if (!this.isSupported()) {
      console.warn('Speech synthesis not supported');
      return;
    }

    this.stop();

    const utterance = new SpeechSynthesisUtterance(text);
    
    const voice = this.config.voice 
      ? this.getVoices().find(v => v.name === this.config.voice)
      : this.getChineseVoice();
    
    if (voice) {
      utterance.voice = voice;
    }
    
    utterance.rate = this.config.rate;
    utterance.pitch = this.config.pitch;
    utterance.volume = this.config.volume;

    utterance.onend = () => {
      this.isSpeaking = false;
      onEnd?.();
    };

    utterance.onerror = () => {
      this.isSpeaking = false;
    };

    this.synth.speak(utterance);
    this.isSpeaking = true;
  }

  /**
   * 停止朗读
   */
  stop(): void {
    this.synth.cancel();
    this.isSpeaking = false;
  }

  /**
   * 暂停朗读
   */
  pause(): void {
    this.synth.pause();
  }

  /**
   * 继续朗读
   */
  resume(): void {
    this.synth.resume();
  }

  /**
   * 获取状态
   */
  getStatus(): { isSpeaking: boolean; config: TTSConfig } {
    return {
      isSpeaking: this.isSpeaking,
      config: this.config,
    };
  }
}

export const ttsService = new TTSService();
