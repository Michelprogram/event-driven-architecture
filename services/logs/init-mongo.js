db = db.getSiblingDB("admin");
db.auth("admin_root", "password_root");

db = db.getSiblingDB("log_db");

db.createUser({
  user: "user_app",
  pwd: "strong_app_password",
  roles: [{ role: "readWrite", db: "log_db" }],
});

db.createCollection("logs");
