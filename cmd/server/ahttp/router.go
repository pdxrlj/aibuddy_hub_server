// Package ahttp 提供应用层 HTTP 路由配置
package ahttp

import (
	anniversaryhandler "aibuddy/cmd/server/ahttp/handler/anniversary"
	devicehandler "aibuddy/cmd/server/ahttp/handler/device"
	emotionhandler "aibuddy/cmd/server/ahttp/handler/emotion"
	filehandler "aibuddy/cmd/server/ahttp/handler/file"
	"aibuddy/cmd/server/ahttp/handler/nfc"
	otahandler "aibuddy/cmd/server/ahttp/handler/ota"
	remindhandler "aibuddy/cmd/server/ahttp/handler/remind"

	ttsvoicehandler "aibuddy/cmd/server/ahttp/handler/tts_voice"

	rolehandler "aibuddy/cmd/server/ahttp/handler/role"
	shophandler "aibuddy/cmd/server/ahttp/handler/shop"

	userhandler "aibuddy/cmd/server/ahttp/handler/user"
	websockethandler "aibuddy/cmd/server/ahttp/handler/websocket"
	"aibuddy/cmd/server/ahttp/middleware"
	"aibuddy/pkg/ahttp"

	"github.com/labstack/echo/v4"
)

// DemmoRequest 演示请求结构
type DemmoRequest struct {
	Name string `json:"name" validate:"required,min=3,max=10"`
	Age  int    `json:"age" validate:"required,min=18,max=100"`
}

