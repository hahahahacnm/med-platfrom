# 题库 Excel 批量导入功能

## 目标
实现管理员后台题库管理功能，支持通过 Upload Excel 批量导入题目，以文件名作为章节名称，自动组织数据。

## 已完成工作
1. **后端开发**:
   - 安装 `xlsx` 库用于解析 Excel 文件。
   - `QuizService` 新增 `importQuestions` 方法：
     - 解析 Excel 文件。
     - 提取文件名作为章节 (`Chapter`) 标题。
     - 自动映射 Excel 列 (`题干`, `选项`, `正确答案`, `解析`) 到 `Question` 实体。
     - 支持 B/S 架构下的批量上传。
   - `QuizController` 新增 `POST /subjects/:id/import` 接口。

2. **前端开发**:
   - `api.ts` 新增 `importQuestions` 方法。
   - `AdminView.tsx` (`ContentManagement` 组件) 新增导入按钮。
   - 实现文件选择与上传交互，支持进度提示与结果反馈。

## 待优化
- 支持更多复杂的题型格式 (如判断题、填空题)。
- 增加导入预览功能。
