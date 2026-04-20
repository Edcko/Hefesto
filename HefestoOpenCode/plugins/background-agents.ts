/**
 * BackgroundAgents — Hefesto OpenCode Plugin
 *
 * Async delegation system for OpenCode. Enables fire-and-forget background
 * tasks that persist to disk and survive context compaction.
 *
 * Features:
 *   1. `delegate` tool — launch background task, returns ID immediately
 *   2. `delegation_read` tool — retrieve persisted results
 *   3. `delegation_list` tool — list all delegations
 *   4. Session idle detection for completion
 *   5. Disk persistence to ~/.local/share/opencode/delegations/
 *   6. Compaction hook — inject delegation context for recovery
 *   7. Batched notifications — wait for ALL delegations to complete
 *   8. Anti-recursion — disable task/delegate/todowrite for sub-agents
 *   9. 15-minute timeout with cleanup
 *
 * Based on Gentleman.Dots background-agents.ts (MIT License)
 * https://github.com/Gentleman-Programming/Gentleman.Dots
 */

import * as crypto from "node:crypto"
import * as fs from "node:fs/promises"
import * as os from "node:os"
import * as path from "node:path"
import { stat } from "node:fs/promises"
import { type Plugin, type ToolContext, tool } from "@opencode-ai/plugin"
import type { createOpencodeClient } from "@opencode-ai/sdk"
import type { Event, Message, Part, TextPart } from "@opencode-ai/sdk"

// ─── Configuration ───────────────────────────────────────────────────────────

const MAX_RUN_TIME_MS = 15 * 60 * 1000 // 15 minutes
const DELEGATION_TIMEOUT_MS = 30_000 // 30 seconds for metadata generation

// Simple word lists for readable IDs (no npm dependency)
const ADJECTIVES = [
  "swift", "calm", "bright", "keen", "vivid", "crisp", "clear", "brave",
  "rapid", "solid", "gentle", "bold", "wise", "sharp", "quick", "steady"
]

const ANIMALS = [
  "falcon", "tiger", "eagle", "wolf", "bear", "lion", "fox", "hawk",
  "raven", "owl", "deer", "otter", "lynx", "seal", "crow", "finch"
]

// ─── Types ───────────────────────────────────────────────────────────────────

type OpencodeClient = ReturnType<typeof createOpencodeClient>

interface SessionMessageItem {
  info: Message
  parts: Part[]
}

interface AssistantSessionMessageItem {
  info: Message & { role: "assistant" }
  parts: Part[]
}

interface DelegationProgress {
  toolCalls: number
  lastUpdate: Date
  lastMessage?: string
  lastMessageAt?: Date
}

interface DelegationResult {
  publicSummary: string
  rawResult: string
}

interface Delegation {
  id: string // Human-readable ID (e.g., "swift-falcon")
  sessionID: string
  parentSessionID: string
  parentMessageID: string
  parentAgent: string
  prompt: string
  agent: string
  status: "running" | "complete" | "error" | "cancelled" | "timeout"
  startedAt: Date
  completedAt?: Date
  progress: DelegationProgress
  error?: string
  title?: string
  description?: string
  result?: string
  publicSummary?: string // Sanitized, concise summary for parent notifications
  rawResult?: string // Complete, unfiltered output for explicit retrieval
}

interface DelegateInput {
  parentSessionID: string
  parentMessageID: string
  parentAgent: string
  prompt: string
  agent: string
}

