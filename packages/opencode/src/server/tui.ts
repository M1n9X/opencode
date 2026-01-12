import { Hono, type Context } from "hono"
import { describeRoute, resolver, validator } from "hono-openapi"
import { z } from "zod"
import { AsyncQueue } from "../util/queue"

const TuiRequestSchema = z.object({
  path: z.string(),
  body: z.unknown(),
})

type TuiRequest = z.infer<typeof TuiRequestSchema>
type TuiResponse = unknown

const reqQueue = new AsyncQueue<TuiRequest>()
const resQueue = new AsyncQueue<TuiResponse>()

export async function callTui(ctx: Context) {
  const body = await ctx.req.json()
  reqQueue.push({
    path: ctx.req.path,
    body,
  })
  return resQueue.next()
}

export const TuiRoute = new Hono()
  .get(
    "/next",
    describeRoute({
      summary: "Get next TUI request",
      description: "Retrieve the next TUI (Terminal User Interface) request from the queue for processing.",
      operationId: "tui.control.next",
      responses: {
        200: {
          description: "Next TUI request",
          content: {
            "application/json": {
              schema: resolver(TuiRequestSchema),
            },
          },
        },
      },
    }),
    async (c) => {
      const req = await reqQueue.next()
      return c.json(req)
    },
  )
  .post(
    "/response",
    describeRoute({
      summary: "Submit TUI response",
      description: "Submit a response to the TUI request queue to complete a pending request.",
      operationId: "tui.control.response",
      responses: {
        200: {
          description: "Response submitted successfully",
          content: {
            "application/json": {
              schema: resolver(z.boolean()),
            },
          },
        },
      },
    }),
    validator("json", z.unknown()),
    async (c) => {
      const body = c.req.valid("json")
      resQueue.push(body)
      return c.json(true)
    },
  )
