db = db.getSiblingDB("admin");
db.auth("admin_root", "password_root");

db = db.getSiblingDB("user_db");

db.createUser({
  user: "user_app",
  pwd: "strong_app_password",
  roles: [{ role: "readWrite", db: "user_db" }],
});

db.createCollection("users");

// Password sha1 for "password123"
db.users.insertMany([
  {
    username: "rick sanchez",
    passwordHash: "cbfdac6008f9cab4083784cbd1874f76618d2a97",
  },
  {
    username: "morty smith",
    passwordHash: "cbfdac6008f9cab4083784cbd1874f76618d2a97",
  },
]);
