FROM oven/bun:alpine

WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run svelte-kit sync

EXPOSE 5173
CMD ["bun", "run", "vite", "dev", "--host", "0.0.0.0", "--port", "5173"]
