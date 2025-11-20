// Connect to the MongoDB instance
db = db.getSiblingDB('orders-db');
db.createCollection('orders');
db.orders.createIndex({ userId: 1 });
print('orders collection created and index on userId added');
