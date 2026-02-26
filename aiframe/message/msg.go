// Package message 实现了好友与留言相关功能
package message

// 好友与留言
// 1. 请求好友列表（设备 → 云端）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "friends_req",
//     "page": 1,
//     "ps": 20
// }
// 2. 好友列表（云端 → 设备）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "friends",
//     "total": 10,
//     "list": [
//         {
//             "id": "AA:BB:CC:DD:EE:FF",
//             "name": "妈妈",
//             "avatar": "https://cdn.com/a.jpg",
//             "rel": "family"
//         }
//     ]
// }
// 3. 发送留言（设备 → 云端）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "send",
//     "mid": "msg_123",
//     "to": "AA:BB:CC:DD:EE:FF",
//     "fmt": "text",
//     "content": "你好吗？",
//     "dur": 0
// }
// 4. 接收留言（云端 → 设备）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "recv",
//     "mid": "msg_123",
//     "from": "AA:BB:CC:DD:EE:FF",
//     "from_name": "妈妈",
//     "fmt": "audio",
//     "content": "https://cdn.com/msg.wav",
//     "dur": 8
// }
// 5. 标记已读（设备 → 云端）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "read",
//     "mids": ["msg_123", "msg_124"]
// }
// 6. 拉取未读（设备 → 云端）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "unread_req",
//     "page": 1,
//     "ps": 10
// }
// 7. 未读列表（云端 → 设备）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "unread",
//     "total": 3,
//     "list": [
//         {
//             "mid": "msg_123",
//             "ts": 1679036000,
//             "from": "AA:BB:CC:DD:EE:FF",
//             "from_name": "妈妈",
//             "fmt": "audio",
//             "content": "https://cdn.com/msg.wav",
//             "dur": 8
//         }
//     ]
// }
// 8. 发送失败（云端 → 设备）
// Topic: aibuddy/{device_id}/msg

// {
//     "type": "send_fail",
//     "mid": "msg_123",
//     "to": "AA:BB:CC:DD:EE:FF",
//     "code": 1003
// }

// MsgType 消息类型
type MsgType string

const (
	// MsgTypeFriendsReq 请求好友列表
	MsgTypeFriendsReq MsgType = "friends_req"
	// MsgTypeFriends 好友列表
	MsgTypeFriends MsgType = "friends"
	// MsgTypeSend 发送留言
	MsgTypeSend MsgType = "send"
	// MsgTypeRecv 接收留言
	MsgTypeRecv MsgType = "recv"
	// MsgTypeRead 标记已读
	MsgTypeRead MsgType = "read"
	// MsgTypeUnreadReq 拉取未读
	MsgTypeUnreadReq MsgType = "unread_req"
	// MsgTypeUnread 未读列表
	MsgTypeUnread MsgType = "unread"
	// MsgTypeSendFail 发送失败
	MsgTypeSendFail MsgType = "send_fail"
)

// IsValid 验证消息类型是否有效
func (m MsgType) IsValid() bool {
	return m == MsgTypeFriendsReq || m == MsgTypeFriends || m == MsgTypeSend || m == MsgTypeRecv || m == MsgTypeRead || m == MsgTypeUnreadReq || m == MsgTypeUnread || m == MsgTypeSendFail
}

// String 转换为字符串
func (m MsgType) String() string {
	return string(m)
}

// Msg 消息
type Msg struct {
	Type MsgType `json:"type"`
}

// FriendsReqMsg 请求好友列表消息
type FriendsReqMsg struct {
	Type MsgType `json:"type"`
	Page int     `json:"page,omitempty"`
	Ps   int     `json:"ps,omitempty"` // page size
}

// FriendsMsg 好友列表消息
type FriendsMsg struct {
	Type  MsgType          `json:"type"`
	Total int              `json:"total,omitempty"`
	List  []FriendsMsgItem `json:"list,omitempty"`
}

// FriendsMsgItem 好友消息项
type FriendsMsgItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Rel    string `json:"rel"`
}

// SendMsg 发送留言消息
type SendMsg struct {
	Type    MsgType `json:"type"`
	Mid     string  `json:"mid,omitempty"`
	To      string  `json:"to,omitempty"`
	Fmt     string  `json:"fmt,omitempty"`
	Content string  `json:"content,omitempty"`
	Dur     int     `json:"dur,omitempty"`
}

// RecvMsg 接收留言消息
type RecvMsg struct {
	Type     MsgType `json:"type"`
	Mid      string  `json:"mid,omitempty"`
	From     string  `json:"from,omitempty"`
	FromName string  `json:"from_name,omitempty"`
	Fmt      string  `json:"fmt,omitempty"`
	Content  string  `json:"content,omitempty"`
	Dur      int     `json:"dur,omitempty"`
}

// MarkReadMsg 标记已读消息
type MarkReadMsg struct {
	Type MsgType  `json:"type"`
	Mids []string `json:"mids,omitempty"`
}

// UnreadReqMsg 拉取未读消息
type UnreadReqMsg struct {
	Type MsgType `json:"type"`
	Page int     `json:"page,omitempty"`
	Ps   int     `json:"ps,omitempty"` // page size
}

// UnreadMsg 未读列表消息
type UnreadMsg struct {
	Type  MsgType         `json:"type"`
	Total int             `json:"total,omitempty"`
	List  []UnreadMsgItem `json:"list,omitempty"`
}

// UnreadMsgItem 未读消息项
type UnreadMsgItem struct {
	Mid      string `json:"mid,omitempty"`
	Ts       int    `json:"ts,omitempty"`
	From     string `json:"from,omitempty"`
	FromName string `json:"from_name,omitempty"`
	Fmt      string `json:"fmt,omitempty"`
	Content  string `json:"content,omitempty"`
	Dur      int    `json:"dur,omitempty"`
}

// SendFailMsg 发送失败消息
type SendFailMsg struct {
	Type MsgType `json:"type"`
	Mid  string  `json:"mid,omitempty"`
	To   string  `json:"to,omitempty"`
	Code int     `json:"code,omitempty"`
}
