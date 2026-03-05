package baidu

import (
	"testing"

	"aibuddy/pkg/config"
)

func TestRoleList(t *testing.T) {
	config.Setup("../../config")

	role := NewRole()
	resp, err := role.RoleList("")
	if err != nil {
		t.Fatalf("RoleList failed: %v", err)
	}

	t.Logf("Role count: %d", len(resp.LLM.Roles))
	for i, r := range resp.LLM.Roles {
		t.Logf("Role %d: ID=%d, Name=%s, Model=%s, DefaultUsage=%v, Description=%s",
			i+1, r.ID, r.Name, r.Model, r.DefaultUsage, r.Description)
	}
}

func TestGetSystemRoles(t *testing.T) {
	config.Setup("../../config")

	role := NewRole()
	resp, err := role.GetSystemRoles()
	if err != nil {
		t.Fatalf("GetSystemRoles failed: %v", err)
	}

	t.Logf("System role count: %d", len(resp.Data))
	for i, r := range resp.Data {
		t.Logf("System Role %d: ID=%d, Name=%s, Description=%s",
			i+1, r.ID, r.Name, r.Description)
	}
}

func TestBindFunctionTemplate(t *testing.T) {
	t.Skip("跳过实际执行，需要真实的 functionTemplateId")

	config.Setup("../../config")

	role := NewRole()
	err := role.BindFunctionTemplate("", &BindFunctionTemplateRequest{
		Type:               "SYSTEM",
		FunctionTemplateID: 1000000, // 替换为真实的模板ID
	})
	if err != nil {
		t.Fatalf("BindFunctionTemplate failed: %v", err)
	}

	t.Log("BindFunctionTemplate success")
}

func TestManageRoles_Add(t *testing.T) {
	config.Setup("../../config")

	role := NewRole()

	// 先清理可能存在的重复角色
	roleList, _ := role.RoleList("")
	for _, r := range roleList.LLM.Roles {
		if r.Name == "北极熊" {
			role.ManageRoles("", &ManageRolesRequest{
				Type: "SYSTEM",
				Role: &RoleOperation{RemoveItems: []int64{r.ID}},
			})
		}
	}

	// 添加系统角色(北极熊 ID=5)
	err := role.ManageRoles("", &ManageRolesRequest{
		Type: "SYSTEM",
		Role: &RoleOperation{
			AddItems: []RoleAddItem{
				{SystemRoleID: 5, DefaultUsage: false},
			},
		},
	})
	if err != nil {
		t.Fatalf("ManageRoles Add failed: %v", err)
	}

	t.Log("ManageRoles Add success")
}

func TestManageRoles_Update(t *testing.T) {
	config.Setup("../../config")

	role := NewRole()

	// 先添加角色
	err := role.ManageRoles("", &ManageRolesRequest{
		Type: "SYSTEM",
		Role: &RoleOperation{
			AddItems: []RoleAddItem{
				{SystemRoleID: 7, DefaultUsage: false}, // 猪八戒
			},
		},
	})
	if err != nil {
		t.Fatalf("Prepare Add failed: %v", err)
	}

	// 获取角色ID
	roleList, err := role.RoleList("")
	if err != nil {
		t.Fatalf("RoleList failed: %v", err)
	}

	var roleID int64
	for _, r := range roleList.LLM.Roles {
		if r.Name == "猪八戒" {
			roleID = r.ID
			break
		}
	}
	if roleID == 0 {
		t.Fatal("找不到刚添加的角色")
	}

	// 更新角色
	err = role.ManageRoles("", &ManageRolesRequest{
		Type: "SYSTEM",
		Role: &RoleOperation{
			UpdateItems: []RoleUpdateItem{{ID: roleID, DefaultUsage: false}},
		},
	})
	if err != nil {
		t.Fatalf("ManageRoles Update failed: %v", err)
	}

	// 清理：删除角色
	role.ManageRoles("", &ManageRolesRequest{
		Type: "SYSTEM",
		Role: &RoleOperation{RemoveItems: []int64{roleID}},
	})

	t.Log("ManageRoles Update success")
}

func TestManageRoles_Remove(t *testing.T) {
	config.Setup("../../config")

	role := NewRole()

	// 先添加角色
	err := role.ManageRoles("", &ManageRolesRequest{
		Type: "SYSTEM",
		Role: &RoleOperation{
			AddItems: []RoleAddItem{
				{SystemRoleID: 8, DefaultUsage: false}, // 哪吒
			},
		},
	})
	if err != nil {
		t.Fatalf("Prepare Add failed: %v", err)
	}

	// 获取角色ID
	roleList, err := role.RoleList("")
	if err != nil {
		t.Fatalf("RoleList failed: %v", err)
	}

	var roleID int64
	for _, r := range roleList.LLM.Roles {
		if r.Name == "哪吒" {
			roleID = r.ID
			break
		}
	}
	if roleID == 0 {
		t.Fatal("找不到刚添加的角色")
	}

	// 删除角色
	err = role.ManageRoles("", &ManageRolesRequest{
		Type: "SYSTEM",
		Role: &RoleOperation{RemoveItems: []int64{roleID}},
	})
	if err != nil {
		t.Fatalf("ManageRoles Remove failed: %v", err)
	}

	t.Log("ManageRoles Remove success")
}

func TestManageRoles_Complex(t *testing.T) {
	config.Setup("../../config")

	role := NewRole()

	// 复杂操作：同时增删改
	err := role.ManageRoles("", &ManageRolesRequest{
		Type: "SYSTEM",
		Role: &RoleOperation{
			AddItems: []RoleAddItem{
				{SystemRoleID: 877, DefaultUsage: false}, // 中英互译
			},
		},
	})
	if err != nil {
		t.Fatalf("ManageRoles Complex Add failed: %v", err)
	}

	// 获取角色ID并清理
	roleList, _ := role.RoleList("")
	for _, r := range roleList.LLM.Roles {
		if r.Name == "中英互译" {
			role.ManageRoles("", &ManageRolesRequest{
				Type: "SYSTEM",
				Role: &RoleOperation{RemoveItems: []int64{r.ID}},
			})
			break
		}
	}

	t.Log("ManageRoles Complex success")
}
