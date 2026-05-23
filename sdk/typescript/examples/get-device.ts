import { WlteClient } from '../src/index.js'

const client = new WlteClient({
  clientId: process.env.WLTE_CLIENT_ID!,
  clientSecret: process.env.WLTE_CLIENT_SECRET!,
  baseUrl: process.env.WLTE_BASE_URL,
})

const device = await client.devices.get(process.env.WLTE_DEVICE_ID!)
console.log(device)

