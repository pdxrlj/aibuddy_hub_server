// Package baidu 百度云API客户端知识库管理
package baidu

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Knowledge 知识库API
type Knowledge struct {
	client *Client
}

// NewKnowledge 创建知识库客户端
func NewKnowledge() *Knowledge {
	return &Knowledge{client: NewClient()}
}

// NewKnowledgeWithAKSK 使用指定AK/SK创建知识库客户端
func NewKnowledgeWithAKSK(ak, sk string) *Knowledge {
	return &Knowledge{client: NewClientWithAKSK(ak, sk)}
}

// KnowledgeStatus 知识库状态
type KnowledgeStatus string

const (
	// KnowledgeStatusProcessing 处理中
	KnowledgeStatusProcessing KnowledgeStatus = "PROCESSING"
	// KnowledgeStatusReady 可用
	KnowledgeStatusReady KnowledgeStatus = "READY"
	// KnowledgeStatusError 异常
	KnowledgeStatusError KnowledgeStatus = "ERROR"
)

// DocumentSourceType 文档来源类型
type DocumentSourceType string

const (
	// DocumentSourceTypeString 字符串内容
	DocumentSourceTypeString DocumentSourceType = "string"
	// DocumentSourceTypeFileURL 文件URL
	DocumentSourceTypeFileURL DocumentSourceType = "fileUrl"
	// DocumentSourceTypeFileData 文件二进制数据
	DocumentSourceTypeFileData DocumentSourceType = "fileData"
)

// DocumentContentType 文档内容类型
type DocumentContentType string

const (
	// DocumentContentTypeText 纯文本
	DocumentContentTypeText DocumentContentType = "text"
)

// IndexConfig 索引配置
type IndexConfig struct {
	ChunkSize     int `json:"chunkSize,omitempty"`     // 切片大小，默认值500
	ChunkOverlap  int `json:"chunkOverlap,omitempty"`  // 切片重叠大小，默认值50
}

// KnowledgeConfig 知识库配置
type KnowledgeConfig struct {
	Index *IndexConfig `json:"index,omitempty"` // 索引配置
}

// Document 文档
type Document struct {
	Name        string              `json:"name"`                  // 文档名称，长度不能超过128
	ContentType DocumentContentType `json:"contentType,omitempty"` // 文档内容类型，暂时只支持text
	SourceType  DocumentSourceType  `json:"sourceType"`            // 文档来源类型：string、fileUrl、fileData
	Content     string              `json:"content,omitempty"`     // 字符串内容（sourceType=string时有意义且必须存在）
	FileURL     string              `json:"fileUrl,omitempty"`     // 文件URL（sourceType=fileUrl时有意义且必须存在）
	FileData    []byte              `json:"fileData,omitempty"`    // 文件二进制数据（sourceType=fileData时有意义且必须存在）
}

// --- 创建知识库 ---

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name        string           `json:"name"`                  // 知识库名称，长度不能超过60
	Description string           `json:"description"`           // 知识库描述，长度不能超过256
	Config      *KnowledgeConfig `json:"config,omitempty"`      // 知识库配置项
	Documents   []Document       `json:"documents,omitempty"`   // 文档列表（可选，创建同时导入文档）
}

// CreateKnowledgeBaseResponse 创建知识库响应
type CreateKnowledgeBaseResponse struct {
	KbID        string   `json:"kbId"`        // 知识库ID
	Name        string   `json:"name"`        // 知识库名称
	Description string   `json:"description"` // 知识库描述
	Status      string   `json:"status"`      // 知识库状态
	DocIDs      []string `json:"docIds"`      // 文档ID列表
	Config      *KnowledgeConfig `json:"config,omitempty"` // 知识库配置
}

