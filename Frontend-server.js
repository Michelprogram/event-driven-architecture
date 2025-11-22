// ================================
// 🌐 47FRONTEND SERVER + KAFKA PRODUCER
// ================================

import express from "express";
import bodyParser from "body-parser";
import cors from "cors";
import { Kafka } from "kafkajs";

// --- Création du serveur Express ---
const app = express();
app.use(cors());
app.use(bodyParser.json());

// --- Sert les fichiers du dossier "public" ---
app.use(express.static("public"));

// --- Configuration Kafka ---
const kafka = new Kafka({
  clientId: "frontend",
  brokers: ["localhost:9092"], // ⚠️ changer en "kafka:9092" quand Docker sera utilisé
});

const producer = kafka.producer();

// --- Connexion Kafka ---
await producer.connect();
console.log("✅ Kafka connecté");

// --- Endpoint principal : POST /buy ---
app.post("/buy", async (req, res) => {
  try {
    const { items, total } = req.body;

    // 1️⃣ Construction de l'événement principal
    const event = {
      event: "commande.initialisee",
      data: {
        idCommande: `cmd-${Date.now()}`,
        idUtilisateur: "user1",
        produits: items || [],
        prixTotal: total || 0,
      },
      timestamp: new Date().toISOString(),
    };

    // 2️⃣ Envoi dans le topic principal
    await producer.send({
      topic: "commande.initialisee",
      messages: [{ value: JSON.stringify(event) }],
    });

    console.log("📤 Event envoyé dans commande.initialisee :", event);

    // 3️⃣ Envoi dans logs.central (obligatoire pour le projet)
    await producer.send({
      topic: "logs.central",
      messages: [
        {
          value: JSON.stringify({
            source: "frontend",
            event: "commande.initialisee",
            payload: event,
            timestamp: new Date().toISOString(),
          }),
        },
      ],
    });

    console.log("📝 Log envoyé dans logs.central");

    return res.json({ success: true, event });

  } catch (err) {
    console.error("❌ Erreur Kafka :", err);
    return res.status(500).json({ error: "Erreur lors de l'envoi Kafka" });
  }
});

// --- Lancement du serveur HTTP ---
const PORT = 8080;
app.listen(PORT, "0.0.0.0", () => {
  console.log(`🚀 Serveur Frontend en ligne sur http://localhost:${PORT}`);
});

