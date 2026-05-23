import { WlteClient } from '../src/index.js'

const client = new WlteClient({
  clientId: process.env.WLTE_CLIENT_ID!,
  clientSecret: process.env.WLTE_CLIENT_SECRET!,
  baseUrl: process.env.WLTE_BASE_URL,
})

const devices = await client.devices.list({ page: 1, pageSize: 50 })
console.log(devices)

