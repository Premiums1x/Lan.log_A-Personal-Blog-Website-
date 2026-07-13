# lancer.log

一个程序员的写字处。

Go + PostgreSQL 服务端渲染博客，带自托管 React 后台。模板和静态资源全部 `//go:embed` 进单一二进制，部署只需要一个可执行文件 + 一份 .env + 一个 PG 库。

---

## 设计语言

开发者 monospace + 单冷调 zinc + 克制蓝 `#0969DA`（GitHub-blue，不是 AI 紫）。

- **字体：** Geist Sans / Geist Mono
- **文章列表：** `git log --oneline` 风，每篇带一个 sha1 commit hash + 日期 + 主题色 tag + `min read`
- **关于页：** 终端 `whoami` / `cat bio.yml` / `uptime` 卡片，`now.txt` 终端输出
- **面包屑：** `~/posts/随笔/enough-is-enough.md` 路径式
- **代码块：** mac 三圆点 + 语言标签 + 浅色语法高亮
- **暗色模式：** 跟随 `prefers-color-scheme`
- **降级：** `prefers-reduced-motion` 关闭动效

---

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.26 + Gin |
| 数据库 | PostgreSQL 14+（pgcrypto 扩展） |
| 模板 | `html/template` 服务端渲染公开页 |
| Markdown | goldmark（GFM + Linkify，保存时转 HTML） |
| 鉴权 | JWT (HS256) + bcrypt |
| 后台 SPA | React 18 + Vite 6 + React Router 6 + TanStack Query + **Ant Design 5** + Tailwind v4（仅作 token 容器） |
| 打包 | `//go:embed` 模板 / 静态资源 / 迁移 / 后台 SPA（build tag `admin` 开启） |

---

## 目录结构

```
.
├── cmd/
│   ├── blog/main.go            # 入口：load config → migrate → bootstrap admin → server
│   └── tplcheck/main.go        # 模板自检（typed 零数据渲染四页）
├── internal/
│   ├── auth/jwt.go             # JWT manager + bcrypt + 中间件
│   ├── config/config.go        # 环境变量读取
│   ├── db/db.go                # pgxpool 封装（Conn 接口 + WithTx）
│   ├── handler/
│   │   ├── api.go              # /api/* REST（login/me/posts/settings/brand）
│   │   ├── helpers.go          # commit hash / splitAccent 等纯函数
│   │   ├── public.go           # 公开页 handler（index/post/about/section/archive/tags/shelf/404）
│   │   ├── templates.go        # 模板加载 + clone 机制
│   │   └── types.go            # DTO + FuncMap
│   ├── markdown/md.go          # goldmark 渲染 + 字数（CJK 感知）+ 阅读时间 + 摘要
│   ├── migrate/migrate.go      # 顺序执行 web/migrations/*.sql
│   ├── model/model.go          # User / Post / Tag / Setting
│   ├── repo/repo.go            # 数据访问层（users/posts/tags/settings）
│   └── server/
│       ├── server.go           # 路由组装 + NoRoute 处理 /admin
│       ├── adminfs_dev.go      # !admin：adminFS() = nil
│       └── adminfs_prod.go     # +admin：embed web/admin-dist
├── web/
│   ├── embed.go                # //go:embed templates / static / migrations
│   ├── admin_embed.go          # //go:embed admin-dist（build tag: admin）
│   ├── migrations/0001_init.sql
│   ├── static/style.css        # 前台样式（design tokens + 暗色模式 + 响应式）
│   ├── templates/              # layout / index / post / about / archive / shelf / notfound
│   └── admin/                  # React 后台源码
└── go.mod                      # module github.com/lancer/log
```

---

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `HTTP_ADDR` | `:8080` | 监听地址 |
| `DATABASE_URL` | `postgres://blog:blog@localhost:5432/blog?sslmode=disable` | PG 连接串 |
| `JWT_SECRET` | `dev-secret-change-me-please-32bytes!` | JWT 签名密钥，**生产必改**，≥ 16 字节 |
| `JWT_TTL_HOURS` | `72` | token 有效期 |
| `ADMIN_USERNAME` | `admin` | 首次启动创建的管理员用户名 |
| `ADMIN_PASSWORD` | （随机） | 首次启动管理员密码；留空则生成随机密码并打印到 stdout |
| `ADMIN_EMAIL` | 空 | 首次启动创建管理员时写入找回邮箱；已有用户不会被覆盖 |
| `SMTP_HOST` | 空 | 密码找回邮件 SMTP 主机 |
| `SMTP_PORT` | `587` | 密码找回邮件 SMTP 端口 |
| `SMTP_USERNAME` | 空 | SMTP 用户名 |
| `SMTP_PASSWORD` | 空 | SMTP 密码 |
| `SMTP_FROM` | 空 | 密码找回邮件发件人地址 |

