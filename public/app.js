// ============================
// 🛹 SkateShop Frontend Logic
// ============================

// ---- Liste des produits (doit correspondre à tes images Figma) ----
const PRODUCTS = {
  "pro-street": {
    id: "pro-street",
    title: "Pro Street Cruiser",
    desc: "Perfect for street skating and tricks",
    price: 89.99,
    img: "assets/pro-street.jpeg"   
  },
  "elite-deck": {
    id: "elite-deck",
    title: "Elite Performance Deck",
    desc: "Professional grade for advanced riders",
    price: 129.99,
    img: "assets/elite-deck.jpeg"    
  },
  "sunset-rider": {
    id: "sunset-rider",
    title: "Sunset Rider",
    desc: "Vibrant design with smooth ride",
    price: 99.99,
    img: "assets/sunset-rider.jpeg"  
  },
  "park-master": {
    id: "park-master",
    title: "Park Master",
    desc: "Built for skate park sessions",
    price: 109.99,
    img: "assets/park-master.jpeg"   
  }
};


// ---- Gestion du panier via localStorage ----
const CART_KEY = "skateshop_cart";

function getCart() {
  return JSON.parse(localStorage.getItem(CART_KEY) || "{}");
}

function saveCart(cart) {
  localStorage.setItem(CART_KEY, JSON.stringify(cart));
}

function getCartCount() {
  return Object.values(getCart()).reduce((sum, qty) => sum + qty, 0);
}

// ---- Mise à jour du badge View Cart ----
function updateCartBadge() {
  const countEl = document.getElementById("cartCount");
  if (countEl) countEl.textContent = getCartCount();
}

// ---- Page : Our Collection ----
function setupCollectionPage() {
  document.querySelectorAll("[data-add]").forEach(btn => {
    btn.addEventListener("click", () => {
      const id = btn.dataset.add;
      const cart = getCart();
      cart[id] = (cart[id] || 0) + 1;
      saveCart(cart);
      updateCartBadge();

      // petit effet visuel
      btn.textContent = "Added ✓";
      btn.disabled = true;
      setTimeout(() => {
        btn.textContent = "Add to Cart";
        btn.disabled = false;
      }, 800);
    });
  });
  updateCartBadge();
}

// ---- Page : Shopping Cart ----
function money(n) {
  return `$${n.toFixed(2)}`;
}

const TAX_RATE = 0.08;

function renderCartPage() {
  const list = document.getElementById("cartList");
  if (!list) return;

  const cart = getCart();
  list.innerHTML = "";
  let subtotal = 0;

  Object.entries(cart).forEach(([id, qty]) => {
    const product = PRODUCTS[id];
    if (!product) return;

    subtotal += product.price * qty;

    const item = document.createElement("div");
    item.className = "item";
    item.innerHTML = `
      <img src="${product.img}" alt="${product.title}" class="item__img" />
      <div>
        <h4 class="item__title">${product.title}</h4>
        <p class="item__desc">${product.desc}</p>
        <div class="item__price">${money(product.price)}</div>
      </div>
      <div class="item__qty">
        <button class="qtybtn" data-dec="${id}">−</button>
        <span>${qty}</span>
        <button class="qtybtn" data-inc="${id}">+</button>
        <span class="trash" data-del="${id}" title="Supprimer">🗑️</span>
      </div>
    `;
    list.appendChild(item);
  });

  // Calcul du total
  const tax = subtotal * TAX_RATE;
  const total = subtotal + tax;
  document.getElementById("sumSubtotal").textContent = money(subtotal);
  document.getElementById("sumTax").textContent = money(tax);
  document.getElementById("sumTotal").textContent = money(total);

  // Boutons quantité et suppression
  list.querySelectorAll("[data-inc]").forEach(btn => {
    btn.onclick = () => {
      const id = btn.dataset.inc;
      const cart = getCart();
      cart[id] = (cart[id] || 0) + 1;
      saveCart(cart);
      renderCartPage();
      updateCartBadge();
    };
  });

  list.querySelectorAll("[data-dec]").forEach(btn => {
    btn.onclick = () => {
      const id = btn.dataset.dec;
      const cart = getCart();
      cart[id] = Math.max(0, (cart[id] || 0) - 1);
      if (cart[id] === 0) delete cart[id];
      saveCart(cart);
      renderCartPage();
      updateCartBadge();
    };
  });

  list.querySelectorAll("[data-del]").forEach(btn => {
    btn.onclick = () => {
      const id = btn.dataset.del;
      const cart = getCart();
      delete cart[id];
      saveCart(cart);
      renderCartPage();
      updateCartBadge();
    };
  });
}

// ---- Fonction Checkout ----
async function checkout() {
  const status = document.getElementById("checkoutStatus");
  status.textContent = "⌛ Sending order to Kafka...";
  try {
    const cart = getCart();
    const items = Object.entries(cart).map(([id, qty]) => ({
      id,
      qty,
    }));
    const total = Object.entries(cart).reduce(
      (acc, [id, qty]) => acc + PRODUCTS[id].price * qty,
      0
    );

    // Envoi vers ton backend Node.js (Kafka)
    const res = await fetch("/buy", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ items, total }),
    });

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    status.textContent = "✅ Order sent! (commande.initialisee)";
    console.log("Event envoyé à Kafka:", data);

    // vider le panier après envoi
    localStorage.removeItem(CART_KEY);
    updateCartBadge();
    renderCartPage();
  } catch (err) {
    console.error(err);
    status.textContent = "❌ Erreur lors de l'envoi";
  }
}

// ---- Initialisation selon la page ----
document.addEventListener("DOMContentLoaded", () => {
  const page = document.body.dataset.page;

  if (page === "collection") {
    setupCollectionPage();
  }

  if (page === "cart") {
    updateCartBadge();
    renderCartPage();
    const btn = document.getElementById("checkoutBtn");
    if (btn) btn.addEventListener("click", checkout);
  }
});
