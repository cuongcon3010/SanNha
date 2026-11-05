# Frontend

Next.js App Router, TypeScript, CSS, Lucide. Run commands inside `FE/`.

- `npm ci`, `npm run dev`, `npm run build`
- `npm run lint`, `npm run typecheck`, `npm test`, `npm run format:check`
- Read `node_modules/next/dist/docs/` before changing Next.js APIs.
- Preserve the Penpot palette, Inter font, responsive layout, Vietnamese text and demo notices.
- Fixture booking pages/dashboard and password forms still use demo behavior. OAuth, `/tai-khoan`, `/dat-san` and `/thanh-toan/[id]` call the Go backend with its real session cookie. Never send localStorage/demo booking IDs to live payments or present demo role navigation as actual authorization. Payment success must come from the backend callback state, not a client button or localStorage.

<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