> 程序启动第一件事是连 PG 跑迁移，连不上会 `log.Fatalf` 退出。没有数据库无法启动。

---

## 本地开发

需要 Go 1.26+、Node 22+、可连的 PostgreSQL。

### 1. 准备数据库

任选一种方式起 PG：

**A. 直接装本机 PostgreSQL**
```bash
sudo -u postgres createuser blog -P
sudo -u postgres createdb blog -O blog
```

**B. 用 Docker 起一次性容器（推荐 Windows / WSL）**
```bash
docker run -d --name blog-pg -p 5432:5432 \
  -e POSTGRES_USER=blog -e POSTGRES_PASSWORD=blog -e POSTGRES_DB=blog \
  postgres:16
```

迁移会在程序首次启动时自动执行（`migrate.Run` 读 `web/migrations/*.sql` 顺序跑），**无需手动跑 SQL**。


### Windows 一键启动脚本

项目根目录提供了两个脚本，避免每次手动敲环境变量：

- `start-backend.cmd`：只启动 Go 后端。
- `start-dev.cmd`：同时打开 Go 后端和 Vite 管理端两个窗口。

首次运行时如果没有 `.env`，脚本会从 `.env.example` 自动复制一份 `.env`，先编辑里面的数据库、管理员账号和 QQ 邮箱 SMTP 授权码，再双击脚本启动。
### 2. 启动后端（dev 模式，admin 不 embed）

```bash
export DATABASE_URL="postgres://blog:blog@localhost:5432/blog?sslmode=disable"
export JWT_SECRET="dev-secret-change-me-please-32bytes!"
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin
export ADMIN_EMAIL=you@example.com
# Optional: enable forgot-password email codes
# export SMTP_HOST=smtp.example.com
# export SMTP_PORT=587
# export SMTP_USERNAME=you@example.com
# export SMTP_PASSWORD=your-smtp-password
# export SMTP_FROM=you@example.com

# 可选：用于后台生成文章摘要的 OpenAI-compatible 接口
LLM_API_URL=https://your-provider.example/v1/chat/completions
LLM_API_KEY=your-llm-api-key
LLM_MODEL=your-model-name
LLM_TIMEOUT_SECONDS=20

go run ./cmd/blog
```

启动后：

- 公开站 → http://localhost:8080
- API → http://localhost:8080/api
- 后台 SPA → dev 模式下二进制里没 embed admin，需另起 Vite（见下）

### 3. 启动后台前端（Vite dev server）

```bash
cd web/admin
npm install
npm run dev
```

Vite 跑在 http://localhost:5174，已在 `vite.config.ts` 配代理把 `/api` 和 `/static` 转发到 `:8080`。

打开 http://localhost:5174/admin/login 用 `admin / admin` 登录。登录后可在「账号安全」里修改密码并设置找回邮箱。

### 4. 模板自检

```bash
go run ./cmd/tplcheck
```

用 typed 零数据渲染四页（index/post/about/notfound），验证模板语法与 FuncMap 是否健全。

### Windows 开发的两个坑

1. **`go build` 在工作区上层无 git 仓库时报 `error obtaining VCS status`**：加 `-buildvcs=false`，或设 `GOFLAGS=-buildvcs=false` 长期生效。
2. **`go get` 拉不下来 proxy.golang.org**： Go 代理设国内源：
   ```powershell
   go env -w GOPROXY=https://goproxy.cn,direct
   ```
3. **PowerShell 在中文 Windows 下写 UTF-8 多字节字符会按 GBK 落地**：写源代码（含中文的模板、ts、md）用编辑器或 git，不要用 `Set-Content` / `Out-File` / `WriteAllText` 不带 `-Encoding utf8` 的写法。

---

## 构建生产二进制

生产用单一二进制：Go 交叉编译，admin SPA 先 `npm run build` 再用 `-tags admin` embed 进二进制。

### 1. 构建后台 SPA

```bash
cd web/admin
npm install
npm run build     # 产物 → web/admin-dist/
```

### 2. 编译 Go 二进制（Linux 服务器）

```bash
# 仓库根目录
GOOS=linux GOARCH=amd64 go build -tags admin -o bin/blog -buildvcs=false ./cmd/blog
```

