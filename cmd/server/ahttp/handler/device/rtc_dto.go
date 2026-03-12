// Package devicehandler provides the device handler for the server.
package devicehandler

// GenerateAIAgentCallRequest is the request for the GenerateAIAgentCall service.
type GenerateAIAgentCallRequest struct {
	AppID        string         `json:"app_id"`
	InstanceType string         `json:"instance_type"`
	Config       *AIAgentConfig `json:"config"`
}

// AIAgentConfig is the config for the AIAgent.
type AIAgentConfig struct {
	Location              string `json:"location"`
	Role                  string `json:"role"`
	SceneRole             string `json:"scene_role"`
	TTSSayHi              string `json:"tts_sayhi"`
	Lang                  string `json:"lang"`
	AudioCodec            string `json:"audio_codec"`
	AudioBitrate          int    `json:"audio_bitrate"`
	DisableVoiceAutoInt   bool   `json:"disable_voice_auto_int"`
	InterruptionWords     string `json:"interruption_words"`
	InterruptionAnswer    string `json:"interruption_answer"`
	AutoPauseSecs         int    `json:"auto_pause_secs"`
	TTS                   string `json:"tts"`
	TTSURL                string `json:"tts_url"`
	TTSEndDelayMs         int    `json:"tts_end_delay_ms"`
	LLM                   string `json:"llm"`
	LLMURL                string `json:"llm_url"`
	LLMCfg                string `json:"llm_cfg"`
	LLMToken              string `json:"llm_token"`
	E2ELLMode             string `json:"e2ellm_mode"`
	E2ELLSampleRate       int    `json:"e2ellm_sample_rate"`
	UserID                string `json:"user_id"`
	ContentUsedDB         string `json:"content_used_db"`
	ASRLongAudioMode      bool   `json:"asr_long_audio_mode"`
	ASRVADAppend          bool   `json:"asr_vad_append"`
	ASRVAD                int    `json:"asr_vad"`
	ASRVADLevel           int    `json:"asr_vad_level"`
	VoiceFPURL            string `json:"voice_fp_url"`
	RemoteMusicPlayer     bool   `json:"remote_music_player"`
	EmotionRecognitionCfg string `json:"emotion_recognition_cfg"`
	Cloud3AURL            string `json:"cloud_3a_url"`
}
