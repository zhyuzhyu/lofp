package engine

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func connectTestDB(t *testing.T) *mongo.Database {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the same MongoDB as the game (from env)
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI not set")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Skipf("Cannot connect to MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Cannot ping MongoDB: %v", err)
	}
	return client.Database("lofp")
}

func TestLoadPlayerTaliesin(t *testing.T) {
	db := connectTestDB(t)
	ctx := context.Background()

	// Direct MongoDB query to find Taliesin
	coll := db.Collection("players")

	// First, find ALL characters with firstName "Taliesin" (including any soft-deleted)
	cursor, err := coll.Find(ctx, bson.M{"firstName": "Taliesin"})
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}
	var allPlayers []Player
	if err := cursor.All(ctx, &allPlayers); err != nil {
		t.Fatalf("Cursor error: %v", err)
	}

	t.Logf("Found %d player(s) with firstName 'Taliesin':", len(allPlayers))
	for _, p := range allPlayers {
		t.Logf("  Name: %s %s, AccountID: %s, Room: %d, Level: %d, DeletedAt: %v",
			p.FirstName, p.LastName, p.AccountID, p.RoomNumber, p.Level, p.DeletedAt)
	}

	if len(allPlayers) == 0 {
		t.Fatal("No character named Taliesin found in database!")
	}

	// Now test the LoadPlayer function (which filters out soft-deleted)
	player := allPlayers[0]
	filter := bson.M{
		"firstName": player.FirstName,
		"lastName":  player.LastName,
		"deletedAt": bson.M{"$exists": false},
	}
	var loaded Player
	err = coll.FindOne(ctx, filter).Decode(&loaded)
	if err != nil {
		t.Logf("LoadPlayer filter FAILED for %s %s: %v", player.FirstName, player.LastName, err)
		t.Logf("This means the character has a deletedAt field or the name doesn't match exactly")

		// Check if deletedAt exists
		var raw bson.M
		coll.FindOne(ctx, bson.M{"firstName": "Taliesin"}).Decode(&raw)
		if raw != nil {
			if da, ok := raw["deletedAt"]; ok {
				t.Logf("  deletedAt field exists: %v", da)
			} else {
				t.Logf("  deletedAt field does NOT exist (should match)")
			}
			t.Logf("  firstName: %q, lastName: %q", raw["firstName"], raw["lastName"])
		}
	} else {
		t.Logf("LoadPlayer succeeded: %s %s (account: %s)", loaded.FirstName, loaded.LastName, loaded.AccountID)
	}

	// Check all accounts in the system
	accountColl := db.Collection("accounts")
	acCursor, _ := accountColl.Find(ctx, bson.M{})
	var accounts []bson.M
	acCursor.All(ctx, &accounts)
	t.Logf("All accounts in system:")
	for _, a := range accounts {
		name, _ := a["name"].(string)
		email, _ := a["email"].(string)
		id := a["_id"]
		t.Logf("  ID=%v name=%s email=%s", id, name, email)
	}

	// Test ListPlayersByAccount
	if player.AccountID != "" {
		accountFilter := bson.M{"accountId": player.AccountID, "deletedAt": bson.M{"$exists": false}}
		cursor2, _ := coll.Find(ctx, accountFilter)
		var accountPlayers []Player
		cursor2.All(ctx, &accountPlayers)
		t.Logf("Characters for account %s: %d", player.AccountID, len(accountPlayers))
		for _, p := range accountPlayers {
			t.Logf("  %s %s (level %d)", p.FirstName, p.LastName, p.Level)
		}
	}
}
