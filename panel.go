package main

import "html/template"

func init() {
	template.Must(pages.Parse(panelHTML))
}

const panelHTML = `
{{define "panel-locked"}}{{template "head"}}<title>Ferry</title></head><body>
<main style="max-width:460px;padding-top:12vh"><div class="card" style="padding:24px">
<h1 style="margin:0 0 8px;font-size:22px">Open Ferry again</h1>
<p class="muted" style="margin:0">This page only works when opened by the Ferry app. Open Ferry from your apps and it will bring you back here.</p>
</div></main></body></html>{{end}}

{{define "panel"}}{{template "head"}}<title>Ferry</title>
<style>
main{max-width:760px}
#intro{position:fixed;inset:0;z-index:10;background:var(--bg);display:none;place-items:center;cursor:pointer;transition:opacity .45s ease}
#intro.play{display:grid}
#intro.out{opacity:0}
#intro svg{width:min(440px,88vw);height:auto;overflow:visible}
#intro .pc{fill:none;stroke:var(--text);stroke-width:3;stroke-linejoin:round;stroke-linecap:round}
#intro .screen{fill:var(--card);transition:fill .35s ease}
#intro.arrived .screen.to{fill:var(--accent)}
#intro .file{fill:var(--accent)}
#intro .cargo{transition:opacity .25s ease}
#intro.arrived .cargo{opacity:0}
#intro .landed{fill:var(--accent-text);opacity:0;transform:translateY(-6px);transition:opacity .3s ease .1s,transform .3s ease .1s}
#intro.arrived .landed{opacity:1;transform:none}
#intro .wave{fill:none;stroke:var(--accent);stroke-width:3;stroke-linecap:round;opacity:.55}
#intro .waves{animation:drift 1.2s linear infinite}
#intro .ferry{animation:cross 1.7s cubic-bezier(.45,0,.3,1) .25s both}
#intro .bob{animation:bob .9s ease-in-out infinite alternate;transform-box:fill-box;transform-origin:50% 100%}
#intro .hull{fill:var(--text)}
#intro .cabin{fill:var(--card);stroke:var(--text);stroke-width:2.5}
#intro .stack{fill:var(--accent)}
#intro .name{font:700 22px/1 system-ui,-apple-system,"Segoe UI",sans-serif;fill:var(--text);letter-spacing:.02em;opacity:0;animation:fadein .4s ease 1.6s forwards}
@keyframes cross{from{transform:translateX(0)}to{transform:translateX(158px)}}
@keyframes bob{from{transform:translateY(0) rotate(-2.5deg)}to{transform:translateY(-3px) rotate(2.5deg)}}
@keyframes drift{from{transform:translateX(0)}to{transform:translateX(-40px)}}
@keyframes fadein{to{opacity:1}}
main.enter{animation:enter .45s ease both}
@keyframes enter{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:none}}
header{display:flex;align-items:center;gap:10px;margin-bottom:20px}
header .logo{width:28px;height:28px;border-radius:6px;background:var(--accent);display:grid;place-items:center;color:var(--accent-text);font-weight:800}
header h1{font-size:20px;margin:0}
header .grow{flex:1}
section.card{padding:20px;margin-bottom:16px}
h2{font-size:17px;margin:0 0 14px}
label.field{display:block;font-weight:600;margin:16px 0 6px}
.row{display:flex;gap:8px}
.row input{flex:1;min-width:0}
input[type=text],input[type=number]{font:inherit;padding:9px 11px;border:1px solid var(--line);border-radius:6px;background:var(--bg);color:var(--text)}
input[type=number]{width:140px}
.seg{display:inline-flex;border:1px solid var(--line);border-radius:6px;overflow:hidden}
.seg button{background:transparent;color:var(--text);border-radius:0;font-weight:600;padding:9px 18px}
.seg button+button{border-left:1px solid var(--line)}
.seg button.on{background:var(--accent);color:var(--accent-text)}
.check{display:flex;gap:10px;align-items:flex-start;margin-top:16px;cursor:pointer}
.check input{margin-top:5px;width:16px;height:16px;accent-color:var(--accent)}
.hint{font-size:14px;color:var(--muted);margin:4px 0 0}
.actions{margin-top:20px;display:flex;gap:10px;align-items:center}
.top{display:flex;gap:12px;align-items:flex-start;justify-content:space-between;flex-wrap:wrap}
.top>div{flex:1;min-width:0}
.what{word-break:break-all;margin:2px 0 0}
.pin{font:700 40px/1.1 ui-monospace,Menlo,Consolas,monospace;letter-spacing:8px;margin:4px 0 0}
.links{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:16px;margin-top:20px}
.link{border:1px solid var(--line);border-radius:8px;padding:16px}
.link .url{font:14px ui-monospace,Menlo,Consolas,monospace;word-break:break-all;margin:6px 0 10px}
.link img{display:block;width:180px;height:180px;margin:12px 0 8px;image-rendering:pixelated;background:#fff;padding:8px;border-radius:6px}
.small{font-size:13px;padding:6px 10px}
#log{list-style:none;margin:0;padding:0;font:13px/1.7 ui-monospace,Menlo,Consolas,monospace;max-height:220px;overflow:auto}
#log li{border-top:1px solid var(--line);padding:2px 0}
#log li:first-child{border-top:0}
#closed{display:none;padding:14px 16px;margin-bottom:16px;border-color:var(--bad)}
#update{border-color:var(--accent)}
#update .top{align-items:center}
.label{font-size:13px;font-weight:700;text-transform:uppercase;letter-spacing:.06em;color:var(--muted)}
main.in-settings>section:not(#settings),main.in-settings>.page-hint{display:none!important}
#settings h2{margin:0}
#settings .group{border-top:1px solid var(--line);margin-top:20px;padding-top:18px}
#settings .top+.group{border-top:0;margin-top:14px;padding-top:0}
#check-status{margin:0}
#port{width:160px}
#prefs-saved{font-size:13px;color:var(--muted)}
button:disabled{opacity:.45;cursor:default}
main.in-transfer{max-width:1180px}
main.in-transfer>section:not(.xfer),main.in-transfer>.page-hint{display:none!important}
.xfer h2{margin:0}
.cform{display:grid;grid-template-columns:minmax(150px,190px) minmax(160px,1fr) 84px minmax(110px,.7fr) minmax(110px,.7fr) auto;gap:10px;align-items:end;margin-top:14px}
.cform label{display:block;font-size:13px;font-weight:600;margin-bottom:5px}
.cform input,.cform select{width:100%}
select,input[type=password]{font:inherit;padding:9px 11px;border:1px solid var(--line);border-radius:6px;background:var(--bg);color:var(--text)}
.cform input[type=number]{width:100%}
.recent{display:flex;flex-wrap:wrap;gap:8px;margin-top:8px}
.recent .item{display:inline-flex;border:1px solid var(--line);border-radius:6px;overflow:hidden}
.recent .item button{background:transparent;color:var(--text);border-radius:0;font-weight:500;padding:6px 10px;font-size:13px}
.recent .item button+button{border-left:1px solid var(--line);color:var(--muted);padding:6px 9px}
.trustbox{border:1px solid var(--accent);border-radius:8px;padding:14px 16px;margin-top:14px}
.trustbox.warn{border-color:var(--bad)}
.fp{font:13px/1.5 ui-monospace,Menlo,Consolas,monospace;word-break:break-all;background:var(--bg);border:1px solid var(--line);border-radius:6px;padding:8px 10px;margin:8px 0}
.panes{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.pane{border:1px solid var(--line);border-radius:8px;display:flex;flex-direction:column;min-width:0;background:var(--card)}
.pane.drop{outline:2px dashed var(--accent);outline-offset:-3px}
.pane-head{padding:10px;border-bottom:1px solid var(--line)}
.pane-head .label{margin-bottom:8px;display:flex;justify-content:space-between;gap:8px}
.pane-head .count{font-weight:500;text-transform:none;letter-spacing:0}
.pane-head .row input{font:13px ui-monospace,Menlo,Consolas,monospace;padding:7px 9px}
.pane-list{height:400px;overflow:auto;position:relative}
.pane-msg{padding:16px;color:var(--muted);font-size:14px}
table.files{width:100%;border-collapse:collapse;font-size:14px;table-layout:fixed}
table.files th{position:sticky;top:0;background:var(--card);text-align:left;font-size:12px;font-weight:600;color:var(--muted);padding:6px 8px;border-bottom:1px solid var(--line)}
table.files td{padding:5px 8px;border-top:1px solid var(--line);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;user-select:none}
table.files tr:first-child td{border-top:0}
table.files tbody tr{cursor:default}
table.files tbody tr:hover{background:color-mix(in srgb,var(--text) 5%,transparent)}
table.files tr.sel,table.files tr.sel:hover{background:color-mix(in srgb,var(--accent) 18%,transparent)}
table.files .ck{width:30px}
table.files .num{text-align:right;color:var(--muted);font-variant-numeric:tabular-nums}
table.files .sz{width:84px}
table.files .dt{width:150px}
table.files td.num{font-size:13px}
.files .dirbtn{background:none;border:0;padding:0;color:var(--text);font-weight:600;cursor:pointer;max-width:calc(100% - 34px);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;text-align:left;vertical-align:bottom}
.files .dirbtn:hover{text-decoration:underline}
.files .kind{display:inline-block;width:34px;font-size:11px;font-weight:700;color:var(--muted)}
.pane-foot{padding:10px;border-top:1px solid var(--line);display:flex;gap:8px;flex-wrap:wrap;align-items:center}
.pane-foot .grow{flex:1}
#queue{list-style:none;margin:0;padding:0;max-height:300px;overflow:auto}
#queue li{display:grid;grid-template-columns:76px minmax(0,1fr) 160px 118px 74px;gap:10px;align-items:center;padding:7px 0;border-top:1px solid var(--line);font-size:14px}
#queue li:first-child{border-top:0}
#queue .dirn{font-size:12px;font-weight:700;color:var(--muted);text-transform:uppercase;letter-spacing:.04em}
#queue .nm{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
#queue .st{font-size:13px;color:var(--muted);text-align:right;font-variant-numeric:tabular-nums}
#queue .st.failed{color:var(--bad)}
#queue .err{grid-column:2/-1;font-size:13px;color:var(--bad);margin-top:-4px}
.bar{height:6px;background:var(--line);border-radius:3px;overflow:hidden}
.bar i{display:block;height:100%;width:0;background:var(--accent);transition:width .3s ease}
.trusted li{display:flex;justify-content:space-between;align-items:center;gap:10px;padding:6px 0;border-top:1px solid var(--line);font-size:14px}
.trusted{list-style:none;margin:8px 0 0;padding:0}
@media (max-width:560px){table.files .dt{display:none}.pane-list{height:300px}}
@media (max-width:860px){.cform{grid-template-columns:1fr 1fr}.panes{grid-template-columns:1fr}#queue li{grid-template-columns:64px minmax(0,1fr) 90px}#queue .bar,#queue .st{display:none}}
</style></head><body>
<div id="intro" aria-hidden="true" title="Click to skip">
<svg viewBox="0 0 420 190">
<defs><clipPath id="sea"><rect x="96" y="100" width="228" height="40"/></clipPath></defs>
<g>
<rect class="screen" x="18" y="40" width="78" height="54" rx="5"/>
<rect class="pc" x="18" y="40" width="78" height="54" rx="5"/>
<path class="pc" d="M57 94v14M40 110h34"/>
<path class="file" d="M44 58h12l4 4h14v18H44z"/>
</g>
<g>
<rect class="screen to" x="324" y="40" width="78" height="54" rx="5"/>
<rect class="pc" x="324" y="40" width="78" height="54" rx="5"/>
<path class="pc" d="M363 94v14M346 110h34"/>
<path class="landed" d="M350 58h12l4 4h14v18H350z"/>
</g>
<g clip-path="url(#sea)"><g class="waves">
<path class="wave" d="M80 122q10-7 20 0t20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0"/>
<path class="wave" d="M60 134q10-7 20 0t20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0"/>
</g></g>
<g class="ferry"><g class="bob">
<rect class="stack" x="124" y="90" width="7" height="12" rx="1.5"/>
<rect class="cabin" x="112" y="100" width="30" height="13" rx="2"/>
<path class="hull" d="M101 113h52l-7 11h-39z"/>
<path class="file cargo" d="M120 103h5l1.5 1.5h6v6.5h-12.5z"/>
</g></g>
<text class="name" x="210" y="172" text-anchor="middle">Ferry</text>
</svg>
</div>
<main>
<header><div class="logo">F</div><h1>Ferry</h1><span class="muted">{{.Version}}</span><span class="grow"></span>
<button class="btn plain small" id="open-transfer">Transfer files</button>
<button class="btn plain small" id="open-settings">Settings</button>
<button class="btn plain small" id="quit">Quit Ferry</button></header>

<div class="card" id="closed">Ferry has closed. Open the app again to keep sharing.</div>

<section class="card" id="update" hidden>
<div class="top"><div><div class="label">Update ready</div><p class="what" id="update-text"></p></div>
<button type="button" id="update-go">Restart to update</button></div>
</section>

<section class="card" id="setup" hidden>
<h2>What do you want to share?</h2>
<div class="seg"><button type="button" data-mode="files">A folder</button><button type="button" data-mode="site">A website</button><button type="button" data-mode="receive">From another computer</button></div>

<div id="files-opts">
<label class="field" for="folder">Folder</label>
<div class="row"><input type="text" id="folder" placeholder="Choose a folder, or paste its path"><button type="button" class="btn plain" id="browse">Browse</button></div>
<label class="check"><input type="checkbox" id="uploads"><span>Let visitors upload files into this folder</span></label>
<label class="check"><input type="checkbox" id="keeppin"><span>Keep the same PIN when sharing starts again
<span class="hint" style="display:block">Needed if another computer keeps a copy of this folder up to date.</span></span></label>
</div>

<div id="site-opts" hidden>
<label class="field" for="target">Port your website is running on</label>
<input type="number" id="target" min="1" max="65535">
<p class="hint">Start the website on this computer first. Common ports: 3000, 5173, 8000.</p>
</div>

<div id="receive-opts" hidden>
<label class="field" for="server">Address of the other computer</label>
<input type="text" id="server" placeholder="For example 192.168.1.20:8080" autocomplete="off">
<p class="hint">Find the computer below, or type the address shown in the Ferry window on it. Both computers need to be on the same Wi-Fi.</p>
<button type="button" class="btn plain small" id="find">Find computers on this network</button>
<div id="found" class="hint"></div>
<label class="field" for="rfolder">Folder to keep up to date</label>
<div class="row"><input type="text" id="rfolder" placeholder="Choose a folder, or paste its path"><button type="button" class="btn plain" id="rbrowse">Browse</button></div>
<label class="field" for="rpin">PIN from the other computer</label>
<input type="text" id="rpin" inputmode="numeric" autocomplete="off" maxlength="12" style="width:160px">
<p class="hint">Files in this folder are kept the same as on the other computer, so changes you make to them here are overwritten. Files removed there move to a folder called .ferry-removed here. Files you put in this folder yourself are left alone.</p>
</div>

<label class="check" id="internet-opt"><input type="checkbox" id="internet"><span>Also allow access from outside this Wi-Fi
<span class="hint" style="display:block">Gives a free link that works anywhere. Visitors still need the PIN.</span></span></label>

<p class="error" id="err" hidden></p>
<div class="actions"><button type="button" id="start">Start sharing</button></div>
</section>

<section class="card" id="mirror-live" hidden>
<div class="top"><div><div class="label">Receiving from</div><p class="what" id="mirror-what"></p><p class="hint" id="mirror-status"></p></div>
<button type="button" class="btn plain" id="mirror-stop">Stop receiving</button></div>
</section>

<section class="card" id="live" hidden>
<div class="top"><div><div class="label">Sharing</div><p class="what" id="what"></p></div>
<button type="button" class="btn plain" id="stop">Stop sharing</button></div>
<div style="margin-top:18px"><div class="label">PIN</div><div class="pin" id="pin"></div>
<p class="hint">Visitors type this when they open the link.</p></div>
<div class="links" id="links"></div>
<p class="hint" id="net" hidden></p>
<p class="hint" style="margin-top:16px">Scanning a code logs in without typing the PIN, so only show codes to people you trust. Keep this computer awake while sharing.</p>
</section>

<section class="card xfer" id="xfer-connect" hidden>
<div class="top"><div><h2>Transfer files</h2>
<p class="hint">Connect to an FTP or SFTP server, like a website host, a NAS drive or a Raspberry Pi, and copy files to and from it.</p></div>
<button type="button" class="btn plain small" id="close-transfer">Done</button></div>

<div id="cbar" hidden>
<div class="top" style="margin-top:14px;align-items:center"><div><div class="label">Connected to</div><p class="what" id="clabel"></p></div>
<button type="button" class="btn plain" id="disconnect">Disconnect</button></div>
</div>

<div id="cform-box">
<form class="cform" id="cform" autocomplete="off">
<div><label for="cproto">Kind of server</label>
<select id="cproto">
<option value="sftp">SFTP (SSH)</option>
<option value="ftpes">FTP over TLS</option>
<option value="ftps">FTPS (implicit TLS)</option>
<option value="ftp">FTP, not encrypted</option>
</select></div>
<div><label for="chost">Server</label><input type="text" id="chost" placeholder="ftp.example.com" spellcheck="false"></div>
<div><label for="cport">Port</label><input type="number" id="cport" min="1" max="65535" placeholder="22"></div>
<div><label for="cuser">User name</label><input type="text" id="cuser" spellcheck="false" autocapitalize="off"></div>
<div><label for="cpass">Password</label><input type="password" id="cpass"></div>
<div><button type="submit" id="connect">Connect</button></div>
</form>
<p class="hint" id="cproto-hint"></p>
<div class="trustbox" id="trustbox" hidden>
<div class="label" id="trust-title"></div>
<p class="what" id="trust-text" style="word-break:normal"></p>
<div class="fp" id="trust-fp"></div>
<p class="hint">Only trust it if this is the server you meant to reach. Ferry remembers your answer and warns you if this ever changes.</p>
<div class="actions" style="margin-top:12px"><button type="button" id="trust-yes">Trust and connect</button><button type="button" class="btn plain" id="trust-no">Cancel</button></div>
</div>
<div id="recent-box" hidden><div class="label" style="margin-top:18px">Recent connections</div>
<div class="recent" id="recent"></div>
<p class="hint">Ferry remembers the server, user name and last folder. Passwords are not saved.</p></div>
</div>
<p class="error" id="cerr" hidden></p>
</section>

<section class="card xfer" id="xfer-files" hidden>
<div class="panes">
<div class="pane" id="pane-local" data-side="local">
<div class="pane-head"><div class="label">This computer <span class="count"></span></div>
<div class="row"><button type="button" class="btn plain small" data-act="up" title="Folder above">Up</button><input type="text" data-act="path" spellcheck="false" aria-label="Folder on this computer"><button type="button" class="btn plain small" data-act="refresh">Refresh</button></div></div>
<div class="pane-list"></div>
<div class="pane-foot"><span class="grow"></span>
<button type="button" class="small" data-act="send">Upload to server</button></div>
</div>
<div class="pane" id="pane-remote" data-side="remote">
<div class="pane-head"><div class="label">Server <span class="count"></span></div>
<div class="row"><button type="button" class="btn plain small" data-act="up" title="Folder above">Up</button><input type="text" data-act="path" spellcheck="false" aria-label="Folder on the server"><button type="button" class="btn plain small" data-act="refresh">Refresh</button></div></div>
<div class="pane-list"></div>
<div class="pane-foot">
<button type="button" class="btn plain small" data-act="mkdir">New folder</button>
<button type="button" class="btn plain small" data-act="rename">Rename</button>
<button type="button" class="btn plain small" data-act="delete">Delete</button><span class="grow"></span>
<button type="button" class="small" data-act="send">Download here</button></div>
</div>
</div>
<p class="hint" style="margin-top:12px">Click to select, double-click a folder to open it, or drag files from one side to the other. Shift-click selects a range.</p>
</section>

<section class="card xfer" id="xfer-queue" hidden>
<div class="top" style="align-items:center"><h2 style="margin:0">Transfers</h2><span class="grow" style="flex:1"></span>
<button type="button" class="btn plain small" id="q-cancel">Cancel all</button>
<button type="button" class="btn plain small" id="q-clear">Clear finished</button></div>
<p class="hint" id="q-sum"></p>
<ul id="queue"></ul>
</section>

<section class="card" id="settings" hidden>
<div class="top"><h2>Settings</h2><span id="prefs-saved"></span><button type="button" class="btn plain small" id="close-settings">Done</button></div>

<div class="group">
<div class="label">Updates</div>
<p class="what">This is Ferry {{.Version}}.</p>
<div class="actions"><button type="button" id="check">Check for updates</button>
<button type="button" id="check-install" style="display:none">Restart to update</button></div>
<p class="hint" id="check-status"></p>
<label class="check"><input type="checkbox" id="autoupdate"><span>Check for updates automatically
<span class="hint" style="display:block">When Ferry opens and every 6 hours while it runs. Nothing is installed until you choose to restart.</span></span></label>
</div>

<div class="group">
<div class="label">Network</div>
<label class="check"><input type="checkbox" id="findable"><span>Let other computers on this Wi-Fi find this one
<span class="hint" style="display:block">Only while you share a folder. They still need the PIN to open it.</span></span></label>
<label class="field" for="port">Sharing port</label>
<input type="number" id="port" min="1024" max="65535" placeholder="Automatic">
<p class="hint">Leave empty to pick one automatically. Set a port if you want the address for this computer to stay the same each time. Used the next time sharing starts.</p>
<p class="error" id="prefs-err" hidden></p>
</div>

<div class="group">
<div class="label">Trusted servers</div>
<p class="hint">Servers you chose to trust on the Transfer files page. Remove one if its owner tells you its key or certificate changed.</p>
<ul class="trusted" id="trusted"></ul>
</div>

<div class="group">
<div class="label">About</div>
<p class="hint">Ferry keeps its settings on this computer only. <a href="https://github.com/S1xinch/Ferry---Local-File-Sharing/releases" target="_blank" rel="noopener">See what changed in each version</a>.</p>
</div>
</section>

<section class="card"><h2>Activity</h2><ul id="log"></ul></section>
<p class="hint page-hint">You can close this tab. Sharing keeps going until you stop it, and opening Ferry again brings this window back.</p>
</main>
<script>
var $ = function (id) { return document.getElementById(id); };
var mode = "files", formFilled = false, lastLinks = "", running = null;

function call(path, body) {
  return fetch(path, {
    method: body === undefined && path === "/api/state" ? "GET" : "POST",
    headers: {"X-Ferry": "1", "Content-Type": "application/json"},
    body: body === undefined ? undefined : JSON.stringify(body)
  }).then(function (r) {
    return r.json().then(function (j) { if (!r.ok) throw new Error(j.error || "Something went wrong"); return j; });
  });
}

function setMode(m) {
  mode = m;
  document.querySelectorAll(".seg button").forEach(function (b) { b.classList.toggle("on", b.dataset.mode === m); });
  $("files-opts").hidden = m !== "files";
  $("site-opts").hidden = m !== "site";
  $("receive-opts").hidden = m !== "receive";
  $("internet-opt").hidden = m === "receive";
  $("start").textContent = m === "receive" ? "Start receiving" : "Start sharing";
}
document.querySelectorAll(".seg button").forEach(function (b) { b.onclick = function () { setMode(b.dataset.mode); }; });

function showError(msg) { msg = msg ? msg.charAt(0).toUpperCase() + msg.slice(1) : ""; $("err").textContent = msg; $("err").hidden = !msg; }

function browseInto(button, input) {
  button.disabled = true;
  call("/api/pick", {}).then(function (r) { if (r.folder) input.value = r.folder; })
    .catch(function (e) { showError(e.message); })
    .finally(function () { button.disabled = false; });
}
$("browse").onclick = function () { browseInto($("browse"), $("folder")); };
$("rbrowse").onclick = function () { browseInto($("rbrowse"), $("rfolder")); };

$("find").onclick = function () {
  $("find").disabled = true;
  $("found").textContent = "Looking...";
  call("/api/discover", {}).then(function (r) {
    var box = $("found");
    box.textContent = "";
    if (!r.computers || !r.computers.length) {
      box.textContent = "No Ferry computers answered. Check that both are on the same Wi-Fi, that the other one is sharing a folder, and that its firewall lets Ferry in. You can also type the address shown on the other computer.";
      return;
    }
    r.computers.forEach(function (c) {
      var b = el("button", "btn plain small", c.name + ": " + c.folder + " (" + c.addr + ")");
      b.style.margin = "4px 6px 0 0";
      b.onclick = function () { $("server").value = c.addr; };
      box.appendChild(b);
    });
  }).catch(function (e) { $("found").textContent = e.message; })
    .finally(function () { $("find").disabled = false; });
};

$("mirror-stop").onclick = function () { call("/api/mirror/stop", {}).then(refresh); };

$("start").onclick = function () {
  showError("");
  $("start").disabled = true;
  call("/api/start", {
    mode: mode, folder: $("folder").value, target: parseInt($("target").value, 10) || 0,
    uploads: $("uploads").checked, internet: $("internet").checked, keepPin: $("keeppin").checked,
    server: $("server").value, mirrorFolder: $("rfolder").value, receivePin: $("rpin").value
  }).then(refresh).catch(function (e) { showError(e.message); })
    .finally(function () { $("start").disabled = false; });
};

$("stop").onclick = function () { call("/api/stop", {}).then(function () { formFilled = false; refresh(); }); };
$("quit").onclick = function () {
  if (running && !confirm("Quitting stops sharing. Quit Ferry?")) return;
  call("/api/quit", {}).finally(function () { $("closed").hidden = false; $("closed").style.display = "block"; });
};

$("update-go").onclick = function () {
  if (running && !confirm("Updating restarts Ferry and stops sharing. Update now?")) return;
  $("update-go").disabled = true;
  call("/api/update", {}).then(function () {
    $("closed").textContent = "Ferry is restarting with the new version. Its window will open in a moment.";
    $("closed").style.display = "block";
  }).catch(function (e) {
    $("update-go").disabled = false;
    // A network error here means Ferry closed partway through.
    var msg = e instanceof TypeError
      ? "Ferry closed while updating. Open Ferry again. If it still shows the old version, download the new one from GitHub."
      : e.message.charAt(0).toUpperCase() + e.message.slice(1);
    $("update-text").textContent = msg;
    $("check-status").textContent = msg;
  });
};

var prefsFilled = false, savedTimer = null;
function showSettings(on) {
  if (on) showTransfer(false);
  document.querySelector("main").classList.toggle("in-settings", on);
  $("settings").hidden = !on;
  window.scrollTo(0, 0);
}
$("open-settings").onclick = function () { showSettings(true); };
$("close-settings").onclick = function () { showSettings(false); };

$("check").onclick = function () {
  $("check").disabled = true;
  $("check-status").textContent = "Checking...";
  call("/api/update/check", {}).then(refresh).catch(function (e) {
    $("check").disabled = false;
    $("check-status").textContent = e.message;
  });
};
$("check-install").onclick = function () { $("update-go").onclick(); };

function savePrefs() {
  $("prefs-err").hidden = true;
  var port = $("port").value.trim();
  call("/api/prefs", {
    autoUpdate: $("autoupdate").checked,
    hidden: !$("findable").checked,
    port: port ? parseInt(port, 10) || -1 : 0
  }).then(function () {
    $("prefs-saved").textContent = "Saved";
    clearTimeout(savedTimer);
    savedTimer = setTimeout(function () { $("prefs-saved").textContent = ""; }, 1500);
  }).catch(function (e) {
    $("prefs-err").textContent = e.message;
    $("prefs-err").hidden = false;
  });
}
$("autoupdate").onchange = savePrefs;
$("findable").onchange = savePrefs;
$("port").onchange = savePrefs;

function renderSettings(s) {
  var p = s.settings.prefs || {};
  if (!prefsFilled) {
    $("autoupdate").checked = !!p.autoUpdate;
    $("findable").checked = !p.hidden;
    $("port").value = p.port || "";
    prefsFilled = true;
  }
  var u = s.update || {};
  $("check").disabled = !!u.checking || !!u.ready;
  $("check-install").style.display = u.ready ? "" : "none";
  $("check-status").textContent = u.status || "";
}

/* ---------- Transfer files ---------- */
var xfer = {
  open: false, connected: false, label: "", busy: false, wasBusy: false,
  recentKey: "", trustedKey: "", pendingTrust: null,
  local: {path: "", up: "", entries: [], sel: {}, anchor: -1, loaded: false},
  remote: {path: "", up: "", entries: [], sel: {}, anchor: -1, loaded: false}
};
var protoHints = {
  sftp: "Works with most Linux servers, NAS drives and Raspberry Pis. Leave the password empty to use your SSH key.",
  ftpes: "Regular FTP with the connection encrypted. The best choice for most web hosts.",
  ftps: "An older way of encrypting FTP. Use it if your host says “implicit”.",
  ftp: "Nothing is encrypted, including your password. Only use this on your own network."
};
var defaultPorts = {sftp: 22, ftpes: 21, ftps: 990, ftp: 21};

function showTransfer(on) {
  if (on) showSettings(false);
  xfer.open = on;
  document.querySelector("main").classList.toggle("in-transfer", on);
  document.querySelectorAll("section.xfer").forEach(function (s) { s.hidden = !on; });
  window.scrollTo(0, 0);
  if (on && !xfer.local.loaded) loadPane("local", "");
  if (on && xfer.connected && !xfer.remote.loaded) loadPane("remote", "");
}
$("open-transfer").onclick = function () { showTransfer(true); };
$("close-transfer").onclick = function () { showTransfer(false); };

function xerr(msg) {
  msg = msg ? msg.charAt(0).toUpperCase() + msg.slice(1) : "";
  $("cerr").textContent = msg;
  $("cerr").hidden = !msg;
}

function protoChanged() {
  var p = $("cproto").value;
  $("cport").placeholder = defaultPorts[p];
  $("cproto-hint").textContent = protoHints[p];
}
$("cproto").onchange = protoChanged;
protoChanged();

function connectNow(trust) {
  xerr("");
  var btn = $("connect");
  btn.disabled = true;
  btn.textContent = "Connecting...";
  var body = {
    protocol: $("cproto").value, host: $("chost").value.trim(),
    port: parseInt($("cport").value, 10) || 0, user: $("cuser").value.trim(),
    password: $("cpass").value, trust: trust || ""
  };
  return call("/api/remote/connect", body).then(function (r) {
    if (r.needsTrust) {
      xfer.pendingTrust = r.fingerprint;
      var ssh = r.kind === "ssh";
      $("trust-title").textContent = "New server";
      $("trust-text").textContent = ssh
        ? "This is the first time Ferry has connected to " + r.host + ". To be sure it is the right server, you can compare its key fingerprint with the one shown by the server's owner or control panel."
        : r.host + " uses a certificate that no known authority vouches for. That is normal for home servers and NAS drives, but check the fingerprint if you can.";
      $("trust-fp").textContent = r.fingerprint;
      $("trustbox").hidden = false;
      return;
    }
    $("trustbox").hidden = true;
    $("cpass").value = "";
    xfer.remote.loaded = false;
    return refresh().then(function () { loadPane("remote", r.path); });
  }).catch(function (e) { xerr(e.message); })
    .finally(function () { btn.disabled = false; btn.textContent = "Connect"; });
}
$("cform").onsubmit = function (ev) { ev.preventDefault(); connectNow(""); };
$("trust-yes").onclick = function () { connectNow(xfer.pendingTrust); };
$("trust-no").onclick = function () { $("trustbox").hidden = true; xfer.pendingTrust = null; };

$("disconnect").onclick = function () {
  if (xfer.busy && !confirm("Transfers are still running. Disconnect and cancel them?")) return;
  call("/api/remote/disconnect", {}).then(refresh);
};

function fillFromRecent(r) {
  $("cproto").value = r.protocol;
  protoChanged();
  $("chost").value = r.host;
  $("cport").value = r.port === defaultPorts[r.protocol] ? "" : r.port;
  $("cuser").value = r.user;
  $("cpass").value = "";
  $("trustbox").hidden = true;
  xerr("");
  $("cpass").focus();
}

function renderRecent(list) {
  var key = JSON.stringify(list.map(function (r) { return r.key; }));
  if (key === xfer.recentKey) return;
  xfer.recentKey = key;
  var box = $("recent");
  box.textContent = "";
  $("recent-box").hidden = !list.length;
  list.forEach(function (r) {
    var item = el("span", "item");
    var go = el("button", "", r.label);
    go.type = "button";
    go.title = "Fill in this connection";
    go.onclick = function () { fillFromRecent(r); };
    var x = el("button", "", "×");
    x.type = "button";
    x.title = "Forget " + r.label;
    x.setAttribute("aria-label", "Forget " + r.label);
    x.onclick = function () { call("/api/remote/forget", {key: r.key}).then(refresh); };
    item.appendChild(go);
    item.appendChild(x);
    box.appendChild(item);
  });
}

function fmtSize(n) {
  if (n < 1024) return n + " B";
  var u = ["KB", "MB", "GB", "TB"], i = -1;
  do { n /= 1024; i++; } while (n >= 1024 && i < u.length - 1);
  return (n < 10 ? n.toFixed(1) : Math.round(n)) + " " + u[i];
}
function fmtDate(s) {
  if (!s || s.indexOf("0001-") === 0) return "";
  var d = new Date(s);
  if (isNaN(d)) return "";
  var now = new Date();
  var same = d.getFullYear() === now.getFullYear();
  return d.toLocaleDateString(undefined, same ? {day: "numeric", month: "short"} : {day: "numeric", month: "short", year: "numeric"})
    + (same ? " " + d.toLocaleTimeString(undefined, {hour: "2-digit", minute: "2-digit"}) : "");
}

function pane(side) { return $("pane-" + side); }
function q(side, act) { return pane(side).querySelector("[data-act=" + act + "]"); }

function paneMessage(side, msg) {
  var list = pane(side).querySelector(".pane-list");
  list.textContent = "";
  list.appendChild(el("div", "pane-msg", msg));
}

function loadPane(side, p, name) {
  var url = side === "local" ? "/api/local/list" : "/api/remote/list";
  var body = side === "local" ? {path: p, name: name || ""} : {path: name ? joinRemote(p, name) : p};
  if (side === "remote" && !xfer.connected) { paneMessage("remote", "Connect to a server to see its files."); return Promise.resolve(); }
  return call(url, body).then(function (l) {
    var st = xfer[side];
    var keep = st.path === l.path ? st.sel : {};
    st.path = l.path; st.up = l.up; st.entries = l.entries || []; st.loaded = true; st.anchor = -1;
    st.sel = {};
    st.entries.forEach(function (e) { if (keep[e.name]) st.sel[e.name] = true; });
    drawPane(side);
  }).catch(function (e) {
    if (!xfer[side].loaded) paneMessage(side, e.message.charAt(0).toUpperCase() + e.message.slice(1));
    else xerr(e.message);
  });
}

function joinRemote(dir, name) { return (dir === "/" ? "" : dir.replace(/\/+$/, "")) + "/" + name; }

function selected(side) {
  var st = xfer[side];
  return st.entries.filter(function (e) { return st.sel[e.name]; });
}

function updateButtons(side) {
  var st = xfer[side], n = selected(side).length, other = side === "local" ? "remote" : "local";
  var foot = pane(side).querySelector(".count");
  var files = st.entries.length;
  foot.textContent = n ? n + " selected" : files + (files === 1 ? " item" : " items");
  var canSend = n > 0 && xfer.connected && xfer[other].loaded && xfer.local.path !== ":drives";
  q(side, "send").disabled = !canSend;
  q(side, "up").disabled = !st.up;
  if (side === "remote") {
    q("remote", "mkdir").disabled = !xfer.connected;
    q("remote", "rename").disabled = n !== 1;
    q("remote", "delete").disabled = n === 0;
  }
}

function drawPane(side) {
  var st = xfer[side];
  var input = q(side, "path");
  if (document.activeElement !== input) {
    input.value = st.path === ":drives" ? "Drives" : st.path;
    input.scrollLeft = input.scrollWidth; // keep the folder's own name in view
  }
  var list = pane(side).querySelector(".pane-list");
  list.textContent = "";
  if (!st.entries.length) {
    list.appendChild(el("div", "pane-msg", "This folder is empty."));
    updateButtons(side);
    return;
  }
  var table = el("table", "files");
  var head = el("thead");
  var hr = el("tr");
  [["", "ck"], ["Name", ""], ["Size", "sz num"], ["Modified", "dt"]].forEach(function (h) {
    var th = el("th", h[1], h[0]);
    hr.appendChild(th);
  });
  head.appendChild(hr);
  table.appendChild(head);
  var body = el("tbody");
  st.entries.forEach(function (e, i) {
    var tr = el("tr", st.sel[e.name] ? "sel" : "");
    tr.draggable = true;
    var ck = el("td", "ck");
    var box = document.createElement("input");
    box.type = "checkbox";
    box.checked = !!st.sel[e.name];
    box.setAttribute("aria-label", "Select " + e.name);
    box.onclick = function (ev) { ev.stopPropagation(); toggle(side, i, ev.shiftKey, true); };
    ck.appendChild(box);
    var nm = el("td");
    if (e.dir) {
      var b = el("button", "dirbtn", e.name);
      b.type = "button";
      b.title = "Open " + e.name;
      b.onclick = function (ev) { ev.stopPropagation(); openFolder(side, e.name); };
      nm.appendChild(el("span", "kind", "DIR"));
      nm.appendChild(b);
    } else {
      nm.appendChild(el("span", "kind", ""));
      nm.appendChild(document.createTextNode(e.name));
      nm.title = e.name;
    }
    tr.appendChild(ck);
    tr.appendChild(nm);
    tr.appendChild(el("td", "num", e.dir ? "" : fmtSize(e.size)));
    tr.appendChild(el("td", "num dt", fmtDate(e.mod)));
    tr.onclick = function (ev) { toggle(side, i, ev.shiftKey, ev.ctrlKey || ev.metaKey); };
    tr.ondblclick = function () { if (e.dir) openFolder(side, e.name); };
    tr.ondragstart = function (ev) {
      if (!st.sel[e.name]) { st.sel = {}; st.sel[e.name] = true; drawPane(side); }
      ev.dataTransfer.setData("text/x-ferry", side);
      ev.dataTransfer.effectAllowed = "copy";
    };
    body.appendChild(tr);
  });
  table.appendChild(body);
  list.appendChild(table);
  updateButtons(side);
}

// toggle selects a row: a plain click picks just it, ctrl/cmd or the
// checkbox adds or removes it, and shift picks a range.
function toggle(side, i, range, add) {
  var st = xfer[side], name = st.entries[i].name;
  if (range && st.anchor >= 0) {
    var a = Math.min(st.anchor, i), b = Math.max(st.anchor, i);
    if (!add) st.sel = {};
    for (var k = a; k <= b; k++) st.sel[st.entries[k].name] = true;
  } else if (add) {
    if (st.sel[name]) delete st.sel[name]; else st.sel[name] = true;
    st.anchor = i;
  } else {
    var only = st.sel[name] && Object.keys(st.sel).length === 1;
    st.sel = {};
    if (!only) st.sel[name] = true;
    st.anchor = i;
  }
  drawPane(side);
}

function openFolder(side, name) {
  var st = xfer[side];
  if (side === "local") loadPane("local", st.path, name);
  else loadPane("remote", st.path, name);
}

function send(from) {
  var to = from === "local" ? "remote" : "local";
  var items = selected(from).map(function (e) { return {name: e.name, dir: e.dir, size: e.size, mod: e.mod}; });
  if (!items.length) return;
  var there = {};
  xfer[to].entries.forEach(function (e) { there[e.name] = true; });
  var clash = items.filter(function (e) { return there[e.name]; }).map(function (e) { return e.name; });
  if (clash.length) {
    var list = clash.slice(0, 5).join(", ") + (clash.length > 5 ? " and " + (clash.length - 5) + " more" : "");
    if (!confirm("Already there: " + list + ".\n\nReplace them? Files with the same name are overwritten.")) return;
  }
  xerr("");
  call("/api/transfer", {
    up: from === "local", localDir: xfer.local.path, remoteDir: xfer.remote.path, items: items
  }).then(function () { xfer[from].sel = {}; drawPane(from); refresh(); })
    .catch(function (e) { xerr(e.message); });
}

["local", "remote"].forEach(function (side) {
  var p = pane(side);
  q(side, "up").onclick = function () { if (xfer[side].up) loadPane(side, xfer[side].up); };
  q(side, "refresh").onclick = function () { loadPane(side, xfer[side].path); };
  q(side, "path").onkeydown = function (ev) {
    if (ev.key === "Enter") { ev.preventDefault(); loadPane(side, this.value.trim()); this.blur(); }
    if (ev.key === "Escape") { this.value = xfer[side].path; this.blur(); }
  };
  q(side, "send").onclick = function () { send(side); };
  p.ondragover = function (ev) {
    if (ev.dataTransfer.types.indexOf("text/x-ferry") < 0) return;
    ev.preventDefault();
    p.classList.add("drop");
  };
  p.ondragleave = function (ev) { if (!p.contains(ev.relatedTarget)) p.classList.remove("drop"); };
  p.ondrop = function (ev) {
    p.classList.remove("drop");
    var from = ev.dataTransfer.getData("text/x-ferry");
    if (!from || from === side) return;
    ev.preventDefault();
    send(from);
  };
});

q("remote", "mkdir").onclick = function () {
  var name = prompt("Name of the new folder");
  if (!name) return;
  call("/api/remote/mkdir", {dir: xfer.remote.path, name: name.trim()})
    .then(function () { loadPane("remote", xfer.remote.path); }).catch(function (e) { xerr(e.message); });
};
q("remote", "rename").onclick = function () {
  var e = selected("remote")[0];
  if (!e) return;
  var to = prompt("New name for " + e.name, e.name);
  if (!to || to === e.name) return;
  call("/api/remote/rename", {dir: xfer.remote.path, from: e.name, to: to.trim()})
    .then(function () { loadPane("remote", xfer.remote.path); }).catch(function (err) { xerr(err.message); });
};
q("remote", "delete").onclick = function () {
  var items = selected("remote");
  if (!items.length) return;
  var dirs = items.filter(function (e) { return e.dir; }).length;
  var what = items.length === 1 ? "“" + items[0].name + "”" : items.length + " items";
  if (!confirm("Delete " + what + " from the server?" + (dirs ? "\n\nFolders are deleted with everything inside them." : "") + "\n\nThis cannot be undone.")) return;
  call("/api/remote/delete", {dir: xfer.remote.path, items: items.map(function (e) { return {name: e.name, dir: e.dir}; })})
    .then(function () { loadPane("remote", xfer.remote.path); }).catch(function (err) { xerr(err.message); });
};

$("q-cancel").onclick = function () { call("/api/transfer/cancel", {id: 0}).then(refresh); };
$("q-clear").onclick = function () { call("/api/transfer/clear", {}).then(refresh); };

function renderQueue(list) {
  var box = $("queue");
  box.textContent = "";
  var active = 0, done = 0, failed = 0, bytes = 0, total = 0;
  list.forEach(function (t) {
    if (t.state === "queued" || t.state === "running") active++;
    if (t.state === "done" && !t.dir) done++;
    if (t.state === "failed") failed++;
    if (!t.dir && (t.state === "queued" || t.state === "running")) { total += t.size; bytes += t.done; }
  });
  $("xfer-queue").hidden = !xfer.open || !list.length;
  $("q-cancel").disabled = !active;
  $("q-clear").disabled = active === list.length;
  var sum = [];
  if (active) sum.push(active + " to go" + (total ? ", " + fmtSize(Math.max(total - bytes, 0)) + " left" : ""));
  if (done) sum.push(done + " done");
  if (failed) sum.push(failed + " failed");
  $("q-sum").textContent = sum.join(" · ");
  // Running and waiting first, then the most recent finished ones.
  var order = list.filter(function (t) { return t.state === "running"; })
    .concat(list.filter(function (t) { return t.state === "queued"; }).slice(0, 50))
    .concat(list.filter(function (t) {
      // A finished folder adds nothing once its files are listed.
      return t.state !== "running" && t.state !== "queued" && !(t.dir && t.state === "done");
    }).reverse().slice(0, 100));
  order.forEach(function (t) {
    var li = el("li");
    li.appendChild(el("span", "dirn", t.up ? "Upload" : "Download"));
    var nm = el("span", "nm", t.name + (t.dir ? "/" : ""));
    nm.title = t.name;
    li.appendChild(nm);
    var bar = el("span", "bar"), fill = el("i");
    var pct = t.state === "done" ? 100 : t.size ? Math.min(100, Math.round(t.done / t.size * 100)) : 0;
    fill.style.width = pct + "%";
    bar.appendChild(fill);
    li.appendChild(bar);
    var label = {queued: "Waiting", done: "Done", failed: "Failed", cancelled: "Cancelled"}[t.state];
    if (t.state === "running") label = t.dir ? "Reading folder" : t.size ? pct + "% of " + fmtSize(t.size) : fmtSize(t.done);
    else if (t.state === "done" && !t.dir && t.size) label = "Done, " + fmtSize(t.size);
    li.appendChild(el("span", "st" + (t.state === "failed" ? " failed" : ""), label));
    var act = el("span");
    if (t.state === "queued" || t.state === "running") {
      var c = el("button", "btn plain small", "Cancel");
      c.type = "button";
      c.onclick = function () { call("/api/transfer/cancel", {id: t.id}).then(refresh); };
      act.appendChild(c);
    }
    li.appendChild(act);
    if (t.error) li.appendChild(el("span", "err", t.error.charAt(0).toUpperCase() + t.error.slice(1)));
    box.appendChild(li);
  });
  xfer.busy = active > 0;
  if (xfer.wasBusy && !xfer.busy) {
    // Show what just arrived on either side.
    if (xfer.local.loaded) loadPane("local", xfer.local.path);
    if (xfer.remote.loaded) loadPane("remote", xfer.remote.path);
  }
  xfer.wasBusy = xfer.busy;
}

function renderTrusted(ids) {
  var key = JSON.stringify(ids);
  if (key === xfer.trustedKey) return;
  xfer.trustedKey = key;
  var box = $("trusted");
  box.textContent = "";
  if (!ids.length) { box.appendChild(el("li", "", "None yet.")); return; }
  ids.forEach(function (id) {
    var sp = id.indexOf(" ");
    var kind = id.slice(0, sp), where = id.slice(sp + 1);
    var li = el("li");
    li.appendChild(el("span", "", (kind === "ssh" ? "SFTP key for " : "Certificate for ") + where));
    var b = el("button", "btn plain small", "Remove");
    b.type = "button";
    b.onclick = function () { call("/api/remote/untrust", {id: id}).then(refresh); };
    li.appendChild(b);
    box.appendChild(li);
  });
}

function renderRemote(r) {
  if (!r) return;
  var wasConnected = xfer.connected;
  xfer.connected = r.connected;
  $("cbar").hidden = !r.connected;
  $("cform-box").hidden = r.connected;
  if (r.connected) {
    $("clabel").textContent = r.label;
    if (xfer.open && (!wasConnected || r.label !== xfer.label) && !xfer.remote.loaded) loadPane("remote", r.path);
  } else if (wasConnected || !xfer.remote.loaded) {
    xfer.remote = {path: "", up: "", entries: [], sel: {}, anchor: -1, loaded: false};
    q("remote", "path").value = "";
    paneMessage("remote", "Connect to a server to see its files.");
    updateButtons("remote");
  }
  xfer.label = r.label || "";
  renderRecent(r.recent || []);
  renderTrusted(r.trusted || []);
  renderQueue(r.transfers || []);
  if (xfer.local.loaded) updateButtons("local");
}
setInterval(function () { if (xfer.busy) refresh(); }, 600);

function el(tag, cls, text) { var e = document.createElement(tag); if (cls) e.className = cls; if (text) e.textContent = text; return e; }

function renderLinks(links) {
  var key = JSON.stringify(links.map(function (l) { return l.url; }));
  if (key === lastLinks) return;
  lastLinks = key;
  var box = $("links");
  box.textContent = "";
  links.forEach(function (l) {
    var card = el("div", "link");
    card.appendChild(el("div", "label", l.label));
    card.appendChild(el("div", "url", l.url));
    var copy = el("button", "btn plain small", "Copy link");
    copy.onclick = function () {
      navigator.clipboard.writeText(l.url).then(function () { copy.textContent = "Copied"; setTimeout(function () { copy.textContent = "Copy link"; }, 1500); });
    };
    card.appendChild(copy);
    var img = el("img"); img.src = l.qr; img.alt = "QR code for " + l.label;
    card.appendChild(img);
    box.appendChild(card);
  });
}

function render(s) {
  running = s.running;
  $("setup").hidden = s.running;
  $("live").hidden = !s.running;
  var mr = s.mirror;
  $("mirror-live").hidden = !mr;
  if (mr) {
    $("mirror-what").textContent = mr.server + " into " + mr.folder;
    $("mirror-status").textContent = mr.lastError ? "Waiting to reconnect: " + mr.lastError
      : mr.lastSync ? "Last checked " + mr.lastSync + ". " + mr.files + " files there, " + mr.changed + " updated and " + mr.removed + " moved to the trash at the last check."
      : "Connecting...";
  }
  renderSettings(s);
  renderRemote(s.remote);
  var u = s.update || {};
  $("update").hidden = !u.ready;
  if (u.ready) $("update-text").textContent = "Ferry " + u.version + " is downloaded. Restart Ferry to switch to it. Your settings are kept.";
  if (!formFilled) {
    var st = s.settings;
    setMode(st.mode || "files");
    $("folder").value = st.folder || "";
    $("target").value = st.target || 3000;
    $("uploads").checked = !!st.uploads;
    $("internet").checked = !!st.internet;
    $("keeppin").checked = !!st.keepPin;
    $("server").value = st.server || "";
    $("rfolder").value = st.mirrorFolder || "";
    formFilled = true;
  }
  if (s.running) {
    $("what").textContent = s.what;
    $("pin").textContent = s.pin;
    renderLinks(s.links || []);
    $("net").textContent = s.internetStatus || "";
    $("net").hidden = !s.internetStatus;
  } else {
    lastLinks = "";
  }
  var log = $("log");
  log.textContent = "";
  (s.activity.length ? s.activity : ["Nothing yet."]).forEach(function (line) { log.appendChild(el("li", "", line)); });
}

function refresh() {
  return call("/api/state").then(function (s) {
    $("closed").style.display = "none";
    render(s);
  }).catch(function () { $("closed").style.display = "block"; });
}
(function intro() {
  var params = new URLSearchParams(location.search);
  if (!params.has("intro")) return;
  history.replaceState(null, "", "/");
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  var box = $("intro"), main = document.querySelector("main"), done = false;
  function finish() {
    if (done) return;
    done = true;
    box.classList.add("out");
    main.classList.add("enter");
    setTimeout(function () { box.classList.remove("play"); }, 460);
  }
  box.classList.add("play");
  box.onclick = finish;
  setTimeout(function () { box.classList.add("arrived"); }, 1900);
  setTimeout(finish, 2600);
})();
refresh();
setInterval(refresh, 1500);
</script>
</body></html>{{end}}
`
