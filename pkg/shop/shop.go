// Package shop 微信商品
package shop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// ProductURL 商品连接
	ProductURL = "https://api.weixin.qq.com/channels/ec/"
)

// MiniShop minishop
type MiniShop struct{}

// NewMiniShop 实例化minishop
func NewMiniShop() *MiniShop {
	return &MiniShop{}
}

// GetProductResp 商品列表响应数据
type GetProductResp struct {
	ErrCode int                     `json:"errcode"`
	ErrMsg  string                  `json:"errmsg"`
	Product *ProductProductResponse `json:"product"`
}

// GoodsListResponse 商品数据
type GoodsListResponse struct {
	ErrCode    int     `json:"errcode"`
	ErrMsg     string  `json:"errmsg"`
	ProductIDs []int64 `json:"product_ids"`
	NextKey    string  `json:"next_key"`
	TotalNum   int     `json:"total_num"`
}

// ProductProductResponse 商品详情
type ProductProductResponse struct {
	ProductID    string              `json:"product_id"`
	OutProductID string              `json:"out_product_id"`
	Title        string              `json:"title"`
	SubTitle     string              `json:"sub_title"`
	HeadImgs     []string            `json:"head_imgs"`
	DescInfo     *GetProductDescInfo `json:"desc_info"`
	Cats         []GetProductCat     `json:"cats"`

	CatsV2          []GetProductCatV2      `json:"cats_v2"`
	Attrs           []GetProductAttr       `json:"attrs"`
	ExpressInfo     *GetProductExpressInfo `json:"express_info"`
	Status          int                    `json:"status"`
	EditStatus      int                    `json:"edit_status"`
	Skus            []GetProductSku        `json:"skus"`
	MinPrice        int                    `json:"min_price"`
	SpuCode         string                 `json:"spu_code"`
	ProductQuaInfos []GetProductQuaInfo    `json:"product_qua_infos"`
}

// GetProductDescInfo 商品描述信息
type GetProductDescInfo struct {
	Imgs []string `json:"imgs"`
}

// GetProductCat 商品分类id
type GetProductCat struct {
	CatID string `json:"cat_id"`
}

// GetProductCatV2  新类目树
type GetProductCatV2 struct {
	CatID string `json:"cat_id"`
}

// GetProductAttr 属性键key
type GetProductAttr struct {
	AttrKey   string `json:"attr_key"`
	AttrValue string `json:"attr_value"`
}

// GetProductExpressInfo 运费信息
type GetProductExpressInfo struct {
	TemplateID string `json:"template_id"`
}

// GetProductSku sku信息
type GetProductSku struct {
	SkuID     string              `json:"sku_id"`
	OutSkuID  string              `json:"out_sku_id"`
	ThumbImg  string              `json:"thumb_img"`
	SalePrice int                 `json:"sale_price"`
	StockNum  int                 `json:"stock_num"`
	SkuCode   string              `json:"sku_code"`
	SkuAttrs  []GetProductSkuAttr `json:"sku_attrs"`
}

// GetProductSkuAttr sku属性
type GetProductSkuAttr struct {
	AttrKey   string `json:"attr_key"`
	AttrValue string `json:"attr_value"`
}

// GetProductQuaInfo 商品资质列表
type GetProductQuaInfo struct {
	QuaID  string   `json:"qua_id"`
	QuaURL []string `json:"qua_url"`
}

// GetGoodsList 获取商品列表
func (m *MiniShop) GetGoodsList(accessToken string, pageSize int, nextKey string) (GoodsListResponse, error) {
	var result GoodsListResponse
	uri := fmt.Sprintf("%sproduct/list/get?access_token=%s", ProductURL, accessToken)

	payload := map[string]any{
		"status":    5,
		"page_size": pageSize,
		"next_key":  nextKey,
	}
	data, _ := json.Marshal(payload)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	response, err := client.Post(uri, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return result, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	if err := response.Body.Close(); err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, err
}

// GetGoodsInfo 获取商品详情
func (m *MiniShop) GetGoodsInfo(accessToken string, id int64) (*ProductProductResponse, error) {
	var result *GetProductResp
	uri := fmt.Sprintf("%sproduct/get?access_token=%s", ProductURL, accessToken)
	payload := map[string]any{
		"product_id": id,
	}
	data, _ := json.Marshal(payload)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	response, err := client.Post(uri, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err := response.Body.Close(); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Product, nil
}
