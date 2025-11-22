// ================================
// 🌐 47FRONTEND SERVER + KAFKA PRODUCER
// ================================

import express from "express";
import bodyParser from "body-parser";
import cors from "cors";

// --- Création du serveur Express ---
const app = express();
app.use(cors());
app.use(bodyParser.json());

// --- Sert les fichiers du dossier "public" ---
app.use(express.static("public"));

// Proxy to users-service to avoid CORS
app.post("/api/register", async (req, res) => {
  try {
    const { username, password } = req.body || {};
    if (!username || !password)
      return res
        .status(400)
        .json({ error: "username and password are required" });

    const upstream = await fetch("http://users-service:3001/register", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({ username, password }),
    });

    const text = await upstream.text();
    res.status(upstream.status).send(text);
  } catch (e) {
    res.status(500).json({ error: "proxy error" });
  }
});

app.post("/api/login", async (req, res) => {
  try {
    const { username, password } = req.body || {};
    if (!username || !password)
      return res
        .status(400)
        .json({ error: "username and password are required" });

    const upstream = await fetch("http://users-service:3001/login", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({ username, password }),
    });

    const text = await upstream.text(); // Go service returns token in body
    res.status(upstream.status).send(text);
  } catch (e) {
    res.status(500).json({ error: "proxy error" });
  }
});

// Add below other /api/* routes
app.get("/api/orders", async (req, res) => {
  try {
    const upstream = await fetch("http://orders-service:3003/orders");
    const data = await upstream.json();
    res.status(upstream.status).json(data);
  } catch (e) {
    res.status(500).json({ error: "proxy error" });
  }
});

// Add below the other /api/* routes
app.post("/api/pay", async (req, res) => {
  try {
    const { itemName, itemQuantity } = req.body || {};
    if (!itemName || !itemQuantity) {
      return res
        .status(400)
        .json({ error: "itemName and itemQuantity are required" });
    }

    const upstream = await fetch("http://payments-service:3002/pay", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        itemName,
        itemQuantity: String(itemQuantity),
      }),
    });

    const text = await upstream.text();
    res.status(upstream.status).send(text);
  } catch (e) {
    res.status(500).json({ error: "proxy error" });
  }
});

// --- Lancement du serveur HTTP ---
const PORT = 8080;
app.listen(PORT, "0.0.0.0", () => {
  console.log(`🚀 Serveur Frontend en ligne sur http://localhost:${PORT}`);
});
