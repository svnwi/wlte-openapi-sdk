import { WlteClient } from '../src/index.js'

const client = new WlteClient({
  clientId: process.env.WLTE_CLIENT_ID!,
  clientSecret: process.env.WLTE_CLIENT_SECRET!,
  baseUrl: process.env.WLTE_BASE_URL,
})

const command = await client.relays.set(process.env.WLTE_DEVICE_ID!, {
  index: 1,
  on: true,
  idempotencyKey: `example-${Date.now()}`,
})
console.log(command)
