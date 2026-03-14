package baidu

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"aibuddy/pkg/config"
)

const testUniqID = "test_user_001"

func TestCreateCloneVoice(t *testing.T) {
	config.Setup("../../config")

	// 读取音频文件
	audioPath := `C:\Users\Administrator\Desktop\aaa.wav`
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("读取音频文件失败: %v", err)
	}

	// 转换为base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)
	t.Logf("音频文件大小: %d bytes, base64长度: %d", len(audioData), len(audioBase64))

	voice := NewTTSVoice()
	resp, err := voice.CreateCloneVoice(&CreateCloneVoiceRequest{
		UniqID:       testUniqID,
		Name:         "测试音色",
		Description:  "这是一个测试音色",
		AuditionText: "这是一段试听文本",
		Audios: []CloneAudio{
			{
				AudioBytes:  audioBase64,
				AudioFormat: "wav",
				Text:        "这是测试音频的文本内容",
			},
		},
		Language: 0,
	})
	if err != nil {
		t.Fatalf("CreateCloneVoice failed: %v", err)
	}

	t.Logf("CreateCloneVoice success, voice_id: %d", resp.VoiceID)
}

func TestRetrainCloneVoice(t *testing.T) {
	t.Skip("跳过实际执行，需要真实的音频数据和voiceID")

	config.Setup("../../config")

	// 读取音频文件
	audioPath := `C:\Users\Administrator\Desktop\aaa.wav`
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("读取音频文件失败: %v", err)
	}
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	voice := NewTTSVoice()
	err = voice.RetrainCloneVoice("testVoiceId", &RetrainCloneVoiceRequest{
		UniqID:       testUniqID,
		Name:         "更新后的音色名称",
		Description:  "更新后的音色描述",
		AuditionText: "更新后的试听文本",
		Audios: []CloneAudio{
			{
				AudioBytes:  audioBase64,
				AudioFormat: "wav",
				Text:        "这是重新训练的音频",
			},
		},
		Language: 0,
	})
	if err != nil {
		t.Fatalf("RetrainCloneVoice failed: %v", err)
	}

	t.Log("RetrainCloneVoice success")
}

func TestGetCloneVoiceList(t *testing.T) {
	config.Setup("../../config")

	voice := NewTTSVoice()
	resp, err := voice.GetCloneVoiceList(&CloneVoiceListRequest{
		UniqID: testUniqID,
	})
	if err != nil {
		t.Fatalf("GetCloneVoiceList failed: %v", err)
	}

	t.Logf("Total count: %d", resp.TotalCount)
	for i, item := range resp.Data {
		t.Logf("Voice %d: VoiceID=%d, Name=%s, Status=%s, Language=%s, UniqID=%s, CreateTime=%s",
			i+1, item.VoiceID, item.Name, item.Status, item.Language, item.UniqID, item.CreateTime)
	}
}

func TestGetCloneVoiceListByApp(t *testing.T) {
	config.Setup("../../config")

	voice := NewTTSVoice()
	// 不传UniqID，获取应用下所有音色
	resp, err := voice.GetCloneVoiceList(&CloneVoiceListRequest{})
	if err != nil {
		t.Fatalf("GetCloneVoiceList failed: %v", err)
	}

	t.Logf("App total count: %d", resp.TotalCount)
	for i, item := range resp.Data {
		t.Logf("Voice %d: VoiceID=%d, Name=%s, Status=%s, UniqID=%s",
			i+1, item.VoiceID, item.Name, item.Status, item.UniqID)
	}
}

func TestDeleteCloneVoice(t *testing.T) {
	t.Skip("跳过实际执行，需要真实的voiceID")

	config.Setup("../../config")

	voice := NewTTSVoice()
	err := voice.DeleteCloneVoice("", testUniqID, "testVoiceId")
	if err != nil {
		t.Fatalf("DeleteCloneVoice failed: %v", err)
	}

	t.Log("DeleteCloneVoice success")
}

// TestCreateAndDeleteCloneVoice 创建并删除音色的完整流程测试
func TestCreateAndDeleteCloneVoice(t *testing.T) {
	t.Skip("跳过实际执行，需要真实的音频数据")

	config.Setup("../../config")

	// 读取音频文件
	audioPath := `C:\Users\Administrator\Desktop\aaa.wav`
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("读取音频文件失败: %v", err)
	}
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	voice := NewTTSVoice()

	// 创建音色
	createResp, err := voice.CreateCloneVoice(&CreateCloneVoiceRequest{
		UniqID:       testUniqID,
		Name:         "流程测试音色",
		Description:  "用于测试完整流程",
		AuditionText: "测试试听",
		Audios: []CloneAudio{
			{
				AudioBytes:  audioBase64,
				AudioFormat: "wav",
				Text:        "测试文本",
			},
		},
		Language: 0,
	})
	if err != nil {
		t.Fatalf("CreateCloneVoice failed: %v", err)
	}

	t.Logf("Created voice_id: %d", createResp.VoiceID)

	// 查询列表确认创建成功
	listResp, err := voice.GetCloneVoiceList(&CloneVoiceListRequest{
		UniqID: testUniqID,
	})
	if err != nil {
		t.Fatalf("GetCloneVoiceList failed: %v", err)
	}

	var found bool
	for _, item := range listResp.Data {
		if item.UniqID == testUniqID {
			found = true
			t.Logf("Found voice in list: VoiceID=%d, Name=%s", item.VoiceID, item.Name)
			break
		}
	}

	if !found {
		t.Fatal("Created voice not found in list")
	}

	// 删除音色
	err = voice.DeleteCloneVoice("", testUniqID, fmt.Sprintf("%d", createResp.VoiceID))
	if err != nil {
		t.Fatalf("DeleteCloneVoice failed: %v", err)
	}

	t.Log("CreateAndDeleteCloneVoice flow success")
}