interface DelegationListItem {
  id: string
  status: string
  title?: string
  description?: string
  agent?: string
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

/**
 * Generate a simple readable ID from adjective + animal.
 * Replaces unique-names-generator dependency.
 */
function generateReadableId(): string {
  const adj = ADJECTIVES[Math.floor(Math.random() * ADJECTIVES.length)]
  const animal = ANIMALS[Math.floor(Math.random() * ANIMALS.length)]
  return `${adj}-${animal}`
}

/**
 * Generate title from prompt (first 80 chars).
 * Replaces small_model metadata generation.
 */
function generateTitle(prompt: string): string {
  const firstLine = prompt.split("\n").find((l) => l.trim().length > 0) || prompt
  return firstLine.slice(0, 77).trim() + (firstLine.length > 77 ? "..." : "")
}

/**
 * Generate description from result content.
 */
function generateDescription(content: string): string {
  return content.slice(0, 147).trim() + (content.length > 147 ? "..." : "")
}

/**
 * Sanitize raw sub-agent output into a concise public summary.
 * Strips reasoning blocks, tool-call noise, and internal patterns
 * to produce a clean summary suitable for parent-thread notifications.
 */
function sanitizePublicSummary(rawText: string): string {
  let text = rawText

  // Strip <thinking>...</thinking> blocks (greedy, multiline)
  text = text.replace(/<thinking>[\s\S]*?<\/thinking>/gi, "")

  // Strip <tool-call>...</tool-call> blocks (greedy, multiline)
  text = text.replace(/<tool-call>[\s\S]*?<\/tool-call>/gi, "")

  // Strip lines starting with common internal reasoning patterns
  text = text
    .split("\n")
    .filter((line) => {
      const trimmed = line.trim()
      // Strip blockquote-style "Thinking" / "Analyzing" lines
      if (/^>\s*\*Thinking\*/i.test(trimmed)) return false
      if (/^>\s*\*Analyzing\*/i.test(trimmed)) return false
      // Strip "Let me..." internal reasoning starts
      if (/^Let me\s/i.test(trimmed)) return false
      return true
    })
    .join("\n")

  // Trim and collapse whitespace
  text = text.trim().replace(/\n{3,}/g, "\n\n")

  // Truncate to 800 characters
  if (text.length > 800) {
    text = text.slice(0, 797).trimEnd() + "..."
  }

  // Fallback if empty after sanitization
  if (!text || text.trim().length === 0) {
    return "Task completed — no structured result produced"
  }

  return text
}

/**
 * Hash a path for project ID fallback.
 */
function hashPath(projectRoot: string): string {
  const hash = crypto.createHash("sha256").update(projectRoot).digest("hex")
  return hash.slice(0, 16)
}

/**
 * Get project ID from git root commit or path hash.
 * Inlined from kdco-primitives.
 */
async function getProjectId(projectRoot: string): Promise<string> {
  if (!projectRoot || typeof projectRoot !== "string") {
    throw new Error("getProjectId: projectRoot is required")
  }

  const gitPath = path.join(projectRoot, ".git")
  const gitStat = await stat(gitPath).catch(() => null)
  if (!gitStat) return hashPath(projectRoot)

  let gitDir = gitPath
  if (gitStat.isFile()) {
    const content = await Bun.file(gitPath).text()
    const match = content.match(/^gitdir:\s*(.+)$/m)
    if (!match) return hashPath(projectRoot)

    const resolvedGitdir = path.resolve(projectRoot, match[1].trim())
    const commondirPath = path.join(resolvedGitdir, "commondir")
    const commondirFile = Bun.file(commondirPath)

    if (await commondirFile.exists()) {
      const commondirContent = (await commondirFile.text()).trim()
      gitDir = path.resolve(resolvedGitdir, commondirContent)
    } else {
      gitDir = path.resolve(resolvedGitdir, "../..")
    }
  }

  // Try to get from cache
  const cacheFile = path.join(gitDir, "opencode")
  const cache = Bun.file(cacheFile)
  if (await cache.exists()) {
    const cached = (await cache.text()).trim()
    if (/^[a-f0-9]{40}$/i.test(cached) || /^[a-f0-9]{16}$/i.test(cached)) {
      return cached
    }
  }

  // Get root commit hash
  try {
    const proc = Bun.spawn(["git", "rev-list", "--max-parents=0", "--all"], {
      cwd: projectRoot,
      stdout: "pipe",
      stderr: "pipe",
    })

    const exitCode = await Promise.race([
      proc.exited,
      new Promise<number>((_, reject) =>
        setTimeout(() => {
          proc.kill()
          reject(new Error("git timeout"))
        }, 5000)
      ),
    ]).catch(() => 1)

    if (exitCode === 0) {
      const output = await new Response(proc.stdout).text()
      const roots = output.split("\n").filter(Boolean).map((x) => x.trim()).sort()
      if (roots.length > 0 && /^[a-f0-9]{40}$/i.test(roots[0])) {
        try { await Bun.write(cacheFile, roots[0]) } catch {}
        return roots[0]
      }
    }
  } catch {}

  return hashPath(projectRoot)
}

// ─── Logger ──────────────────────────────────────────────────────────────────

function createLogger(client: OpencodeClient) {
  const log = (level: "debug" | "info" | "warn" | "error", message: string) =>
    client.app.log({ body: { service: "background-agents", level, message } }).catch(() => {})
  return {
    debug: (msg: string) => log("debug", msg),
    info: (msg: string) => log("info", msg),
    warn: (msg: string) => log("warn", msg),
    error: (msg: string) => log("error", msg),
  }
}

type Logger = ReturnType<typeof createLogger>

// ─── Delegation Manager ──────────────────────────────────────────────────────

class DelegationManager {
  private delegations: Map<string, Delegation> = new Map()
  private client: OpencodeClient
  private baseDir: string
  private log: Logger
  private pendingByParent: Map<string, Set<string>> = new Map()

