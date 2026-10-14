function json(data, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: { "content-type": "application/json; charset=utf-8" },
  });
}

function makeCode() {
  const bytes = crypto.getRandomValues(new Uint8Array(8));
  return Array.from(bytes, (byte) => byte.toString(36).padStart(2, "0")).join("").slice(0, 8);
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname === "/healthz") return new Response("ok");

    if (url.pathname === "/shorten" && request.method === "POST") {
      let input;
      try { input = await request.json(); } catch { return json({ error: "请求格式错误" }, 400); }
      if (typeof input.url !== "string" || input.url.length > 2048) return json({ error: "url 必须是有效网址" }, 400);
      let target;
      try { target = new URL(input.url); } catch { return json({ error: "url 必须是有效网址" }, 400); }
      if (!["http:", "https:"].includes(target.protocol)) return json({ error: "仅支持 http/https" }, 400);

      for (let attempt = 0; attempt < 3; attempt++) {
        const code = makeCode();
        try {
          await env.DB.prepare("INSERT INTO links (short_code, original_url, created_at) VALUES (?, ?, ?)")
            .bind(code, target.href, new Date().toISOString()).run();
          return json({ code: 200, msg: "success", shortcode: code });
        } catch (error) {
          if (attempt === 2) return json({ error: "数据库保存失败" }, 500);
        }
      }
    }

    if (url.pathname === "/api/links" && request.method === "GET") {
      const result = await env.DB.prepare("SELECT short_code, original_url, created_at FROM links ORDER BY created_at DESC LIMIT 1000").all();
      return json({ code: 200, msg: "success", data: result.results });
    }

    const code = decodeURIComponent(url.pathname.slice(1));
    if (!code || code.includes("/")) return new Response("404页面迷路啦～", { status: 404 });
    const row = await env.DB.prepare("SELECT original_url FROM links WHERE short_code = ?").bind(code).first();
    if (!row) return new Response("404页面迷路啦～", { status: 404 });
    return Response.redirect(row.original_url, 302);
  },
};
