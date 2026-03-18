// Package role 服务层
package role

import (
	"aibuddy/internal/model"
	"aibuddy/internal/repository"
	"aibuddy/internal/services/agent"
	"aibuddy/internal/services/websocket"
	"aibuddy/pkg/baidu"
	"aibuddy/pkg/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var tracer = func() trace.Tracer {
	return otel.Tracer(config.Instance.Tracer.ServiceName)
}

var defaultTTS = func() string {
	vcn := config.Instance.Baidu.TTS.Vcn
	if vcn == "" {
		vcn = "1000578"
	}
	return fmt.Sprintf(`DEFAULT{"vcn":"%s"}`, vcn)
}

// Service 角色服务
type Service struct {
	AgentRepo  *repository.AgentRepo
	DeviceRepo *repository.DeviceRepo
	RoleAPI    *baidu.Role
	SwitchRole *baidu.SwitchRole

	UserAgentRepo *repository.UserAgentRepository

	RoleAgentService *agent.RoleAgentService

	// 生成中的任务缓存，防止重复生成
	generatingMu sync.Mutex
	generating   map[string]bool
}

// NewRoleService 实例化服务
func NewRoleService() *Service {
	return &Service{
		AgentRepo:     repository.NewAgentRepo(),
		DeviceRepo:    repository.NewDeviceRepo(),
		RoleAPI:       baidu.NewRole(),
		SwitchRole:    baidu.NewSwitchRole(),
		UserAgentRepo: repository.NewUserAgentRepository(),

		RoleAgentService: agent.NewRoleAgentService(),
		generating:       make(map[string]bool),
	}
}

