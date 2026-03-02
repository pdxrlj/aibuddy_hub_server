// Package management 提供设备管理接口定义
package management

import "encoding/json"

// 4. 状态上报（设备 → 云端）
// Topic: aibuddy/{device_id}/status

// {
//     "type": "stat",
//     "bat": 90,
//     "charging": true,
//     "csq": 30,
//     "storage_free": 11800000,
//     "mem_free": 262144,
//     "iccid": "898608171524D0000303" // SIM卡ICCID
// }

// 8. 状态查询（云端 → 设备）
// Topic: aibuddy/{device_id}/cmd/mgmt

// {
//     "type": "qry_stat"
// }
// 9. 挂失（云端 → 设备）
// Topic: aibuddy/{device_id}/cmd/mgmt

// {
//     "type": "lost",
//     "contact": "妈妈",
//     "phone": "18888888888"
// }
// 10. 解除挂失（云端 → 设备）
// Topic: aibuddy/{device_id}/cmd/mgmt

// {
//     "type": "recover"
// }
// 11. 绑定通知（云端 → 设备）
// Topic: aibuddy/{device_id}/cmd/mgmt

// {
//     "type": "bound",
//     "user": "张三",
//     "avatar": "https://cdn.com/avatar.jpg"
// }
// 12. 解绑通知（云端 → 设备）
// Topic: aibuddy/{device_id}/cmd/mgmt

// {
//     "type": "unbind"
// }

// MgmtType 设备管理类型
type MgmtType string

// MgmtTypeQryStat 查询设备状态
const (
	MgmtTypeQryStat MgmtType = "qry_stat"
	// MgmtTypeLost 挂失
	MgmtTypeLost MgmtType = "lost"
	// MgmtTypeRecover 解除挂失
	MgmtTypeRecover MgmtType = "recover"
	// MgmtTypeBound 绑定
	MgmtTypeBound MgmtType = "bound"
	// MgmtTypeUnbind 解绑
	MgmtTypeUnbind MgmtType = "unbind"
)

// IsValid 验证设备管理类型是否有效
func (m MgmtType) IsValid() bool {
	return m == MgmtTypeQryStat || m == MgmtTypeLost || m == MgmtTypeRecover || m == MgmtTypeBound || m == MgmtTypeUnbind
}

// String 转换为字符串
func (m MgmtType) String() string {
	return string(m)
}

// Mgmt 设备管理消息
type Mgmt struct {
	Type    MgmtType `json:"type"`
	Contact string   `json:"contact"`
	Phone   string   `json:"phone"`

	User   string `json:"user"`
	Avatar string `json:"avatar"`
	Sn     string `json:"sn"`
}

// Encode 编码设备管理消息
func (m *Mgmt) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode 解码设备管理消息
func (m *Mgmt) Decode(data []byte) error {
	return json.Unmarshal(data, m)
}