  constructor(client: OpencodeClient, baseDir: string, log: Logger) {
    this.client = client
    this.baseDir = baseDir
    this.log = log
  }

  /**
   * Resolve root session ID by walking up parent chain.
   */
  async getRootSessionID(sessionID: string): Promise<string> {
    let currentID = sessionID
    for (let depth = 0; depth < 10; depth++) {
      try {
        const session = await this.client.session.get({ path: { id: currentID } })
        if (!session.data?.parentID) return currentID
        currentID = session.data.parentID
      } catch {
        return currentID
      }
    }
    return currentID
  }

  private async getDelegationsDir(sessionID: string): Promise<string> {
    const rootID = await this.getRootSessionID(sessionID)
    return path.join(this.baseDir, rootID)
  }

  private async ensureDelegationsDir(sessionID: string): Promise<string> {
    const dir = await this.getDelegationsDir(sessionID)
    await fs.mkdir(dir, { recursive: true })
    return dir
  }

  /**
   * Delegate a task to an agent.
   */
  async delegate(input: DelegateInput): Promise<Delegation> {
    // Generate unique readable ID
    let id = generateReadableId()
    let attempts = 0
    while (this.delegations.has(id) && attempts < 10) {
      id = generateReadableId()
      attempts++
    }
    if (this.delegations.has(id)) {
      throw new Error("Failed to generate unique delegation ID")
    }

    // Validate agent exists
    const agentsResult = await this.client.app.agents({})
    const agents = (agentsResult.data ?? []) as { name: string; description?: string; mode?: string }[]
    const validAgent = agents.find((a) => a.name === input.agent)

    if (!validAgent) {
      const available = agents
        .filter((a) => a.mode === "subagent" || a.mode === "all" || !a.mode)
        .map((a) => `• ${a.name}${a.description ? ` - ${a.description}` : ""}`)
        .join("\n")
      throw new Error(`Agent "${input.agent}" not found.\n\nAvailable agents:\n${available || "(none)"}`)
    }

    // Create isolated session
    const sessionResult = await this.client.session.create({
      body: {
        title: `Delegation: ${id}`,
        parentID: input.parentSessionID,
      },
    })

    if (!sessionResult.data?.id) {
      throw new Error("Failed to create delegation session")
    }

    const delegation: Delegation = {
      id,
      sessionID: sessionResult.data.id,
      parentSessionID: input.parentSessionID,
      parentMessageID: input.parentMessageID,
      parentAgent: input.parentAgent,
      prompt: input.prompt,
      agent: input.agent,
      status: "running",
      startedAt: new Date(),
      progress: { toolCalls: 0, lastUpdate: new Date() },
      // Generate title immediately from prompt (no small_model)
      title: generateTitle(input.prompt),
    }

    this.delegations.set(delegation.id, delegation)

    // Track for batched notification
    const parentId = input.parentSessionID
    if (!this.pendingByParent.has(parentId)) {
      this.pendingByParent.set(parentId, new Set())
    }
    this.pendingByParent.get(parentId)?.add(delegation.id)

    // Timeout handler
    setTimeout(() => {
      const current = this.delegations.get(delegation.id)
      if (current && current.status === "running") {
        this.handleTimeout(delegation.id)
      }
    }, MAX_RUN_TIME_MS + 5000)

    // Ensure directory exists
    await this.ensureDelegationsDir(input.parentSessionID)

    // Resolve agent's configured model
    const agentModel = await this.resolveAgentModel(input.agent)

    // Fire the prompt (anti-recursion: disable nested delegations)
    this.client.session
      .prompt({
        path: { id: delegation.sessionID },
        body: {
          agent: input.agent,
          ...(agentModel && { model: agentModel }),
          parts: [{ type: "text", text: input.prompt }],
          tools: {
            task: false,
            delegate: false,
            todowrite: false,
            plan_save: false,
          },
        },
      })
      .catch((error: Error) => {
        delegation.status = "error"
        delegation.error = error.message
        delegation.completedAt = new Date()
        this.persistOutput(delegation, `Error: ${error.message}`)
        this.notifyParent(delegation)
      })

    return delegation
  }

