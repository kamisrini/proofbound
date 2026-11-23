// The Audit room's only script. It never builds HTML from data: text goes in through textContent.
(function () {
  "use strict";
  var base = document.body.getAttribute("data-base") || "";

  function post(path, body) {
    return fetch(path, {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    }).then(function (r) {
      return r.json().catch(function () { return { error: "unexpected answer (" + r.status + ")" }; })
        .then(function (data) { if (!r.ok && !data.error) { data.error = "refused (" + r.status + ")"; } return data; });
    });
  }

  function say(el, text, bad) {
    if (!el) { return; }
    el.textContent = text;
    el.className = "msg" + (bad ? " bad" : " closed");
  }

  // Filter.
  document.querySelectorAll("button[data-filter]").forEach(function (b) {
    b.addEventListener("click", function () {
      var exc = b.getAttribute("data-filter") === "exc";
      document.body.classList.toggle("only-exc", exc);
      document.querySelectorAll("button[data-filter]").forEach(function (o) {
        o.setAttribute("aria-pressed", String(o === b));
      });
    });
  });

  // Print.
  document.querySelectorAll("[data-act=print]").forEach(function (b) {
    b.addEventListener("click", function () {
      document.querySelectorAll("details.control").forEach(function (d) { d.open = true; });
      window.print();
    });
  });

  // Exception outcomes.
  document.querySelectorAll("article.exception").forEach(function (art) {
    var id = art.getAttribute("data-id");
    var start = art.querySelector("[data-act=open]");
    var form = art.querySelector("form.outcome");
    var preview = art.querySelector(".preview");
    var msg = art.querySelector(".msg");
    if (!start || !form || !preview) { return; }
    var pending = null;

    function request() {
      var data = new FormData(form);
      var req = { kind: "outcome", base: base, exception: id, outcome: data.get("outcome") || "", reason: (data.get("reason") || "").trim() };
      if (req.outcome === "fixed") { req.fixed_by = (data.get("fixed_by") || "").trim(); }
      return req;
    }
    start.addEventListener("click", function () { start.hidden = true; form.hidden = false; say(msg, ""); });
    form.addEventListener("change", function () {
      form.querySelector(".fix").hidden = request().outcome !== "fixed";
    });
    art.querySelector("[data-act=cancel]").addEventListener("click", function () {
      form.hidden = true; start.hidden = false; say(msg, "");
    });
    form.addEventListener("submit", function (ev) {
      ev.preventDefault();
      var req = request();
      if (!req.outcome) { say(msg, "Choose what happened to it.", true); return; }
      say(msg, "");
      post("/act/preview", req).then(function (res) {
        if (res.error) { say(msg, res.error, true); return; }
        pending = req; pending.at = res.at; pending.digest = res.digest;
        preview.querySelector("pre").textContent = res.path + "\n" + res.record + "\nsigned by " + res.signer;
        form.hidden = true; preview.hidden = false;
      });
    });
    art.querySelector("[data-act=back]").addEventListener("click", function () {
      preview.hidden = true; form.hidden = false;
    });
    art.querySelector("[data-act=sign]").addEventListener("click", function (ev) {
      ev.target.disabled = true;
      post("/act/sign", pending).then(function (res) {
        ev.target.disabled = false;
        if (res.error) { say(msg, res.error, true); return; }
        say(msg, "Signed and committed as " + res.path + ". Reloading.");
        window.setTimeout(function () { window.location.reload(); }, 800);
      });
    });
  });

  // Now: mark commits newer than the last visit (kept only in this browser).
  var head = document.body.getAttribute("data-head");
  if (head) {
    var last = null;
    try { last = window.localStorage.getItem("proofbound.lastHead"); } catch (e) { last = null; }
    if (last && last !== head) {
      var rows = document.querySelectorAll("tr[data-commit]");
      var count = 0;
      for (var i = rows.length - 1; i >= 0; i--) {
        if (rows[i].getAttribute("data-commit") === last) { break; }
        rows[i].querySelector(".new").hidden = false;
        count++;
      }
      var since = document.querySelector(".since");
      if (since && count) { since.textContent = count + " commit(s) since you last looked."; since.hidden = false; }
    }
    try { window.localStorage.setItem("proofbound.lastHead", head); } catch (e) { /* private window */ }
  }

  // Review bench.
  var bench = document.querySelector(".bench-form");
  if (bench) {
    var commit = document.body.getAttribute("data-commit");
    var work = document.querySelector(".work");
    var bpv = work.querySelector(".preview");
    var bmsg = work.querySelector(".msg");
    var decisions = function () {
      var out = [];
      bench.querySelectorAll("fieldset.promise").forEach(function (f) {
        var part = f.getAttribute("data-part"), g = +f.getAttribute("data-group"), i = +f.getAttribute("data-index");
        var picked = f.querySelector("input[type=radio]:checked");
        var d = { part: part, group: g, index: i, obligation_id: f.getAttribute("data-obligation"), outcome: picked ? picked.value : "" };
        var finding = f.querySelector("input[name^=f-]");
        if (finding && finding.value.trim()) { d.finding = finding.value.trim(); }
        if (part === "area") { d.path = f.getAttribute("data-path"); d.lines = f.getAttribute("data-lines"); }
        out.push(d);
      });
      return out;
    };
    var send = function (act) {
      return post("/act/" + act, { base: base, commit: commit, decisions: decisions() });
    };
    work.querySelector("[data-act=review-draft]").addEventListener("click", function () {
      send("review-draft").then(function (res) {
        if (res.error) { say(bmsg, res.error, true); return; }
        say(bmsg, "Draft saved to " + res.saved + " (on this machine only, not the record).");
      });
    });
    work.querySelector("[data-act=review-preview]").addEventListener("click", function () {
      send("review-preview").then(function (res) {
        if (res.error) { say(bmsg, res.error + (res.missing ? ": " + res.missing.join("; ") : ""), true); return; }
        bpv.querySelector("pre").textContent = "Reviewer: " + res.reviewer + "\nFiles:\n  " + res.files.join("\n  ") + "\nCommit message: " + res.commit_message;
        bpv.hidden = false; say(bmsg, "");
      });
    });
    work.querySelector("[data-act=review-back]").addEventListener("click", function () { bpv.hidden = true; });
    work.querySelector("[data-act=review-seal]").addEventListener("click", function (ev) {
      ev.target.disabled = true;
      send("review-seal").then(function (res) {
        ev.target.disabled = false;
        if (res.error) { say(bmsg, res.error, true); return; }
        say(bmsg, "Sealed, signed and committed: " + res.files.join(", ") + ". Reloading.");
        window.setTimeout(function () { window.location.href = "/review"; }, 1200);
      });
    });
  }

  // Proof in the browser (U3, CLI-INV-51): re-verify each signature with WebCrypto Ed25519 against the
  // declared key, from the exact bytes the page carries. Nothing is sent anywhere.
  var fromB64 = function (b) { var s = atob(b), out = new Uint8Array(s.length); for (var i = 0; i < s.length; i++) { out[i] = s.charCodeAt(i); } return out; };
  var fromHex = function (h) { if (!/^([0-9a-f]{2})*$/.test(h)) { return null; } var out = new Uint8Array(h.length / 2); for (var i = 0; i < out.length; i++) { out[i] = parseInt(h.substr(i * 2, 2), 16); } return out; };
  var utf8 = function (s) { return new TextEncoder().encode(s); };
  var concat = function (a, b) { var out = new Uint8Array(a.length + b.length); out.set(a, 0); out.set(b, a.length); return out; };
  // Canonical JSON of an all-string object: keys in UTF-16 order, JSON.stringify for every string
  // (RFC 8785 for this shape).
  var canonical = function (obj) {
    return "{" + Object.keys(obj).sort().map(function (k) { return JSON.stringify(k) + ":" + JSON.stringify(obj[k]); }).join(",") + "}";
  };
  var mark = function (el, state, text) {
    el.textContent = text;
    el.className = "bcheck tag " + (state === "ok" ? "tag-passed" : state === "bad" ? "tag-exception" : "tag-none");
  };
  var verifyOne = function (el) {
    var key = fromHex(el.getAttribute("data-key") || "");
    if (!key || key.length !== 32) { mark(el, "bad", "no declared key: not verified in this browser"); return; }
    var bytes = fromB64(el.getAttribute("data-b64") || ""), tag = utf8(el.getAttribute("data-tag") || ""), msg, sig;
    if (el.getAttribute("data-kind") === "record") {
      var obj;
      try { obj = JSON.parse(new TextDecoder().decode(bytes)); } catch (e) { mark(el, "bad", "malformed record: does not verify in this browser"); return; }
      sig = fromHex(obj.signature || "");
      delete obj.signature;
      msg = concat(tag, utf8(canonical(obj)));
    } else {
      var f = (el.getAttribute("data-sig") || "").split(" ");
      if (f.length !== 3 || f[0] !== "ed25519" || f[1] !== el.getAttribute("data-key")) { mark(el, "bad", "signature key is not the declared key"); return; }
      sig = fromHex(f[2]);
      msg = concat(tag, bytes);
    }
    if (!sig || sig.length !== 64) { mark(el, "bad", "malformed signature: does not verify in this browser"); return; }
    if (!window.crypto || !crypto.subtle) { mark(el, "none", "this browser cannot check"); return; }
    crypto.subtle.importKey("raw", key, { name: "Ed25519" }, false, ["verify"]).then(function (k) {
      return crypto.subtle.verify({ name: "Ed25519" }, k, sig, msg);
    }).then(function (ok) {
      mark(el, ok ? "ok" : "bad", ok ? "verified in this browser" : "does NOT verify in this browser");
    }, function () { mark(el, "none", "this browser cannot check Ed25519"); });
  };
  document.querySelectorAll(".bcheck").forEach(verifyOne);

  // Countersign the transcript.
  var attest = document.querySelector(".attest");
  if (attest) {
    var btn = attest.querySelector("[data-act=attest]");
    var pv = attest.querySelector(".preview");
    var m = attest.querySelector(".msg");
    var pendingAt = "", pendingDigest = "";
    btn.addEventListener("click", function () {
      post("/act/preview", { kind: "attestation", base: base }).then(function (res) {
        if (res.error) { say(m, res.error, true); return; }
        pendingAt = res.at; pendingDigest = res.digest;
        pv.querySelector("pre").textContent = res.path + "\n" + res.record + "\nsigned by " + res.signer;
        pv.hidden = false; btn.hidden = true; say(m, "");
      });
    });
    attest.querySelector("[data-act=attest-back]").addEventListener("click", function () { pv.hidden = true; btn.hidden = false; });
    attest.querySelector("[data-act=attest-sign]").addEventListener("click", function (ev) {
      ev.target.disabled = true;
      post("/act/sign", { kind: "attestation", base: base, at: pendingAt, digest: pendingDigest }).then(function (res) {
        ev.target.disabled = false;
        if (res.error) { say(m, res.error, true); return; }
        say(m, "Countersigned and committed as " + res.path + ". Reloading.");
        window.setTimeout(function () { window.location.reload(); }, 800);
      });
    });
  }
})();
