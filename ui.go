package main

import (
	"html/template"
	"net/http"
)

type loginPage struct{ Next, Error string }

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if err := pages.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var pages = template.Must(template.New("").Parse(`
{{define "head"}}<!doctype html>
<html lang="en"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex">
<link rel="icon" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Crect width='32' height='32' rx='6' fill='%231f5fbf'/%3E%3Cpath d='M11 7.5v17M11 7.5h11M11 16h8' stroke='white' stroke-width='3.5' stroke-linecap='round'/%3E%3C/svg%3E">
<style>
/* Calm Blue: neutral until something needs attention. The accent is blue so a
   button never reads as a status; red (out) and amber (low) are only for
   problems, always with a word and a shape. Every text colour passes WCAG AA. */
:root{--bg:#f6f6f4;--panel:#fff;--line:#e2e1dc;--text:#1d1d1b;--dim:#5f5e59;--accent:#1f5fbf;--accent-text:#fff;
--out:#b42318;--out-bg:#fdecea;--out-fill:#c4320a;--out-ink:#fff;--low:#8a4a00;--low-bg:#fff4e0;--low-fill:#f5b347;--low-ink:#2b1a00;
--ok:#2e8540;--track:#ecebe6;--hover:#f0efea;--sel:#e7eefa;--input:#fff;
--mono:ui-monospace,"Cascadia Mono",Consolas,Menlo,monospace}
@media (prefers-color-scheme:dark){:root{--bg:#161716;--panel:#1f201e;--line:#34352f;--text:#ecebe6;--dim:#a3a29b;--accent:#7aa7ff;--accent-text:#0b1730;
--out:#ff8f80;--out-bg:#3a1c19;--low:#f3b453;--low-bg:#382910;--low-fill:#f3b453;--ok:#4caf68;--track:#2c2d29;--hover:#272824;--sel:#1e2a3e;--input:#191a18;color-scheme:dark}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:14px/1.45 "Segoe UI",system-ui,-apple-system,sans-serif}
main{max-width:860px;margin:0 auto;padding:24px 16px 48px}
a{color:var(--accent)}
.card{background:var(--panel);border:1px solid var(--line);border-radius:6px}
button,.btn{font:inherit;line-height:1.45;color:var(--text);background:var(--panel);border:1px solid var(--line);border-radius:4px;padding:6px 12px;cursor:pointer;text-decoration:none;display:inline-block;white-space:nowrap}
button:hover,.btn:hover{border-color:var(--dim)}
button.primary,.btn.primary{background:var(--accent);border-color:var(--accent);color:var(--accent-text);font-weight:600}
button.danger{color:var(--out)}
button.small,.btn.small{padding:2px 9px;font-size:13px}
button:disabled{opacity:.45;cursor:default;border-color:var(--line)}
input,select,textarea{font:inherit;color:var(--text);background:var(--input);border:1px solid var(--line);border-radius:4px;padding:6px 9px}
textarea{resize:vertical}
:focus-visible,input:focus,select:focus,textarea:focus{outline:2px solid var(--accent);outline-offset:1px}
progress,input[type=checkbox]{accent-color:var(--accent)}
.muted{color:var(--dim)}
.error{color:var(--out)}
</style>
{{end}}

{{define "login"}}{{template "head"}}<title>Enter PIN</title></head><body>
<main style="max-width:380px;padding-top:12vh">
<form class="card" method="post" style="padding:24px">
<h1 style="margin:0 0 4px;font-size:22px">Enter the PIN</h1>
<p class="muted" style="margin:0 0 16px">It is shown in the Ferry window on the computer you are connecting to.</p>
<input name="pin" inputmode="numeric" autocomplete="one-time-code" autofocus required maxlength="12"
 style="width:100%;font:600 28px/1 var(--mono);letter-spacing:6px;padding:12px">
{{if .Error}}<p class="error" style="margin:12px 0 0">{{.Error}}</p>{{end}}
<button class="primary" style="width:100%;margin-top:16px;padding:10px">Open</button>
</form>
</main></body></html>{{end}}

{{define "offline"}}{{template "head"}}<title>Website not running</title></head><body>
<main><div class="card" style="padding:24px">
<h1 style="margin:0 0 8px;font-size:22px">The website is not running</h1>
<p class="muted" style="margin:0">Nothing is answering on port {{.Port}} of the host computer. Start the website there, then reload this page.</p>
</div></main></body></html>{{end}}

{{define "visitor-css"}}<style>
nav{display:flex;flex-wrap:wrap;gap:4px;align-items:center;font-size:18px;font-weight:600;margin-bottom:16px}
nav a{text-decoration:none;color:var(--text)}
nav span{color:var(--dim)}
.bar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:16px;align-items:center}
.bar .grow{flex:1}
#drop{border:2px dashed var(--line);border-radius:6px;padding:20px;text-align:center;margin-bottom:16px;background:var(--panel)}
#drop.over{border-color:var(--accent)}
progress{width:100%;margin-top:10px}
.notice{padding:10px 14px;margin-bottom:16px}
.texts{padding:12px 14px;margin-bottom:16px}
.texts label{display:block;font-weight:600;margin-bottom:6px}
.trow{display:flex;gap:8px;align-items:flex-start}
.trow textarea{flex:1;min-width:0;min-height:38px}
.texts .note{font-size:12px;margin:6px 0 0}
#tlist{list-style:none;margin:0;padding:0}
#tlist li{display:flex;gap:10px;align-items:flex-start;padding:8px 0;border-top:1px solid var(--line)}
#tform[hidden]+#terr+#tlist li:first-child{border-top:0;padding-top:0}
#tlist .msg{flex:1;min-width:0}
#tlist .meta{font-size:12px}
.ttext{white-space:pre-wrap;overflow-wrap:anywhere;max-height:12em;overflow:auto}
</style>{{end}}

{{define "texts"}}<div class="card texts" id="texts"{{if .ReadOnly}} hidden{{end}}>
<form id="tform"{{if .ReadOnly}} hidden{{end}}>
<label for="tnew">Send a link or a note</label>
<div class="trow"><textarea id="tnew" rows="1" maxlength="10000" placeholder="Paste or type it here"></textarea><button class="primary">Send</button></div>
<p class="muted note">It shows up in the Ferry window on the computer you are connected to. Everyone with the PIN can read texts sent here.</p>
</form>
<p id="terr" class="error" hidden></p>
<ul id="tlist"></ul>
</div>
<script>
(function(){
const box=document.getElementById('texts'),form=document.getElementById('tform'),input=document.getElementById('tnew'),list=document.getElementById('tlist'),err=document.getElementById('terr');
let key='';
const cap=s=>s?s.charAt(0).toUpperCase()+s.slice(1):s;
function linkify(el,text){
  text.split(/(https?:\/\/[^\s<>"]+)/).forEach((part,i)=>{
    if(i%2){const a=document.createElement('a');a.href=part;a.textContent=part;a.target='_blank';a.rel='noopener noreferrer';el.appendChild(a)}
    else if(part)el.appendChild(document.createTextNode(part));
  });
}
// Copying needs a secure page, which a plain Wi-Fi link is not, so fall
// back to selecting the text and copying it the old way.
function copy(text,btn){
  const done=()=>{btn.textContent='Copied';setTimeout(()=>{btn.textContent='Copy'},1500)};
  const fallback=()=>{
    const t=document.createElement('textarea');t.value=text;t.setAttribute('readonly','');t.style.position='fixed';t.style.opacity='0';
    document.body.appendChild(t);t.select();t.setSelectionRange(0,text.length);
    let ok=false;try{ok=document.execCommand('copy')}catch(e){}
    t.remove();if(ok)done();else btn.textContent='Select it and copy';
  };
  if(navigator.clipboard&&window.isSecureContext)navigator.clipboard.writeText(text).then(done,fallback);else fallback();
}
function draw(d){
  const ts=d.texts||[];
  if(ts.length)box.hidden=false;
  const k=ts.map(t=>t.id).join(',');
  if(k===key)return;
  key=k;list.textContent='';
  for(const t of ts){
    const li=document.createElement('li'),msg=document.createElement('div'),meta=document.createElement('div'),body=document.createElement('div'),b=document.createElement('button');
    msg.className='msg';meta.className='muted meta';body.className='ttext';
    meta.textContent=(t.host?'From the computer sharing this':'From '+t.from)+', '+t.time;
    linkify(body,t.text);
    b.type='button';b.className='small';b.textContent='Copy';b.onclick=()=>copy(t.text,b);
    msg.append(meta,body);li.append(msg,b);list.appendChild(li);
  }
}
function load(){
  if(document.hidden)return;
  fetch('/__ferry/texts',{cache:'no-store'}).then(r=>r.ok?r.json():null).then(d=>{if(d)draw(d)}).catch(()=>{});
}
form.onsubmit=e=>{
  e.preventDefault();
  const text=input.value;
  if(!text.trim()){input.focus();return}
  fetch('/__ferry/texts',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({text})})
    .then(r=>r.json().then(d=>{if(!r.ok)throw new Error(d.error);input.value='';err.hidden=true;draw(d)}))
    .catch(e=>{err.textContent=e instanceof SyntaxError||e instanceof TypeError?'Could not send. Check the connection and try again.':cap(e.message);err.hidden=false});
};
input.onkeydown=e=>{if(e.key==='Enter'&&(e.ctrlKey||e.metaKey))form.requestSubmit()};
load();setInterval(load,3000);document.addEventListener('visibilitychange',load);
})();
</script>{{end}}

{{define "upload"}}<form id="drop" method="post" enctype="multipart/form-data">
<input id="pick" type="file" name="files" multiple hidden>
<p style="margin:0 0 10px">Drop files here to send them to this folder</p>
<button type="button" class="primary" onclick="pick.click()">Choose files</button>
<progress id="prog" value="0" max="100" hidden></progress>
<p id="status" class="muted" style="margin:8px 0 0"></p>
</form>
<script>
const drop=document.getElementById('drop'),pick=document.getElementById('pick'),prog=document.getElementById('prog'),status=document.getElementById('status');
function send(files){
  if(!files.length)return;
  const fd=new FormData();for(const f of files)fd.append('files',f);
  const x=new XMLHttpRequest();x.open('POST',location.pathname);
  prog.hidden=false;status.textContent='Uploading '+files.length+' file(s)...';
  x.upload.onprogress=e=>{if(e.lengthComputable)prog.value=e.loaded/e.total*100};
  x.onload=()=>{if(x.status<400)location.href=x.responseURL;else status.textContent='Upload failed: '+x.responseText};
  x.onerror=()=>status.textContent='Upload failed, connection lost.';
  x.send(fd);
}
pick.onchange=()=>send(pick.files);
['dragenter','dragover'].forEach(t=>drop.addEventListener(t,e=>{e.preventDefault();drop.classList.add('over')}));
['dragleave','drop'].forEach(t=>drop.addEventListener(t,e=>{e.preventDefault();drop.classList.remove('over')}));
drop.addEventListener('drop',e=>send(e.dataTransfer.files));
</script>{{end}}

{{define "list"}}{{template "head"}}<title>{{.Title}}</title>
{{template "visitor-css"}}
<style>
table{width:100%;border-collapse:collapse}
td{padding:8px 12px;border-top:1px solid var(--line)}
tr:first-child td{border-top:0}
tbody tr:hover{background:var(--hover)}
td.name a{text-decoration:none;color:var(--text);word-break:break-word}
td.thumb{width:56px;padding-right:0}
td.thumb a{display:block;width:56px;height:56px;text-decoration:none}
td.thumb img{display:block;width:56px;height:56px;object-fit:cover;border-radius:6px;background:var(--bg)}
.tile{display:grid;place-items:center;width:56px;height:56px;border-radius:4px;background:var(--bg);border:1px solid var(--line);color:var(--dim);font:700 11px/1 "Segoe UI",system-ui,sans-serif;letter-spacing:.04em;text-transform:uppercase}
.tile.dir{background:var(--sel);color:var(--accent)}
td.meta{color:var(--dim);white-space:nowrap;font-size:13px;text-align:right;font-variant-numeric:tabular-nums}
td.act{width:1%;white-space:nowrap}
td.act a{font-size:13px}
@media (max-width:560px){td.meta.date{display:none}}
</style></head><body><main>
<nav>{{range $i, $c := .Crumbs}}{{if $i}}<span>/</span>{{end}}<a href="{{$c.Href}}">{{$c.Name}}</a>{{end}}</nav>
{{if .Notice}}<div class="card notice">{{.Notice}}</div>{{end}}
<div class="bar"><span class="muted grow">{{len .Entries}} items</span>
<a class="btn" href="?zip">Download all as zip</a>
<a class="btn" href="/__ferry/logout">Log out</a></div>
{{template "texts" .}}
{{if not .ReadOnly}}{{template "upload" .}}{{end}}
<div class="card">{{if .Entries}}<table>
{{range .Entries}}<tr>
<td class="thumb"><a href="{{.Href}}" tabindex="-1" aria-hidden="true">{{if .Thumb}}<img src="{{.Href}}?thumb" alt="" loading="lazy" decoding="async">{{else if .IsDir}}<span class="tile dir">Folder</span>{{else}}<span class="tile">{{.Badge}}</span>{{end}}</a></td>
<td class="name"><a href="{{.Href}}">{{.Name}}</a></td>
<td class="meta">{{.Size}}</td>
<td class="meta date">{{.Modified}}</td>
<td class="act">{{if .IsDir}}<a href="{{.Href}}?zip">zip</a>{{else}}<a href="{{.Href}}?dl">download</a>{{end}}</td>
</tr>{{end}}
</table>{{else}}<p class="muted" style="padding:14px;margin:0">This folder is empty.</p>{{end}}</div>
</main></body></html>{{end}}

{{define "drop"}}{{template "head"}}<title>Send files to {{.Title}}</title>
{{template "visitor-css"}}</head><body><main>
<nav><span style="color:var(--text)">Send files to {{.Title}}</span></nav>
{{if .Notice}}<div class="card notice">{{.Notice}} Thank you.</div>{{end}}
<div class="bar"><span class="muted grow">Files you send go into the folder {{.Title}} on the computer running Ferry. You cannot see what is already in it, and nothing there is replaced.</span>
<a class="btn" href="/__ferry/logout">Log out</a></div>
{{template "upload" .}}
{{template "texts" .}}
</main></body></html>{{end}}
`))
