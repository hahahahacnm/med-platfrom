# 前端界面重构实施计划

## 目标
重构前端界面，实现现代优美、动效优雅的风格，自动适配桌面与移动端。主界面采用左侧边栏布局，风格参考 `other-platfrom`。

## 已完成工作
1.  **创建主布局 (`MainLayout.vue`)**:
    -   实现了响应式的左侧边栏 (桌面端固定，移动端抽屉式)。
    -   集成了 Logo、导航菜单、用户信息区域。
    -   使用 Vanilla CSS 实现现代化 UI (蓝/白/灰配色, 阴影, 圆角)。
    -   移动端顶部导航栏。

2.  **配置路由 (`router/index.ts`)**:
    -   将主要用户页面 (`Home`, `Mistakes`, `Favorites`, `MyNotes`, `Profile`) 包装在 `MainLayout` 路由下。

3.  **重构核心页面**:
    -   **首页 (`Home.vue`)**: 移除旧的顶部导航，将搜索和题库切换整合到页面内容顶部的控制栏中。保留了原有的“侧边栏目录”与“答题卡”布局，调整为适应 flex 容器的高度。
    -   **错题本 (`Mistakes.vue`)**: 移除顶部导航，统一改为页面内控制栏，样式卡片化，高度自适应。
    -   **收藏夹 (`Favorites.vue`)**: 同上，统一 UI 风格。
    -   **笔记本 (`MyNotes.vue`)**: 同上，移除冗余导航链接。

## 待执行任务
1.  **优化个人中心 (`Profile.vue`)**:
    -   确保其容器支持滚动 (`overflow-y: auto`)，以适应新的固定高度布局。
2.  **验证与测试**:
    -   检查各页面在桌面与移动端的表现。
    -   确保构建无错误。

## 详细步骤
- [x] 创建 `src/layout/MainLayout.vue`
- [x] 修改 `src/router/index.ts`
- [x] 重构 `src/views/Home.vue`
- [x] 重构 `src/views/Mistakes.vue`
- [x] 重构 `src/views/Favorites.vue`
- [x] 重构 `src/views/MyNotes.vue`
- [x] 优化 `src/views/personal/Profile.vue`

## 仪表盘优化 (当前任务)
- [x] 全局样式优化 (`style.css`): 引入现代化字体与配色变量。
- [x] 仪表盘重构 (`Dashboard.vue`):
    - [x] 头部欢迎区: 增加玻璃拟态效果与更加友好的问候。
    - [x] 统计卡片: 优化视觉层次，增加图标与趋势指示。
    - [x] 学习强度图 (Heatmap): 优化配色与交互提示 (Tooltip)。
    - [x] 学科透视: 改进折叠面板样式与进度条视觉。
    - [x] 快捷入口: 设计为更加显眼的功能入口卡片。
    - [x] 卷王榜: 优化列表展示样式。
- [x] 由于用户要求，将仪表盘改为统一的蓝色主题，减少杂色。
- [x] 统一全局圆角为 16px (卡片) 和 12px (内部元素)。
- [x] 二次迭代优化:
    - [x] 深度适配移动端 (响应式 Grid, Banner 布局调整)。
    - [x] 增加页面入场 staggered 动画。
    - [x] 优化卡片色彩与阴影质感。

