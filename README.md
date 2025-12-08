# 🧭 SZTU-Trip-Planner Backend

Golang + Gin 后端服务
提供注册登录、行程管理、路线规划等 RESTful API

## 🚀 快速启动

### 准备环境
- [Docker](https://www.docker.com/get-started/)

### 运行

将项目克隆到本地：

```bash
git clone https://github.com/Aminorsh/sztu-trip-planner-backend.git
cd sztu-trip-planner-backend
```

进入`go-service-driver`目录：

```bash
mv .env.example .env
```
编辑`.env`文件，配置数据库连接等参数。

使用 Docker Compose 启动服务：

```bash
docker compose up -d
```
服务启动后，API 将在 `http://localhost:8080` 可用。

## 📚 API 文档

### Postman Documentation:
- [SZTU Trip Planner Backend Go](https://github.com/Aminorsh/sztu-trip-planner-backend/blob/dev/postman/collections/SZTU%20Trip%20Planner%20Backend%20Go.postman_collection.json)

