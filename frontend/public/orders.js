async function fetchOrders() {
  const res = await fetch("/api/orders");
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return await res.json();
}

function renderOrders(ordersRaw) {
  const tbody = document.getElementById("ordersTbody");
  const empty = document.getElementById("ordersEmpty");
  if (!tbody) return;

  // Normalize possible shapes
  const orders = Array.isArray(ordersRaw)
    ? ordersRaw
    : (ordersRaw && (ordersRaw.data || ordersRaw.orders)) || [];

  tbody.innerHTML = "";
  if (!orders.length) {
    if (empty) empty.style.display = "";
    return;
  }
  if (empty) empty.style.display = "none";

  orders.forEach((o) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
        <td class="p-2">${o.userId ?? ""}</td>
        <td class="p-2">${o.productId ?? ""}</td>
        <td class="p-2">${o.quantity ?? ""}</td>
      `;
    tbody.appendChild(tr);
  });
}

async function loadOrders() {
  try {
    const data = await fetchOrders();
    renderOrders(data);
  } catch (e) {
    console.error("orders load error", e);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const page = document.body.dataset.page;

  if (page === "orders") {
    updateCartBadge();
    loadOrders();
    // poll every 5s for new orders
    setInterval(loadOrders, 5000);
  }
});