  /**
   * Resolve the model configured for an agent.
   */
  private async resolveAgentModel(
    agentName: string,
  ): Promise<{ providerID: string; modelID: string } | undefined> {
    try {
      const config = await this.client.config.get()
      const configData = config.data as { agent?: Record<string, { model?: string }> } | undefined
      const modelStr = configData?.agent?.[agentName]?.model
      if (!modelStr) return undefined

      const slashIndex = modelStr.indexOf("/")
      if (slashIndex === -1) return undefined

      return {
        providerID: modelStr.substring(0, slashIndex),
        modelID: modelStr.substring(slashIndex + 1),
      }
    } catch {
      return undefined
    }
  }

  /**
   * Handle delegation timeout.
   */
  private async handleTimeout(delegationId: string): Promise<void> {
    const delegation = this.delegations.get(delegationId)
    if (!delegation || delegation.status !== "running") return

    delegation.status = "timeout"
    delegation.completedAt = new Date()
    delegation.error = `Delegation timed out after ${MAX_RUN_TIME_MS / 1000}s`

    try {
      await this.client.session.delete({ path: { id: delegation.sessionID } })
    } catch {}

    const { publicSummary, rawResult } = await this.getResult(delegation)
    delegation.publicSummary = publicSummary
    delegation.rawResult = rawResult
    delegation.result = rawResult
    await this.persistOutput(delegation, `${rawResult}\n\n[TIMEOUT REACHED]`)
    await this.notifyParent(delegation)
  }

  /**
   * Handle session.idle event.
   */
  async handleSessionIdle(sessionID: string): Promise<void> {
    const delegation = this.findBySession(sessionID)
    if (!delegation || delegation.status !== "running") return

    delegation.status = "complete"
    delegation.completedAt = new Date()

    const { publicSummary, rawResult } = await this.getResult(delegation)
    delegation.publicSummary = publicSummary
    delegation.rawResult = rawResult
    delegation.result = rawResult // Backward compatibility
    delegation.description = generateDescription(publicSummary)

    await this.persistOutput(delegation, rawResult)
    await this.notifyParent(delegation)
  }

  /**
   * Get result from delegation's session.
   * Returns a two-tier result: publicSummary (cleaned) and rawResult (full).
   */
  private async getResult(delegation: Delegation): Promise<DelegationResult> {
    try {
      const messages = await this.client.session.messages({
        path: { id: delegation.sessionID },
      })

      const messageData = messages.data as SessionMessageItem[] | undefined
      if (!messageData || messageData.length === 0) {
        return {
          rawResult: "Delegation completed but produced no output.",
          publicSummary: "Task completed — no structured result produced",
        }
      }

      const assistantMessages = messageData.filter(
        (m): m is AssistantSessionMessageItem => m.info.role === "assistant"
      )

      if (assistantMessages.length === 0) {
        return {
          rawResult: "Delegation completed but produced no assistant response.",
          publicSummary: "Task completed — no structured result produced",
        }
      }

      // Raw result: all assistant messages' text parts joined (full output)
      const allTextParts: string[] = []
      for (const msg of assistantMessages) {
        const textParts = msg.parts.filter((p): p is TextPart => p.type === "text")
        for (const p of textParts) {
          allTextParts.push(p.text)
        }
      }
      const rawResult = allTextParts.join("\n")

      // Public summary: sanitize the last assistant message only (the actual answer)
      const lastMessage = assistantMessages[assistantMessages.length - 1]
      const lastTextParts = lastMessage.parts.filter((p): p is TextPart => p.type === "text")
      const lastText = lastTextParts.map((p) => p.text).join("\n")

      if (!lastText.trim()) {
        return {
          rawResult: rawResult || "Delegation completed but produced no text content.",
          publicSummary: "Task completed — no structured result produced",
        }
      }

      const publicSummary = sanitizePublicSummary(lastText)

      return { publicSummary, rawResult }
    } catch (error) {
      const fallback = `Delegation completed but result could not be retrieved: ${
        error instanceof Error ? error.message : "Unknown error"
      }`
      return {
        rawResult: fallback,
        publicSummary: "Task completed — result could not be retrieved",
      }
    }
  }

