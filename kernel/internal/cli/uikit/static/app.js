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
        if (part === "area" || part === "tests") { d.path = f.getAttribute("data-path"); d.lines = f.getAttribute("data-lines"); }
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
        var text = "Reviewer: " + res.reviewer + "\nFiles:\n  " + res.files.join("\n  ") + "\nCommit message: " + res.commit_message;
        if (res.warnings && res.warnings.length) { text += "\n\nBefore you seal:\n  " + res.warnings.join("\n  "); }
        bpv.querySelector("pre").textContent = text;
        bpv.hidden = false; say(bmsg, "");
      });
    });
    work.querySelector("[data-act=review-back]").addEventListener("click", function () { bpv.hidden = true; });
    // The bench's background job (ui-reviewer-spec §1.1): the tests, or a seal that runs them first.
    // The page polls a read-only endpoint; closing the tab does not stop the job.
    var jobLog = work.querySelector(".job-log");
    var busy = function (on) { work.querySelectorAll("button[data-act]").forEach(function (b) { if (!b.hasAttribute("data-refused")) { b.disabled = on; } }); };
    var follow = function () {
      busy(true); jobLog.hidden = false;
      var tick = function () {
        fetch("/review/job?commit=" + encodeURIComponent(commit), { credentials: "same-origin" })
          .then(function (r) { return r.json(); })
          .then(function (job) {
            if (job.none) { busy(false); return; }
            jobLog.textContent = (job.lines || []).join("\n") + (job.error ? "\n" + job.error : "");
            if (job.running) { window.setTimeout(tick, 1000); return; }
            busy(false);
            if (job.error) { say(bmsg, ({ seal: "Not sealed: ", agent: "The agent did not answer: " }[job.kind] || "The tests did not finish: ") + job.error, true); return; }
            if (job.kind === "agent") {
              say(bmsg, "The agent answered. Reloading to show its answers beside each question.");
              window.setTimeout(function () { window.location.reload(); }, 800);
            } else if (job.kind === "seal") {
              say(bmsg, "Sealed, signed and committed: " + job.result.files.join(", ") + ". Loading the check result.");
              window.setTimeout(function () { window.location.href = "/review?sealed=" + encodeURIComponent(commit); }, 1200);
            } else {
              say(bmsg, "Tests finished. Reloading to show each part.");
              window.setTimeout(function () { window.location.reload(); }, 800);
            }
          })
          .catch(function () { window.setTimeout(tick, 2000); });
      };
      tick();
    };
    var runTests = work.querySelector("[data-act=review-tests]");
    if (runTests) {
      runTests.addEventListener("click", function () {
        send("review-tests").then(function (res) {
          if (res.error) { say(bmsg, res.error, true); return; }
          say(bmsg, "Running the tests. You may leave this page; the run goes on.");
          follow();
        });
      });
    }
    var askAgent = work.querySelector("[data-act=review-agent]");
    if (askAgent) {
      askAgent.addEventListener("click", function () {
        var confirmBox = work.querySelector(".agent-confirm");
        if (confirmBox && !confirmBox.checked) { say(bmsg, "Read the command above and tick the box first: it is not Claude Code.", true); return; }
        post("/act/review-agent", { base: base, commit: commit, decisions: decisions(), confirm: !!(confirmBox && confirmBox.checked) }).then(function (res) {
          if (res.error) { say(bmsg, res.error, true); return; }
          say(bmsg, "Asking the agent. You may leave this page; it goes on.");
          follow();
        });
      });
    }
    if (document.body.getAttribute("data-job") === "running") { follow(); }
    work.querySelector("[data-act=review-seal]").addEventListener("click", function (ev) {
      ev.target.disabled = true;
      send("review-seal").then(function (res) {
        ev.target.disabled = false;
        if (res.error) { say(bmsg, res.error + (res.missing ? ": " + res.missing.join("; ") : ""), true); return; }
        if (res.job) { bpv.hidden = true; say(bmsg, "Running the tests again, then sealing."); follow(); return; }
        say(bmsg, "Sealed, signed and committed: " + res.files.join(", ") + ". Loading the check result.");
        window.setTimeout(function () { window.location.href = "/review?sealed=" + encodeURIComponent(commit); }, 1200);
      });
    });
  }

  // Show in the code (ui-reviewer-spec §10.2): scroll the change to a decision's part and mark it.
  document.querySelectorAll("a.jump").forEach(function (a) {
    a.addEventListener("click", function (ev) {
      var path = a.getAttribute("data-jump-path"), line = a.getAttribute("data-jump-line");
      var target = null, header = null;
      document.querySelectorAll(".diff .dl[data-path]").forEach(function (d) {
        if (d.getAttribute("data-path") !== path) { return; }
        if (!target && d.getAttribute("data-line") === line) { target = d; }
        if (!header && d.getAttribute("data-line") === "0") { header = d; }
      });
      // A part with no added line of its own (removed lines) lands on its file's first line.
      target = target || header;
      if (!target) { a.textContent = "This file is not in the view (it is past the end the page shows): read it with git show"; ev.preventDefault(); return; }
      ev.preventDefault();
      var change = document.getElementById("the-change");
      if (change) { change.open = true; }
      document.querySelectorAll(".diff .dl-jump").forEach(function (d) { d.classList.remove("dl-jump"); });
      target.classList.add("dl-jump");
      target.scrollIntoView({ block: "center", behavior: "smooth" });
    });
  });

  // Issue the warranty (ui-reviewer-spec §2.2): preview first; the server refuses a partial release.
  var issue = document.querySelector(".issue");
  if (issue && issue.querySelector("[data-act=issue-preview]")) {
    var ipv = issue.querySelector(".issue-preview"), imsg = issue.querySelector(".msg");
    var issueSend = function (confirm) { return post("/act/issue", { base: base, confirm: confirm }); };
    issue.querySelector("[data-act=issue-preview]").addEventListener("click", function () {
      issueSend(false).then(function (res) {
        ipv.hidden = true;
        if (res.issued_already) { say(imsg, "Already issued at this point: " + res.issued_already + "."); return; }
        if (res.error) { say(imsg, res.error + (res.not_warranted ? ":\n" + res.not_warranted.join("\n") : ""), true); return; }
        ipv.querySelector("pre").textContent = "Signer: " + res.signer + "\nDescribes: " + res.describes + " (" + res.warranted + " change(s) warranted)\nFiles:\n  " + res.files.join("\n  ") + "\nCommit message: " + res.commit_message;
        ipv.hidden = false; say(imsg, "");
      });
    });
    issue.querySelector("[data-act=issue-confirm]").addEventListener("click", function (ev) {
      ev.target.disabled = true;
      issueSend(true).then(function (res) {
        ev.target.disabled = false;
        if (res.error) { say(imsg, res.error + (res.not_warranted ? ":\n" + res.not_warranted.join("\n") : ""), true); return; }
        if (res.issued_already) { say(imsg, "Already issued at this point: " + res.issued_already + "."); return; }
        ipv.hidden = true;
        say(imsg, "Issued and committed: " + res.folder + ". It warrants the period up to " + res.describes.slice(0, 12) + ", the commit before this release commit. Send the buyer this folder and, separately, your key fingerprint.");
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
