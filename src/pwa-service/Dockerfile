FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./

RUN npm ci --only=production

FROM node:18-alpine

WORKDIR /app

RUN apk add --no-cache dumb-init

COPY --from=builder /app/node_modules ./node_modules
COPY server.js ./
COPY service-worker.js ./
COPY public ./public

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001 && \
    chown -R nodejs:nodejs /app

USER nodejs

EXPOSE 8095

ENV NODE_ENV=production
ENV PORT=8095

ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "server.js"]
