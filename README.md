# 考古发掘出土文物编目系统（DigCatalog）

面向考古工地出土文物登记与编目的全栈演示项目：支持发掘工地、探方/发掘单位、出土文物、材质字典的 CRUD，以及围绕**出土文物**的测年送检（碳十四 / 热释光）全流程管理与概览统计。

## 技术栈

- **前端**: Vue 3 + Vite + Pinia + Vue Router（Composition API + `<script setup>`）
- **后端**: Go 1.21+ + Gin + GORM
- **数据库**: MySQL 8.0
- **认证**: JWT + bcrypt

## 一键启动

```bash
docker compose up --build
```

启动完成后访问：

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3200 |
| 后端 API | http://localhost:8200/api |
| MySQL | localhost:3307（用户 `root` / 密码 `root`，库名 `digcatalog`） |

停止服务：

```bash
docker compose down
```

清除数据卷后重建：

```bash
docker compose down -v
docker compose up --build
```

## 测试账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `admin` | `123456` | 管理员 |
| `recorder` | `123456` | 记录员 |

## 功能模块

1. **登录认证** — 管理员 / 记录员角色，JWT 鉴权
2. **发掘工地 Site** — 名称、时代、经纬度、负责人
3. **探方/发掘单位 Unit** — 所属工地、编号、深度区间、地层简述
4. **出土文物 Find** — 所属探方、登记号、器物类型、材质、完整度、出土日期、描述、存放位置
5. **材质分类 Material** — 名称、描述（字典表）
6. **测年送检 DatingSubmission** — 每单必须关联一件出土文物，记录承测实验室、测年方法（`c14` 碳十四 / `tl` 热释光），并按状态机流转：`draft`（草稿）→ `submitted`（已送检）→ `resulted`（已出结果）/ `void`（已作废）。支持状态/方法筛选、详情中流转并展示测年结果，可从「出土文物」页一键以指定文物创建草稿。
7. **概览页** — 工地数、探方数、文物总数、按器物类型统计

> **送检对象说明**：设计上送检对象为「出土文物 Find」或「采样记录 Sample」二选一且必须恰选一个（`linkedFindId` 与 `linkedSampleId` 互斥，后端已做双向校验）。由于当前版本尚未建立 Sample 表，**本期仅支持挂接 Find**；提交 `linkedSampleId` 会收到 400 错误提示。新增 Sample 表后放开互斥校验的另一分支即可，无需改动状态机与前端流转。

### 测年送检状态机

```
draft ──submit──▶ submitted ──result──▶ resulted（终态）
                      └──────void──────▶ void（终态）
```

- 仅 `draft` 可编辑/删除；`submit` 时写入 `submittedAt`
- 仅 `submitted` 可登记结果（必填 `resultText`，写入 `resultedAt`）或作废
- `resulted` / `void` 为终态，文物关联与单据内容均锁定
- 任何非法跳转返回 **409** 与中文错误说明（如「非法状态流转：仅草稿（draft）可送检…」）

## API 前缀

所有接口以 `/api` 开头：

- `POST /api/auth/login`
- `GET|POST|PUT|DELETE /api/sites`
- `GET|POST|PUT|DELETE /api/units`
- `GET|POST|PUT|DELETE /api/finds`
- `GET|POST /api/dating-submissions`、`GET|PUT|DELETE /api/dating-submissions/:id`
- `POST /api/dating-submissions/:id/transition`（body：`{"action":"submit|result|void","resultText":"..."}`；支持 `?status=`、`?method=`、`?findId=` 筛选）
- `GET|POST|PUT|DELETE /api/materials`
- `GET /api/overview`

前端经 Nginx 将 `/api` 反代至后端容器 `http://backend:8080`。

## 端口映射

| 服务 | 宿主机 | 容器内 |
|------|--------|--------|
| Frontend | 3200 | 80 |
| Backend | 8200 | 8080 |
| MySQL | 3307 | 3306 |

## 目录结构

```
DigCatalog/
├── docker-compose.yml
├── README.md
├── .gitignore
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── internal/
│       ├── config/
│       ├── models/
│       ├── handlers/
│       ├── middleware/
│       └── seed/
└── frontend/
    ├── Dockerfile
    ├── nginx.conf
    ├── package.json
    ├── vite.config.js
    ├── index.html
    └── src/
```

## 本地开发（可选）

### 后端

```bash
cd backend
go mod tidy
# 确保 MySQL 已启动且环境变量正确
go run .
```

### 前端

```bash
cd frontend
npm install --registry=https://registry.npmmirror.com
npm run dev
```
