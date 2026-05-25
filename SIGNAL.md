# SIGNAL

> the marketing and outreach brief for this project.
> drop this file into any repo to define the signal it puts out into the world.

---

## product

- **description**: single-binary terminal file manager with dual-pane layout, git badges, and syntax-highlighted previews
- **audience**: terminal-first developers — individual and small teams
- **stage**: launch-ready (v0.8.x — homebrew tap live, demo gif present, docs complete)
- **model**: free and open source

---

## positioning

- **what makes it different**: git integration + zero config out of the box; single binary, no plugins, no rc files
- **for**: developers who want immediate file and git situational awareness in the terminal; small teams where zero-setup matters for consistency across machines
- **not for**: users who need file operations (copy/move/delete); GUI or mouse-driven workflows

---

## channels

four channels. that's it. do these well before adding more.

| channel          | audience                      | priority | rationale                                              |
| :--------------- | :---------------------------- | :------- | :----------------------------------------------------- |
| github           | developers; contributors      | high     | stars = trust signal; issues = usage signal            |
| awesome lists    | developers evaluating tools   | high     | one PR to `awesome-tui` / `awesome-go`; permanent traffic |
| hacker news      | senior engineers; open-source | high     | show hn at launch; highest single-post leverage        |
| reddit           | broad developer community     | high     | r/commandline, r/golang, r/unixporn; sustained reach   |

**later, when the above are done:**

| channel              | priority | note                                         |
| :------------------- | :------- | :------------------------------------------- |
| developer newsletters | medium  | Console.dev, Go Weekly, TLDR; pitch once     |
| x / twitter          | medium   | demo gif; low effort, low ceiling            |
| dev.to               | low      | "building a tui in go" article; long-tail    |

---

## messaging

**tone**: lowercase, no hype, let the tool speak. a good gif beats any copy.

**key themes**:
- single binary, zero config, works on any machine
- see your files and git status together, instantly
- stays out of your way

**phrases to use**: "zero config", "immediate situational awareness", "stays out of your way"

**phrases to avoid**: "blazing fast", "powerful", "intuitive", feature-list dumps without showing the tool

---

## content calendar

| step | channel          | format       | topic / asset                                      | status   | unblocked    |
| :--- | :--------------- | :----------- | :------------------------------------------------- | :------- | :----------- |
| 1    | awesome-tuis     | PR           | PR #687 open at rothgar/awesome-tuis               | done     | now          |
| 2    | hacker news      | show hn post | launch post — git badges + zero config angle       | planned  | now          |
| 3    | reddit           | text post    | r/commandline + r/golang same day as HN            | planned  | now          |
| 4    | awesome-cli-apps | PR           | add to file managers section                       | blocked  | 2026-07-17   |
| 5    | awesome-go       | PR           | add to advanced console UIs or utilities section   | blocked  | 2026-11-18   |
| 6    | newsletters      | pitch email  | Console.dev, Go Weekly — one-liner + gif           | backlog  | after 20 stars |

**blockers (revisit mid-July 2026):**
- `awesome-cli-apps`: requires 90 days old (eligible 2026-07-17) and 20+ stars
- `awesome-go`: requires 5 months old (eligible 2026-09-18) + OSI-approved license (BUSL-1.1 not accepted) + 80% test coverage
- **license note**: BUSL-1.1 blocks most curated lists. switching to MIT or Apache-2.0 removes this permanently.

**milestones to amplify**: first external star, 20 stars (unlocks list submissions), 100 stars, first roundup mention, v1.0

---

## metrics

| channel       | success signal                              | threshold   |
| :------------ | :------------------------------------------ | :---------- |
| github        | stars from non-network accounts             | 100 stars   |
| awesome lists | PR merged; referral traffic in star history | merged      |
| hacker news   | front page; comments from terminal users    | top 10      |
| reddit        | upvotes; constructive comments              | 50 upvotes  |

---

## links

- repo: https://github.com/mirageglobe/scout
- support: https://buymeacoffee.com/mirageglobe
