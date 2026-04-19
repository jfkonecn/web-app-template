const path = require("path");
const { toMatchImageSnapshot } = require("jest-image-snapshot");
const { startFixtureServer } = require("./helpers/fixture-server.cjs");

const goldensDir = path.join(__dirname, "__goldens__");
const artifactsDir = path.join(__dirname, "__artifacts__");
const screenshotViewport = {
  width: 1440,
  height: 1100,
  deviceScaleFactor: 1,
};

expect.extend({ toMatchImageSnapshot });

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
    { name: "admin-example", route: "/admin-example" },
    { name: "forbidden", route: "/forbidden" },
  ])('$name matches the committed screenshot', async ({ name, route }) => {
    await page.setViewport(screenshotViewport);
    await page.goto(`${baseUrl}${route}`, { waitUntil: "networkidle0" });
    const screenshot = await page.screenshot({ fullPage: true });

    expect(screenshot).toMatchImageSnapshot({
      customSnapshotsDir: goldensDir,
      customDiffDir: artifactsDir,
      customReceivedDir: artifactsDir,
      storeReceivedOnFailure: true,
      customReceivedPostfix: ".actual",
      customSnapshotIdentifier: name,
      customDiffConfig: {
        threshold: 0.1,
      },
    });
  });
});