// RegisterRoutes 注册认证路由
func RegisterRoutes(base *ahttp.Base) {
	base.Group("/api/v1", nil, func(group *ahttp.Group) {
		group.Group("/device", nil, func(deviceGroup *ahttp.Group) {
			device := devicehandler.NewDevice()
			deviceGroup.POST("/firstonline", device.FirstOnline)

			// 获取设备的位置信息
			deviceGroup.GET("/location", device.GetLocation)

			// 获取好友列表
			deviceGroup.GET("/:device_id/friends", device.GetFriends)

			// 查询好友信息
			deviceGroup.GET("/:device_id/device_info", device.GetDeviceInfo)

			// 添加好友
			deviceGroup.POST("/:device_id/add_friend", device.AddFriend)

			// 删除好友
			deviceGroup.DELETE("/:device_id/delete_friend", device.DeleteFriend)

			// 消息发送 文本/语音
			// TODO 缺少对家长端的websocket消息发送
			deviceGroup.POST("/:device_id/send_message", device.SendMessage)

			// 获取设备消息列表
			deviceGroup.GET("/:device_id/message_list", device.MessageList)

			// 大模型互动实例
			deviceGroup.Group("/aiagent/:device_id", nil, func(rtcGroup *ahttp.Group) {
				rtc := devicehandler.NewRtcHandler()
				// 与端侧SDK交互的接口
				rtcGroup.POST("/generateAIAgentCall", rtc.GenerateAIAgentCall)
				rtcGroup.POST("/stopAIAgentInstance", rtc.StopAIAgentInstance)
				rtcGroup.POST("/switchSceneRole", rtc.SwitchSceneRole)

				// userserver接口
				rtcGroup.POST("/userserver/instance/generate", rtc.InstanceGenerate)
				rtcGroup.POST("/userserver/instance/stop", rtc.InstanceStop)
				rtcGroup.POST("/userserver/auth/generate", rtc.AuthGenerate)
				rtcGroup.POST("/userserver/instance/baidu", rtc.InstanceBaidu)
				rtcGroup.POST("/userserver/instance/qianwen", rtc.InstanceQianwen)
				rtcGroup.POST("/userserver/instance/volc", rtc.InstanceVolc)
			})
		})

		group.Group("/file", nil, func(fileGroup *ahttp.Group) {
			f := filehandler.NewFile()
			// 上传文件（表单方式）
			fileGroup.POST("/:device_id/upload_file", f.UploadFile)

			// 流式上传文件
			fileGroup.POST("/:device_id/upload_stream", f.UploadStream)

			// 文件代理
			fileGroup.GET("/:device_id/file_proxy", f.FileProxy)
		})

		group.Group("/ota", nil, func(otaGroup *ahttp.Group) {
			ota := otahandler.NewOta()
			otaGroup.POST("/send_to_device", ota.SendToDevice)
		})

		group.Group("/user", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(userGroup *ahttp.Group) {
			h := userhandler.New()
			userGroup.POST("/send_code", h.SendCode)         // 发送验证码
			userGroup.POST("/login", h.Login)                // 登录接口
			userGroup.POST("/refresh_token", h.RefreshToken) // 刷新token
			userGroup.GET("/info", h.GetUserInfo)            // 获取用户信息
			userGroup.POST("/update", h.UpdateInfo)          // 修改用户信息
			userGroup.POST("/logout", h.Logout)              // 退出接口

			// 完善用户信息，扫描绑定后完善用户信息
			userGroup.POST("/profile", h.CompleteProfile)

			// 用户是否已经绑定了设备
			userGroup.GET("/have_device", h.HaveDevice)

			// 设备列表
			userGroup.GET("/device_list", h.DeviceList)

			// 发送挂失消息给设备
			userGroup.POST("/lost", h.Lost)

			// 发送解除挂失消息给设备
			userGroup.POST("/unlost", h.Unlost)

			// 发送解绑消息给设备
			userGroup.POST("/unbind", h.Unbind)

			// 分析用户成长报告
			userGroup.GET("/analysis_growth_report", h.AnalysisGrowthReport)
			userGroup.GET("/get_growth_report_list", h.GetGrowthReportList)

			// 小程序留言
			userGroup.POST("/message", h.LeavaMessage)
			userGroup.GET("/message_list", h.MessageList)
		})

		// 情绪
		group.Group("/emotion", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(emotionGroup *ahttp.Group) {
			h := emotionhandler.NewHandler()
			emotionGroup.GET("/:device_id/list", h.GetEmotions)
			emotionGroup.GET("/:device_id/latest", h.GetLatestEmotion)
		})

		group.Group("/role", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(role *ahttp.Group) {
			r := rolehandler.NewRoleHandler()
			role.GET("/list", r.RoleList)
			role.POST("/change", r.ChangeRole) // 切换角色

			role.GET("/:device_id/get_chat_analysis", r.GetChatAnalysis)
			role.GET("/:device_id/refresh_chat_analysis", r.RefreshChatAnalysis)
		})

		group.Group("/remind", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(group *ahttp.Group) {
			m := remindhandler.NewRemindHandler()
			group.POST("/create", m.CreateRemind) // 添加提醒事件
			group.POST("/update", m.UpdateRemind) // 更新提醒事件
			group.POST("/delete", m.DeleteRemind) // 删除提醒事件
			group.GET("/list", m.ListRemind)      // 提醒事件列表
		})

		group.Group("/anniversary", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(group *ahttp.Group) {
			m := anniversaryhandler.NewAnniversaryHandler()
			group.POST("/create", m.CreateAnniversary)  // 添加纪念日
			group.POST("/update", m.UpdateAnniversary)  // 更新纪念日
			group.POST("/delete", m.DeleateAnniversary) // 删除纪念日
			group.GET("/list", m.ListAnniversary)       // 纪念日列表
		})

		group.GET("/websocket", websockethandler.NewHandler().HandleConnect, middleware.UnifiedAuthMiddleware())

		group.Group("/nfc", nil, func(group *ahttp.Group) {
			nfcHandler := nfc.NewHandler()
			group.POST("/create", nfcHandler.CreateNFC, middleware.UnifiedAuthMiddleware())

			// 获取NFC信息
			group.GET("/:nfc_id/info", nfcHandler.GetNFCInfo)

			// 获取NFC列表
			group.GET("/list/:device_id", nfcHandler.GetNFCList, middleware.UnifiedAuthMiddleware())

			// 更新NFC
			group.PUT("/:cid", nfcHandler.UpdateNFC, middleware.UnifiedAuthMiddleware())

			// 删除NFC
			group.DELETE("/:cid", nfcHandler.DeleteNFC, middleware.UnifiedAuthMiddleware())
		})

		// TTS语音复刻
		group.Group("/tts_voice", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(group *ahttp.Group) {
			tts := ttsvoicehandler.New()
			group.POST("/create", tts.CreateCloneVoice)               // 创建复刻音色
			group.PUT("/:voice_id/retrain", tts.RetrainCloneVoice)    // 重新训练复刻音色
			group.GET("/list", tts.GetCloneVoiceList)                 // 获取音色列表
			group.DELETE("/:uniq_id/:voice_id", tts.DeleteCloneVoice) // 删除音色
		})

		group.Group("/minishop", []echo.MiddlewareFunc{middleware.UnifiedAuthMiddleware()}, func(group *ahttp.Group) {
			s := shophandler.NewShopHandler()
			group.GET("/goods_list", s.GoodsList)
		})
	})
}