`-tags admin` 开启 `web/admin_embed.go` 的 `//go:embed all:admin-dist`，把后台 SPA 打进二进制。

不带 `-tags admin` 编译时 `adminFS()` 返回 nil，`/admin/*` 由 NoRoute 兜底（适合只想部署公开站、后台另起的场景）。

### 3. 上传 + 运行

```bash
scp bin/blog user@server:/opt/blog/blog
ssh user@server 'chmod +x /opt/blog/blog'
```

样式或模板改动都要重 build + 替换二进制 + 重启服务。

---

## 服务器部署（推荐 2 核 / 2G RAM / 40G 盘）

### PostgreSQL

```bash
sudo apt update && sudo apt install -y postgresql

# 建库 + 用户（注意从 /tmp 跑，避开 /root permission denied）
cd /tmp
sudo -u postgres createuser blog -P
sudo -u postgres createdb blog -O blog

# 调小 shared_buffers 适配 2G 内存
ls /etc/postgresql/   # 查实际版本号目录（如 14 / 15 / 16）
sudo sed -i 's/^#shared_buffers.*/shared_buffers = 128MB/' /etc/postgresql/<VER>/main/postgresql.conf
sudo systemctl restart postgresql
```

### 创建运行用户

```bash
sudo useradd -r -s /bin/false blog
sudo mkdir -p /opt/blog
sudo chown -R blog:blog /opt/blog
```

### .env 文件

`/opt/blog/.env`（systemd 的 `EnvironmentFile` 用 `KEY=value` 格式）：

```
HTTP_ADDR=127.0.0.1:8088
DATABASE_URL=postgres://blog:强密码@127.0.0.1:5432/blog?sslmode=disable
JWT_SECRET=至少16字节的随机字符串
JWT_TTL_HOURS=72
ADMIN_USERNAME=admin
ADMIN_PASSWORD=强密码
ADMIN_EMAIL=you@example.com
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=you@example.com
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=you@example.com

# 可选：用于后台生成文章摘要的 OpenAI-compatible 接口
LLM_API_URL=https://your-provider.example/v1/chat/completions
LLM_API_KEY=your-llm-api-key
LLM_MODEL=your-model-name
LLM_TIMEOUT_SECONDS=20
```

> `HTTP_ADDR=127.0.0.1:8088` 只绑回环，只让 Nginx 反代访问，比 `:8088` 暴露公网更安全。
> `JWT_SECRET` 用 `openssl rand -hex 32` 生成一个 32 字节随机串。

```bash
sudo install -m 600 -o blog -g blog /dev/stdin /opt/blog/.env <<'EOF'
（粘贴上面 .env 内容）
EOF
```

### systemd 服务

`/etc/systemd/system/blog.service`：

```ini
[Unit]
Description=lancer.log blog
After=network.target postgresql.service

[Service]
Type=simple
User=blog
WorkingDirectory=/opt/blog
EnvironmentFile=/opt/blog/.env
ExecStart=/opt/blog/blog
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now blog
sudo journalctl -u blog -f --no-pager
```

看到 `listening on :8088 — admin at /admin, api at /api` 表示启动成功。

### Nginx 反代

`/etc/nginx/sites-available/blog`：

```nginx
server {
    listen 80;
    server_name _;                # 公网 IP 部署用 _；有域名写域名
    client_max_body_size 20M;

    location / {
        proxy_pass http://127.0.0.1:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
sudo ln -sf /etc/nginx/sites-available/blog /etc/nginx/sites-enabled/blog
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx
```

云控制台安全组放行 `80/tcp` 入站即可（PG 用 127.0.0.1 不用放，8088 也绑了 127.0.0.1 不用放）。

> **HTTPS（有域名才上）**： certbot 走 Let's Encrypt 签证书，把 `listen 80` 那个 server block 加一行 `return 301 https://$host$request_uri;`，新加一个 `listen 443 ssl http2;` 块即可。**Let's Encrypt 不签纯 IP**，没有域名只能走 HTTP。

### 旧的容器反代占用 :80 / :8080 怎么办

阿里云 / 腾讯云镜像常预装 Typecho / openresty 跑在 Docker 容器里，开机自启抢 :80 或 :8080：

```bash
sudo ss -tlnp | grep ':80\b'          # 看是谁占了
sudo docker ps | grep -iE 'typecho|rest'  # 找占用容器
sudo docker stop <容器ID>
sudo docker rm  <容器ID>
sudo docker update --restart=no <容器ID>   # 防止 always 策略再拉起（rm 之前做）
```

