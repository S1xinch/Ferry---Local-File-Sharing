package main

import "html/template"

func init() {
	template.Must(pages.Parse(panelHTML))
}

const panelHTML = `
{{define "panel-locked"}}{{template "head"}}<title>Ferry</title></head><body>
<main style="max-width:460px;padding-top:12vh"><div class="card" style="padding:24px">
<h1 style="margin:0 0 8px;font-size:20px">Open Ferry again</h1>
<p class="muted" style="margin:0">This page only works when opened by the Ferry app. Open Ferry from your apps and it will bring you back here.</p>
</div></main></body></html>{{end}}

{{define "panel"}}{{template "head"}}<title>Ferry</title>
<style>
/* The window is the app: a header with the pages, then a list on the left and
   a detail panel on the right, and a status bar along the bottom. */
[hidden]{display:none!important}
html,body{height:100%}
body{overflow:hidden}
#intro{position:fixed;inset:0;z-index:40;background:var(--bg);display:none;place-items:center;cursor:pointer;transition:opacity .45s ease}
#intro.play{display:grid}
#intro.out{opacity:0}
#intro svg{width:min(440px,88vw);height:auto;overflow:visible}
#intro .pc{fill:none;stroke:var(--text);stroke-width:3;stroke-linejoin:round;stroke-linecap:round}
#intro .screen{fill:var(--panel);transition:fill .35s ease}
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
#intro .cabin{fill:var(--panel);stroke:var(--text);stroke-width:2.5}
#intro .stack{fill:var(--accent)}
#intro .name{font:700 22px/1 "Segoe UI",system-ui,-apple-system,sans-serif;fill:var(--text);letter-spacing:.02em;opacity:0;animation:fadein .4s ease 1.6s forwards}
@keyframes cross{from{transform:translateX(0)}to{transform:translateX(158px)}}
@keyframes bob{from{transform:translateY(0) rotate(-2.5deg)}to{transform:translateY(-3px) rotate(2.5deg)}}
@keyframes drift{from{transform:translateX(0)}to{transform:translateX(-40px)}}
@keyframes fadein{to{opacity:1}}
.app.enter{animation:enter .45s ease both}
@keyframes enter{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:none}}

.app{height:100%;display:flex;flex-direction:column}
.app>header{display:flex;gap:6px;align-items:center;padding:10px 16px;border-bottom:1px solid var(--line);background:var(--panel);flex:none}
.app>header h1{font-size:16px;margin:0 14px 0 0;display:flex;gap:8px;align-items:center}
.logo{width:20px;height:20px;flex:none}
.logo rect{fill:var(--accent)}
.logo path{fill:none;stroke:var(--accent-text);stroke-width:3.5;stroke-linecap:round}
.grow{flex:1}
button.on{background:var(--sel);border-color:var(--accent);color:var(--accent);font-weight:600}
.banner{display:flex;gap:10px;align-items:center;padding:8px 16px;background:var(--sel);border-bottom:1px solid var(--line);flex:none}
.banner span{flex:1}
.app>main.page{display:none;flex:1;min-height:0;max-width:none;margin:0;padding:0}
body[data-page=share] #page-share,body[data-page=receive] #page-receive{display:grid;grid-template-columns:minmax(0,1fr) 380px}
body[data-page=transfer] #page-transfer{display:flex;flex-direction:column}
.list{overflow:auto;min-width:0;min-height:0}
.bar{display:flex;gap:10px;align-items:center;padding:8px 12px;border-bottom:1px solid var(--line);min-height:41px}
.label{font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:.04em;color:var(--dim);white-space:nowrap}
.sub{color:var(--dim);font-size:12px}
.mono{font-family:var(--mono)}
.empty{padding:48px 24px;text-align:center;color:var(--dim)}
table.files{width:100%;border-collapse:collapse;table-layout:fixed}
table.files th{position:sticky;top:0;z-index:1;background:var(--bg);text-align:left;font-weight:600;color:var(--dim);font-size:12px;text-transform:uppercase;letter-spacing:.04em;padding:7px 12px;border-bottom:1px solid var(--line)}
table.files td{padding:7px 12px;border-bottom:1px solid var(--line);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
table.files .num{text-align:right;font-variant-numeric:tabular-nums}
table.files td.num{color:var(--dim);font-size:13px}
table.files .sz{width:84px}
table.files .dt{width:140px}
table.files .ck{width:34px;padding-right:0}
table.files .addr{width:190px}
table.files td.mono{font-size:13px}
table.files tbody tr:hover{background:var(--hover)}
table.files tr.sel,table.files tr.sel:hover{background:var(--sel)}
table.pick tbody tr,.pane table.files tbody tr{cursor:pointer;user-select:none}
.ico{width:15px;height:15px;vertical-align:-3px;margin-right:8px;color:var(--dim)}
tr.dir .nm{font-weight:600}
aside{border-left:1px solid var(--line);background:var(--panel);overflow:auto;padding:16px;min-width:0;overflow-wrap:anywhere}
aside h2{margin:0 0 2px;font-size:18px}
.stat{display:flex;gap:28px;margin:12px 0 2px;flex-wrap:wrap}
.stat div{color:var(--dim);font-size:12px}
.stat b{display:block;font-size:20px;color:var(--text)}
.stat .pin b{font-family:var(--mono);letter-spacing:.14em;font-size:26px}
.section{font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:.04em;color:var(--dim);margin:20px 0 6px}
.field{margin:12px 0}
.field>label{display:block;font-weight:600;margin-bottom:3px}
.field input,.field select{width:100%}
.hint{color:var(--dim);font-size:12px;margin:3px 0 0}
.check{display:flex;gap:9px;align-items:flex-start;margin:12px 0;cursor:pointer}
.check input{margin:3px 0 0;flex:none}
.check .hint{display:block;margin:0}
.row{display:flex;gap:6px}
.row input{flex:1;min-width:0}
.row-btns{display:flex;gap:8px;flex-wrap:wrap;margin-top:18px}
.seg{display:flex;gap:6px;margin:14px 0 4px}
.dot{width:8px;height:8px;border-radius:50%;background:var(--ok);display:inline-block;flex:none}
.state{display:flex;gap:6px;align-items:center;font-size:12px;font-weight:600;color:var(--dim);text-transform:uppercase;letter-spacing:.04em;margin-bottom:4px}
.badge{display:inline-flex;align-items:center;gap:3px;font-size:11px;font-weight:700;line-height:1;padding:3px 5px;border-radius:3px;text-transform:uppercase;letter-spacing:.03em}
.badge svg{width:11px;height:11px;flex:none}
.badge.low{background:var(--low-fill);color:var(--low-ink)}
.badge.out{background:var(--out-fill);color:var(--out-ink)}
.lowtxt{color:var(--low)}
.linkrow{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:14px;align-items:start}
.url{font:13px/1.4 var(--mono);overflow-wrap:anywhere;margin-bottom:8px}
.linkrow img{width:120px;height:120px;display:block;background:#fff;padding:6px;border:1px solid var(--line);border-radius:4px;image-rendering:pixelated}
table.log{width:100%;border-collapse:collapse}
table.log td{padding:5px 4px;font-size:12px;border-bottom:1px solid var(--line);vertical-align:top;overflow-wrap:anywhere}
table.log td:first-child{color:var(--dim);font-variant-numeric:tabular-nums;width:1%;white-space:nowrap;padding-right:10px}
table.log tr:last-child td{border-bottom:0}
.app>footer{display:flex;gap:6px;align-items:center;padding:4px 16px;color:var(--dim);font-size:12px;border-top:1px solid var(--line);background:var(--panel);white-space:nowrap;overflow:hidden;flex:none}
.app>footer span:last-child{overflow:hidden;text-overflow:ellipsis}
.error{margin:10px 0 0;font-size:13px}

.connbar{display:flex;gap:10px;align-items:center;padding:8px 16px;border-bottom:1px solid var(--line);flex:none;min-height:45px}
.connbar>span{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.connbar .on-line{display:flex;gap:6px;align-items:center}
.panes{flex:1;min-height:0;display:grid;grid-template-columns:minmax(0,1fr) minmax(0,1fr)}
.pane{display:flex;flex-direction:column;min-height:0;min-width:0;background:var(--panel)}
.pane+.pane{border-left:1px solid var(--line)}
.pane.drop{outline:2px dashed var(--accent);outline-offset:-3px}
.pane-head{display:flex;gap:6px;align-items:center;padding:8px 12px;border-bottom:1px solid var(--line)}
.pane-head .label{margin-right:4px}
.pane-head input{flex:1;min-width:0;font:12px var(--mono);padding:4px 7px}
.pane-list{flex:1;overflow:auto;min-height:0}
.pane table.files th{background:var(--panel)}
.pane-foot{display:flex;gap:6px;align-items:center;padding:8px 12px;border-top:1px solid var(--line);flex-wrap:wrap}
.pane-msg{padding:32px 16px;color:var(--dim);text-align:center}
.files .dirbtn{background:none;border:0;padding:0;color:var(--text);font-weight:600;cursor:pointer;max-width:calc(100% - 23px);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;text-align:left;vertical-align:bottom;border-radius:2px}
.files .dirbtn:hover{text-decoration:underline}
.queue{flex:none;border-top:1px solid var(--line);background:var(--panel);max-height:200px;overflow:auto}
.queue .bar{position:sticky;top:0;background:var(--panel);z-index:1}
.queue h3{margin:0;font-size:14px}
table.q{width:100%;border-collapse:collapse;table-layout:fixed;font-size:13px}
table.q td{padding:6px 12px;border-bottom:1px solid var(--line);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;vertical-align:middle}
table.q .dirn{width:100px;font-size:12px;font-weight:600;color:var(--dim);text-transform:uppercase;letter-spacing:.04em}
table.q .mt{width:150px}
table.q .st{width:140px;text-align:right;color:var(--dim);font-variant-numeric:tabular-nums}
table.q .act{width:84px;text-align:right}
table.q tr.has-err td{border-bottom:0}
table.q tr.err td{padding-top:0;color:var(--out);font-size:12px;white-space:normal}
.meter{height:4px;background:var(--track);border-radius:2px;overflow:hidden}
.meter i{display:block;height:100%;width:0;background:var(--ok);border-radius:2px}

dialog{border:1px solid var(--line);border-radius:6px;background:var(--panel);color:var(--text);padding:0;width:min(480px,92vw);max-height:88vh;overflow-wrap:anywhere}
dialog::backdrop{background:rgba(0,0,0,.35)}
dialog form{padding:18px 20px}
dialog h3{margin:0 0 10px;font-size:16px}
.two{display:grid;grid-template-columns:1fr 1fr;gap:10px}
.two.port{grid-template-columns:minmax(0,1fr) 96px}
.two .field{margin:0 0 12px}
.dlg-btns{display:flex;gap:8px;align-items:center;margin-top:18px}
.fp{font:12px/1.5 var(--mono);overflow-wrap:anywhere;background:var(--bg);border:1px solid var(--line);border-radius:4px;padding:8px 10px;margin:8px 0}
.recent{display:flex;flex-wrap:wrap;gap:6px}
.recent .item{display:inline-flex;border:1px solid var(--line);border-radius:4px;overflow:hidden}
.recent .item button{border:0;border-radius:0;font-size:13px;padding:3px 9px}
.recent .item button+button{border-left:1px solid var(--line);color:var(--dim)}
.plain-list{list-style:none;margin:6px 0 0;padding:0}
.plain-list li{display:flex;justify-content:space-between;align-items:center;gap:10px;padding:6px 0;border-top:1px solid var(--line);font-size:13px}
.plain-list li:last-child{border-bottom:1px solid var(--line)}
.logbox{max-height:150px;overflow:auto;border:1px solid var(--line);border-radius:4px;padding:2px 8px}
#toast{position:fixed;bottom:36px;left:50%;transform:translateX(-50%);background:var(--text);color:var(--bg);padding:8px 14px;border-radius:4px;max-width:80vw;display:none;z-index:20}
#toast.show{display:block}
#toast.err{background:var(--out-fill);color:var(--out-ink)}
#offline{position:fixed;inset:0;background:rgba(0,0,0,.45);display:none;place-items:center;z-index:30;padding:16px}
#offline.show{display:grid}
#offline>div{background:var(--panel);border:1px solid var(--line);border-radius:6px;padding:18px 20px;max-width:420px}

@media (max-width:760px){
  body{overflow:auto}
  .app{height:auto;min-height:100%}
  .app>header{flex-wrap:wrap}
  body[data-page=share] #page-share,body[data-page=receive] #page-receive{grid-template-columns:1fr}
  .list{max-height:45vh}
  aside{border-left:0;border-top:1px solid var(--line)}
  .panes{grid-template-columns:1fr}
  .pane{height:360px}
  .pane+.pane{border-left:0;border-top:1px solid var(--line)}
  table.files .dt,table.q .mt{display:none}
}
</style></head>
<body data-page="share">
<svg width="0" height="0" style="position:absolute" aria-hidden="true">
<symbol id="i-folder" viewBox="0 0 16 16"><path d="M1.75 3.5h4.1l1.6 1.6h6.8v7.4H1.75z" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></symbol>
<symbol id="i-file" viewBox="0 0 16 16"><path d="M3.75 1.75h5.5l3 3v9.5h-8.5zM9.25 1.75v3h3" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></symbol>
</svg>
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

<div class="app">
<header>
<h1><svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="6"/><path d="M11 7.5v17M11 7.5h11M11 16h8"/></svg>Ferry</h1>
<button type="button" data-go="share" class="on">Share</button>
<button type="button" data-go="receive">Receive</button>
<button type="button" data-go="transfer">Transfer files</button>
<span class="grow"></span>
<button type="button" id="open-settings">Settings</button>
</header>

<div class="banner" id="update" hidden><span id="update-text"></span>
<button type="button" class="primary" id="update-go">Restart to update</button><button type="button" id="update-later">Later</button></div>

<main class="page" id="page-share">
<div class="list">
<div class="bar"><span class="label">What visitors see</span><span class="sub grow" id="preview-count"></span><button type="button" class="small" id="preview-refresh">Refresh</button></div>
<div id="preview"></div>
</div>
<aside>
<div id="setup">
<h2 id="setup-title">Share a folder</h2>
<div class="sub">Anyone with the link and the PIN can open it in a browser.</div>
<div class="seg"><button type="button" data-mode="files">A folder</button><button type="button" data-mode="site">A website</button></div>
<div id="files-opts">
<div class="field"><label for="folder">Folder</label>
<div class="row"><input type="text" id="folder" placeholder="Choose a folder, or paste its path"><button type="button" id="browse">Browse</button></div></div>
<label class="check"><input type="checkbox" id="uploads"><span>Let visitors upload files into this folder</span></label>
<label class="check"><input type="checkbox" id="keeppin"><span>Keep the same PIN when sharing starts again<span class="hint">Needed if another computer keeps a copy of this folder up to date.</span></span></label>
</div>
<div id="site-opts" hidden>
<div class="field"><label for="target">Port your website is running on</label>
<input type="number" id="target" min="1" max="65535" style="width:140px">
<p class="hint">Start the website on this computer first. Common ports: 3000, 5173, 8000.</p></div>
</div>
<label class="check"><input type="checkbox" id="internet"><span>Also allow access from outside this Wi-Fi<span class="hint">Gives a free link that works anywhere. Visitors still need the PIN.</span></span></label>
<p class="error" id="err" hidden></p>
<div class="row-btns"><button type="button" class="primary" id="start">Start sharing</button></div>
</div>

<div id="live" hidden>
<div class="state"><span class="dot"></span>Sharing now</div>
<h2 id="live-name"></h2>
<div class="sub mono" id="what"></div>
<div class="stat"><div class="pin">PIN<b id="pin"></b></div><div id="uploads-stat">Uploads<b id="uploads-val"></b></div></div>
<p class="hint">Visitors type the PIN when they open the link.</p>
<div id="links"></div>
<p class="hint" id="net" hidden></p>
<p class="hint" style="margin-top:10px">Scanning a code logs in without the PIN, so only show codes to people you trust. Keep this computer awake while sharing.</p>
<div class="section">Recent activity</div>
<table class="log"><tbody id="log"></tbody></table>
<div class="row-btns"><button type="button" class="danger" id="stop">Stop sharing</button></div>
</div>
</aside>
</main>

<main class="page" id="page-receive">
<div class="list">
<div class="bar"><span class="label">Computers on this Wi-Fi</span><span class="sub grow" id="found-sum"></span><button type="button" class="small" id="find">Look again</button></div>
<table class="files pick" id="found-table" hidden><thead><tr><th style="width:30%">Computer</th><th>Shared folder</th><th class="addr">Address</th></tr></thead><tbody id="found"></tbody></table>
<div class="empty" id="found-empty">Ferry looks for other computers sharing a folder on this Wi-Fi.</div>
</div>
<aside>
<div id="receive-setup">
<h2>Keep a folder up to date</h2>
<div class="sub">Copies a folder shared by another Ferry computer and keeps it the same.</div>
<div class="field"><label for="server">Address of the other computer</label>
<input type="text" id="server" class="mono" placeholder="For example 192.168.1.20:8080" autocomplete="off" spellcheck="false">
<p class="hint">Pick a computer from the list, or type the address shown in the Ferry window on it. Both computers need to be on the same Wi-Fi.</p></div>
<div class="field"><label for="rfolder">Folder to keep up to date</label>
<div class="row"><input type="text" id="rfolder" placeholder="Choose a folder, or paste its path"><button type="button" id="rbrowse">Browse</button></div></div>
<div class="field"><label for="rpin">PIN from the other computer</label>
<input type="text" id="rpin" class="mono" inputmode="numeric" autocomplete="off" maxlength="12" style="width:140px"></div>
<p class="hint">Files in this folder are kept the same as on the other computer, so changes you make to them here are overwritten. Files removed there move to a folder called .ferry-removed here. Files you put in this folder yourself are left alone.</p>
<p class="error" id="rerr" hidden></p>
<div class="row-btns"><button type="button" class="primary" id="rstart">Start receiving</button></div>
</div>

<div id="mirror-live" hidden>
<div class="state" id="mirror-state"></div>
<h2 id="mirror-name"></h2>
<div class="sub mono" id="mirror-what"></div>
<div class="stat"><div>Files there<b id="m-files"></b></div><div>Updated<b id="m-changed"></b></div><div>Moved aside<b id="m-removed"></b></div></div>
<p class="hint" id="mirror-status"></p>
<div class="row-btns"><button type="button" class="danger" id="mirror-stop">Stop receiving</button></div>
</div>
</aside>
</main>

<main class="page" id="page-transfer">
<div class="connbar">
<span class="grow sub" id="conn-off">Copy files to and from an FTP or SFTP server, like a website host, a NAS drive or a Raspberry Pi.</span>
<span class="grow on-line" id="conn-on" hidden><span class="dot"></span><span>Connected to <b id="clabel"></b> <span class="sub" id="cproto-label"></span></span></span>
<button type="button" class="primary small" id="open-connect">Connect to a server</button>
<button type="button" class="small" id="disconnect" hidden>Disconnect</button>
</div>
<div class="panes">
<div class="pane" id="pane-local" data-side="local">
<div class="pane-head"><span class="label">This computer</span><button type="button" class="small" data-act="up" title="Folder above">Up</button><input type="text" data-act="path" spellcheck="false" aria-label="Folder on this computer"><button type="button" class="small" data-act="refresh">Refresh</button></div>
<div class="pane-list"></div>
<div class="pane-foot"><span class="sub grow count"></span><button type="button" class="primary small" data-act="send">Upload to server</button></div>
</div>
<div class="pane" id="pane-remote" data-side="remote">
<div class="pane-head"><span class="label">Server</span><button type="button" class="small" data-act="up" title="Folder above">Up</button><input type="text" data-act="path" spellcheck="false" aria-label="Folder on the server"><button type="button" class="small" data-act="refresh">Refresh</button></div>
<div class="pane-list"></div>
<div class="pane-foot"><button type="button" class="small" data-act="mkdir">New folder</button><button type="button" class="small" data-act="rename">Rename</button><button type="button" class="small danger" data-act="delete">Delete</button><span class="sub grow count" style="text-align:right"></span><button type="button" class="primary small" data-act="send">Download here</button></div>
</div>
</div>
<div class="queue" id="xfer-queue" hidden>
<div class="bar"><h3>Transfers</h3><span class="sub grow" id="q-sum"></span><button type="button" class="small" id="q-cancel">Cancel all</button><button type="button" class="small" id="q-clear">Clear finished</button></div>
<table class="q"><tbody id="queue"></tbody></table>
</div>
</main>

<footer id="foot"></footer>
</div>

<dialog id="settings"><form method="dialog">
<div style="display:flex;gap:10px;align-items:baseline"><h3 class="grow">Settings</h3><span class="sub" id="prefs-saved"></span></div>
<div class="section" style="margin-top:4px">Updates</div>
<div style="display:flex;gap:8px;align-items:center"><span class="grow">This is Ferry {{.Version}}.</span>
<button type="button" class="small" id="check">Check for updates</button><button type="button" class="small primary" id="check-install" hidden>Restart to update</button></div>
<p class="hint" id="check-status"></p>
<label class="check"><input type="checkbox" id="autoupdate"><span>Check for updates automatically<span class="hint">When Ferry opens and every 6 hours while it runs. Nothing is installed until you choose to restart.</span></span></label>
<div class="section">Network</div>
<label class="check"><input type="checkbox" id="findable"><span>Let other computers on this Wi-Fi find this one<span class="hint">Only while you share a folder. They still need the PIN to open it.</span></span></label>
<div class="field"><label for="port">Sharing port</label>
<input type="number" id="port" min="1024" max="65535" placeholder="Automatic" style="width:160px">
<p class="hint">Leave empty to pick one automatically. Set a port if you want the address for this computer to stay the same each time. Used the next time sharing starts.</p></div>
<p class="error" id="prefs-err" hidden></p>
<div class="section">Trusted servers</div>
<p class="hint">Servers you chose to trust on the Transfer files page. Remove one if its owner tells you its key or certificate changed.</p>
<ul class="plain-list" id="trusted"></ul>
<div class="section">Activity</div>
<div class="logbox"><table class="log"><tbody id="all-log"></tbody></table></div>
<div class="section">About</div>
<p class="hint">Ferry keeps its settings on this computer only. <a href="https://github.com/S1xinch/Ferry---Local-File-Sharing/releases" target="_blank" rel="noopener">See what changed in each version</a>.</p>
<div class="dlg-btns"><button type="button" class="danger" id="quit">Quit Ferry</button><span class="grow"></span><button class="primary" value="done">Done</button></div>
</form></dialog>

<dialog id="connect-dlg"><form id="cform" autocomplete="off">
<h3>Connect to a server</h3>
<div class="field"><label for="cproto">Kind of server</label>
<select id="cproto">
<option value="sftp">SFTP (SSH)</option>
<option value="ftpes">FTP over TLS</option>
<option value="ftps">FTPS (implicit TLS)</option>
<option value="ftp">FTP, not encrypted</option>
</select>
<p class="hint" id="cproto-hint"></p></div>
<div class="two port"><div class="field"><label for="chost">Server</label><input type="text" id="chost" placeholder="ftp.example.com" spellcheck="false"></div>
<div class="field"><label for="cport">Port</label><input type="number" id="cport" min="1" max="65535" placeholder="22"></div></div>
<div class="two"><div class="field"><label for="cuser">User name</label><input type="text" id="cuser" spellcheck="false" autocapitalize="off"></div>
<div class="field"><label for="cpass">Password</label><input type="password" id="cpass"></div></div>
<div id="recent-box" hidden><div class="section" style="margin-top:4px">Recent connections</div>
<div class="recent" id="recent"></div>
<p class="hint" style="margin-top:6px">Ferry remembers the server, user name and last folder. Passwords are not saved.</p></div>
<p class="error" id="connect-err" hidden></p>
<div class="dlg-btns"><span class="grow"></span><button type="button" id="connect-cancel">Cancel</button><button type="submit" class="primary" id="connect">Connect</button></div>
</form></dialog>

<dialog id="trust-dlg"><form method="dialog">
<h3 id="trust-title">New server</h3>
<p id="trust-text" style="margin:0"></p>
<div class="fp" id="trust-fp"></div>
<p class="hint">Only trust it if this is the server you meant to reach. Ferry remembers your answer and warns you if this ever changes.</p>
<div class="dlg-btns"><span class="grow"></span><button type="button" id="trust-no">Cancel</button><button type="button" class="primary" id="trust-yes">Trust and connect</button></div>
</form></dialog>

<dialog id="ask"><form onsubmit="return false">
<h3 id="ask-title"></h3>
<p id="ask-text" style="margin:0;white-space:pre-line"></p>
<input type="text" id="ask-input" spellcheck="false" style="width:100%;margin-top:10px" aria-labelledby="ask-title" hidden>
<div class="dlg-btns"><span class="grow"></span><button type="button" id="ask-no">Cancel</button><button type="button" id="ask-ok">OK</button></div>
</form></dialog>

<div id="toast" role="status"></div>
<div id="offline"><div><b>Ferry is not running</b><p id="offline-text" style="margin:6px 0 0"></p></div></div>

<script>
var $ = function (id) { return document.getElementById(id); };
var version = {{.Version}};
var mode = "files", formFilled = false, firstRender = true, lastLinks = "", running = false, receiving = false, last = null;
var previewFolder = "", closing = "", laterVersion = "", lastActivity = "";
var BADGE_LOW = '<span class="badge low"><svg viewBox="0 0 12 12" aria-hidden="true"><path d="M6 1.2L11.2 10.5H0.8Z" fill="currentColor" opacity=".25"/><path d="M6 4.3v3M6 8.9v.1" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>Waiting</span>';
var ICON_OUT = '<svg viewBox="0 0 12 12" aria-hidden="true"><circle cx="6" cy="6" r="5.5" fill="currentColor" opacity=".25"/><path d="M3.8 3.8l4.4 4.4M8.2 3.8L3.8 8.2" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>';

function cap(s) { return s ? s.charAt(0).toUpperCase() + s.slice(1) : ""; }
function el(tag, cls, text) { var e = document.createElement(tag); if (cls) e.className = cls; if (text) e.textContent = text; return e; }
function base(p) { var parts = String(p || "").split(/[\\\/]/).filter(Boolean); return parts.length ? parts[parts.length - 1] : String(p || ""); }
function ico(dir) {
  var ns = "http://www.w3.org/2000/svg", s = document.createElementNS(ns, "svg"), u = document.createElementNS(ns, "use");
  s.setAttribute("class", "ico"); s.setAttribute("aria-hidden", "true");
  u.setAttribute("href", dir ? "#i-folder" : "#i-file");
  s.appendChild(u);
  return s;
}

function call(path, body) {
  return fetch(path, {
    method: body === undefined && path === "/api/state" ? "GET" : "POST",
    headers: {"X-Ferry": "1", "Content-Type": "application/json"},
    body: body === undefined ? undefined : JSON.stringify(body)
  }).then(function (r) {
    return r.json().then(function (j) { if (!r.ok) throw new Error(j.error || "Something went wrong"); return j; });
  });
}

var toastTimer = null;
function toast(msg, err) {
  var t = $("toast");
  t.textContent = cap(msg);
  t.className = err ? "show err" : "show";
  clearTimeout(toastTimer);
  toastTimer = setTimeout(function () { t.className = ""; }, err ? 6000 : 2200);
}

// Links to websites open in the person's own browser, not inside the window.
document.addEventListener("click", function (ev) {
  var a = ev.target.closest && ev.target.closest("a[target=_blank]");
  if (!a) return;
  ev.preventDefault();
  call("/api/open", {url: a.href}).catch(function () { window.open(a.href); });
});

function offline(msg) { $("offline-text").textContent = msg; $("offline").classList.add("show"); }

// ask shows a question inside the window. It resolves to true, or the typed
// text when o.value is set, for OK, and to null for Cancel or Escape.
function ask(o) {
  return new Promise(function (done) {
    var d = $("ask"), input = $("ask-input"), ok = $("ask-ok"), typed = o.value !== undefined;
    // The answer is settled here rather than on the dialog's close event,
    // which Chrome does not always fire when the dialog is reused.
    function finish(yes) { if (d.open) d.close(); done(yes ? (typed ? input.value.trim() : true) : null); }
    $("ask-title").textContent = o.title;
    $("ask-text").textContent = o.text || "";
    $("ask-text").hidden = !o.text;
    input.hidden = !typed;
    input.value = o.value || "";
    ok.textContent = o.ok || "OK";
    ok.className = o.danger ? "danger" : "primary";
    ok.onclick = function () { finish(true); };
    input.onkeydown = function (ev) { if (ev.key === "Enter") { ev.preventDefault(); finish(true); } };
    $("ask-no").onclick = function () { finish(false); };
    d.oncancel = function () { finish(false); };
    d.showModal();
    if (typed) { input.focus(); input.select(); } else ok.focus();
  });
}
function askIf(needed, o) { return needed ? ask(o) : Promise.resolve(true); }

/* ---------- pages ---------- */
var page = "share", visited = {};
function go(p) {
  page = p;
  document.body.dataset.page = p;
  document.querySelectorAll("[data-go]").forEach(function (b) { b.classList.toggle("on", b.dataset.go === p); });
  if (p === "transfer") {
    if (!xfer.local.loaded) loadPane("local", "");
    if (xfer.connected && !xfer.remote.loaded) loadPane("remote", "");
  }
  if (p === "receive" && !visited.receive && !receiving) findComputers();
  visited[p] = true;
  if (last) renderFoot(last);
}
document.querySelectorAll("[data-go]").forEach(function (b) { b.onclick = function () { go(b.dataset.go); }; });

function renderFoot(s) {
  var f = $("foot"), parts = [], mr = s.mirror;
  var active = ((s.remote && s.remote.transfers) || []).filter(function (t) { return t.state === "queued" || t.state === "running"; }).length;
  if (s.running) parts.push("Sharing " + (s.settings.mode === "site" ? "a website" : base(s.what)));
  if (mr) parts.push(mr.lastError ? "Receiving, waiting to reconnect" : "Receiving into " + base(mr.folder));
  if (active) parts.push(active + (active === 1 ? " transfer" : " transfers") + " to go");
  if (!parts.length) parts.push("Not sharing");
  parts.push("Version " + version);
  if (page === "transfer") parts.push("Click to select, double-click a folder to open it, drag files from one side to the other, shift-click for a range");
  f.textContent = "";
  if (s.running || (mr && !mr.lastError)) f.appendChild(el("span", "dot"));
  f.appendChild(el("span", "", parts.join(" \u00a0|\u00a0 ")));
}

/* ---------- share ---------- */
function setMode(m) {
  mode = m;
  document.querySelectorAll(".seg button").forEach(function (b) { b.classList.toggle("on", b.dataset.mode === m); });
  $("files-opts").hidden = m !== "files";
  $("site-opts").hidden = m !== "site";
  $("setup-title").textContent = m === "site" ? "Share a website" : "Share a folder";
  showPreview();
}
document.querySelectorAll(".seg button").forEach(function (b) { b.onclick = function () { setMode(b.dataset.mode); }; });

function showError(id, msg) { $(id).textContent = cap(msg); $(id).hidden = !msg; }

function browseInto(button, input, done) {
  button.disabled = true;
  call("/api/pick", {}).then(function (r) { if (r.folder) { input.value = r.folder; if (done) done(); } })
    .catch(function (e) { toast(e.message, true); })
    .finally(function () { button.disabled = false; });
}
function folderChanged() { previewFolder = $("folder").value.trim().replace(/^["']+|["']+$/g, ""); showPreview(); }
$("browse").onclick = function () { browseInto($("browse"), $("folder"), folderChanged); };
$("folder").onchange = folderChanged;
$("target").onchange = function () { showPreview(); };
$("preview-refresh").onclick = function () { showPreview(true); };

// The list shows what visitors will see: the folder being shared, or the one
// chosen in the form.
var previewKey = "", previewSeq = 0;
function showPreview(force) {
  var site = running ? last.settings.mode === "site" : mode === "site";
  var folder = running ? last.what : previewFolder;
  var port = running ? last.settings.target : parseInt($("target").value, 10) || 0;
  var key = site ? "site:" + port : "folder:" + folder;
  if (!force && key === previewKey) return;
  previewKey = key;
  var box = $("preview"), seq = ++previewSeq;
  function message(text) { box.textContent = ""; box.appendChild(el("div", "empty", text)); $("preview-count").textContent = ""; }
  $("preview-refresh").hidden = site;
  if (site) { message(port ? "Visitors see the website running on port " + port + " of this computer." : "Type the port your website is running on."); return; }
  if (!folder) { message("Choose a folder to see what visitors will see."); return; }
  call("/api/local/list", {path: folder, name: ""}).then(function (l) {
    if (seq !== previewSeq) return;
    var es = l.entries || [];
    $("preview-count").textContent = es.length + (es.length === 1 ? " item" : " items");
    if (!es.length) { box.textContent = ""; box.appendChild(el("div", "empty", "This folder is empty.")); return; }
    var table = el("table", "files"), head = el("thead"), hr = el("tr"), body = el("tbody");
    [["Name", ""], ["Size", "sz num"], ["Modified", "dt num"]].forEach(function (h) { hr.appendChild(el("th", h[1], h[0])); });
    head.appendChild(hr);
    table.appendChild(head);
    es.forEach(function (e) {
      var tr = el("tr", e.dir ? "dir" : ""), nm = el("td", "nm");
      nm.appendChild(ico(e.dir));
      nm.appendChild(document.createTextNode(e.name));
      nm.title = e.name;
      tr.appendChild(nm);
      tr.appendChild(el("td", "num", e.dir ? "" : fmtSize(e.size)));
      tr.appendChild(el("td", "num dt", fmtDate(e.mod)));
      body.appendChild(tr);
    });
    table.appendChild(body);
    box.textContent = "";
    box.appendChild(table);
  }).catch(function (e) {
    if (seq === previewSeq) message(cap(e.message));
  });
}

$("start").onclick = function () {
  showError("err", "");
  $("start").disabled = true;
  call("/api/start", {
    mode: mode, folder: $("folder").value, target: parseInt($("target").value, 10) || 0,
    uploads: $("uploads").checked, internet: $("internet").checked, keepPin: $("keeppin").checked
  }).then(refresh).catch(function (e) { showError("err", e.message); })
    .finally(function () { $("start").disabled = false; });
};
$("stop").onclick = function () { call("/api/stop", {}).then(function () { formFilled = false; refresh(); }); };

function copyLink(url) {
  navigator.clipboard.writeText(url).then(function () { toast("Link copied"); },
    function () { toast("Could not copy. Select the link and copy it yourself.", true); });
}

function renderLinks(links) {
  var key = JSON.stringify(links.map(function (l) { return l.url; }));
  if (key === lastLinks) return;
  lastLinks = key;
  var box = $("links");
  box.textContent = "";
  links.forEach(function (l) {
    box.appendChild(el("div", "section", l.label));
    var row = el("div", "linkrow"), left = el("div");
    left.appendChild(el("div", "url", l.url));
    var copy = el("button", "small", "Copy link");
    copy.type = "button";
    copy.onclick = function () { copyLink(l.url); };
    left.appendChild(copy);
    row.appendChild(left);
    var img = el("img");
    img.src = l.qr;
    img.alt = "QR code for the " + l.label + " link";
    row.appendChild(img);
    box.appendChild(row);
  });
}

function drawLog(tbody, lines, n) {
  tbody.textContent = "";
  if (!lines.length) lines = ["Nothing yet."];
  lines.slice(0, n).forEach(function (line) {
    var i = line.indexOf("  "), tr = el("tr");
    tr.appendChild(el("td", "", i > 0 ? line.slice(0, i) : ""));
    tr.appendChild(el("td", "", i > 0 ? line.slice(i + 2) : line));
    tbody.appendChild(tr);
  });
}

/* ---------- receive ---------- */
var found = [];
function findComputers() {
  $("find").disabled = true;
  $("found-sum").textContent = "Looking...";
  call("/api/discover", {}).then(function (r) { found = r.computers || []; drawFound(); })
    .catch(function (e) { found = []; drawFound(cap(e.message)); })
    .finally(function () { $("find").disabled = false; });
}
$("find").onclick = findComputers;

function drawFound(errMsg) {
  var n = found.length, body = $("found");
  $("found-sum").textContent = errMsg ? "" : n ? n + (n === 1 ? " computer found" : " computers found") : "None found";
  $("found-table").hidden = !n;
  $("found-empty").hidden = !!n;
  $("found-empty").textContent = errMsg || "No Ferry computers answered. Check that both are on the same Wi-Fi, that the other one is sharing a folder, and that its firewall lets Ferry in. You can also type the address shown on the other computer.";
  body.textContent = "";
  found.forEach(function (c) {
    var tr = el("tr", $("server").value.trim() === c.addr ? "sel" : "");
    tr.appendChild(el("td", "", c.name));
    var f = el("td", "", c.folder);
    f.title = c.folder;
    tr.appendChild(f);
    tr.appendChild(el("td", "mono", c.addr));
    tr.onclick = function () { $("server").value = c.addr; drawFound(); $("rpin").focus(); };
    body.appendChild(tr);
  });
}

$("rbrowse").onclick = function () { browseInto($("rbrowse"), $("rfolder")); };
$("rstart").onclick = function () {
  showError("rerr", "");
  var b = $("rstart");
  b.disabled = true;
  b.textContent = "Connecting...";
  call("/api/start", {mode: "receive", server: $("server").value, mirrorFolder: $("rfolder").value, receivePin: $("rpin").value})
    .then(function () { $("rpin").value = ""; return refresh(); })
    .catch(function (e) { showError("rerr", e.message); })
    .finally(function () { b.disabled = false; b.textContent = "Start receiving"; });
};
$("mirror-stop").onclick = function () { call("/api/mirror/stop", {}).then(refresh); };

function renderMirror(mr) {
  receiving = !!mr;
  $("receive-setup").hidden = !!mr;
  $("mirror-live").hidden = !mr;
  if (!mr) return;
  $("mirror-state").innerHTML = mr.lastError ? BADGE_LOW : '<span class="dot"></span>Receiving now';
  $("mirror-name").textContent = base(mr.folder);
  $("mirror-what").textContent = "From " + mr.server + " into " + mr.folder;
  $("m-files").textContent = mr.lastSync ? mr.files : "-";
  $("m-changed").textContent = mr.lastSync ? mr.changed : "-";
  $("m-removed").textContent = mr.lastSync ? mr.removed : "-";
  var st = $("mirror-status");
  st.className = mr.lastError ? "hint lowtxt" : "hint";
  st.textContent = mr.lastError ? "Waiting to reconnect: " + mr.lastError
    : mr.lastSync ? "Last checked " + mr.lastSync + ". Updated and moved aside count the files changed at that check. Files removed there are in .ferry-removed here."
    : "Connecting...";
}

/* ---------- updates, settings, quit ---------- */
$("quit").onclick = function () {
  askIf(running || receiving || xfer.busy, {title: "Quit Ferry?", text: "Quitting stops sharing, receiving and any transfers.", ok: "Quit Ferry", danger: true})
    .then(function (yes) {
      if (!yes) return;
      closing = "Ferry has closed. You can close this window.";
      call("/api/quit", {}).finally(function () { offline(closing); });
    });
};

$("update-go").onclick = function () {
  askIf(running, {title: "Update now?", text: "Updating restarts Ferry and stops sharing.", ok: "Restart to update"})
    .then(function (yes) { if (yes) updateNow(); });
};
function updateNow() {
  $("update-go").disabled = true;
  closing = "Ferry is restarting with the new version. Its window will open in a moment.";
  call("/api/update", {}).then(function () { offline(closing); }).catch(function (e) {
    $("update-go").disabled = false;
    // A network error here means Ferry closed partway through.
    var msg = e instanceof TypeError
      ? "Ferry closed while updating. Open Ferry again. If it still shows the old version, download the new one from GitHub."
      : cap(e.message);
    closing = e instanceof TypeError ? msg : "";
    $("update-text").textContent = msg;
    $("check-status").textContent = msg;
  });
}
$("update-later").onclick = function () { laterVersion = (last && last.update.version) || ""; $("update").hidden = true; };

var prefsFilled = false, savedTimer = null;
$("open-settings").onclick = function () { $("settings").showModal(); };

$("check").onclick = function () {
  $("check").disabled = true;
  $("check-status").textContent = "Checking...";
  call("/api/update/check", {}).then(refresh).catch(function (e) {
    $("check").disabled = false;
    $("check-status").textContent = cap(e.message);
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
    $("prefs-err").textContent = cap(e.message);
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
  $("check-install").hidden = !u.ready;
  $("check-status").textContent = u.status || "";
}

/* ---------- transfer files ---------- */
var xfer = {
  connected: false, label: "", busy: false, wasBusy: false,
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
var protoNames = {sftp: "SFTP", ftpes: "FTP over TLS", ftps: "FTPS, implicit TLS", ftp: "FTP, not encrypted"};
var defaultPorts = {sftp: 22, ftpes: 21, ftps: 990, ftp: 21};

function xerr(msg) { if (msg) toast(msg, true); }
function connErr(msg) { showError("connect-err", msg); }

function protoChanged() {
  var p = $("cproto").value;
  $("cport").placeholder = defaultPorts[p];
  $("cproto-hint").textContent = protoHints[p];
}
$("cproto").onchange = protoChanged;
protoChanged();

$("open-connect").onclick = function () {
  connErr("");
  $("connect-dlg").showModal();
  ($("chost").value ? $("cpass") : $("chost")).focus();
};
$("connect-cancel").onclick = function () { $("connect-dlg").close(); };

function connectNow(trust) {
  connErr("");
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
      $("trust-dlg").showModal();
      return;
    }
    $("trust-dlg").close();
    $("connect-dlg").close();
    $("cpass").value = "";
    xfer.remote.loaded = false;
    return refresh().then(function () { loadPane("remote", r.path); });
  }).catch(function (e) { connErr(e.message); })
    .finally(function () { btn.disabled = false; btn.textContent = "Connect"; });
}
$("cform").onsubmit = function (ev) { ev.preventDefault(); connectNow(""); };
$("trust-yes").onclick = function () { $("trust-dlg").close(); connectNow(xfer.pendingTrust); };
$("trust-no").onclick = function () { $("trust-dlg").close(); xfer.pendingTrust = null; };

$("disconnect").onclick = function () {
  askIf(xfer.busy, {title: "Disconnect?", text: "Transfers are still running. Disconnecting cancels them.", ok: "Disconnect", danger: true})
    .then(function (yes) { if (yes) call("/api/remote/disconnect", {}).then(refresh); });
};

function fillFromRecent(r) {
  $("cproto").value = r.protocol;
  protoChanged();
  $("chost").value = r.host;
  $("cport").value = r.port === defaultPorts[r.protocol] ? "" : r.port;
  $("cuser").value = r.user;
  $("cpass").value = "";
  connErr("");
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
    var pick = el("button", "", r.label);
    pick.type = "button";
    pick.title = "Fill in this connection";
    pick.onclick = function () { fillFromRecent(r); };
    var x = el("button", "", "×");
    x.type = "button";
    x.title = "Forget " + r.label;
    x.setAttribute("aria-label", "Forget " + r.label);
    x.onclick = function () { call("/api/remote/forget", {key: r.key}).then(refresh); };
    item.appendChild(pick);
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
    if (!xfer[side].loaded) paneMessage(side, cap(e.message));
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
  var files = st.entries.length;
  pane(side).querySelector(".count").textContent = !st.loaded ? "" : n ? n + " selected" : files + (files === 1 ? " item" : " items");
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
  [["", "ck"], ["Name", ""], ["Size", "sz num"], ["Modified", "dt num"]].forEach(function (h) { hr.appendChild(el("th", h[1], h[0])); });
  head.appendChild(hr);
  table.appendChild(head);
  var body = el("tbody");
  st.entries.forEach(function (e, i) {
    var tr = el("tr", (e.dir ? "dir " : "") + (st.sel[e.name] ? "sel" : ""));
    tr.draggable = true;
    var ck = el("td", "ck");
    var box = document.createElement("input");
    box.type = "checkbox";
    box.checked = !!st.sel[e.name];
    box.setAttribute("aria-label", "Select " + e.name);
    box.onclick = function (ev) { ev.stopPropagation(); toggle(side, i, ev.shiftKey, true); };
    ck.appendChild(box);
    var nm = el("td", "nm");
    nm.appendChild(ico(e.dir));
    if (e.dir) {
      var b = el("button", "dirbtn", e.name);
      b.type = "button";
      b.title = "Open " + e.name;
      b.onclick = function (ev) { ev.stopPropagation(); openFolder(side, e.name); };
      nm.appendChild(b);
    } else {
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

function openFolder(side, name) { loadPane(side, xfer[side].path, name); }

function send(from) {
  var to = from === "local" ? "remote" : "local";
  var items = selected(from).map(function (e) { return {name: e.name, dir: e.dir, size: e.size, mod: e.mod}; });
  if (!items.length) return;
  var there = {};
  xfer[to].entries.forEach(function (e) { there[e.name] = true; });
  var clash = items.filter(function (e) { return there[e.name]; }).map(function (e) { return e.name; });
  var list = clash.slice(0, 5).join(", ") + (clash.length > 5 ? " and " + (clash.length - 5) + " more" : "");
  askIf(clash.length, {title: "Replace files?", text: "Already there: " + list + ".\nFiles with the same name are overwritten.", ok: "Replace"})
    .then(function (yes) {
      if (!yes) return;
      call("/api/transfer", {
        up: from === "local", localDir: xfer.local.path, remoteDir: xfer.remote.path, items: items
      }).then(function () { xfer[from].sel = {}; drawPane(from); refresh(); })
        .catch(function (e) { xerr(e.message); });
    });
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
  ask({title: "New folder", text: "Name of the new folder on the server.", value: "", ok: "Create"}).then(function (name) {
    if (!name) return;
    call("/api/remote/mkdir", {dir: xfer.remote.path, name: name})
      .then(function () { loadPane("remote", xfer.remote.path); }).catch(function (e) { xerr(e.message); });
  });
};
q("remote", "rename").onclick = function () {
  var e = selected("remote")[0];
  if (!e) return;
  ask({title: "Rename", text: "New name for " + e.name + ".", value: e.name, ok: "Rename"}).then(function (to) {
    if (!to || to === e.name) return;
    call("/api/remote/rename", {dir: xfer.remote.path, from: e.name, to: to})
      .then(function () { loadPane("remote", xfer.remote.path); }).catch(function (err) { xerr(err.message); });
  });
};
q("remote", "delete").onclick = function () {
  var items = selected("remote");
  if (!items.length) return;
  var dirs = items.filter(function (e) { return e.dir; }).length;
  var what = items.length === 1 ? "“" + items[0].name + "”" : items.length + " items";
  ask({title: "Delete " + what + " from the server?", text: (dirs ? "Folders are deleted with everything inside them. " : "") + "This cannot be undone.", ok: "Delete", danger: true})
    .then(function (yes) {
      if (!yes) return;
      call("/api/remote/delete", {dir: xfer.remote.path, items: items.map(function (e) { return {name: e.name, dir: e.dir}; })})
        .then(function () { loadPane("remote", xfer.remote.path); }).catch(function (err) { xerr(err.message); });
    });
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
  $("xfer-queue").hidden = !list.length;
  $("q-cancel").disabled = !active;
  $("q-clear").disabled = active === list.length;
  var sum = [];
  if (active) sum.push(active + " to go" + (total ? ", " + fmtSize(Math.max(total - bytes, 0)) + " left" : ""));
  if (done) sum.push(done + " done");
  if (failed) sum.push(failed + " failed");
  $("q-sum").textContent = sum.join(" \u00a0|\u00a0 ");
  // Running and waiting first, then the most recent finished ones.
  var order = list.filter(function (t) { return t.state === "running"; })
    .concat(list.filter(function (t) { return t.state === "queued"; }).slice(0, 50))
    .concat(list.filter(function (t) {
      // A finished folder adds nothing once its files are listed.
      return t.state !== "running" && t.state !== "queued" && !(t.dir && t.state === "done");
    }).reverse().slice(0, 100));
  order.forEach(function (t) {
    var tr = el("tr", t.error ? "has-err" : "");
    tr.appendChild(el("td", "dirn", t.up ? "Upload" : "Download"));
    var nm = el("td", "", t.name + (t.dir ? "/" : ""));
    nm.title = t.name;
    tr.appendChild(nm);
    var pct = t.state === "done" ? 100 : t.size ? Math.min(100, Math.round(t.done / t.size * 100)) : 0;
    var mt = el("td", "mt");
    if (t.state === "failed") {
      var badge = el("span", "badge out");
      badge.innerHTML = ICON_OUT;
      badge.appendChild(document.createTextNode("Failed"));
      mt.appendChild(badge);
    } else if (t.state !== "cancelled") {
      var bar = el("div", "meter"), fill = el("i");
      fill.style.width = pct + "%";
      bar.appendChild(fill);
      mt.appendChild(bar);
    }
    tr.appendChild(mt);
    var label = {queued: "Waiting", done: "Done", failed: t.size ? fmtSize(t.size) : "", cancelled: "Cancelled"}[t.state];
    if (t.state === "running") label = t.dir ? "Reading folder" : t.size ? pct + "% of " + fmtSize(t.size) : fmtSize(t.done);
    else if (t.state === "done" && !t.dir && t.size) label = "Done, " + fmtSize(t.size);
    tr.appendChild(el("td", "st", label));
    var act = el("td", "act");
    if (t.state === "queued" || t.state === "running") {
      var c = el("button", "small", "Cancel");
      c.type = "button";
      c.onclick = function () { call("/api/transfer/cancel", {id: t.id}).then(refresh); };
      act.appendChild(c);
    }
    tr.appendChild(act);
    box.appendChild(tr);
    if (t.error) {
      var er = el("tr", "err"), msg = el("td", "", cap(t.error));
      er.appendChild(el("td"));
      msg.colSpan = 4;
      er.appendChild(msg);
      box.appendChild(er);
    }
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
    var b = el("button", "small", "Remove");
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
  $("conn-on").hidden = !r.connected;
  $("conn-off").hidden = r.connected;
  $("open-connect").hidden = r.connected;
  $("disconnect").hidden = !r.connected;
  if (r.connected) {
    $("clabel").textContent = r.label;
    $("cproto-label").textContent = protoNames[r.protocol] || "";
    if (page === "transfer" && (!wasConnected || r.label !== xfer.label) && !xfer.remote.loaded) loadPane("remote", r.path);
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

/* ---------- state ---------- */
function render(s) {
  last = s;
  running = s.running;
  if (!formFilled) {
    var st = s.settings;
    $("folder").value = st.folder || "";
    previewFolder = st.folder || "";
    $("target").value = st.target || 3000;
    $("uploads").checked = !!st.uploads;
    $("internet").checked = !!st.internet;
    $("keeppin").checked = !!st.keepPin;
    if (!$("server").value) $("server").value = st.server || "";
    if (!$("rfolder").value) $("rfolder").value = st.mirrorFolder || "";
    setMode(st.mode === "site" ? "site" : "files");
    formFilled = true;
  }
  if (firstRender) {
    firstRender = false;
    if (!s.running && (s.mirror || s.settings.mode === "receive")) go("receive");
  }
  $("setup").hidden = s.running;
  $("live").hidden = !s.running;
  if (s.running) {
    var site = s.settings.mode === "site";
    $("live-name").textContent = site ? "Website on port " + s.settings.target : base(s.what);
    $("what").textContent = site ? "Visitors see the website running on this computer." : s.what;
    $("pin").textContent = s.pin;
    $("uploads-stat").hidden = site;
    $("uploads-val").textContent = s.settings.uploads ? "Allowed" : "Off";
    renderLinks(s.links || []);
    $("net").textContent = s.internetStatus || "";
    $("net").hidden = !s.internetStatus;
  } else {
    lastLinks = "";
  }
  // Refresh the list when something happens, like a visitor uploading a file.
  var top = s.activity[0] || "";
  showPreview(s.running && top !== lastActivity && top.indexOf("Received ") >= 0);
  if (top !== lastActivity) {
    lastActivity = top;
    drawLog($("log"), s.activity, 5);
    drawLog($("all-log"), s.activity, 30);
  }
  renderMirror(s.mirror);
  renderSettings(s);
  renderRemote(s.remote);
  var u = s.update || {};
  $("update").hidden = !u.ready || laterVersion === u.version;
  if (u.ready && !$("update-go").disabled) $("update-text").textContent = "Ferry " + u.version + " is downloaded. Restart Ferry to switch to it. Your settings are kept.";
  renderFoot(s);
}

function refresh() {
  return call("/api/state").then(function (s) {
    if (!closing) $("offline").classList.remove("show");
    render(s);
  }).catch(function () { offline(closing || "Ferry has closed. Open Ferry again to keep sharing."); });
}

(function intro() {
  var params = new URLSearchParams(location.search);
  if (!params.has("intro")) return;
  history.replaceState(null, "", "/");
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  var box = $("intro"), app = document.querySelector(".app"), done = false;
  function finish() {
    if (done) return;
    done = true;
    box.classList.add("out");
    app.classList.add("enter");
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