// CreateKnowledgeBase 创建知识库
func (k *Knowledge) CreateKnowledgeBase(req *CreateKnowledgeBaseRequest) (*CreateKnowledgeBaseResponse, error) {
	path := "/api/v1/rag/knowledgeBase"

	var result CreateKnowledgeBaseResponse
	if err := k.client.Request("POST", path, nil, req, &result); err != nil {
		return nil, fmt.Errorf("创建知识库失败: %w", err)
	}
	return &result, nil
}

// --- 查询知识库列表 ---

// KnowledgeBaseItem 知识库列表项
type KnowledgeBaseItem struct {
	KbID        string `json:"kbId"`        // 知识库ID
	Name        string `json:"name"`        // 知识库名称
	Description string `json:"description"` // 知识库描述
	Status      string `json:"status"`      // 知识库状态
}

// ListKnowledgeBasesRequest 查询知识库列表请求
type ListKnowledgeBasesRequest struct {
	PageNum  int `json:"pageNum,omitempty"`  // 页码，默认值为1
	PageSize int `json:"pageSize,omitempty"` // 每页数量，默认值为10
}

// ListKnowledgeBasesResponse 查询知识库列表响应
type ListKnowledgeBasesResponse struct {
	List []KnowledgeBaseItem `json:"list"` // 知识库列表
}

// ListKnowledgeBases 查询知识库列表
func (k *Knowledge) ListKnowledgeBases(req *ListKnowledgeBasesRequest) (*ListKnowledgeBasesResponse, error) {
	path := "/api/v1/rag/knowledgeBase/list"

	query := url.Values{}
	if req.PageNum > 0 {
		query.Set("pageNum", strconv.Itoa(req.PageNum))
	}
	if req.PageSize > 0 {
		query.Set("pageSize", strconv.Itoa(req.PageSize))
	}

	var result ListKnowledgeBasesResponse
	if err := k.client.Request("GET", path, query, nil, &result); err != nil {
		return nil, fmt.Errorf("查询知识库列表失败: %w", err)
	}
	return &result, nil
}

// --- 查询知识库详情 ---

// GetKnowledgeBaseResponse 查询知识库详情响应
type GetKnowledgeBaseResponse struct {
	KbID        string           `json:"kbId"`                  // 知识库ID
	Name        string           `json:"name"`                  // 知识库名称
	Description string           `json:"description"`           // 知识库描述
	Status      string           `json:"status"`                // 知识库状态
	Config      *KnowledgeConfig `json:"config,omitempty"`      // 知识库配置
}

// GetKnowledgeBase 查询知识库详情
func (k *Knowledge) GetKnowledgeBase(kbID string) (*GetKnowledgeBaseResponse, error) {
	path := fmt.Sprintf("/api/v1/rag/knowledgeBase/%s", kbID)

	var result GetKnowledgeBaseResponse
	if err := k.client.Request("GET", path, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("查询知识库详情失败: %w", err)
	}
	return &result, nil
}

// --- 删除知识库 ---

// DeleteKnowledgeBase 删除知识库
func (k *Knowledge) DeleteKnowledgeBase(kbID string) error {
	path := fmt.Sprintf("/api/v1/rag/knowledgeBase/%s", kbID)

	if err := k.client.Request("DELETE", path, nil, nil, nil); err != nil {
		return fmt.Errorf("删除知识库失败: %w", err)
	}
	return nil
}

// --- 上传文档到知识库 ---

// UploadDocumentRequest 上传文档到知识库请求
type UploadDocumentRequest struct {
	KbID      string     // 知识库ID（路径参数）
	Documents []Document `json:"documents"` // 文档列表
}

// UploadDocumentResponse 上传文档到知识库响应
type UploadDocumentResponse struct {
	KbID       string   `json:"kbId"`       // 知识库ID
	AddedCount int      `json:"addedCount"` // 新增文档数量
	DocIDs     []string `json:"docIds"`     // 文档ID列表
	Status     string   `json:"status"`     // 知识库状态
}

