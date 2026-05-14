# Stage 1: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
# This creates the /dist folder
RUN npm run build

# Stage 2: Runtime
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
# Install only production dependencies (no devDependencies like typescript)
RUN npm install --omit=dev
# Copy only the compiled JS from the builder stage
COPY --from=builder /app/dist ./dist

CMD ["node", "dist/index.js"]