  /**
   * Persist delegation output to disk.
   */
  private async persistOutput(delegation: Delegation, content: string): Promise<void> {
    try {
      const dir = await this.ensureDelegationsDir(delegation.parentSessionID)
      const filePath = path.join(dir, `${delegation.id}.md`)

      const header = `# ${delegation.title || delegation.id}

${delegation.description || "(No description)"}

**ID:** ${delegation.id}
**Agent:** ${delegation.agent}
**Status:** ${delegation.status}
**Started:** ${delegation.startedAt.toISOString()}
**Completed:** ${delegation.completedAt?.toISOString() || "N/A"}

---

`
      await fs.writeFile(filePath, header + content, "utf8")
    } catch (error) {
      this.log.warn(`Failed to persist output: ${error instanceof Error ? error.message : "Unknown"}`)
    }
  }

  /**
   * Notify parent session (batched).
   * Uses publicSummary (clean, concise) instead of raw full output.
   */
  private async notifyParent(delegation: Delegation): Promise<void> {
    try {
      const pendingSet = this.pendingByParent.get(delegation.parentSessionID)
      if (pendingSet) {
        pendingSet.delete(delegation.id)
      }

      const allComplete = !pendingSet || pendingSet.size === 0
      if (allComplete && pendingSet) {
        this.pendingByParent.delete(delegation.parentSessionID)
      }

      // Use publicSummary for notification — clean, concise, no raw reasoning noise
      const summary = delegation.publicSummary || delegation.error || "(No result)"

      // Send completion notification with clean public summary
      const notification = `[TASK NOTIFICATION]
ID: ${delegation.id}
Status: ${delegation.status}
Agent: ${delegation.title || delegation.id}${delegation.error ? `\nError: ${delegation.error}` : ""}

Summary:

${summary}

> Use \`delegation_read("${delegation.id}")\` to retrieve full output.`

      await this.client.session.prompt({
        path: { id: delegation.parentSessionID },
        body: {
          noReply: true,
          agent: delegation.parentAgent,
          parts: [{ type: "text", text: notification }],
        },
      })

      // If all complete, trigger response
      if (allComplete) {
        await this.client.session.prompt({
          path: { id: delegation.parentSessionID },
          body: {
            noReply: false,
            agent: delegation.parentAgent,
            parts: [{ type: "text", text: "[TASK NOTIFICATION] All delegations complete." }],
          },
        })
      }
    } catch (error) {
      this.log.warn(`Failed to notify parent: ${error instanceof Error ? error.message : "Unknown"}`)
    }
  }

  /**
   * Read delegation output (blocks if running).
   */
  async readOutput(sessionID: string, id: string): Promise<string> {
    // Try file first
    try {
      const dir = await this.getDelegationsDir(sessionID)
      const filePath = path.join(dir, `${id}.md`)
      return await fs.readFile(filePath, "utf8")
    } catch {}

    // Check if running
    const delegation = this.delegations.get(id)
    if (delegation?.status === "running") {
      // Wait for completion
      const startTime = Date.now()
      while (delegation.status === "running" && Date.now() - startTime < MAX_RUN_TIME_MS + 10000) {
        await new Promise((r) => setTimeout(r, 1000))
      }

      // Try file again
      try {
        const dir = await this.getDelegationsDir(sessionID)
        const filePath = path.join(dir, `${id}.md`)
        return await fs.readFile(filePath, "utf8")
      } catch {}

      // Return status if still no file
      if (delegation.status !== "running") {
        return `Delegation "${delegation.title || delegation.id}" ended with status: ${delegation.status}. ${delegation.error || ""}`
      }
    }

    throw new Error(
      `Delegation "${id}" not found.\n\nUse delegation_list() to see available delegations.`
    )
  }

