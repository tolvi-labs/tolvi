# syntax=docker/dockerfile:1.6
#
# Multi-stage build for the Tolvi server.
#
# Build context is the REPO ROOT (not server/), so this file is invoked as:
#   docker build -f server/Dockerfile -t tolvilabs/tolvi-server .
#
# That lets us COPY spec/schemas into the image alongside the server code.

# Stage 1: install all deps (incl. dev) for the build
FROM node:20-alpine AS deps
WORKDIR /app
COPY server/package.json server/package-lock.json ./
RUN npm ci

# Stage 2: compile TypeScript
FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY server/ ./
# Bring spec/schemas in alongside the server source so the runtime image
# can ship them at a stable relative path.
COPY spec/schemas ./spec-schemas
RUN npm run build

# Stage 3: lean runtime image (production deps only)
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY server/package.json server/package-lock.json ./
RUN npm ci --omit=dev && npm cache clean --force
COPY --from=build /app/dist ./dist
COPY --from=build /app/src/db/migrations ./src/db/migrations
COPY --from=build /app/spec-schemas ./spec/schemas
EXPOSE 3000
CMD ["node", "dist/server.js"]
