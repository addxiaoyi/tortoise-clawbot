import Foundation
import AVFoundation
import Speech
import NaturalLanguage

/// Voice Wake Forward Manager
/// 语音唤醒转发功能 - 当检测到唤醒词时，自动转发到 Tortoise
class VoiceWakeManager: NSObject, ObservableObject {
    
    // MARK: - Properties
    
    @Published var isListening = false
    @Published var isWakeWordDetected = false
    @Published var lastTranscription = ""
    @Published var audioLevel: Float = 0
    
    private var audioEngine: AVAudioEngine?
    private var speechRecognizer: SFSpeechRecognizer?
    private var recognitionRequest: SFSpeechAudioBufferRecognitionRequest?
    private var recognitionTask: SFSpeechRecognitionTask?
    
    private let wakeWord = "Hey Tortoise"
    private let commandTemplate = "openclaw-mac agent --message \"{text}\" --thinking low"
    
    // 配置
    var serverURL: String = "http://localhost:8080"
    var shellPath: String = "/bin/bash"
    
    // 回调
    var onWakeWordDetected: (() -> Void)?
    var onCommandExecuted: ((String) -> Void)?
    var onError: ((Error) -> Void)?
    
    // MARK: - Lifecycle
    
    override init() {
        super.init()
        setupAudioSession()
        setupSpeechRecognizer()
    }
    
    // MARK: - Setup
    
    private func setupAudioSession() {
        do {
            let session = AVAudioSession.sharedInstance()
            try session.setCategory(.playAndRecord, mode: .default, options: [.defaultToSpeaker, .allowBluetooth])
            try session.setActive(true)
        } catch {
            print("Failed to setup audio session: \(error)")
        }
    }
    
    private func setupSpeechRecognizer() {
        speechRecognizer = SFSpeechRecognizer(locale: Locale(identifier: "en-US"))
        speechRecognizer?.delegate = self
    }
    
    // MARK: - Public Methods
    
    /// 开始监听
    func startListening() throws {
        guard !isListening else { return }
        
        // 请求权限
        SFSpeechRecognizer.requestAuthorization { [weak self] status in
            DispatchQueue.main.async {
                switch status {
                case .authorized:
                    try? self?.startRecognition()
                case .denied, .restricted:
                    self?.onError?(VoiceWakeError.speechRecognitionNotAuthorized)
                case .notDetermined:
                    self?.onError?(VoiceWakeError.speechRecognitionNotDetermined)
                @unknown default:
                    break
                }
            }
        }
        
        // 请求麦克风权限
        AVCaptureDevice.requestAccess(for: .audio) { [weak self] granted in
            if !granted {
                DispatchQueue.main.async {
                    self?.onError?(VoiceWakeError.microphoneAccessDenied)
                }
            }
        }
    }
    
    /// 停止监听
    func stopListening() {
        isListening = false
        audioEngine?.stop()
        recognitionRequest?.endAudio()
        recognitionTask?.cancel()
        recognitionTask = nil
        recognitionRequest = nil
    }
    
    /// 手动发送命令
    func sendCommand(_ text: String) {
        executeCommand(text)
    }
    
    // MARK: - Private Methods
    
    private func startRecognition() throws {
        // 取消之前的任务
        recognitionTask?.cancel()
        recognitionTask = nil
        
        // 创建音频引擎
        audioEngine = AVAudioEngine()
        guard let audioEngine = audioEngine else {
            throw VoiceWakeError.audioEngineFailed
        }
        
        let inputNode = audioEngine.inputNode
        let recordingFormat = inputNode.outputFormat(forBus: 0)
        
        // 创建识别请求
        recognitionRequest = SFSpeechAudioBufferRecognitionRequest()
        guard let recognitionRequest = recognitionRequest else {
            throw VoiceWakeError.recognitionRequestFailed
        }
        
        recognitionRequest.shouldReportPartialResults = true
        recognitionRequest.requiresOnDeviceRecognition = true
        
        // 开始识别
        recognitionTask = speechRecognizer?.recognitionTask(with: recognitionRequest) { [weak self] result, error in
            guard let self = self else { return }
            
            if let result = result {
                let transcription = result.bestTranscription.formattedString
                self.lastTranscription = transcription
                
                // 检查是否包含唤醒词
                if transcription.lowercased().contains(self.wakeWord.lowercased()) {
                    self.handleWakeWordDetected(transcription)
                }
            }
            
            if error != nil || result?.isFinal == true {
                // 重新开始监听
                if self.isListening {
                    try? self.startRecognition()
                }
            }
        }
        
        // 配置音频输入
        inputNode.installTap(onBus: 0, bufferSize: 1024, format: recordingFormat) { [weak self] buffer, _ in
            self?.recognitionRequest?.append(buffer)
            
            // 计算音频级别
            let level = self?.calculateAudioLevel(buffer: buffer) ?? 0
            DispatchQueue.main.async {
                self?.audioLevel = level
            }
        }
        
        // 启动引擎
        audioEngine.prepare()
        try audioEngine.start()
        isListening = true
    }
    
