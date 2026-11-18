# Ferry

A small Go app that shares a folder or a local website from one computer to any other device's browser (same Wi-Fi, or anywhere via a Cloudflare tunnel), keeps a folder in sync from another Ferry computer, and transfers files with FTP/FTPS/SFTP servers. One binary, no account, Windows and Mac first (Linux builds too). Owner: S1xinch.

## Commands

```
go run .                 # open the app window (no args = app mode)
go run . files ~/Pics    # terminal version, see `go run . -h`
go vet ./... && go test -race ./...
GOOS=windows go vet ./... && GOOS=darwin go vet ./...   # always check both, there is per-OS code
scripts/build.sh [windows] [mac] [linux]                # release files into dist/
```

CI (`.github/workflows/ci.yml`) runs the tests on Ubuntu, Windows and macOS on every push to main.

## How it fits together

- `main.go`: no args (or macOS `-psn_`) runs the app; otherwise the terminal version (`parseArgs`, `runCLI`).
- `app.go`: the app. A local HTTP server on 127.0.0.1:47615 (or a random port) serves the window, guarded by a token cookie plus an `X-Ferry: 1` header (`guard`, `api`). `routes()` lists every endpoint. `state()` is polled by the window every 1.5s. Settings live in `settings.json` under `configDir()`.
- `panel.go`: the whole app window as one Go template (HTML, CSS, JS inline, no build step). Full-height layout: header with the pages, a list on the left and a 380px detail `aside` on the right, status `footer`. `body[data-page]` (`share`, `receive`, `transfer`) picks the page via `go()`. Settings, Connect, the trust prompt and `ask()` (in-window confirm/prompt) are `<dialog>`s; `ask()` settles on its buttons and the `cancel` event, not `close`, which Chrome does not always fire on reuse. `toast()` for short messages, `offline()` for the "Ferry is not running" overlay. `ui.go` has the shared `head` template with the Calm Blue colour tokens (copied from the Inventory app: blue accent, red/amber only for problems, always with a word and a shape), dark mode, and the visitor pages.
- `window_*.go`: the app window is native so the taskbar and Dock show Ferry. Windows: WebView2 via `jchv/go-webview2` (pure Go, no cgo), data in `configDir()/webview`, icon set from the exe's "APP" resource. Mac: WKWebView in Objective-C through cgo (`window_darwin.go`, export in `window_darwin_export.go`), so the Mac build in `scripts/build.sh` uses `CGO_ENABLED=1` with `clang -arch`; closing the window keeps Ferry running, the Dock icon reopens it, Cmd+Q goes through `stopApp`. Linux and non-cgo Mac builds use `window_other.go` and fall back. A second launch calls `GET /api/show?k=<token>` on the running copy. Links with `target=_blank` go to `POST /api/open` so they open in the real browser.
- `platform.go`: the fallback, a Chrome/Edge `--app` window or the browser. Do NOT call `hideWindow` on the browser launch (that caused the blank white window on Windows).
- `share.go` / `files.go` / `site.go` / `auth.go` / `network.go`: what visitors see. PIN login with lockouts, folder listing, uploads (never overwrite), zip download, reverse proxy for a website, LAN addresses, Cloudflare tunnel.
- `discover.go` (+ `broadcast_*.go`): UDP "who is sharing?" on port 47616. Asks every address on each local network (capped to a /22) plus broadcast, twice. Mac needs `NSLocalNetworkUsageDescription` in `packaging/Info.plist`.
- `mirror.go`: receive mode, keeps a local folder the same as another Ferry's shared folder.
- `update.go`: checks GitHub releases on open and every 6 hours, downloads in the background, swaps itself on "Restart to update". Mac: refuses up front when running from App Translocation or an unwritable folder; the old bundle is left as `Ferry.app.old` and cleaned next launch.
- Transfer files (FileZilla-style):
  - `remote.go`: `remoteFS` interface with `ftpFS` (jlaffaye/ftp: plain, explicit TLS, implicit TLS) and `sftpFS` (pkg/sftp + x/crypto/ssh, password, keyboard-interactive, agent, default keys). Trust-on-first-use for SSH host keys and TLS certificates (`trustStore`, `trusted-servers.json`), `~/.ssh/known_hosts` honoured, changed keys refused. `safeName` guards every name coming from a server.
  - `transfer.go`: one-at-a-time queue on its own connection; folders expand into children when their turn comes; downloads write `.<name>.ferry-part` then rename; cancel via context.
  - `remote_api.go`: the HTTP handlers, browsing session (reconnects once if dropped), recent connections (`connections.json`, last 10, never passwords), local file listing (`:drives` on Windows).

## Conventions

- Copy in the UI, errors and README is plain, friendly English for non-technical people. No em dashes. No emoji icons, no pill-shaped buttons, no gradients. No browser `confirm()`/`prompt()`: use `ask()`.
- Buttons are outlined by default; `primary` is the one filled main action, `danger` is red text, `small` for toolbars, `on` for the selected page or choice.
- Errors returned to the window are lowercase sentences; the JS capitalises the first letter.
- Keep `go.mod` on Go 1.24. Newer `golang.org/x/crypto` bumps the go directive, so pin with `GOTOOLCHAIN=local go get ...@version` (crypto is on v0.48.0).
- Tests that open files must close them before `t.TempDir` cleanup, or Windows CI fails. Skip tests that rely on Unix paths on Windows rather than weakening them.
- Commits: short imperative subject, a body explaining why in plain words.

## Releasing

The Release workflow runs when a `v*` tag is pushed: it tests, builds Windows/Linux on Ubuntu and the universal signed Mac app on macOS, then publishes the release with generated notes. The version comes from the tag (`-X main.version`), so no files change for a release.

```
git tag -a v0.2.1 -m "Ferry 0.2.1" && git push origin v0.2.1
```

Creating the release in the GitHub web UI with a new tag also triggers it.

## Current state (2026-11-18)

- v0.1.6 (Transfer files) and v0.2.0 (Calm Blue redesign: Share, Receive and Transfer pages, dialogs, blue icon) are published.
- The app icons in `packaging/icons` were recoloured from the green originals by script; there is no source artwork file.
- Not yet tested against real servers: implicit FTPS (same code path as FTP over TLS, which is tested), and the Windows drives list in the local pane.
- Ideas not done: saving passwords in the OS keychain, resuming interrupted transfers, editing local files (rename/delete) from the Transfer page.

## Testing the window by hand

`go run .` opens it. For a headless check, serve `a.routes()` from a throwaway test on a spare port (the `app.port` must match, the guard checks the Host header) with `token: "secret"`, open `http://127.0.0.1:<port>/?k=secret` in headless Chrome, and drive it with the DevTools protocol. To see states that need a real share, server or second computer, replace `window.fetch` in the page to return a made-up `/api/state` (and `/api/remote/list`, `/api/discover`); starting a real share from a test binary triggers a Windows Firewall prompt. `pyftpdlib` gives quick FTP and FTP-over-TLS servers for manual checks.
