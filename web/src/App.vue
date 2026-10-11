<script setup lang="ts">
import { ref } from 'vue'
import { useWhipEvents } from './useWhipEvents'
import { playWhip } from './whipSound'
import bloodClawd from '../assets/blood-clawd.svg'

const HIT_MS = 800

// The row of the latest whipper flashes instead of a separate popup, so
// the overlay stays within the leaderboard, and Claude recoils from the
// lash.
const hit = ref<string>()
let hitTimer: ReturnType<typeof setTimeout> | undefined

const { command, leaderboard } = useWhipEvents((whip) => {
  playWhip()

  hit.value = undefined
  clearTimeout(hitTimer)
  // Restart the animation even when the same viewer whips again.
  requestAnimationFrame(() => {
    hit.value = whip.name
    hitTimer = setTimeout(() => {
      hit.value = undefined
    }, HIT_MS)
  })
})
</script>

<template>
  <main class="board">
    <h1>
      <img :src="bloodClawd" alt="" class="emote" :class="{ struck: hit !== undefined }" />
    </h1>

    <TransitionGroup tag="ol" name="row">
      <li v-for="(entry, i) in leaderboard.top" :key="entry.name" :class="{ hit: entry.name === hit }">
        <span class="rank">{{ i + 1 }}</span>
        <span class="name">{{ entry.name }}</span>
        <span class="count">{{ entry.count }}</span>
      </li>
    </TransitionGroup>
    <p v-if="command && leaderboard.top.length === 0" class="empty">
      <span class="command">{{ command }}</span> in chat
    </p>
  </main>
</template>

<style>
:root {
  --bg: #121214;
  --text: #e6e6ea;
  --muted: #9a9aa2;
  --accent: #d97757;
  --hit: rgb(217 119 87 / 0.3);

  /* Everything scales with the browser source size. Ten rows of 1.58em,
     the 2.8em header and 1.6em of padding take about 20.2em, so 4.5vh
     leaves a margin; long names are cut with an ellipsis. */
  font-size: min(4.5vh, 8vw);
  font-family: system-ui, sans-serif;
  color: var(--text);
}

html,
body {
  margin: 0;
  background: var(--bg);
  overflow: hidden;
}
</style>

<style scoped>
.board {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  width: 100vw;
  height: 100vh;
  padding: 0.8em 1em;
  background: var(--bg);
}

h1 {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4em;
  margin: 0 0 0.5em;
  font-size: 1.4em;
  color: var(--accent);
}

.emote {
  transform-origin: 50% 90%;
  height: 2.6em;
  /* Overlap the padding instead of pushing the rows down, so all ten
     still fit; more overlap below sits the emote a bit lower. */
  margin: -0.35em 0 -0.75em;
  image-rendering: pixelated;
}

ol {
  flex: 1;
  margin: 0;
  padding: 0;
  list-style: none;
}

li {
  display: flex;
  align-items: baseline;
  gap: 0.6em;
  line-height: 1.2;
  padding: 0.1em 0.3em;
  border-radius: 0.3em;
}

.emote.struck {
  animation: struck 0.8s ease-out;
}

/* A sharp jolt away from the lash, then a fading sway back. */
@keyframes struck {
  8% {
    transform: translateX(-0.08em) rotate(-14deg) scale(0.92, 1.06);
    filter: brightness(1.5);
  }
  20% {
    transform: translateX(0.04em) rotate(10deg);
    filter: none;
  }
  35% {
    transform: rotate(-7deg);
  }
  50% {
    transform: rotate(5deg);
  }
  65% {
    transform: rotate(-3deg);
  }
  80% {
    transform: rotate(1.5deg);
  }
}

li.hit {
  animation: hit 0.8s ease-out;
}

@keyframes hit {
  from {
    background: var(--hit);
  }
}

.rank {
  flex-shrink: 0;
  width: 1.5em;
  color: var(--accent);
  font-weight: 700;
}

.name {
  flex: 1;
  min-width: 0;
  font-size: 1.15em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.count {
  flex-shrink: 0;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
}

.empty {
  flex: 1;
  margin: 0;
  padding-top: 2em;
  font-size: 1.25em;
  text-align: center;
  color: var(--muted);
}

.command {
  color: #cfcfd4;
  font-weight: 700;
}

.row-move {
  transition: transform 0.4s ease;
}
</style>
