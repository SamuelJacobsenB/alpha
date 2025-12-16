const header = document.querySelector("header"); // Novo: Pega o elemento header
const menu = document.querySelector("#menu");
const burguer = document.querySelector("#burguer");

// Função para gerenciar o scroll
function handleScroll() {
  const scrollThreshold = 50;

  if (window.scrollY > scrollThreshold) {
    header.classList.add("scrolled");
  } else {
    header.classList.remove("scrolled");
  }
}

// Listener de scroll e chama no carregamento
window.addEventListener("scroll", handleScroll);
window.addEventListener("load", handleScroll);

// Lógica do menu hamburguer
burguer.addEventListener("click", () => {
  const expanded = menu.classList.contains("open");
  if (expanded) {
    menu.classList.remove("open");
    burguer.classList.remove("open");
    burguer.setAttribute("aria-expanded", "false");
  } else {
    menu.classList.add("open");
    burguer.classList.add("open");
    burguer.setAttribute("aria-expanded", "true");
  }
});
