# pinakatype.sh: Terminal-Based Typing Test Architecture

## 1. Overview
pinakatype.sh is a minimalistic, fast, and customizable terminal-based typing test inspired by Monkeytype. Unlike traditional CLI applications that require installation, pinakatype.sh is designed to be accessible instantly over SSH (e.g., `ssh pinakatype.sh`). It features real-time WPM (Words Per Minute) tracking, accuracy calculation, configurable test modes, theming, and global leaderboards.

## 2. Tech Stack Definition

*   **Programming Language:** Go (Golang) - chosen for its excellent concurrency, performance, and the robust Charmbracelet ecosystem for CLI/TUI apps.
*   **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The Elm architecture for terminal applications. Handles state, events, and rendering.
*   **SSH Server:** [Wish](https://github.com/charmbracelet/wish) - An SSH server for Go that makes building SSH apps easy. It integrates seamlessly with Bubble Tea.
*   **Styling & UI Components:** [Lip Gloss](https://github.com/charmbracelet/lipgloss) (for styling) and [Bubbles](https://github.com/charmbracelet/bubbles) (for common UI components like text inputs, spinners, lists).
*   **Database:** PostgreSQL - For storing robust user histories and leaderboards concurrently.
*   **Authentication:** SSH Public Key - Users don't need to create accounts with passwords; their SSH key naturally acts as their unique identifier.

## 3. Physical Architecture

```mermaid
graph TD
    User([User Terminal]) -- "ssh pinakatype.sh" --> SSHGateway[SSH Gateway / Wish Server]

    subgraph "Application Server (Go)"
        SSHGateway --> SessionManager[Session Manager]
        SessionManager --> BubbleTeaApp[Bubble Tea App Instance]

        subgraph "TUI Components"
            BubbleTeaApp --> Menu[Main Menu]
            BubbleTeaApp --> TypeEngine[Typing Engine]
            BubbleTeaApp --> LeaderboardUI[Leaderboard View]
            BubbleTeaApp --> ResultUI[Results View]
        End

        BubbleTeaApp --> WordGenerator[Word Generator]
        BubbleTeaApp --> DAL[Data Access Layer]
    End

    DAL -- TCP --> Postgres[(PostgreSQL DB)]

    subgraph "Database Schema"
        Postgres -.-> UsersTable(Users Table)
        Postgres -.-> TestsTable(Test Results Table)
    End
```

## 4. Core System Components

### A. The SSH Gateway (Wish)
Instead of serving HTTP, the application listens on port 22 (or another designated SSH port). When a user connects, `Wish` intercepts the SSH handshake, extracts the user's public key (for authentication), and spins up a dedicated `BubbleTea` application instance for that specific connection.

### B. Typing Engine (State Machine)
The core typing logic uses a state machine driven by keystrokes:
*   **State: Waiting** -> Waiting for the first keystroke to start the timer.
*   **State: Typing** -> Real-time comparison of input array vs target word array. Calculates instantaneous WPM and Accuracy.
*   **State: Finished** -> Timer ends or word count reached. Calculates final stats and dispatches a save event to the database.

### C. Word Generation
A module that loads word lists (e.g., top 200 English words, programming keywords) into memory and generates deterministic or random sequences based on the user's selected configuration (Time-based vs. Word-based).

### D. Theming Engine
Uses `Lip Gloss` to define color palettes. Themes can be selected via a configuration menu and are saved to the user's profile in the database.

## 5. Database Schema

**`users` table**
*   `id` (UUID, Primary Key)
*   `ssh_pub_key` (String, Unique) - Used for seamless auth.
*   `username` (String, Unique) - User-claimed display name.
*   `active_theme` (String) - Currently selected visual theme.
*   `created_at` (Timestamp)

**`test_results` table**
*   `id` (UUID, Primary Key)
*   `user_id` (UUID, Foreign Key)
*   `wpm` (Decimal) - Final Words Per Minute.
*   `accuracy` (Decimal) - Percentage (0-100).
*   `mode` (String) - e.g., "time_15", "word_50".
*   `raw_wpm` (Decimal) - WPM including errors.
*   `test_date` (Timestamp)

## 6. Access and User Flow

1.  **Connection:** User types `ssh pinakatype.sh` in their terminal.
2.  **Auth:** Server accepts the connection, identifies the user via their local SSH key automatically.
3.  **UI Render:** Present the main view: typing area in the center, configuration (time/words) at the bottom.
4.  **Interaction:** User starts typing. Timer starts on the first keypress. Errors are highlighted in red (customizable theme color), correct letters in dim white/green.
5.  **Completion:** Test completes. Results (WPM, Accuracy) are shown. Data is asynchronously saved to PostgreSQL.
6.  **Leaderboard:** User presses `tab` to switch to the global global leaderboard view.

## 7. Deployment Strategy

*   **Hosting:** Fly.io or DigitalOcean Droplet. Fly.io is particularly well-suited for Go-based TCP/SSH edge routing.
*   **Containerization:** The Go binary is packaged in a minimal Docker scratch image to ensure a tiny deployment footprint.
*   **Scaling:** The app is stateless (state belongs to the specific SSH session instance in memory). Horizontal scaling simply requires a TCP load balancer in front of multiple Go application instances.