document.addEventListener("DOMContentLoaded", () => {
  const links = document.querySelectorAll(".docs-sidebar a");
  const sections = document.querySelectorAll(
    ".docs-content h2, .docs-content h1"
  );

  // Função para destacar o link ativo na sidebar
  function highlightSidebar() {
    let scrollY = window.scrollY;

    sections.forEach((current) => {
      const sectionHeight = current.offsetHeight;
      const sectionTop = current.offsetTop - 150; // Ajuste fino para o header fixo
      const sectionId = current.getAttribute("id");

      if (scrollY > sectionTop && scrollY <= sectionTop + sectionHeight + 200) {
        // +200 pega o conteúdo abaixo do título
        links.forEach((link) => {
          link.classList.remove("active");
          if (link.getAttribute("href").includes(sectionId)) {
            link.classList.add("active");
          }
        });
      }
    });
  }

  window.addEventListener("scroll", highlightSidebar);
});
