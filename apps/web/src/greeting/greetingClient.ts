// Talks to the hello service's GreetingService, with code generated from
// proto/hello/v1/greeting.proto. Learn more: docs/learn/connect-rpc.md
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { GreetingService } from '../gen/hello/v1/greeting_pb.ts'

// A greeting, as the rest of the web app sees it.
export type Greeting = {
  id: bigint
  message: string
}

// API calls go to the same address as the page itself: the gateway, which
// forwards them to the hello service (see gateway/Caddyfile).
const transport = createConnectTransport({ baseUrl: window.location.origin })
const client = createClient(GreetingService, transport)

// fetchNewestGreeting asks the hello service for the newest greeting.
export async function fetchNewestGreeting(): Promise<Greeting> {
  const response = await client.getGreeting({})
  return { id: response.id, message: response.message }
}
