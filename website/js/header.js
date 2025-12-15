document.addEventListener("DOMContentLoaded", () => {
  const burger = document.getElementById("burguer");
  const nav = document.getElementById("menu");
  const header = document.querySelector("header");

  if (burger && nav) {
    burger.addEventListener("click", () => {
      // Alterna classes para animação do ícone e visibilidade do menu
      burger.classList.toggle("open");
      nav.classList.toggle("open");

      // Impede rolagem do corpo quando menu está aberto (opcional)
      document.body.style.overflow = nav.classList.contains("open")
        ? "hidden"
        : "auto";
    });
  }

  // Fecha o menu ao clicar em um link
  nav.querySelectorAll("a").forEach((link) => {
    link.addEventListener("click", () => {
      burger.classList.remove("open");
      nav.classList.remove("open");
      document.body.style.overflow = "auto";
    });
  });

  // Efeito de sombra no header ao rolar
  window.addEventListener("scroll", () => {
    if (window.scrollY > 10) {
      header.style.boxShadow = "0 4px 6px -1px rgba(0, 0, 0, 0.1)";
    } else {
      header.style.boxShadow = "2px 2px 4px rgba(0, 0, 0, 0.1)";
    }
  });
});
