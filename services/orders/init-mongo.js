// Connect to the MongoDB instance
db = db.getSiblingDB("admin");
db.auth("admin_root", "password_root");

db = db.getSiblingDB("orders-db");

db.createUser({
  user: "user_app",
  pwd: "strong_app_password",
  roles: [{ role: "readWrite", db: "orders-db" }],
});

db.createCollection("orders");
db.orders.createIndex({ userId: 1 });