    private func handleWakeWordDetected(_ transcription: String) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            self.isWakeWordDetected = true
            self.onWakeWordDetected?()
            
            // 提取唤醒词后的命令
            var command = transcription
            if let range = command.lowercased().range(of: self.wakeWord.lowercased()) {
                command = String(command[range.upperBound...]).trimmingCharacters(in: .whitespaces)
            }
            
            if !command.isEmpty {
                self.executeCommand(command)
            }
            
            // 延迟重置状态
            DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
                self.isWakeWordDetected = false
            }
        }
    }
    
    private func executeCommand(_ text: String) {
        let command = commandTemplate.replacingOccurrences(of: "{text}", with: text)
        
        // 使用 NSTask 执行命令 (在 macOS 上)
        let task = Process()
        task.launchPath = shellPath
        task.arguments = ["-c", command]
        
        let pipe = Pipe()
        task.standardOutput = pipe
        task.standardError = pipe
        
        do {
            try task.run()
            task.waitUntilExit()
            
            let data = pipe.fileHandleForReading.readDataToEndOfFile()
            let output = String(data: data, encoding: .utf8) ?? ""
            
            DispatchQueue.main.async { [weak self] in
                self?.onCommandExecuted?(output)
            }
        } catch {
            DispatchQueue.main.async { [weak self] in
                self?.onError?(error)
            }
        }
    }
    
    private func calculateAudioLevel(buffer: AVAudioPCMBuffer) -> Float {
        guard let channelData = buffer.floatChannelData?[0] else { return 0 }
        let frameLength = UInt32(buffer.frameLength)
        
        var sum: Float = 0
        for i in 0..<Int(frameLength) {
            sum += channelData[i] * channelData[i]
        }
        
        let rms = sqrt(sum / Float(frameLength))
        let db = 20 * log10(rms)
        let normalizedLevel = max(0, min(1, (db + 50) / 50))
        
        return normalizedLevel
    }
}

// MARK: - SFSpeechRecognizerDelegate

extension VoiceWakeManager: SFSpeechRecognizerDelegate {
    func speechRecognizer(_ speechRecognizer: SFSpeechRecognizer, availabilityDidChange available: Bool) {
        if !available && isListening {
            stopListening()
            onError?(VoiceWakeError.speechRecognizerUnavailable)
        }
    }
}

// MARK: - Errors

enum VoiceWakeError: LocalizedError {
    case speechRecognitionNotAuthorized
    case speechRecognitionNotDetermined
    case microphoneAccessDenied
    case audioEngineFailed
    case recognitionRequestFailed
    case speechRecognizerUnavailable
    
    var errorDescription: String? {
        switch self {
        case .speechRecognitionNotAuthorized:
            return "语音识别未授权，请在系统设置中授权"
        case .speechRecognitionNotDetermined:
            return "语音识别权限未确定"
        case .microphoneAccessDenied:
            return "麦克风访问被拒绝，请在系统设置中授权"
        case .audioEngineFailed:
            return "音频引擎启动失败"
        case .recognitionRequestFailed:
            return "语音识别请求创建失败"
        case .speechRecognizerUnavailable:
            return "语音识别服务不可用"
        }
    }
}
