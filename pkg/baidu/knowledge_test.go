package baidu

import (
	"testing"

	"aibuddy/pkg/config"
)

// TestKnowledge_CreateKnowledgeBase 测试创建知识库
func TestKnowledge_CreateKnowledgeBase(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	resp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "测试知识库",
		Description: "用于接口测试的知识库",
		Config: &KnowledgeConfig{
			Index: &IndexConfig{
				ChunkSize:    500,
				ChunkOverlap: 50,
			},
		},
		Documents: []Document{
			{
				Name:        "测试文档1",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "这是一个测试文档的内容，用于验证知识库创建接口。",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase failed: %v", err)
	}

	if resp.KbID == "" {
		t.Error("KbID should not be empty")
	}

	t.Logf("CreateKnowledgeBase succeeded: kbId=%s, name=%s, status=%s, docIds=%v",
		resp.KbID, resp.Name, resp.Status, resp.DocIDs)

	// 清理：删除知识库
	err = kb.DeleteKnowledgeBase(resp.KbID)
	if err != nil {
		t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
	}
}

// TestKnowledge_CreateKnowledgeBase_Empty 测试创建空知识库（不含文档）
func TestKnowledge_CreateKnowledgeBase_Empty(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	resp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "空测试知识库",
		Description: "不含文档的测试知识库",
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase empty failed: %v", err)
	}

	if resp.KbID == "" {
		t.Error("KbID should not be empty")
	}

	t.Logf("CreateKnowledgeBase empty succeeded: kbId=%s, status=%s", resp.KbID, resp.Status)

	// 清理
	err = kb.DeleteKnowledgeBase(resp.KbID)
	if err != nil {
		t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
	}
}

// TestKnowledge_ListKnowledgeBases 测试查询知识库列表
func TestKnowledge_ListKnowledgeBases(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 先创建一个知识库确保列表不为空
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "列表测试知识库",
		Description: "用于列表查询测试",
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for list test failed: %v", err)
	}
	t.Logf("Created knowledge base: kbId=%s", createResp.KbID)

	defer func() {
		err := kb.DeleteKnowledgeBase(createResp.KbID)
		if err != nil {
			t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
		}
	}()

	// 查询知识库列表
	resp, err := kb.ListKnowledgeBases(&ListKnowledgeBasesRequest{
		PageNum:  1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListKnowledgeBases failed: %v", err)
	}

	t.Logf("ListKnowledgeBases succeeded: count=%d", len(resp.List))
	for i, item := range resp.List {
		t.Logf("  [%d] kbId=%s, name=%s, status=%s", i, item.KbID, item.Name, item.Status)
	}
}

// TestKnowledge_GetKnowledgeBase 测试查询知识库详情
func TestKnowledge_GetKnowledgeBase(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 先创建知识库
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "详情测试知识库",
		Description: "用于详情查询测试",
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for get test failed: %v", err)
	}

	defer func() {
		err := kb.DeleteKnowledgeBase(createResp.KbID)
		if err != nil {
			t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
		}
	}()

	// 查询知识库详情
	resp, err := kb.GetKnowledgeBase(createResp.KbID)
	if err != nil {
		t.Fatalf("GetKnowledgeBase failed: %v", err)
	}

	t.Logf("GetKnowledgeBase succeeded: kbId=%s, name=%s, description=%s, status=%s",
		resp.KbID, resp.Name, resp.Description, resp.Status)

	if resp.KbID != createResp.KbID {
		t.Errorf("KbID mismatch: expected=%s, got=%s", createResp.KbID, resp.KbID)
	}
}

// TestKnowledge_DeleteKnowledgeBase 测试删除知识库
func TestKnowledge_DeleteKnowledgeBase(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 先创建知识库
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "删除测试知识库",
		Description: "用于删除测试",
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for delete test failed: %v", err)
	}

	t.Logf("Created knowledge base for deletion: kbId=%s", createResp.KbID)

	// 删除知识库
	err = kb.DeleteKnowledgeBase(createResp.KbID)
	if err != nil {
		t.Fatalf("DeleteKnowledgeBase failed: %v", err)
	}

	t.Log("DeleteKnowledgeBase succeeded")

	// 验证删除后查询应返回错误
	_, err = kb.GetKnowledgeBase(createResp.KbID)
	if err != nil {
		t.Logf("GetKnowledgeBase after deletion returned error as expected: %v", err)
	} else {
		t.Log("Warning: GetKnowledgeBase after deletion succeeded (unexpected)")
	}
}

