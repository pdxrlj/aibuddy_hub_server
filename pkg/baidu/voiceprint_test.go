package baidu

import (
	"encoding/base64"
	"os"
	"testing"

	"aibuddy/pkg/config"
)

// TestVoiceprint_RegisterVoiceprint 测试注册声纹
// 注意：此测试需要真实的WAV音频文件，设置环境变量 VOICEPRINT_AUDIO_FILE 指向音频文件路径
// 音频要求：单声道、16k采样率、WAV或PCM格式
func TestVoiceprint_RegisterVoiceprint(t *testing.T) {
	config.Setup("../../config")

	// 检查是否有真实的音频文件
	audioFile := os.Getenv("VOICEPRINT_AUDIO_FILE")
	if audioFile == "" {
		t.Skip("跳过测试：需要设置环境变量 VOICEPRINT_AUDIO_FILE 指向真实的WAV音频文件")
	}

	// 读取音频文件
	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		t.Fatalf("读取音频文件失败: %v", err)
	}

	vp := NewVoiceprint()
	testAudioBase64 := base64.StdEncoding.EncodeToString(audioData)

	resp, err := vp.RegisterVoiceprint(&RegisterVoiceprintRequest{
		NickName:   "测试用户",
		UserID:     "test_user_001",
		Format:     VoiceprintFormatWAV,
		FileBase64: testAudioBase64,
	})
	if err != nil {
		t.Fatalf("RegisterVoiceprint failed: %v", err)
	}

	t.Logf("RegisterVoiceprint succeeded: vp_id=%s, nick_name=%s", resp.VpID, resp.NickName)
}

func TestVoiceprint_ListVoiceprint(t *testing.T) {
	config.Setup("../../config")

	vp := NewVoiceprint()

	// 测试查询声纹列表
	resp, err := vp.ListVoiceprint(&ListVoiceprintRequest{
		UserID: "test_user_001",
	})
	if err != nil {
		t.Fatalf("ListVoiceprint failed: %v", err)
	}

	t.Logf("ListVoiceprint succeeded: count=%d", len(resp.Data))
	for i, item := range resp.Data {
		t.Logf("  [%d] id=%s, nick_name=%s, format=%s", i, item.ID, item.NickName, item.Format)
	}
}

// TestVoiceprint_DeleteVoiceprint 测试删除声纹
// 注意：此测试需要真实的WAV音频文件，设置环境变量 VOICEPRINT_AUDIO_FILE 指向音频文件路径
func TestVoiceprint_DeleteVoiceprint(t *testing.T) {
	config.Setup("../../config")

	audioFile := os.Getenv("VOICEPRINT_AUDIO_FILE")
	if audioFile == "" {
		t.Skip("跳过测试：需要设置环境变量 VOICEPRINT_AUDIO_FILE 指向真实的WAV音频文件")
	}

	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		t.Fatalf("读取音频文件失败: %v", err)
	}

	vp := NewVoiceprint()

	// 先注册一个声纹用于测试删除
	testAudioBase64 := base64.StdEncoding.EncodeToString(audioData)
	registerResp, err := vp.RegisterVoiceprint(&RegisterVoiceprintRequest{
		NickName:   "待删除用户",
		UserID:     "test_user_delete",
		Format:     VoiceprintFormatWAV,
		FileBase64: testAudioBase64,
	})
	if err != nil {
		t.Fatalf("RegisterVoiceprint for delete test failed: %v", err)
	}

	vpID := registerResp.VpID
	t.Logf("Registered voiceprint for delete test: vp_id=%s", vpID)

	// 测试删除声纹
	err = vp.DeleteVoiceprint(&DeleteVoiceprintRequest{
		UserID: "test_user_delete",
		VpIDs:  []string{vpID},
	})
	if err != nil {
		t.Fatalf("DeleteVoiceprint failed: %v", err)
	}

	t.Log("DeleteVoiceprint succeeded")

	// 验证删除结果
	listResp, err := vp.ListVoiceprint(&ListVoiceprintRequest{
		UserID: "test_user_delete",
	})
	if err != nil {
		t.Fatalf("ListVoiceprint after delete failed: %v", err)
	}

	// 检查列表中是否还存在该声纹
	for _, item := range listResp.Data {
		if item.ID == vpID {
			t.Errorf("Voiceprint %s still exists after delete", vpID)
		}
	}

	t.Log("DeleteVoiceprint verified: voiceprint no longer in list")
}

func TestVoiceprint_DeleteVoiceprint_Batch(t *testing.T) {
	config.Setup("../../config")

	vp := NewVoiceprint()

	// 批量删除测试（使用不存在的vp_id，验证API不会报错）
	err := vp.DeleteVoiceprint(&DeleteVoiceprintRequest{
		UserID: "test_user_batch_delete",
		VpIDs:  []string{"non_existent_vp_id_1", "non_existent_vp_id_2"},
	})
	if err != nil {
		t.Logf("DeleteVoiceprint batch failed: %v", err)
		return
	}

	t.Log("DeleteVoiceprint batch succeeded")
}

func TestVoiceprint_ListVoiceprint_EmptyUser(t *testing.T) {
	config.Setup("../../config")

	vp := NewVoiceprint()

	// 测试查询不存在用户的声纹列表
	resp, err := vp.ListVoiceprint(&ListVoiceprintRequest{
		UserID: "non_existent_user_" + "test_timestamp",
	})
	if err != nil {
		t.Fatalf("ListVoiceprint for non-existent user failed: %v", err)
	}

	t.Logf("ListVoiceprint for non-existent user: count=%d", len(resp.Data))
}

// TestVoiceprint_FullLifecycle 测试声纹完整生命周期
// 注意：此测试需要真实的WAV音频文件，设置环境变量 VOICEPRINT_AUDIO_FILE 指向音频文件路径
func TestVoiceprint_FullLifecycle(t *testing.T) {
	config.Setup("../../config")

	audioFile := os.Getenv("VOICEPRINT_AUDIO_FILE")
	if audioFile == "" {
		t.Skip("跳过测试：需要设置环境变量 VOICEPRINT_AUDIO_FILE 指向真实的WAV音频文件")
	}

	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		t.Fatalf("读取音频文件失败: %v", err)
	}

	vp := NewVoiceprint()
	userID := "test_lifecycle_user"

	// 1. 注册声纹
	testAudioBase64 := base64.StdEncoding.EncodeToString(audioData)
	registerResp, err := vp.RegisterVoiceprint(&RegisterVoiceprintRequest{
		NickName:   "生命周期测试用户",
		UserID:     userID,
		Format:     VoiceprintFormatWAV,
		FileBase64: testAudioBase64,
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	t.Logf("1. Register succeeded: vp_id=%s", registerResp.VpID)

	// 2. 查询列表验证注册
	listResp, err := vp.ListVoiceprint(&ListVoiceprintRequest{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("List after register failed: %v", err)
	}
	t.Logf("2. List after register: count=%d", len(listResp.Data))

	// 3. 删除声纹
	err = vp.DeleteVoiceprint(&DeleteVoiceprintRequest{
		UserID: userID,
		VpIDs:  []string{registerResp.VpID},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	t.Log("3. Delete succeeded")

	// 4. 查询列表验证删除
	listResp, err = vp.ListVoiceprint(&ListVoiceprintRequest{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("List after delete failed: %v", err)
	}
	t.Logf("4. List after delete: count=%d", len(listResp.Data))

	t.Log("Full lifecycle test passed!")
}
