import { onScopeDispose, ref } from 'vue'

export interface Entry {
  name: string
  count: number
}

export interface Leaderboard {
  total: number
  top: Entry[]
}

export interface State {
  command: string
  leaderboard: Leaderboard
}

export interface Whip {
  name: string
  count: number
  leaderboard: Leaderboard
}

// Subscribes to the backend event stream. EventSource reconnects on its
// own, and the backend resends the state on every connect.
export function useWhipEvents(onWhip: (whip: Whip) => void) {
  const command = ref('')
  const leaderboard = ref<Leaderboard>({ total: 0, top: [] })
  const source = new EventSource('/events')

  source.addEventListener('state', (e) => {
    const state: State = JSON.parse(e.data)
    command.value = state.command
    leaderboard.value = state.leaderboard
  })
  source.addEventListener('whip', (e) => {
    const whip: Whip = JSON.parse(e.data)
    leaderboard.value = whip.leaderboard
    onWhip(whip)
  })

  onScopeDispose(() => {
    source.close()
  })

  return { command, leaderboard }
}
