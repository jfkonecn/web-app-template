const path = require("path");
const { mkdir, readFile, writeFile } = require("fs/promises");
const { startFixtureServer } = require("./helpers/fixture-server.cjs");

const goldensDir = path.join(__dirname, "__goldens__");
const artifactsDir = path.join(__dirname, "__artifacts__");
const shouldUpdateGoldens = process.env.UPDATE_SCREENSHOTS === "1";

async function assertMatchesGolden(name, screenshot) {
  const goldenPath = path.join(goldensDir, `${name}.png`);

  if (shouldUpdateGoldens) {
    await mkdir(goldensDir, { recursive: true });
    await writeFile(goldenPath, screenshot);
    return;
  }

  let expected;

  try {
    expected = await readFile(goldenPath);
  } catch (error) {
    if (error.code !== "ENOENT") {
      throw error;
    }

    throw new Error(
      `Missing golden screenshot for "${name}". Run "npm run test:screenshots:update" to create it.`,
    );
  }

  if (Buffer.compare(screenshot, expected) === 0) {
    return;
  }

  await mkdir(artifactsDir, { recursive: true });
  const actualPath = path.join(artifactsDir, `${name}.actual.png`);
  await writeFile(actualPath, screenshot);

  throw new Error(
    `Screenshot mismatch for "${name}". Review ${actualPath} and refresh with "npm run test:screenshots:update" if the change is expected.`,
  );
}

describe("HTML page screenshots", () => {
  let server;
  let baseUrl;

  beforeAll(async () => {
    ({ server, baseUrl } = await startFixtureServer());
  });

  afterAll(async () => {
    await new Promise((resolve, reject) => {
      server.close((error) => {
        if (error) {
          reject(error);
          return;
        }

        resolve();
      });
    });
  });

  test.each([
    { name: "index", route: "/" },
    { name: "user", route: "/user" },
  ])('$name matches the committed screenshot', async ({ name, route }) => {
    await page.goto(`${baseUrl}${route}`, { waitUntil: "networkidle0" });
    const screenshot = await page.screenshot({ fullPage: true });
    await assertMatchesGolden(name, screenshot);
  });
});
