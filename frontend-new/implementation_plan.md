# 前端架构重构实施计划 (Astro + React + TailwindCSS)

## 🎯 重构目标
彻底改变原有 Vue 3 前端的技术栈与视觉风格。
- **技术栈跃迁**: 拥抱 `Astro` (追求极致构建与加载速度) + `React` (强大的组件生态) + `TailwindCSS` (原子化 CSS) + `TypeScript` (强类型保障)。
- **视觉升级**: 对齐现代 SaaS 体验，蓝白主色调，大量应用卡片式(Card-based) 设计、玻璃拟态(Glassmorphism)、呼吸感排版与细致的微交互表现。
- **动效引入**: 使用 `framer-motion` 实现页面切换与元素的 Staggered 优雅呈现，打破沉闷的医疗平台刻板印象。

## 📊 阶段执行详情

### 阶段一：基础设施构建 (正在进行/已完成)
1. **项目初始化**：
   - 清理并初始化 `frontend-new` 目录下 Astro 环境。
   - 集成 `@astrojs/react` 与 `@astrojs/tailwind`。
   - 安装动效与图标库：`framer-motion`, `lucide-react`。
2. **构建设计系统 (Design System)**：
   - 配置 `tailwind.config.mjs`，定义 `primary`, `muted`, `background`, `card` 等语义化变量。
   - 创建 `src/styles/global.css` 实现 CSS 变量，确保支持深浅色模式的无缝扩展。
3. **全局布局组件搭建 (`Layout.astro`)**：
   - 响应式外框。左侧固定侧边栏 (`Sidebar.tsx`)，顶部带抽屉菜单栏 (`Topbar.tsx`)。
   - 主体内容区滚动，保持原生的 App 级体验。

### 阶段二：核心模块研发与视觉落地
1. **仪表盘 (Dashboard)**：
   - 全新的欢迎区 (Banner)，带动态渐变与问候语。
   - 数据概览卡片，使用弹簧动效加载体验。
   - 模拟的热力图学习打卡组件。
2. **题库中心 (QuizBank)**：
   - 网格列表卡片布局，带环形或条形进度指示器。
   - 各科类目的进入按钮与悬浮交互过渡。
3. **其他沉淀流页面 (错题本, 收藏夹, 笔记本)**：
   - 快速构建通用列表卡片容器，统一头部 Actions、搜索区及过滤区，维持视觉一致性。

### 阶段三：迭代与完善
1. **个人中心模块重构**：
   - 将原有的 Vue 布局重写为现代化的个人资料页 (Profile.astro)。
2. **移动端深层适配优化**：
   - 侧边栏抽屉 Drawer 的手势交互优化。
   - 卡片在较宽与较小屏幕间的断点平滑。
3. **状态管理对接**：
   - 将后端 API 请求或状态流引入 React (如使用 SWR 或 React Query，视需求而定)。

## 📈 当前进度与下一步
- [x] 完成了 TailwindCSS 系统底层设计与配置。
- [x] 完成了 Main Layout 通用响应式框架 (Sidebar / Topbar)。
- [x] 完成了 Dashboard / Quiz 体验的框架搭建。
- [ ] 补齐 Mistakes, Favorites, Notes, Profile 界面的建设。
- [ ] 等待 npm 依赖彻底解析完成，测试页面在浏览器中的实际呈现。
