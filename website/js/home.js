document.addEventListener("DOMContentLoaded", () => {
  const runBtn = document.querySelector(".code-header .btn-primary"); // Botão Executar
  const resetBtn = document.querySelector(".code-header .btn-gray"); // Botão Reset
  const outputArea = document.querySelector(".output");

  if (runBtn && outputArea) {
    runBtn.addEventListener("click", () => {
      outputArea.value = "Compiling...";
      outputArea.style.color = "#aaa";

      // Simula um delay de processamento
      setTimeout(() => {
        outputArea.value = "Hello Alpha\n\n> Program exited with code 0";
        outputArea.style.color = "#4ade80"; // Verde terminal
      }, 800);
    });

    resetBtn.addEventListener("click", () => {
      outputArea.value = "";
    });
  }
});