确认 :80 没人占后再 `systemctl reload nginx`。

---

## 路由与 API

### 公开页（无需鉴权）

| 路径 | 说明 |
|---|---|
| `GET /` | 首页（hero + 置顶 + git-log 风格文章列表 + 技术栈） |
| `GET /posts/:slug` | 文章详情 |
| `GET /about` | 关于页（终端 bio + now.txt + contact） |
| `GET /archive` | 归档 |
| `GET /tags` | 标签云 |
| `GET /tags/:tag` | 按 tag 筛选 |
| `GET /section/:section` | 按分区筛选 |
| `GET /shelf` | 书架 |
| `GET /static/*` | 静态资源（CSS / icon） |

### 后端 API

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/api/login` | 公开 | `{username, password}` → `{token}` |
| GET  | `/api/brand` | 公开 | 当前站点品牌名（后台侧栏用） |
| GET  | `/api/me` | JWT | 当前用户 |
| GET  | `/api/posts` | JWT | 列出全部文章（含 draft） |
| GET  | `/api/posts/:id` | JWT | 单篇 |
| POST | `/api/posts` | JWT | 新建 |
| PUT  | `/api/posts/:id` | JWT | 更新 |
| DELETE | `/api/posts/:id` | JWT | 删除 |
| GET  | `/api/settings` | JWT | 列出所有分区 |
| GET  | `/api/settings/:key` | JWT | 取单分区 |
| PUT  | `/api/settings/:key` | JWT | 写单分区（`{value: any}`） |

文章字段：`slug / title / excerpt / body_md / cover_url / section / status(draft|published) / pinned / tag_names[]`。

保存时 `body_md` 自动用 goldmark 转 `body_html`，并计算 `words`（CJK 感知）和 `read_minutes`，`excerpt` 留空时自动取正文前若干字。每篇生成一个 sha1 commit hash 写进 `posts.commit_hash`。

### 后台路由

| 路径 | 页面 |
|---|---|
| /admin/login | 登录 |
| /admin/forgot-password | 邮箱验证码找回密码 |
| /admin/account | 账号安全：修改密码 / 设置找回邮箱 |
| `/admin` | Dashboard（统计 + 最近文章） |
| `/admin/posts` | 文章列表 |
| `/admin/posts/new` | 新建 |
| `/admin/posts/:id` | 编辑 |
| `/admin/settings` | 分区列表 |
| `/admin/settings/:section` | 编辑单分区 |

---

## 后台使用

浏览器打开 `http(s)://你的域名或IP/admin`，用 `.env` 里的 `ADMIN_USERNAME` / `ADMIN_PASSWORD` 登录（首次启动如果 `ADMIN_PASSWORD` 留空，随机密码会打到 systemd journal）。登录后进入「账号安全」设置找回邮箱；配置 SMTP 后，可在 `/admin/forgot-password` 通过邮箱验证码重置密码。

- **概览**： 已发布 / 草稿 / 置顶 统计 + 最近 6 篇。
- **文章**： 搜索 + 状态过滤 + 表格列表，可置顶 / 删除。
- **文章编辑**： slug 自动从标题生成；Tag 回车添加；状态切 draft / publish；置顶开关；客户端简易预览。
- **站点设置**： 8 个分区卡片，每个分区是一个 JSON，改完保存即生效，公开页下次请求即读到新值。

8 个分区： **品牌与页脚 / 导航 / 首页主视觉 / 技术栈 / 关于页 / now.txt / 联系 / 页脚列**。

分区 JSON 结构见 `web/migrations/0001_init.sql` 末尾的 seed 数据。后台保存时只校验是合法 JSON，不校验字段——所以改分区前先看一眼 seed 里的结构。

### 改品牌名（侧栏标题）

后台 → 站点设置 → **品牌与页脚**（branding）分区，改 `brand` 字段，保存后前台 + 后台侧栏标题同步生效。

---

## 设计原型

`prototype-dev/` 是早期纯 HTML 原型（Geist Sans/Mono + zinc + GitHub-blue），生产模板 `web/templates/*.tmpl` 和样式 `web/static/style.css` 均由此移植。保留作设计参考。

```
prototype-dev/
├── index.html      # 首页 + git-log 文章列表
├── article.html    # 文章详情
├── about.html      # 关于 + 终端 now + contact
└── style.css
```

---

## License

MIT