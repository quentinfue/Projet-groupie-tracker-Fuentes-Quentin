console.log("✅ app.js chargé");

document.addEventListener("click", async (e) => {
  const btn = e.target.closest(".fav-btn");
  if (!btn) return;

  console.log("⭐ click fav");
  const id = btn.getAttribute("data-fav-id");
  console.log("id =", id);

  try {
    const res = await fetch(`/api/favorites/toggle?id=${encodeURIComponent(id)}`, { method: "POST" });
    console.log("status =", res.status);

    const txt = await res.text();
    console.log("body =", txt);

    let data = null;
    try { data = JSON.parse(txt); } catch {}

    if (data && data.favorite === true) btn.classList.add("is-fav");
    else btn.classList.remove("is-fav");

  } catch (err) {
    console.error("fetch error:", err);
  }
});
