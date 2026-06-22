package main

type apiEntry struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	SampleBody        string `json:"sampleBody"`
	Unavailable       bool   `json:"unavailable,omitempty"` // Playground 中无法正常调试使用
	UnavailableNote   string `json:"unavailableNote,omitempty"`
}

// categoryOrder 左侧模块显示顺序
var categoryOrder = []string{
	"店铺",
	"商品管理",
	"商品分类与属性",
	"商品尺码与认证",
	"商品价格",
	"备货单",
	"发货台",
	"发货单",
	"全托管物流",
	"半托管-订单",
	"半托管-物流",
	"半托管-包裹",
	"库存",
	"JIT",
	"广告",
	"图片与工具",
	"爆款邀约",
}

var apiCatalog = []apiEntry{
	// ── 店铺 ──
	{ID: "mall-type", Category: "店铺", Name: "店铺类型", Type: "bg.mall.info.get", Description: "查询店铺类型（全托 CN 网关）", SampleBody: `{}`},
	{ID: "mall-permission", Category: "店铺", Name: "Token 权限", Type: "bg.open.accesstoken.info.get.global", Description: "查询 Access Token 权限（Partner 网关，需 env=prod）", SampleBody: `{}`},
	{ID: "mall-token-create", Category: "店铺", Name: "创建 Token", Type: "bg.open.accesstoken.create", Description: "通过 OAuth 回调 code 换取 Access Token", SampleBody: `{"accessToken": "", "code": ""}`},
	{ID: "mall-address-get", Category: "店铺", Name: "发货地址列表", Type: "bg.mall.address.get", Description: "查询店铺发货地址", SampleBody: `{}`},
	{ID: "mall-address-add", Category: "店铺", Name: "新增发货地址", Type: "bg.mall.address.add", Description: "新增店铺发货地址（含四级街道 townCode）", SampleBody: mallAddressAddSample},

	// ── 商品管理 ──
	{ID: "goods-list", Category: "商品管理", Name: "商品列表", Type: "temu.goods.list.get", Description: "分页查询商品列表（Partner 网关）", SampleBody: `{"pageNo": 1, "pageSize": 10}`},
	{ID: "goods-detail", Category: "商品管理", Name: "商品详情", Type: "temu.goods.detail.get", Description: "根据 productId 查询商品详情", SampleBody: `{"productId": 141911679}`},
	{ID: "goods-add", Category: "商品管理", Name: "创建商品", Type: "temu.goods.add", Description: "创建新商品（全托样例）", SampleBody: goodsAddFullSample},
	{ID: "goods-update", Category: "商品管理", Name: "更新商品", Type: "temu.goods.update", Description: "更新商品信息", SampleBody: `{}`},
	{ID: "goods-migrate", Category: "商品管理", Name: "迁移商品", Type: "temu.goods.migrate", Description: "商品迁移", SampleBody: `{}`},
	{ID: "goods-edit-sensitive", Category: "商品管理", Name: "编辑敏感属性", Type: "temu.goods.edit.sensitive.attr", Description: "编辑商品敏感属性", SampleBody: `{}`},
	{ID: "goods-edit-property", Category: "商品管理", Name: "编辑商品属性", Type: "temu.goods.edit.property", Description: "编辑商品属性", SampleBody: `{}`},
	{ID: "goods-image-upload", Category: "商品管理", Name: "上传商品图片", Type: "bg.goods.image.upload.global", Description: "上传商品图片", SampleBody: `{}`},
	{ID: "goods-sales", Category: "商品管理", Name: "商品销量", Type: "bg.goods.sales.get", Description: "查询商品销量（CN 网关）", SampleBody: `{"productSkcIds": [7469668867]}`},
	{ID: "goods-life-search", Category: "商品管理", Name: "商品生命周期搜索", Type: "bg.glo.product.search", Description: "搜索商品生命周期", SampleBody: `{"pageNo": 1, "pageSize": 10}`},
	{ID: "goods-topselling", Category: "商品管理", Name: "爆款售罄", Type: "temu.goods.topselling.soldout.get", Description: "查询爆款售罄商品", SampleBody: `{}`},
	{ID: "goods-label", Category: "商品管理", Name: "商品标签", Type: "temu.goods.labelv2.get", Description: "查询商品标签/条码", SampleBody: `{"productSkcId": 7469668867}`},
	{ID: "goods-custom-label", Category: "商品管理", Name: "定制品标签", Type: "temu.goods.custom.label.get", Description: "查询定制品标签", SampleBody: `{}`},
	{ID: "goods-boxmark", Category: "商品管理", Name: "箱唛信息", Type: "bg.logistics.boxmarkinfo.get", Description: "查询箱唛信息", SampleBody: `{}`},

	// ── 商品分类与属性 ──
	{ID: "goods-cats", Category: "商品分类与属性", Name: "类目列表", Type: "bg.goods.cats.get", Description: "查询商品类目", SampleBody: `{"parentCatId": 0}`},
	{ID: "goods-attrs", Category: "商品分类与属性", Name: "类目属性", Type: "bg.goods.attrs.get", Description: "查询类目属性", SampleBody: `{"catId": 9711}`},
	{ID: "goods-brand", Category: "商品分类与属性", Name: "品牌列表", Type: "bg.glo.goods.brand.get", Description: "查询可绑定品牌", SampleBody: `{}`},
	{ID: "goods-parentspec", Category: "商品分类与属性", Name: "父规格", Type: "bg.glo.goods.parentspec.get", Description: "查询父规格", SampleBody: `{}`},
	{ID: "goods-spec-create", Category: "商品分类与属性", Name: "创建规格", Type: "bg.glo.goods.spec.create", Description: "创建商品规格", SampleBody: `{}`},

	// ── 商品尺码与认证 ──
	{ID: "goods-sizecharts", Category: "商品尺码与认证", Name: "尺码表", Type: "bg.goods.sizecharts.get", Description: "查询尺码表", SampleBody: `{}`},
	{ID: "goods-sizecharts-class", Category: "商品尺码与认证", Name: "尺码表分类", Type: "bg.goods.sizecharts.class.get", Description: "查询尺码表分类", SampleBody: `{}`},
	{ID: "goods-sizecharts-settings", Category: "商品尺码与认证", Name: "尺码表设置", Type: "bg.goods.sizecharts.settings.get", Description: "查询尺码表设置", SampleBody: `{}`},
	{ID: "goods-sizecharts-template", Category: "商品尺码与认证", Name: "创建尺码模板", Type: "bg.goods.sizecharts.template.create", Description: "创建尺码表模板", SampleBody: `{}`},
	{ID: "goods-cert-query", Category: "商品尺码与认证", Name: "商品认证查询", Type: "bg.arbok.open.product.cert.query", Description: "查询商品认证信息", SampleBody: `{}`},
	{ID: "goods-cert-need-upload", Category: "商品尺码与认证", Name: "待上传认证项", Type: "bg.arbok.open.cert.queryNeedUploadItems", Description: "查询需上传的认证项", SampleBody: `{}`},
	{ID: "goods-cert-upload-file", Category: "商品尺码与认证", Name: "上传认证文件", Type: "bg.arbok.open.upload.uploadFile", Description: "上传认证文件", SampleBody: `{}`},
	{ID: "goods-cert-upload-product", Category: "商品尺码与认证", Name: "上传商品认证", Type: "bg.arbok.open.cert.uploadProductCert", Description: "上传商品认证", SampleBody: `{}`},

	// ── 商品价格 ──
	{ID: "goods-price-list", Category: "商品价格", Name: "价格列表", Type: "temu.goods.price.list.get", Description: "查询商品价格列表", SampleBody: `{}`},
	{ID: "goods-price-adjust-query", Category: "商品价格", Name: "调价单查询", Type: "bg.full.adjust.price.page.query", Description: "分页查询调价单", SampleBody: `{"pageNo": 1, "pageSize": 10}`},
	{ID: "goods-price-adjust-review", Category: "商品价格", Name: "批量审核调价", Type: "bg.full.adjust.price.batch.review", Description: "批量审核调价单", SampleBody: `{}`},
	{ID: "goods-price-review-query", Category: "商品价格", Name: "半托核价查询", Type: "bg.semi.price.review.page.query.order", Description: "半托管核价单查询", SampleBody: `{"pageNo": 1, "pageSize": 10}`},
	{ID: "goods-price-review-confirm", Category: "商品价格", Name: "半托核价确认", Type: "bg.semi.price.review.confirm.order", Description: "确认半托管核价", SampleBody: `{}`},
	{ID: "goods-price-review-reject", Category: "商品价格", Name: "半托核价拒绝", Type: "bg.semi.price.review.reject.order", Description: "拒绝半托管核价", SampleBody: `{}`},

	// ── 备货单 ──
	{ID: "purchase-query", Category: "备货单", Name: "备货单查询", Type: "bg.purchaseorderv2.get", Description: "分页查询备货单列表", SampleBody: `{"pageNo": 1, "pageSize": 10}`},
	{ID: "purchase-apply", Category: "备货单", Name: "申请备货", Type: "bg.purchaseorder.apply", Description: "申请备货单", SampleBody: `{}`},
	{ID: "purchase-edit", Category: "备货单", Name: "编辑备货单", Type: "bg.purchaseorder.edit", Description: "编辑备货单", SampleBody: `{}`},
	{ID: "purchase-cancel", Category: "备货单", Name: "取消备货单", Type: "bg.purchaseorder.cancel", Description: "取消备货单", SampleBody: `{}`},

	// ── 发货台 ──
	{ID: "ship-staging-query", Category: "发货台", Name: "发货台查询", Type: "bg.shiporder.staging.get", Description: "查询发货台数据", SampleBody: `{"pageNo": 1, "pageSize": 10}`},
	{ID: "ship-staging-add", Category: "发货台", Name: "加入发货台", Type: "bg.shiporder.staging.add", Description: "将备货单加入发货台", SampleBody: `{}`},

	// ── 发货单 ──
	{ID: "ship-order-query", Category: "发货单", Name: "发货单查询", Type: "bg.shiporderv2.get", Description: "分页查询发货单", SampleBody: `{"pageNo": 1, "pageSize": 10}`},
	{ID: "ship-order-create", Category: "发货单", Name: "创建发货单", Type: "bg.shiporderv3.create", Description: "创建发货单 V3", SampleBody: `{}`},
	{ID: "ship-order-cancel", Category: "发货单", Name: "取消发货单", Type: "bg.shiporder.cancel", Description: "取消发货单", SampleBody: `{}`},
	{ID: "ship-order-logistics-get", Category: "发货单", Name: "发货单物流", Type: "bg.shiporder.logistics.get", Description: "查询发货单物流信息", SampleBody: `{}`},
	{ID: "ship-order-logistics-match", Category: "发货单", Name: "物流匹配 V3", Type: "bg.shiporderv3.logisticsmatch.get", Description: "发货单物流商匹配 V3", SampleBody: `{}`},
	{ID: "ship-order-logistics-order-match", Category: "发货单", Name: "物流单匹配", Type: "bg.shiporder.logisticsorder.match", Description: "物流单匹配", SampleBody: `{}`},
	{ID: "ship-order-logistics-change", Category: "发货单", Name: "变更物流", Type: "bg.shiporder.logistics.change", Description: "变更发货单物流", SampleBody: `{}`},
	{ID: "ship-receive-address", Category: "发货单", Name: "收货地址", Type: "bg.shiporder.receiveaddressv2.get", Description: "查询收货地址", SampleBody: `{}`},
	{ID: "ship-package-get", Category: "发货单", Name: "包裹查询", Type: "bg.shiporder.package.get", Description: "查询发货单包裹", SampleBody: `{}`},
	{ID: "ship-package-edit", Category: "发货单", Name: "编辑包裹", Type: "bg.shiporder.package.edit", Description: "编辑发货单包裹", SampleBody: `{}`},
	{ID: "ship-packing-send", Category: "发货单", Name: "装箱发货", Type: "bg.shiporder.packing.send", Description: "装箱发货", SampleBody: `{}`},
	{ID: "ship-packing-match", Category: "发货单", Name: "装箱匹配", Type: "bg.shiporder.packing.match", Description: "装箱匹配", SampleBody: `{}`},

	// ── 全托管物流 ──
	{ID: "logistics-companies", Category: "全托管物流", Name: "快递公司列表", Type: "bg.logistics.company.get", Description: "查询发货快递公司（CN 网关）", SampleBody: `{}`},

	// ── 半托管-订单 ──
	{ID: "semi-order-list", Category: "半托管-订单", Name: "订单列表", Type: "bg.order.list.v2.get", Description: "半托管订单列表（需 region=US/EU）", SampleBody: `{"pageNumber": 1, "pageSize": 10}`},
	{ID: "semi-order-shipping", Category: "半托管-订单", Name: "订单物流信息", Type: "bg.order.shippinginfo.v2.get", Description: "查询订单物流信息", SampleBody: `{}`},
	{ID: "semi-order-customization", Category: "半托管-订单", Name: "定制品信息", Type: "bg.order.customization.get", Description: "查询定制品定制信息", SampleBody: `{}`},

	// ── 半托管-物流 ──
	{ID: "semi-logistics-companies", Category: "半托管-物流", Name: "物流商列表", Type: "bg.logistics.companies.get", Description: "查询半托物流商", SampleBody: `{}`},
	{ID: "semi-logistics-template", Category: "半托管-物流", Name: "物流模板", Type: "temu.logistics.template.get", Description: "查询物流模板", SampleBody: `{}`},
	{ID: "semi-logistics-warehouse", Category: "半托管-物流", Name: "发货仓库", Type: "bg.logistics.warehouse.list.get", Description: "查询发货仓库列表", SampleBody: `{}`},
	{ID: "semi-logistics-shipping-services", Category: "半托管-物流", Name: "物流服务商", Type: "bg.logistics.shippingservices.get", Description: "根据仓库和订单获取物流服务商", SampleBody: `{}`},
	{ID: "semi-logistics-shipment-create", Category: "半托管-物流", Name: "物流下单", Type: "bg.logistics.shipment.create", Description: "创建物流运单", SampleBody: `{}`},
	{ID: "semi-logistics-shipment-result", Category: "半托管-物流", Name: "下单结果", Type: "bg.logistics.shipment.result.get", Description: "查询物流下单结果", SampleBody: `{}`},
	{ID: "semi-logistics-shipment-update", Category: "半托管-物流", Name: "更新运单", Type: "bg.logistics.shipment.update", Description: "更新物流运单", SampleBody: `{}`},
	{ID: "semi-logistics-shipment-shippingtype", Category: "半托管-物流", Name: "更新运输类型", Type: "bg.logistics.shipment.shippingtype.update", Description: "更新运输类型", SampleBody: `{}`},
	{ID: "semi-logistics-shipment-document", Category: "半托管-物流", Name: "获取面单", Type: "bg.logistics.shipment.document.get", Description: "获取物流面单", SampleBody: `{}`},
	{ID: "semi-logistics-shipment-v2-get", Category: "半托管-物流", Name: "运单查询 V2", Type: "bg.logistics.shipment.v2.get", Description: "查询运单 V2（US 网关）", SampleBody: `{}`},
	{ID: "semi-logistics-shipment-v2-confirm", Category: "半托管-物流", Name: "运单确认 V2", Type: "bg.logistics.shipment.v2.confirm", Description: "确认运单 V2（US 网关）", SampleBody: `{}`},
	{ID: "semi-logistics-scanform-create", Category: "半托管-物流", Name: "创建 ScanForm", Type: "temu.logistics.scanform.create", Description: "创建 ScanForm（US 网关）", SampleBody: `{}`},
	{ID: "semi-logistics-scanform-document", Category: "半托管-物流", Name: "ScanForm 面单", Type: "temu.logistics.scanform.document.get", Description: "获取 ScanForm 面单", SampleBody: `{}`},

	// ── 半托管-包裹 ──
	{ID: "semi-unshipped", Category: "半托管-包裹", Name: "待发货包裹", Type: "bg.order.unshipped.package.get", Description: "获取待发货包裹", SampleBody: `{"pageNumber": 1, "pageSize": 10}`},
	{ID: "semi-shipped-confirm", Category: "半托管-包裹", Name: "确认已发货", Type: "bg.logistics.shipped.package.confirm", Description: "确认包裹已发货", SampleBody: `{}`},

	// ── 库存 ──
	{ID: "stock-quantity-get", Category: "库存", Name: "库存查询", Type: "bg.btg.goods.stock.quantity.get", Description: "查询商品库存（Partner 网关）", SampleBody: `{}`},
	{ID: "stock-quantity-update", Category: "库存", Name: "库存更新", Type: "bg.btg.goods.stock.quantity.update", Description: "更新商品库存", SampleBody: `{}`},
	{ID: "stock-warehouse-list", Category: "库存", Name: "库存仓库列表", Type: "bg.btg.goods.stock.warehouse.list.get", Description: "查询库存仓库列表", SampleBody: `{}`},

	// ── JIT ──
	{ID: "jit-activate", Category: "JIT", Name: "激活 JIT 模式", Type: "temu.jitmode.activate", Description: "激活 JIT 模式", SampleBody: `{}`},
	{ID: "jit-inventory-get", Category: "JIT", Name: "虚拟库存查询", Type: "bg.qtg.stock.virtualinventoryjit.get", Description: "查询 JIT 虚拟库存", SampleBody: `{"productSkcId": 7469668867}`},
	{ID: "jit-inventory-edit", Category: "JIT", Name: "虚拟库存编辑", Type: "bg.qtg.stock.virtualinventoryjit.edit", Description: "编辑 JIT 虚拟库存", SampleBody: `{}`},
	{ID: "jit-rule-get", Category: "JIT", Name: "预售规则查询", Type: "bg.virtualinventoryjit.rule.get", Description: "查询 JIT 预售规则", SampleBody: `{}`},
	{ID: "jit-rule-sign", Category: "JIT", Name: "签署预售规则", Type: "bg.virtualinventoryjit.rule.sign", Description: "签署 JIT 预售规则", SampleBody: `{}`},

	// ── 广告 ──
	{ID: "ad-detail", Category: "广告", Name: "广告详情", Type: "bg.glo.searchrec.ad.detail.query", Description: "查询广告详情", SampleBody: `{}`},
	{ID: "ad-create", Category: "广告", Name: "创建广告", Type: "bg.glo.searchrec.ad.create", Description: "创建广告", SampleBody: `{}`},
	{ID: "ad-batch-modify", Category: "广告", Name: "批量修改广告", Type: "bg.glo.searchrec.ad.batch.modify", Description: "批量修改广告", SampleBody: `{}`},
	{ID: "ad-roas-pred", Category: "广告", Name: "ROAS 预测", Type: "bg.glo.searchrec.ad.roas.pred", Description: "广告 ROAS 预测", SampleBody: `{}`},
	{ID: "ad-log-query", Category: "广告", Name: "广告日志", Type: "bg.glo.searchrec.ad.log.query", Description: "查询广告操作日志", SampleBody: `{}`},
	{ID: "ad-reports-goods", Category: "广告", Name: "商品广告报表", Type: "bg.glo.searchrec.ad.reports.goods.query", Description: "查询商品广告报表", SampleBody: `{}`},

	// ── 图片与工具 ──
	{ID: "picture-compress", Category: "图片与工具", Name: "图片压缩", Type: "temu.picturecompression.get", Description: "获取图片压缩结果", SampleBody: `{}`},

	// ── 爆款邀约 ──
	{ID: "best-seller-invitation", Category: "爆款邀约", Name: "爆款邀约查询", Type: "temu.best.seller.invitation.query", Description: "查询爆款邀约", SampleBody: `{}`},
}