  /**
   * List all delegations for a session.
   */
  async listDelegations(sessionID: string): Promise<DelegationListItem[]> {
    const results: DelegationListItem[] = []

    // Add in-memory delegations
    for (const delegation of this.delegations.values()) {
      results.push({
        id: delegation.id,
        status: delegation.status,
        title: delegation.title || "(generating...)",
        description: delegation.description || "(generating...)",
      })
    }

    // Add persisted delegations
    try {
      const dir = await this.getDelegationsDir(sessionID)
      const files = await fs.readdir(dir)

      for (const file of files) {
        if (file.endsWith(".md")) {
          const id = file.replace(".md", "")
          if (!results.find((r) => r.id === id)) {
            let title = "(loaded from storage)"
            let description = ""
            let agent: string | undefined

            try {
              const content = await fs.readFile(path.join(dir, file), "utf8")
              const titleMatch = content.match(/^# (.+)$/m)
              if (titleMatch) title = titleMatch[1]
              const agentMatch = content.match(/^\*\*Agent:\*\* (.+)$/m)
              if (agentMatch) agent = agentMatch[1]
              const lines = content.split("\n")
              if (lines.length > 2 && lines[2]) {
                description = lines[2].slice(0, 150)
              }
            } catch {}

            results.push({ id, status: "complete", title, description, agent })
          }
        }
      }
    } catch {}

    return results
  }

  findBySession(sessionID: string): Delegation | undefined {
    return Array.from(this.delegations.values()).find((d) => d.sessionID === sessionID)
  }

  getPendingCount(parentSessionID: string): number {
    return this.pendingByParent.get(parentSessionID)?.size ?? 0
  }

  getRunningDelegations(): Delegation[] {
    return Array.from(this.delegations.values()).filter((d) => d.status === "running")
  }

  handleMessageEvent(sessionID: string, messageText?: string): void {
    const delegation = this.findBySession(sessionID)
    if (!delegation || delegation.status !== "running") return

    delegation.progress.lastUpdate = new Date()
    if (messageText) {
      delegation.progress.lastMessage = messageText
      delegation.progress.lastMessageAt = new Date()
    }
  }
}

// ─── Tool Creators ───────────────────────────────────────────────────────────

interface DelegateArgs {
  prompt: string
  agent: string
}

function createDelegate(manager: DelegationManager): ReturnType<typeof tool> {
  return tool({
    description: `Delegate a task to an agent. Returns immediately with a readable ID.

Use this for:
- Research tasks (will be auto-saved)
- Parallel work that can run in background
- Any task where you want persistent, retrievable output

Hybrid UX: On completion, you receive a clean, concise summary in the notification.
The full raw output (including reasoning, tool calls) is persisted to disk.
Use \`delegation_read(id)\` only when you need the complete output.
Results survive compaction.`,
    args: {
      prompt: tool.schema
        .string()
        .describe("The full detailed prompt for the agent. Must be in English."),
      agent: tool.schema
        .string()
        .describe(
          'Agent to delegate to. Available agents: "sdd-init", "sdd-explore", "sdd-propose", "sdd-spec", "sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive", "sdd-plan", "remote-exec", or "general" for generic tasks.',
        ),
    },
    async execute(args: DelegateArgs, toolCtx: ToolContext): Promise<string> {
      if (!toolCtx?.sessionID) {
        return "❌ delegate requires sessionID. This is a system error."
      }
      if (!toolCtx?.messageID) {
        return "❌ delegate requires messageID. This is a system error."
      }

      try {
        const delegation = await manager.delegate({
          parentSessionID: toolCtx.sessionID,
          parentMessageID: toolCtx.messageID,
          parentAgent: toolCtx.agent,
          prompt: args.prompt,
          agent: args.agent,
        })

        const totalActive = manager.getPendingCount(toolCtx.sessionID)
        let response = `Delegation started: ${delegation.id}\nAgent: ${args.agent}`
        if (totalActive > 1) {
          response += `\n\n${totalActive} delegations now active.`
        }
        response += `\nYou WILL be notified when ${totalActive > 1 ? "ALL complete" : "complete"}. Do NOT poll.`

        return response
      } catch (error) {
        return `❌ Delegation failed:\n\n${error instanceof Error ? error.message : "Unknown error"}`
      }
    },
  })
}

function createDelegationRead(manager: DelegationManager): ReturnType<typeof tool> {
  return tool({
    description: `Read the full output of a delegation by its ID.
Returns the complete, unfiltered result including all reasoning chains and tool-call details.
Notifications only show a concise summary — use this to access the full output.`,
    args: {
      id: tool.schema.string().describe("The delegation ID (e.g., 'swift-falcon')"),
    },
    async execute(args: { id: string }, toolCtx: ToolContext): Promise<string> {
      if (!toolCtx?.sessionID) {
        return "❌ delegation_read requires sessionID. This is a system error."
      }

      // Guard: delegation IDs are human-readable (e.g., "swift-falcon"), NOT session IDs (ses_*)
      if (args.id.startsWith("ses_")) {
        return [
          "❌ You passed a session ID (`ses_*`) to `delegation_read`.",
          "",
          "Session IDs come from the sync `task` tool, which returns results inline.",
          "You already have the result — it was in the task tool's response.",
          "",
          "`delegation_read` only accepts delegation IDs like `swift-falcon`,",
          "which come from the async `delegate` tool.",
          "",
          "Do NOT call delegation_read for task results. Use the inline response directly."
        ].join("\n");
      }

      return await manager.readOutput(toolCtx.sessionID, args.id)
    },
  })
}

function createDelegationList(manager: DelegationManager): ReturnType<typeof tool> {
  return tool({
    description: `List all delegations for the current session.
Shows both running and completed delegations.`,
    args: {},
    async execute(_args: Record<string, never>, toolCtx: ToolContext): Promise<string> {
      if (!toolCtx?.sessionID) {
        return "❌ delegation_list requires sessionID. This is a system error."
      }

      const delegations = await manager.listDelegations(toolCtx.sessionID)
      if (delegations.length === 0) {
        return "No delegations found for this session."
      }

      const lines = delegations.map((d) => {
        const titlePart = d.title ? ` | ${d.title}` : ""
        const descPart = d.description ? `\n  → ${d.description}` : ""
        return `- **${d.id}**${titlePart} [${d.status}]${descPart}`
      })

      // Add stale data warning if there are completed delegations from disk
      const hasCompleted = delegations.some(item => item.status === "complete")
      if (hasCompleted) {
        lines.push("")
        lines.push("> ⚠️ Completed delegations may include results from previous sessions.")
        lines.push("> Verify the ID matches your current delegation before relying on the result.")
      }

      return `## Delegations\n\n${lines.join("\n")}`
    },
  })
}

// ─── Delegation Rules (injected into system prompt) ──────────────────────────

const DELEGATION_RULES = `<task-notification>
<delegation-system>

## Async Background Delegation

You have tools for parallel background work:
- \`delegate(prompt, agent)\` - Launch background task, returns ID immediately
- \`delegation_read(id)\` - Retrieve full completed result
- \`delegation_list()\` - List delegations (use sparingly)

## Hybrid UX: Clean Summaries + On-Demand Full Output

Delegation notifications use a **two-tier output model**:
1. **Public summary** — A concise, sanitized notification arrives in the parent thread. This strips internal reasoning chains, thinking blocks, and tool-call noise.
2. **Full raw output** — The complete, unfiltered result is persisted to disk. Use \`delegation_read(id)\` to retrieve it when you need the full detail.

Notifications include the delegation ID so you can retrieve the full output at any time.

## When to Use delegate vs task

| Tool | Behavior | Use When |
|------|----------|----------|
| \`delegate\` | Async, background, persisted to disk | You want to continue working while it runs |
| \`task\` | Synchronous, blocks until complete | You need the result before continuing |

Any agent can be used with \`delegate\`. Results survive context compaction.

## How It Works

1. Call \`delegate(prompt, agent)\` with a detailed prompt and agent name
2. Continue productive work while it runs in the background
3. Receive a \`<task-notification>\` with a **clean summary** when complete
4. If you need the full output (reasoning, tool calls, everything), use \`delegation_read(id)\`

## Critical Constraints

**NEVER poll \`delegation_list\` to check completion.**
You WILL be notified via \`<task-notification>\`. Polling wastes tokens.

**The \`task\` tool returns results INLINE — you already have them. Do NOT use \`delegation_read\` for task results.**

**NEVER wait idle.** Always have productive work while delegations run.

**NOTE:** Background delegations run in isolated sessions. Changes made by write-capable
agents in background sessions are NOT tracked by OpenCode's undo/branching system.

</delegation-system>
</task-notification>`

// ─── Compaction Context Formatting ───────────────────────────────────────────

interface DelegationForContext {
  id: string
  agent?: string
  title?: string
  description?: string
  status: string
  startedAt?: Date
  prompt?: string
}

function formatDelegationContext(
  running: DelegationForContext[],
  completed: DelegationForContext[],
): string {
  const sections: string[] = ["<delegation-context>"]

  if (running.length > 0) {
    sections.push("## Running Delegations", "")
    for (const d of running) {
      sections.push(`### \`${d.id}\`${d.agent ? ` (${d.agent})` : ""}`)
      if (d.startedAt) sections.push(`**Started:** ${d.startedAt.toISOString()}`)
      if (d.prompt) {
        const truncated = d.prompt.length > 200 ? `${d.prompt.slice(0, 200)}...` : d.prompt
        sections.push(`**Prompt:** ${truncated}`)
      }
      sections.push("")
    }
    sections.push(
      "> **Note:** You WILL be notified via a **Task Notification** blockquote when delegations complete.",
      "> Do NOT poll `delegation_list` - continue productive work.",
      ""
    )
  }

  if (completed.length > 0) {
    sections.push("## Recent Completed Delegations", "")
    for (const d of completed) {
      sections.push(`- \`${d.id}\` [${d.status}]`)
    }
    sections.push("", "> Use `delegation_read(id)` to get full output.", "")
  }

  sections.push("## Retrieval")
  sections.push('Use `delegation_read("id")` to access full delegation output.')
  sections.push("</delegation-context>")

  return sections.join("\n")
}

// ─── Plugin Export ───────────────────────────────────────────────────────────

interface SystemTransformInput {
  agent?: string
  sessionID?: string
}

export const BackgroundAgents: Plugin = async (ctx) => {
  const { client, directory } = ctx
  const log = createLogger(client as OpencodeClient)

  // Project-level storage directory
  const projectId = await getProjectId(directory)
  const baseDir = path.join(os.homedir(), ".local", "share", "opencode", "delegations", projectId)
  await fs.mkdir(baseDir, { recursive: true })

  const manager = new DelegationManager(client as OpencodeClient, baseDir, log)

  return {
    // ─── Tools ─────────────────────────────────────────────────────────────

    tool: {
      delegate: createDelegate(manager),
      delegation_read: createDelegationRead(manager),
      delegation_list: createDelegationList(manager),
    },

    // ─── System Prompt Injection ───────────────────────────────────────────

    "experimental.chat.system.transform": async (_input: SystemTransformInput, output) => {
      output.system.push(DELEGATION_RULES)
    },

    // ─── Compaction Hook ───────────────────────────────────────────────────
    // CRITICAL: Inject delegation context for context recovery after compaction

    "experimental.session.compacting": async (
      input: { sessionID: string },
      output: { context: string[]; prompt?: string },
    ) => {
      const rootSessionID = await manager.getRootSessionID(input.sessionID)

      const running = manager
        .getRunningDelegations()
        .filter((d) => d.parentSessionID === input.sessionID || d.parentSessionID === rootSessionID)
        .map((d) => ({
          id: d.id,
          agent: d.agent,
          title: d.title,
          description: d.description,
          status: d.status,
          startedAt: d.startedAt,
          prompt: d.prompt,
        }))

      const allDelegations = await manager.listDelegations(input.sessionID)
      const completed = allDelegations
        .filter((d) => d.status !== "running")
        .slice(-10)
        .map((d) => ({
          id: d.id,
          agent: d.agent,
          title: d.title,
          description: d.description,
          status: d.status,
        }))

      if (running.length === 0 && completed.length === 0) return

      output.context.push(formatDelegationContext(running, completed))
    },

    // ─── Event Hook ────────────────────────────────────────────────────────

    event: async ({ event }: { event: Event }): Promise<void> => {
      // Session idle = delegation complete
      if (event.type === "session.idle") {
        const sessionID = event.properties.sessionID
        if (manager.findBySession(sessionID)) {
          await manager.handleSessionIdle(sessionID)
        }
      }

      // Track message progress
      if (event.type === "message.updated") {
        const sessionID = event.properties.info.sessionID
        if (sessionID) {
          manager.handleMessageEvent(sessionID)
        }
      }
    },
  }
}

export default BackgroundAgents
