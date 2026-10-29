(() => {
    const root = document.documentElement;
    const storageKey = "todo-theme";
    let theme = "light";

    // Read before the page renders to avoid flashing the wrong theme.
    try {
        const saved = localStorage.getItem(storageKey);
        if (saved === "light" || saved === "dark") theme = saved;
    } catch {
        // Theme switching still works when browser storage is unavailable.
    }
    root.dataset.theme = theme;

    document.addEventListener("DOMContentLoaded", () => {
        const button = document.getElementById("theme-toggle");
        if (!button) return;
        const updateButton = () => {
            const dark = root.dataset.theme === "dark";
            button.textContent = dark ? "Switch to light" : "Switch to dark";
            button.setAttribute("aria-pressed", String(dark));
        };
        updateButton();
        button.hidden = false;
        button.addEventListener("click", () => {
            theme = root.dataset.theme === "dark" ? "light" : "dark";
            root.dataset.theme = theme;
            updateButton();
            try {
                localStorage.setItem(storageKey, theme);
            } catch {
                // Keep the current selection even if it cannot be persisted.
            }
        });
    });
})();
