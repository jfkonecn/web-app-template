const http = require("http");
const path = require("path");
const { readFile } = require("fs/promises");

const rootDir = path.resolve(__dirname, "../../..");
const templatesDir = path.join(rootDir, "web", "templates");
const staticDir = path.join(rootDir, "web", "static");

function renderUserTemplate(template) {
  return template
    .replace(/{{\s*\.name\s*}}/g, "Ada Lovelace")
    .replace(/{{\s*\.email\s*}}/g, "ada@example.com");
}

function renderAdminExampleTemplate(template) {
  return template
    .replace(/{{\s*\.name\s*}}/g, "Admin User")
    .replace(/{{\s*\.email\s*}}/g, "admin@example.com")
    .replace(/{{\s*\.requiredPermission\s*}}/g, "read:admin");
}

function renderForbiddenTemplate(template) {
  return template.replace(/{{\s*\.requiredPermission\s*}}/g, "read:admin");
}

async function loadFixtures() {
  const [indexHtml, userTemplate, adminExampleTemplate, forbiddenTemplate, stylesCss] = await Promise.all([
    readFile(path.join(templatesDir, "index.html"), "utf8"),
    readFile(path.join(templatesDir, "user.html"), "utf8"),
    readFile(path.join(templatesDir, "admin-example.html"), "utf8"),
    readFile(path.join(templatesDir, "403.html"), "utf8"),
    readFile(path.join(staticDir, "styles.css"), "utf8"),
  ]);

  return {
    "/": { body: indexHtml, contentType: "text/html; charset=utf-8" },
    "/user": {
      body: renderUserTemplate(userTemplate),
      contentType: "text/html; charset=utf-8",
    },
    "/admin-example": {
      body: renderAdminExampleTemplate(adminExampleTemplate),
      contentType: "text/html; charset=utf-8",
    },
    "/forbidden": {
      body: renderForbiddenTemplate(forbiddenTemplate),
      contentType: "text/html; charset=utf-8",
    },
    "/static/styles.css": {
      body: stylesCss,
      contentType: "text/css; charset=utf-8",
    },
  };
}

async function startFixtureServer() {
  const fixtures = await loadFixtures();

  const server = http.createServer((req, res) => {
    const fixture = fixtures[req.url];

    if (!fixture) {
      res.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
      res.end("Not found");
      return;
    }

    res.writeHead(200, { "content-type": fixture.contentType });
    res.end(fixture.body);
  });

  await new Promise((resolve, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", resolve);
  });

  const address = server.address();
  const baseUrl = `http://127.0.0.1:${address.port}`;

  return { baseUrl, server };
}

module.exports = { startFixtureServer };
