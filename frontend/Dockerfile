# =========================
# STAGE 1: DEPENDENCIES
# =========================
FROM node:20-alpine AS deps
WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

# =========================
# STAGE 2: BUILDER
# =========================
FROM deps AS builder
WORKDIR /app
COPY frontend/ ./
RUN npm run build

# =========================
# STAGE 3: DEV MODE (Vite Dev Server)
# =========================
FROM node:20-alpine AS dev
WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
COPY --from=deps /app/node_modules ./node_modules
COPY frontend/ ./

EXPOSE 3000
CMD ["npm", "run", "dev", "--", "--host"]

# =========================
# STAGE 4: PROD RUNTIME (Nginx Minimal Static Web Server)
# =========================
FROM nginx:alpine AS prod

COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