// GetRoleListByUID 获取role列表
func (r *Service) GetRoleListByUID(ctx context.Context, uid int64, page int, size int) ([]*model.Agent, int64, error) {
	ctx, span := tracer().Start(ctx, "UpsertUser")
	defer span.End()
	data, count, err := r.AgentRepo.GetAgentListByUID(ctx, uid, page, size)

	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

// ChangeRoleName 切换设备角色
func (r *Service) ChangeRoleName(ctx context.Context, uid int64, deviceID, instanceID string, roleName string) error {
	ctx, span := tracer().Start(ctx, "ChangeRoleName")
	defer span.End()

	if deviceID == "" {
		return errors.New("设备ID和实例ID不能同时为空")
	}

	if !r.DeviceRepo.CheckDeviceAuth(ctx, uid, deviceID) {
		return errors.New("无设置该设备的权限")
	}

	if err := r.DeviceRepo.ChangeDeviceRole(ctx, uid, deviceID, roleName); err != nil {
		return fmt.Errorf("切换角色失败: %w", err)
	}

	if instanceID != "" {
		if err := r.SwitchRole.SwitchSceneRole(&baidu.SwitchRoleRequest{
			AiAgentInstanceID: cast.ToUint64(instanceID),
			SceneRole:         roleName,
			TTS:               defaultTTS(),
		}); err != nil {
			return fmt.Errorf("切换角色失败: %w", err)
		}
	}

	return nil
}

// GetDeviceAgentName 获取设备角色名称
func (r *Service) GetDeviceAgentName(ctx context.Context, deviceID string) (string, error) {
	ctx, span := tracer().Start(ctx, "GetDeviceAgentName")
	defer span.End()

	device, err := r.DeviceRepo.GetDeviceInfo(ctx, deviceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("[RoleService] 获取设备失败", "error", err)
		return "", errors.New("获取设备失败")
	}
	if device == nil {
		// slog.Error("[RoleService] 设备不存在", "device_id", deviceID)
		return "", errors.New("设备不存在")
	}

	return device.AgentName, nil
}

// DeviceInstanceSwitchDefRole 切换设备实例到默认角色
func (r *Service) DeviceInstanceSwitchDefRole(ctx context.Context, instanceID uint64, deviceID string) error {
	ctx, span := tracer().Start(ctx, "DeviceInstanceSwitchDefRole")
	defer span.End()

	agentName, err := r.GetDeviceAgentName(ctx, deviceID)
	if err != nil {
		slog.Info("[RoleService] 设备角色不存在，不进行角色切换", "device_id", deviceID)
		return nil
	}

	if agentName != "" {
		if err := r.SwitchRole.SwitchSceneRole(&baidu.SwitchRoleRequest{
			AiAgentInstanceID: instanceID,
			SceneRole:         agentName,
			TTS:               `DEFAULT{"vcn":"1000454"}`,
		}); err != nil {
			return errors.New("切换角色失败:" + err.Error())
		}
	}

	return nil
}

// GetRoleByID 查看角色信息
func (r *Service) GetRoleByID(ctx context.Context, uid int64, roleID int64) (*model.Agent, error) {
	ctx, span := tracer().Start(ctx, "GetRoleByID")
	defer span.End()

	data, err := r.AgentRepo.GetAgentByID(ctx, uid, roleID)
	if err != nil {
		span.RecordError(errors.New("角色信息为空"))
		return nil, errors.New("角色信息为空")
	}

	return data, nil
}

// GetRoleListByAPI 通过API接口拉取角色列表
func (r *Service) GetRoleListByAPI(ctx context.Context) ([]*model.Agent, error) {
	_, span := tracer().Start(ctx, "GetRoleListByAPI")
	defer span.End()
	resp, err := r.RoleAPI.RoleList("")
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	result := make([]*model.Agent, 0, len(resp.LLM.Roles))
	for _, r := range resp.LLM.Roles {
		result = append(result, &model.Agent{
			ID:               r.ID,
			AgentName:        r.Name,
			DefaultUsage:     r.DefaultUsage,
			RoleIntroduction: r.Description,
		})
	}

	return result, nil
}

// setSpanAttrs 设置 span 公共属性
func setSpanAttrs(span trace.Span, deviceID string, startTime, endTime time.Time, agentName string) {
	span.SetAttributes(
		attribute.String("device_id", deviceID),
		attribute.String("start_time", startTime.Format(time.DateTime)),
		attribute.String("end_time", endTime.Format(time.DateTime)),
		attribute.String("agent_name", agentName),
	)
}

// sendReport 发送报告给前端
func (r *Service) sendReport(uid int64, deviceID, agentName string, msg []byte, frameType websocket.FrameType) {
	frame := &websocket.RoleGenerateReportFrame{
		Type:      frameType,
		DeviceID:  deviceID,
		AgentName: agentName,
		Message:   msg,
	}
	websocket.SendMessage(cast.ToString(uid), frame)
}

// sendError 发送错误消息给前端
func (r *Service) sendError(uid int64, deviceID, agentName, errMsg string) {
	frame := &websocket.RoleGenerateReportFrame{
		Type:      websocket.FrameTypeRoleGenerateFailure,
		DeviceID:  deviceID,
		AgentName: agentName,
		Error:     errMsg,
	}
	websocket.SendMessage(cast.ToString(uid), frame)
}

// GetChatAnalysis 获取聊天分析
func (r *Service) GetChatAnalysis(ctx context.Context, deviceID string, agentName string) (*model.UserAgent, error) {
	ctx, span := tracer().Start(ctx, "GetChatAnalysis")
	defer span.End()
	setSpanAttrs(span, deviceID, time.Time{}, time.Time{}, agentName)

	data, err := r.UserAgentRepo.GetUserAgent(ctx, deviceID, agentName)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return data, nil
}

// RefreshChatAnalysis 刷新聊天分析
func (r *Service) RefreshChatAnalysis(ctx context.Context, uid int64, deviceID string, startTime, endTime time.Time, agentName string) (*model.UserAgent, error) {
	_, span := tracer().Start(ctx, "ChatAnalysis")
	defer span.End()
	setSpanAttrs(span, deviceID, startTime, endTime, agentName)

	taskKey := fmt.Sprintf("%s:%s", deviceID, agentName)

	r.generatingMu.Lock()
	if r.generating[taskKey] {
		r.generatingMu.Unlock()
		return nil, errors.New("当前任务正在生成中，请勿重复提交")
	}
	r.generating[taskKey] = true
	r.generatingMu.Unlock()

	go func() {
		defer r.clearGeneratingTask(taskKey)

		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		ctx, agentSpan := tracer().Start(bgCtx, "RoleChatAgent")
		defer agentSpan.End()
		setSpanAttrs(agentSpan, deviceID, startTime, endTime, agentName)

		report, err := r.RoleAgentService.RoleChatAgent(deviceID, startTime, endTime, agentName)
		if err != nil {
			agentSpan.RecordError(err)
			r.sendError(uid, deviceID, agentName, "生成分析报告失败了")
			return
		}

		conversationAnalysis, err := report.ConversationAnalysis.Encode()
		if err != nil {
			agentSpan.RecordError(err)
			r.sendError(uid, deviceID, agentName, "生成对话分析失败了")
			return
		}

		emotionAnalysis, err := report.EmotionAnalysis.Encode()
		if err != nil {
			agentSpan.RecordError(err)
			r.sendError(uid, deviceID, agentName, "生成情绪分析失败了")
			return
		}
		slog.Info("[RoleService] 生成分析报告成功", "device_id", deviceID, "agent_name", agentName)
		if err := r.UserAgentRepo.CreateUserAgent(ctx, &model.UserAgent{
			DeviceID:             deviceID,
			AgentName:            agentName,
			ConversationAnalysis: conversationAnalysis,
			EmotionAnalysis:      emotionAnalysis,
		}); err != nil {
			agentSpan.RecordError(err)
			r.sendError(uid, deviceID, agentName, "保存记录失败了")
			return
		}

		data, err := r.UserAgentRepo.GetUserAgent(ctx, deviceID, agentName)
		if err != nil {
			agentSpan.RecordError(err)
			r.sendError(uid, deviceID, agentName, "获取记录失败了")
			return
		}

		message, err := json.Marshal(data)
		if err != nil {
			agentSpan.RecordError(err)
			r.sendError(uid, deviceID, agentName, "序列化记录失败了")
			return
		}

		r.sendReport(uid, deviceID, agentName, message, websocket.FrameTypeRoleGenerateReport)
	}()

	return nil, nil
}

// clearGeneratingTask 清除正在生成的任务标记
func (r *Service) clearGeneratingTask(taskKey string) {
	r.generatingMu.Lock()
	delete(r.generating, taskKey)
	r.generatingMu.Unlock()
}
