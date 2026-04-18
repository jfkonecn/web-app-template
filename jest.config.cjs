module.exports = {
  preset: "jest-puppeteer",
  rootDir: __dirname,
  testMatch: ["<rootDir>/test/screenshots/**/*.test.cjs"],
  testTimeout: 30000,
  verbose: true,
};
