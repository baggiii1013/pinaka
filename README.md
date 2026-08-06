<p align="center">
  <h1 align="center">pinakatype.sh</h1>
  <p align="center">
    A minimalistic, SSH-native typing test. Inspired by <a href="https://monkeytype.com">Monkeytype</a>, built for the terminal.
  </p>
</p>

<p align="center">
  <a href="#features">Features</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#how-it-works">How It Works</a> &bull;
  <a href="#keybindings">Keybindings</a> &bull;
  <a href="#deployment">Deployment</a>
</p>

---

```
ssh pinakatype.sh
```

No install. No browser. No account. Open your terminal and start typing.

---

## Features

**Test Modes**

| Mode | Options | Description |
|------|---------|-------------|
| Time | 15 / 30 / 60 / 120s | Race against the clock with unlimited words |
| Words | 10 / 25 / 50 / 100 | Complete a fixed set of words as fast as you can |
| Quote | - | Type real quotes and passages verbatim |
| Zen | - | Freeform mode with no end condition -- just type |

**423 Word Lists & Languages** -- English, Spanish, French, German, Chinese, Japanese, Korean, Hindi, Arabic, Russian, and 200+ more. Includes 69 programming language vocabularies: Go, Python, Rust, JavaScript, TypeScript, C, C++, Java, Ruby, Bash, SQL, Dart, Elixir, Haskell, and many others.

**Live Feedback** -- Real-time WPM, accuracy percentage, and countdown/word-progress tracking update as you type. Correct characters light up, errors turn red, and the cursor follows your position.

**Results Dashboard** -- After each test, see a full breakdown:
- **WPM** -- net words per minute (correct chars only, 5 chars = 1 word)
- **Raw WPM** -- gross speed including errors
- **Accuracy** -- percentage of correctly typed characters
- **Consistency** -- coefficient-of-variation metric showing how steady your speed was
- **Character stats** -- correct / incorrect / extra / missed counts
- **Sparkline chart** -- per-second WPM history rendered as a Unicode block chart

**Language Picker** -- searchable modal with three tabs (Popular, Code, All). Fuzzy-search across every dictionary. Switch without restarting your test.

**Toggles** -- add punctuation marks and/or number insertion on any word list for harder tests.

**SSH-First Identity** -- your SSH public key is your identity. Connect from any machine, anywhere. No passwords, no sign-ups.

## Quick Start

**Try it now** (public server):

```bash
ssh pinakatype.sh
```

**Run locally**:

```bash
# Clone
git clone https://github.com/yourusername/pinakatype.git
cd pinakatype

# Run (defaults to port 2222)
go run .

# Or specify a custom port
PORT=3333 go run .

# Connect from another terminal
ssh localhost -p 2222
```

### Prerequisites

- **Go 1.26+**
- **PostgreSQL** -- for user profiles, test history, and leaderboards

## How It Works

```
┌──────────┐       SSH        ┌──────────────┐      ┌───────────────┐
│ Terminal  │ ──────────────>  │  Wish Server │ ───> │  Bubble Tea   │
│ (client)  │  ed25519 auth   │  (port 2222) │      │  TUI Session  │
└──────────┘                  └──────────────┘      └───────┬───────┘
                                                            │
                                                    ┌───────┴───────┐
                                                    │ Typing Engine │
                                                    │  state.go     │
                                                    │  - WPM calc   │
                                                    │  - accuracy   │
                                                    │  - consistency│
                                                    │  - char stats │
                                                    └───────────────┘
```

When you SSH in, the [Wish](https://github.com/charmbracelet/wish) server spawns a dedicated [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI session for your connection. The typing engine tracks every keystroke in real time and calculates WPM using the standard 5-characters-per-word metric. Tests start on the first keypress and end when time runs out or all words are completed. In time and zen modes, new words are generated on the fly as you exhaust the current set.

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | **Go 1.26** |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) -- Elm architecture for terminals |
| SSH Server | [Wish](https://github.com/charmbracelet/wish) -- SSH apps made easy |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) -- CSS-like terminal styling |
| Word Data | 423 JSON files embedded via `go:embed` (Monkeytype-compatible format) |
| Quotes | 87 quote collections across languages and programming domains |
| Database | PostgreSQL -- user profiles, test history, leaderboards |
| Auth | SSH public key -- seamless, passwordless |

## Keybindings

### Before Test

| Key | Action |
|-----|--------|
| `1` `2` `3` `4` | Switch mode: time / words / quote / zen |
| `Left` / `Right` | Cycle time or word count options |
| `p` | Toggle punctuation |
| `n` | Toggle numbers |
| `l` or `d` | Open language picker |
| `Tab` | Restart test |
| `Esc` | Quit |

### During Test

| Key | Action |
|-----|--------|
| Any character | Type |
| `Backspace` | Delete last character (rewinds to previous word if incorrect) |
| `Space` | Submit current word and advance |
| `Tab` | Restart test |
| `Esc` | Quit |

### Results Screen

| Key | Action |
|-----|--------|
| `Tab` | Restart with same settings |
| `Esc` | Quit |

### Language Picker

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle category tabs |
| `Up` / `Down` | Navigate list |
| Type characters | Fuzzy search |
| `Enter` | Select language |
| `Esc` | Cancel |

## Project Structure

```
pinakatype/
  main.go                 Entry point -- reads PORT env, starts SSH server
  server/
    server.go             Wish SSH server setup, session bridging
  engine/
    state.go              Core typing state machine (WPM, accuracy, consistency)
    mode.go               Test mode definitions and defaults
    words.go              Word generation, quote selection, punctuation/numbers
    loader.go             Embedded asset loader with lazy caching
    loader_test.go        Tests for asset loading
  ui/
    model.go              Bubble Tea model -- Update/View lifecycle
    modeselect.go         Mode selector bar component
    langpicker.go         Language picker modal with tabs and search
    results.go            Results dashboard with sparkline charts
    styles.go             Lip Gloss style definitions (Monkeytype-inspired palette)
  data/
    embed.go              go:embed directives for all JSON assets
    languages/            423 language and word list JSON files
    quotes/               87 quote collection JSON files
  database/
    db.go                 PostgreSQL data access layer
  models/                 Shared data structures
```

## Deployment

The Go binary compiles to a single static executable with all word lists and quotes embedded. Package it in a `scratch` Docker image for a minimal footprint:

```dockerfile
FROM golang:1.26 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o pinakatype .

FROM scratch
COPY --from=build /app/pinakatype /pinakatype
EXPOSE 2222
ENTRYPOINT ["/pinakatype"]
```

Deploy on [Fly.io](https://fly.io), Railway, or any platform that supports TCP/SSH edge routing. The app is stateless per session -- horizontal scaling only requires a TCP load balancer.

## Theme

Built with a Monkeytype-inspired color palette:

| Element | Color | Hex |
|---------|-------|-----|
| Background | Dark charcoal | `#323437` |
| Untyped text | Dim grey | `#737373` |
| Correct text | Warm white | `#d1d0c5` |
| Incorrect text | Subtle red | `#ca4754` |
| Accent (stats, cursor, chart) | Golden yellow | `#e2b714` |

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[MIT](LICENSE)
