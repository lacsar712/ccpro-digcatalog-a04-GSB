# 考古发掘出土文物编目系统（DigCatalog）

面向考古工地出土文物登记与编目的全栈演示项目：支持发掘工地、探方/发掘单位、出土文物、材质字典的 CRUD，以及概览统计。

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
6. **概览页** — 工地数、探方数、文物总数、按器物类型统计
7. **测年送检 DatingSubmission** — 围绕出土文物发起碳十四（c14）/ 热释光（tl）测年，跟踪草稿→送检→结果/作废全流程

### 测年送检模块说明

- **送检对象与文物强绑定**：每张送检单必须且只能关联**一件出土文物（Find）**，禁止创建与文物脱节的空白工单。后端对 `linkedFindId` / `linkedSampleId` 做互斥校验：两者都传或都不传均返回 400。
- **样品（Sample）声明**：当前系统**尚未建立 Sample 表**，因此本期送检仅支持挂 Find；模型已预留 `linkedSampleId` 字段，待 Sample 表落地后开放，届时仍与 `linkedFindId` 互斥（二选一恰选一个）。
- **字段**：`labName`（承测实验室）、`method`（仅 `c14` 或 `tl`）、`status`、`submittedAt`、`resultedAt`、`resultText`（结果文本，可空）。
- **状态机**：

  ```
  draft ──提交──▶ submitted ──回填结果──▶ resulted（终态）
                        └────作废──────▶ void（终态）
  ```

  - 仅 `draft` 可编辑/删除/提交；`submitted` 可回填结果或作废。
  - `resulted`、`void` 为终态，不可再流转、不可修改送检关联。
  - 任何非法状态跳转返回 **409** 及中文错误信息（如「非法流转：仅草稿（draft）状态可提交送检，当前状态为 submitted」）。
- **前端**：侧栏「测年送检」支持按状态、方法筛选；详情弹窗可执行流转并展示结果；出土文物列表每行提供「送检测年」按钮，跳转后以该文物预填创建草稿。
- **引用保护**：文物若存在未作废的送检单，禁止删除，避免工单悬空。
- **种子数据**：内置 draft、submitted、resulted 各 1 张，另附 1 张 void 作废单，均挂在具体文物上。

## API 前缀

所有接口以 `/api` 开头：

- `POST /api/auth/login`
- `GET|POST|PUT|DELETE /api/sites`
- `GET|POST|PUT|DELETE /api/units`
- `GET|POST|PUT|DELETE /api/finds`
- `GET|POST|PUT|DELETE /api/materials`
- `GET|POST|PUT|DELETE /api/dating-submissions`
- `POST /api/dating-submissions/:id/transition`（body：`{"action":"submit|result|void","resultText":"..."}`，非法跳转返回 409）
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
