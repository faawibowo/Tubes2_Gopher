FROM node:18-alpine AS base

WORKDIR /app

# ─────────────────────────────
# 1. Dependencies layer
# ─────────────────────────────
FROM base AS deps

RUN apk add --no-cache libc6-compat

COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* .npmrc* ./

RUN \
    if [ -f yarn.lock ]; then yarn install --frozen-lockfile; \
    elif [ -f package-lock.json ]; then npm ci; \
    elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm install --frozen-lockfile; \
    else echo "Lockfile not found." && exit 1; \
    fi

# ─────────────────────────────
# 2. Builder layer
# ─────────────────────────────
FROM base AS builder

# ARG SITE_URL
# ARG BACKEND_URL
ARG NEXT_PUBLIC_BACKEND_HTTP
ARG NEXT_PUBLIC_BACKEND_WS

# ENV SITE_URL=$SITE_URL
# ENV BACKEND_URL=$BACKEND_URL
ENV NEXT_PUBLIC_BACKEND_HTTP=$NEXT_PUBLIC_BACKEND_HTTP
ENV NEXT_PUBLIC_BACKEND_WS=$NEXT_PUBLIC_BACKEND_WS

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN \
    if [ -f yarn.lock ]; then yarn build; \
    elif [ -f package-lock.json ]; then npm run build; \
    elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm run build; \
    else echo "Lockfile not found." && exit 1; \
    fi

# ─────────────────────────────
# 3. Runtime layer
# ─────────────────────────────
FROM base AS runner

WORKDIR /usr/local/rest-api

ENV NODE_ENV=production
ENV HOSTNAME=0.0.0.0

RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
CMD ["node", "server.js"]
