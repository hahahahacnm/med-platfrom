# 后端部署指南 (Debian/Linux)

如果在服务器上启动失败并出现 `DriverPackageNotInstalledError: SQLite package has not been found installed` 错误，通常是因为 `node_modules` 包含与当前操作系统不兼容的二进制文件（例如从 Windows 复制过去的）。

## 修复步骤

请在服务器的后端目录 (`backend/`) 执行以下操作：

1. **清理旧依赖** (非常重要)：
   ```bash
   rm -rf node_modules package-lock.json
   ```

2. **重新安装依赖**：
   ```bash
   npm install
   ```
   > 注意：`sqlite3` 是原生模块，必须在服务器上编译或下载对应的二进制包。重新安装会自动处理。

3. **重新构建项目**：
   ```bash
   npm run build
   ```

4. **启动服务**：
   ```bash
   # 开发模式
   npm run start:dev
   
   # 或生产模式
   npm run start:prod
   ```

## 常见问题

如果 `npm install` 安装 `sqlite3` 失败，可能需要安装构建工具：

```bash
sudo apt-get update
sudo apt-get install build-essential python3
```

然后再运行 `npm install`。



rm -rf node_modules package-lock.json && npm install && npm run build && npm run start:prod