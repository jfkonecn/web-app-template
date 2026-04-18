module.exports = {
  launch: {
    headless: true,
    defaultViewport: {
      width: 1440,
      height: 1100,
      deviceScaleFactor: 1,
    },
    args: [
      "--no-sandbox",
      "--disable-setuid-sandbox",
      "--window-size=1440,1100",
      "--force-device-scale-factor=1",
    ],
  },
};
