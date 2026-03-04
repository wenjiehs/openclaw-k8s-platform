# 前端 Dockerfile - 多阶段构建
# 阶段 1: 开发阶段（用于 docker-compose 开发模式）
FROM node:18-alpine AS development

WORKDIR /app

# 复制依赖文件
COPY package.json package-lock.json* ./
RUN npm install

# 暴露开发服务器端口
EXPOSE 3000

# 开发模式启动命令（挂载源码后执行）
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# =========================================
# 阶段 2: 构建阶段
FROM node:18-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY package.json package-lock.json* ./
RUN npm install

# 复制源代码
COPY . .

# 构建生产版本
RUN npm run build

# =========================================
# 阶段 3: 生产阶段（使用 nginx 提供静态文件）
FROM nginx:alpine AS production

# 复制构建产物到 nginx 目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置（支持 React Router 的 history 模式）
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
