// Package nfc 定义NFC相关的消息类型
package nfc

// NFC 功能
// 1. NFC查询好友（设备 → 云端）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "find",
//     "nfc": "NFC1234567890",
//     "target": "AA:BB:CC:DD:EE:FF"
// }
// 2. 查询结果（云端 → 设备）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "find_res",
//     "nfc": "NFC1234567890",
//     "id": "AA:BB:CC:DD:EE:FF",
//     "name": "小小",
//     "avatar": "https://cdn.com/contact.jpg",
//     "rel": "friend"
// }
// 3. 添加好友请求（设备 → 云端）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "add_req",
//     "target": "AA:BB:CC:DD:EE:FF"
// }
// 4. 添加结果（云端 → 设备）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "add_res",
//     "target": "AA:BB:CC:DD:EE:FF",
//     "res": "success",
//     "name": "小小",
//     "avatar": "https://cdn.com/contact.jpg"
// }
// 5. 删除好友（设备 → 云端）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "del",
//     "target": "AA:BB:CC:DD:EE:FF"
// }
// 6. 被删除通知（云端 → 设备）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "deleted",
//     "from": "AA:BB:CC:DD:EE:FF"
// }
// 7. NFC卡制作（云端 → 设备）
// Topic: aibuddy/{device_id}/cmd/nfc

// {
//     "type": "create",
//     "nfc": "NFC1234567890",
//     "ctype": "birthday",
//     "cid": "content_123"
// }
// 8. 制作完成（设备 → 云端）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "created",
//     "nfc": "NFC1234567890",
//     "cid": "content_123",
//     "res": "success"
// }
// 9. 查询NFC信息（设备 → 云端）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "info_req",
//     "nfc": "NFC1234567890"
// }
// 10. NFC信息（云端 → 设备）
// Topic: aibuddy/{device_id}/nfc

// {
//     "type": "info",
//     "nfc": "NFC1234567890",
//     "ctype": "birthday",
//     "title": "生日快乐",
//     "content": "祝你快乐！",
//     "audio": "https://cdn.com/birthday.mp3"
// }

// Type 定义NFC消息类型
type Type string

const (
	// TypeFind 查询好友
	TypeFind Type = "find"
	// TypeFindRes 查询结果
	TypeFindRes Type = "find_res"
	// TypeAddReq 添加好友请求
	TypeAddReq Type = "add_req"
	// TypeAddRes 添加结果
	TypeAddRes Type = "add_res"
	// TypeDel 删除好友
	TypeDel Type = "del"
	// TypeDeleted 被删除通知
	TypeDeleted Type = "deleted"
	// TypeCreate NFC卡制作
	TypeCreate Type = "create"
	// TypeCreated 制作完成
	TypeCreated Type = "created"
	// TypeInfoReq 查询NFC信息
	TypeInfoReq Type = "info_req"
	// TypeInfo NFC信息
	TypeInfo Type = "info"
)

// BaseMsg 基础消息结构
type BaseMsg struct {
	Type Type `json:"type"`
}

// FindMsg 查询好友消息
type FindMsg struct {
	BaseMsg
	NFC    string `json:"nfc"`
	Target string `json:"target"`
}

// FindResMsg 查询结果消息
type FindResMsg struct {
	BaseMsg
	NFC    string `json:"nfc"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Rel    string `json:"rel"`
}

// AddReqMsg 添加好友请求消息
type AddReqMsg struct {
	BaseMsg
	Target string `json:"target"`
}

// AddResMsg 添加结果消息
type AddResMsg struct {
	BaseMsg
	Target string `json:"target"`
	Res    string `json:"res"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// DelMsg 删除好友消息
type DelMsg struct {
	BaseMsg
	Target string `json:"target"`
}

// DeletedMsg 被删除通知消息
type DeletedMsg struct {
	BaseMsg
	From string `json:"from"`
}

// CreateMsg 制作NFC卡消息
type CreateMsg struct {
	BaseMsg
	NFC   string `json:"nfc"`
	Ctype string `json:"ctype"`
	Cid   string `json:"cid"`
}

// CreatedMsg 制作完成消息
type CreatedMsg struct {
	BaseMsg
	NFC string `json:"nfc"`
	Cid string `json:"cid"`
	Res string `json:"res"`
}

// InfoReqMsg 请求NFC信息消息
type InfoReqMsg struct {
	BaseMsg
	NFC string `json:"nfc"`
}

// InfoMsg NFC信息消息
type InfoMsg struct {
	BaseMsg
	NFC     string `json:"nfc"`
	Ctype   string `json:"ctype"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Audio   string `json:"audio"`
}
