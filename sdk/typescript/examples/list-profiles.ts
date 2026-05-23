import { WlteClient } from '../src/index.js'

const client = new WlteClient({
  clientId: process.env.WLTE_CLIENT_ID!,
  clientSecret: process.env.WLTE_CLIENT_SECRET!,
  baseUrl: process.env.WLTE_BASE_URL,
})

const profiles = await client.profiles.list()
console.log(profiles)

