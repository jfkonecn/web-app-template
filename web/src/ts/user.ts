export {};

const root = document.querySelector<HTMLElement>("[data-page='user']");

if (root) {
  root.dataset.js = "ready";

  const userCard = root.querySelector<HTMLElement>("[data-user-card]");
  if (userCard) {
    userCard.addEventListener("mouseenter", () => {
      userCard.dataset.hovered = "true";
    });

    userCard.addEventListener("mouseleave", () => {
      delete userCard.dataset.hovered;
    });
  }
}
