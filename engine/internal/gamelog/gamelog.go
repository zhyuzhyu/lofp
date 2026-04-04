package gamelog

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type EventType string

const (
	EventLogin           EventType = "login"            // User authenticated via Google
	EventLogout          EventType = "logout"           // User session ended
	EventGameEnter       EventType = "game_enter"       // Character entered the game world
	EventGameExit        EventType = "game_exit"        // Character left the game world
	EventCharacterCreate EventType = "character_create"  // New character created
	EventLevelUp         EventType = "level_up"
	EventGMGrant         EventType = "gm_grant"
	EventGMRevoke        EventType = "gm_revoke"
	EventReport          EventType = "report"            // Player REPORT command
)

type LogEntry struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
	Event     EventType     `bson:"event" json:"event"`
	Player    string        `bson:"player" json:"player"`
	AccountID string        `bson:"accountId,omitempty" json:"accountId,omitempty"`
	Details   string        `bson:"details,omitempty" json:"details,omitempty"`
	RoomNum   int           `bson:"roomNum,omitempty" json:"roomNum,omitempty"`
	RoomName  string        `bson:"roomName,omitempty" json:"roomName,omitempty"`
}

type Logger struct {
	coll *mongo.Collection
}

func New(db *mongo.Database) *Logger {
	if db == nil {
		return &Logger{}
	}
	coll := db.Collection("game_logs")

	// Create TTL index to auto-expire old logs (90 days)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "timestamp", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(90 * 24 * 60 * 60),
	})
	// Index for querying by event type and player
	coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "event", Value: 1}, {Key: "timestamp", Value: -1}},
	})
	coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "player", Value: 1}, {Key: "timestamp", Value: -1}},
	})

	return &Logger{coll: coll}
}

func (l *Logger) Log(event EventType, player, accountID, details string, roomNum int, roomName string) {
	if l.coll == nil {
		return
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Event:     event,
		Player:    player,
		AccountID: accountID,
		Details:   details,
		RoomNum:   roomNum,
		RoomName:  roomName,
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := l.coll.InsertOne(ctx, entry); err != nil {
			log.Printf("gamelog: failed to write log: %v", err)
		}
	}()
}

func (l *Logger) Query(ctx context.Context, eventFilter string, playerFilter string, limit int) ([]LogEntry, error) {
	if l.coll == nil {
		return nil, nil
	}
	filter := bson.M{}
	if eventFilter != "" {
		filter["event"] = eventFilter
	}
	if playerFilter != "" {
		filter["player"] = bson.M{"$regex": playerFilter, "$options": "i"}
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(int64(limit))
	cursor, err := l.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var entries []LogEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
