const mongoHost = process.env.PHARMACY_API_MONGODB_HOST
const mongoPort = process.env.PHARMACY_API_MONGODB_PORT

const mongoUser = process.env.PHARMACY_API_MONGODB_USERNAME
const mongoPassword = process.env.PHARMACY_API_MONGODB_PASSWORD

const database = process.env.PHARMACY_API_MONGODB_DATABASE
const collection = process.env.PHARMACY_API_MONGODB_COLLECTION

const retrySeconds = parseInt(process.env.RETRY_CONNECTION_SECONDS || "5") || 5;

let connection;
while (true) {
    try {
        connection = Mongo(`mongodb://${mongoUser}:${mongoPassword}@${mongoHost}:${mongoPort}`);
        break;
    } catch (exception) {
        print(`Cannot connect to mongoDB: ${exception}`);
        print(`Will retry after ${retrySeconds} seconds`);
        sleep(retrySeconds * 1000);
    }
}

// Idempotent: if the collection already exists, exit cleanly.
const databases = connection.getDBNames();
if (databases.includes(database)) {
    const dbInstance = connection.getDB(database);
    const collections = dbInstance.getCollectionNames();
    if (collections.includes(collection)) {
        print(`Collection '${collection}' already exists in database '${database}'`);
        process.exit(0);
    }
}

const db = connection.getDB(database);
db.createCollection(collection);
db[collection].createIndex({ "id": 1 });

const result = db[collection].insertMany([
    {
        "id": "lekaren-centrum",
        "name": "Lekáreň Centrum",
        "address": "Námestie 1, Bratislava",
        "predefinedCategories": [
            { "code": "analgesics",   "value": "Analgetiká" },
            { "code": "antibiotics",  "value": "Antibiotiká" },
            { "code": "vitamins",     "value": "Vitamíny" },
            { "code": "hygiene",      "value": "Hygiena" },
            { "code": "supplements",  "value": "Výživové doplnky" }
        ],
        "products": [
            {
                "id": "p-paralen-500",
                "name": "Paralen 500 mg",
                "stock": 42,
                "active": true,
                "category": { "code": "analgesics", "value": "Analgetiká" }
            },
            {
                "id": "p-vitamin-c",
                "name": "Vitamín C 1000",
                "stock": 17,
                "active": true,
                "category": { "code": "vitamins", "value": "Vitamíny" }
            }
        ]
    }
]);

if (result.writeError) {
    console.error(result);
    print(`Error when writing the data: ${result.errmsg}`);
}

process.exit(0);