// TestKnowledge_UploadDocument 测试上传文档到知识库
func TestKnowledge_UploadDocument(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 先创建空知识库
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "文档上传测试知识库",
		Description: "用于文档上传测试",
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for upload test failed: %v", err)
	}

	defer func() {
		err := kb.DeleteKnowledgeBase(createResp.KbID)
		if err != nil {
			t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
		}
	}()

	// 上传文档
	uploadResp, err := kb.UploadDocument(&UploadDocumentRequest{
		KbID: createResp.KbID,
		Documents: []Document{
			{
				Name:        "上传测试文档",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "这是通过上传接口添加的测试文档内容。",
			},
		},
	})
	if err != nil {
		t.Fatalf("UploadDocument failed: %v", err)
	}

	t.Logf("UploadDocument succeeded: kbId=%s, addedCount=%d, docIds=%v, status=%s",
		uploadResp.KbID, uploadResp.AddedCount, uploadResp.DocIDs, uploadResp.Status)
}

// TestKnowledge_ListDocuments 测试查询知识库文档列表
func TestKnowledge_ListDocuments(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 创建知识库并上传文档
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "文档列表测试知识库",
		Description: "用于文档列表查询测试",
		Documents: []Document{
			{
				Name:        "文档列表测试文档1",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "测试文档1的内容。",
			},
			{
				Name:        "文档列表测试文档2",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "测试文档2的内容。",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for list documents test failed: %v", err)
	}

	defer func() {
		err := kb.DeleteKnowledgeBase(createResp.KbID)
		if err != nil {
			t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
		}
	}()

	// 查询文档列表
	resp, err := kb.ListDocuments(createResp.KbID)
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	t.Logf("ListDocuments succeeded: kbId=%s, count=%d", resp.KbID, len(resp.List))
	for i, doc := range resp.List {
		t.Logf("  [%d] docId=%s, name=%s, status=%s", i, doc.DocID, doc.Name, doc.Status)
	}
}

// TestKnowledge_GetDocument 测试查询知识库文档详情
func TestKnowledge_GetDocument(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 创建知识库并上传文档
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "文档详情测试知识库",
		Description: "用于文档详情查询测试",
		Documents: []Document{
			{
				Name:        "文档详情测试文档",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "用于详情查询的测试文档内容。",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for get document test failed: %v", err)
	}

	defer func() {
		err := kb.DeleteKnowledgeBase(createResp.KbID)
		if err != nil {
			t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
		}
	}()

	if len(createResp.DocIDs) == 0 {
		t.Fatal("No documents were created with knowledge base")
	}

	docID := createResp.DocIDs[0]

	// 查询文档详情
	resp, err := kb.GetDocument(createResp.KbID, docID)
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}

	t.Logf("GetDocument succeeded: docId=%s, name=%s, contentType=%s, sourceType=%s, status=%s",
		resp.DocID, resp.Name, resp.ContentType, resp.SourceType, resp.Status)
}

// TestKnowledge_DeleteDocument 测试删除文档
func TestKnowledge_DeleteDocument(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 创建知识库并上传文档
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "文档删除测试知识库",
		Description: "用于文档删除测试",
		Documents: []Document{
			{
				Name:        "待删除测试文档",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "用于删除测试的文档内容。",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for delete document test failed: %v", err)
	}

	defer func() {
		err := kb.DeleteKnowledgeBase(createResp.KbID)
		if err != nil {
			t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
		}
	}()

	if len(createResp.DocIDs) == 0 {
		t.Fatal("No documents were created with knowledge base")
	}

	docID := createResp.DocIDs[0]

	// 删除文档
	delResp, err := kb.DeleteDocument(createResp.KbID, docID)
	if err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}

	t.Logf("DeleteDocument succeeded: kbId=%s, deletedCount=%d, docIds=%v",
		delResp.KbID, delResp.DeletedCount, delResp.DocIDs)
}

// TestKnowledge_BatchDeleteDocuments 测试批量删除文档
func TestKnowledge_BatchDeleteDocuments(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 创建知识库并上传多个文档
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "批量删除测试知识库",
		Description: "用于批量删除文档测试",
		Documents: []Document{
			{
				Name:        "批量删除文档1",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "批量删除测试文档1的内容。",
			},
			{
				Name:        "批量删除文档2",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "批量删除测试文档2的内容。",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateKnowledgeBase for batch delete test failed: %v", err)
	}

	defer func() {
		err := kb.DeleteKnowledgeBase(createResp.KbID)
		if err != nil {
			t.Logf("Warning: DeleteKnowledgeBase cleanup failed: %v", err)
		}
	}()

	if len(createResp.DocIDs) < 2 {
		t.Fatalf("Expected at least 2 documents, got %d", len(createResp.DocIDs))
	}

	// 批量删除文档
	delResp, err := kb.BatchDeleteDocuments(&BatchDeleteDocumentsRequest{
		KbID:   createResp.KbID,
		DocIDs: createResp.DocIDs,
	})
	if err != nil {
		t.Fatalf("BatchDeleteDocuments failed: %v", err)
	}

	t.Logf("BatchDeleteDocuments succeeded: kbId=%s, deletedCount=%d, docIds=%v",
		delResp.KbID, delResp.DeletedCount, delResp.DocIDs)
}

// TestKnowledge_FullWorkflow 测试完整工作流：创建→查询→上传文档→查询文档→删除文档→删除知识库
func TestKnowledge_FullWorkflow(t *testing.T) {
	config.Setup("../../config")

	kb := NewKnowledge()

	// 1. 创建空知识库
	createResp, err := kb.CreateKnowledgeBase(&CreateKnowledgeBaseRequest{
		Name:        "完整工作流测试知识库",
		Description: "用于完整工作流测试",
	})
	if err != nil {
		t.Fatalf("Step 1 CreateKnowledgeBase failed: %v", err)
	}
	t.Logf("Step 1: CreateKnowledgeBase succeeded, kbId=%s", createResp.KbID)

	// 2. 查询知识库详情
	getResp, err := kb.GetKnowledgeBase(createResp.KbID)
	if err != nil {
		t.Fatalf("Step 2 GetKnowledgeBase failed: %v", err)
	}
	t.Logf("Step 2: GetKnowledgeBase succeeded, name=%s, status=%s", getResp.Name, getResp.Status)

	// 3. 上传文档
	uploadResp, err := kb.UploadDocument(&UploadDocumentRequest{
		KbID: createResp.KbID,
		Documents: []Document{
			{
				Name:        "工作流测试文档",
				ContentType: DocumentContentTypeText,
				SourceType:  DocumentSourceTypeString,
				Content:     "完整工作流测试文档内容。",
			},
		},
	})
	if err != nil {
		t.Fatalf("Step 3 UploadDocument failed: %v", err)
	}
	t.Logf("Step 3: UploadDocument succeeded, addedCount=%d, docIds=%v", uploadResp.AddedCount, uploadResp.DocIDs)

	// 4. 查询文档列表
	listResp, err := kb.ListDocuments(createResp.KbID)
	if err != nil {
		t.Fatalf("Step 4 ListDocuments failed: %v", err)
	}
	t.Logf("Step 4: ListDocuments succeeded, count=%d", len(listResp.List))

	// 5. 查询文档详情
	if len(uploadResp.DocIDs) > 0 {
		docResp, err := kb.GetDocument(createResp.KbID, uploadResp.DocIDs[0])
		if err != nil {
			t.Fatalf("Step 5 GetDocument failed: %v", err)
		}
		t.Logf("Step 5: GetDocument succeeded, name=%s, status=%s", docResp.Name, docResp.Status)

		// 6. 删除文档
		delResp, err := kb.DeleteDocument(createResp.KbID, uploadResp.DocIDs[0])
		if err != nil {
			t.Fatalf("Step 6 DeleteDocument failed: %v", err)
		}
		t.Logf("Step 6: DeleteDocument succeeded, deletedCount=%d", delResp.DeletedCount)
	}

	// 7. 查询知识库列表
	kbListResp, err := kb.ListKnowledgeBases(&ListKnowledgeBasesRequest{
		PageNum:  1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Step 7 ListKnowledgeBases failed: %v", err)
	}
	t.Logf("Step 7: ListKnowledgeBases succeeded, count=%d", len(kbListResp.List))

	// 8. 删除知识库
	err = kb.DeleteKnowledgeBase(createResp.KbID)
	if err != nil {
		t.Fatalf("Step 8 DeleteKnowledgeBase failed: %v", err)
	}
	t.Log("Step 8: DeleteKnowledgeBase succeeded")

	t.Log("Full workflow test passed!")
}
