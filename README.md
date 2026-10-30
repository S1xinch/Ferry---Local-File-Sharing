# Ferry

Share a folder or a website from your computer, and open it in a browser on any other device: a phone, a laptop, a friend's PC. It works on the same Wi-Fi, and from anywhere if you turn that on. It's free, needs no account, and the other device doesn't need to install anything.

<p align="center"><img src="docs/app.png" width="640" alt="The Ferry window while sharing a folder, showing the PIN, link and QR code"></p>

## Download

Get the latest version from the [Releases page](../../releases/latest).

| Computer | File |
| --- | --- |
| Windows | `ferry-windows-x64.exe` (`ferry-windows-arm64.exe` for ARM laptops like Surface Pro X / Snapdragon) |
| Mac (Apple Silicon or Intel) | `ferry-mac.zip` |

### First time on Windows

Windows shows **"Windows protected your PC"** because the app isn't from a paid developer account. Click **More info**, then **Run anyway**.

When you start sharing, Windows asks whether to allow Ferry on networks. Choose **Allow** (private networks), or other devices won't be able to connect.

### First time on Mac

1. Unzip `ferry-mac.zip` and drag **Ferry** into your **Applications** folder.
2. Open it. macOS says it can't check the app for malware. Click **Done**.
3. Open **System Settings > Privacy & Security**, scroll down, and click **Open Anyway** next to Ferry. Confirm with your password.

You only need to do this once. If macOS asks whether Ferry may accept incoming connections, click **Allow**.

## How to use it

1. Open Ferry. Its window opens in your browser.
2. Choose **A folder** or **A website**.
3. Click **Start sharing**.
4. On the other device, scan the QR code, or type the link and the PIN.

Ferry has no Dock or taskbar icon. Close the tab whenever you like, and sharing keeps going. Open Ferry again to get the window back, and use **Stop sharing** or **Quit Ferry** when you're done. If nothing is being shared and the window is closed, Ferry quits by itself after a few minutes.

### What visitors can do

**A folder:** browse it, open or download files, download a whole folder as a zip, and upload files by dragging them onto the page (you can turn uploads off). Uploads never overwrite anything: a second `photo.jpg` is saved as `photo (1).jpg`.

**A website:** use a website running on your computer (for example a dev server on port 3000) as if they were sitting at it. Live reload works too.

### Access from anywhere

Tick **Also allow access from outside this Wi-Fi**. Ferry gives you a second link that works from any network, using a free Cloudflare tunnel. The first time, it downloads Cloudflare's official `cloudflared` helper (about 40 MB) and keeps it for next time. You get a new random link each time you start sharing, and visitors still need the PIN.

## Safety

- Nothing can be opened without the PIN or a QR code. Each time you start sharing you get a new PIN.
- After 5 wrong PINs a device is locked out for a minute. After 30 wrong PINs from anyone, all logins pause for 10 minutes.
- Visitors can't reach anything outside the folder you chose.
- The QR codes log in without the PIN, so only show them to people you trust.
- The Ferry window only opens on your own computer. Other devices can't control it.
- On the same Wi-Fi the connection isn't encrypted, so anyone on that network could in theory see what's being sent. Use it on networks you trust. The anywhere link is encrypted (HTTPS).

## Terminal version

The same program also runs from a terminal. Download `ferry-cli-windows-x64.exe`, `ferry-cli-mac-arm64`, `ferry-cli-mac-intel` or `ferry-linux-x64`, then:

```
ferry files                  share the current folder
ferry files ~/Pictures       share a specific folder
ferry site 3000              share a website running on port 3000

  --internet     also give a link that works from anywhere
  --pin 123456   choose the PIN yourself
  --port 8080    choose the port (default: first free one from 8080)
  --read-only    visitors can download but not upload
```

## Troubleshooting

- **The other device can't connect.** Check both are on the same Wi-Fi (not a guest network) and that you allowed Ferry through the firewall. On Windows: Settings > Privacy & security > Windows Security > Firewall > Allow an app through firewall.
- **The website option says "The website is not running".** Start your website on this computer first, check the port number, then reload.
- **The anywhere link doesn't appear.** It needs a working internet connection and can take up to 30 seconds. Some school and work networks block it.
- **Sharing stops when the computer sleeps.** Keep it awake while sharing.

## Building it yourself

You need [Go](https://go.dev/dl/) 1.24 or newer.

```
go run .                     run the app
go test ./...                run the tests
scripts/build.sh             build everything into dist/
```

The Mac app is only built as a universal, signed app when `scripts/build.sh mac` runs on a Mac. Pushing a tag like `v0.2.0` builds everything on GitHub and publishes a release automatically.

## License

MIT