// UploadDocument 上传文档到知识库
func (k *Knowledge) UploadDocument(req *UploadDocumentRequest) (*UploadDocumentResponse, error) {
	path := fmt.Sprintf("/api/v1/rag/knowledgeBase/%s/document", req.KbID)

	var result UploadDocumentResponse
	if err := k.client.Request("POST", path, nil, map[string]any{
		"documents": req.Documents,
	}, &result); err != nil {
		return nil, fmt.Errorf("上传文档到知识库失败: %w", err)
	}
	return &result, nil
}

// --- 查询知识库文档列表 ---

// DocumentItem 文档列表项
type DocumentItem struct {
	DocID       string `json:"docId"`       // 文档ID
	Name        string `json:"name"`        // 文档名称
	ContentType string `json:"contentType"` // 文档内容类型
	SourceType  string `json:"sourceType"`  // 文档来源类型
	Status      string `json:"status"`      // 文档状态
}

// ListDocumentsResponse 查询知识库文档列表响应
type ListDocumentsResponse struct {
	KbID string         `json:"kbId"` // 知识库ID
	List []DocumentItem `json:"list"` // 文档列表
}

// ListDocuments 查询知识库文档列表
func (k *Knowledge) ListDocuments(kbID string) (*ListDocumentsResponse, error) {
	path := fmt.Sprintf("/api/v1/rag/knowledgeBase/%s/document/list", kbID)

	var result ListDocumentsResponse
	if err := k.client.Request("GET", path, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("查询知识库文档列表失败: %w", err)
	}
	return &result, nil
}

// --- 查询知识库文档详情 ---

// GetDocumentResponse 查询知识库文档详情响应
type GetDocumentResponse struct {
	DocID       string `json:"docId"`       // 文档ID
	Name        string `json:"name"`        // 文档名称
	ContentType string `json:"contentType"` // 文档类型
	SourceType  string `json:"sourceType"`  // 文档来源
	Status      string `json:"status"`      // 文档状态
}

// GetDocument 查询知识库文档详情
func (k *Knowledge) GetDocument(kbID, docID string) (*GetDocumentResponse, error) {
	path := fmt.Sprintf("/api/v1/rag/knowledgeBase/%s/document/%s", kbID, docID)

	var result GetDocumentResponse
	if err := k.client.Request("GET", path, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("查询知识库文档详情失败: %w", err)
	}
	return &result, nil
}

// --- 删除文档 ---

// DeleteDocumentResponse 删除文档响应
type DeleteDocumentResponse struct {
	KbID         string   `json:"kbId"`         // 知识库ID
	DeletedCount int      `json:"deletedCount"` // 删除成功的文档数目
	DocIDs       []string `json:"docIds"`       // 删除成功的文档ID列表
}

// DeleteDocument 删除文档
func (k *Knowledge) DeleteDocument(kbID, docID string) (*DeleteDocumentResponse, error) {
	path := fmt.Sprintf("/api/v1/rag/knowledgeBase/%s/document/%s", kbID, docID)

	var result DeleteDocumentResponse
	if err := k.client.Request("DELETE", path, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("删除文档失败: %w", err)
	}
	return &result, nil
}

// --- 批量删除文档 ---

// BatchDeleteDocumentsRequest 批量删除文档请求
type BatchDeleteDocumentsRequest struct {
	KbID   string   // 知识库ID（路径参数）
	DocIDs []string // 文档ID列表（查询参数）
}

// BatchDeleteDocuments 批量删除文档
func (k *Knowledge) BatchDeleteDocuments(req *BatchDeleteDocumentsRequest) (*DeleteDocumentResponse, error) {
	path := fmt.Sprintf("/api/v1/rag/knowledgeBase/%s/document/batch", req.KbID)

	query := url.Values{}
	query.Set("docIds", strings.Join(req.DocIDs, ","))

	var result DeleteDocumentResponse
	if err := k.client.Request("DELETE", path, query, nil, &result); err != nil {
		return nil, fmt.Errorf("批量删除文档失败: %w", err)
	}
	return &result, nil
}
