# 实施计划 - 支付网关集成

## 当前阶段: 已完成

### 后端 (Backend)
- [x] 创建 `PaymentModule`，包含 `PaymentService` 和 `PaymentController`。
- [x] 在 `PaymentService` 中封装 SDK 调用逻辑。
- [x] 在 `SettingsService` 中添加支付相关设置。
- [x] 更新 `StoreService` 以处理带支付的结账流程。
- [x] 在 `PaymentController` 中实现回调处理（更新 `Transaction` 状态，发放订阅权限）。
- [x] 解决 `StoreModule` 和 `PaymentModule` 之间的循环依赖。

### 前端 (Frontend)
- [x] 更新 `api.ts` 以支持带 `payType`（支付方式）的 `checkout` 方法。
- [x] 更新 `App.tsx` 上下文以处理支付链接跳转。
- [x] 更新 `StoreView.tsx`，在购物车抽屉中增加支付方式选择器。
- [x] 更新 `AdminView.tsx`，在系统设置中增加支付参数配置界面。

### 下一步
- [ ] 用户验收测试（购买产品，验证权限）。
- [ ] 在管理后台配置真实的商户凭证。
