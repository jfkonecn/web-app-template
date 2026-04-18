export {};

const root = document.querySelector<HTMLElement>("[data-page='index']");

if (root) {
  root.dataset.js = "ready";

  const healthLink = root.querySelector<HTMLAnchorElement>("[data-health-link]");
  if (healthLink) {
    healthLink.addEventListener("click", () => {
      healthLink.dataset.clicked = "true";
    });
  }
}
