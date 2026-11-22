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
