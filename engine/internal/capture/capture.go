// Package capture provides session recording/playback for game sessions.
package capture

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Session represents a captured game session.
type Session struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	AccountID string        `bson:"accountId" json:"accountId"`
	Player    string        `bson:"player" json:"player"`
	StartedAt time.Time     `bson:"startedAt" json:"startedAt"`
	EndedAt   *time.Time    `bson:"endedAt,omitempty" json:"endedAt,omitempty"`
	Lines     []Line        `bson:"lines" json:"lines"`
}

// Line is a single line in a captured session.
type Line struct {
	Time time.Time `bson:"time" json:"time"`
	Type string    `bson:"type" json:"type"` // "input", "output", "system", "broadcast"
	Text string    `bson:"text" json:"text"`
}

// Store handles session capture persistence.
type Store struct {
	coll *mongo.Collection
}

// NewStore creates a new capture store.
func NewStore(db *mongo.Database) *Store {
	if db == nil {
		return &Store{}
	}
	return &Store{coll: db.Collection("session_captures")}
}

// Start creates a new capture session.
func (s *Store) Start(ctx context.Context, accountID, player string) (string, error) {
	if s.coll == nil {
		return "", nil
	}
	sess := Session{
		AccountID: accountID,
		Player:    player,
		StartedAt: time.Now(),
		Lines:     []Line{}, // must be empty array, not nil, for $push to work
	}
	result, err := s.coll.InsertOne(ctx, sess)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(bson.ObjectID).Hex(), nil
}

// AppendLines adds lines to an active capture.
func (s *Store) AppendLines(ctx context.Context, captureID string, lines []Line) error {
	if s.coll == nil || captureID == "" {
		return nil
	}
	oid, err := bson.ObjectIDFromHex(captureID)
	if err != nil {
		return err
	}
	_, err = s.coll.UpdateByID(ctx, oid, bson.M{
		"$push": bson.M{"lines": bson.M{"$each": lines}},
	})
	return err
}

// Stop marks a capture as ended.
func (s *Store) Stop(ctx context.Context, captureID string) error {
	if s.coll == nil || captureID == "" {
		return nil
	}
	oid, err := bson.ObjectIDFromHex(captureID)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = s.coll.UpdateByID(ctx, oid, bson.M{"$set": bson.M{"endedAt": now}})
	return err
}

// List returns captures for an account, newest first.
func (s *Store) List(ctx context.Context, accountID string) ([]Session, error) {
	if s.coll == nil {
		return nil, nil
	}
	opts := options.Find().SetSort(bson.D{{Key: "startedAt", Value: -1}}).SetLimit(50)
	// Only return metadata, not full lines
	opts.SetProjection(bson.M{"lines": 0})
	cursor, err := s.coll.Find(ctx, bson.M{"accountId": accountID}, opts)
	if err != nil {
		return nil, err
	}
	var sessions []Session
	cursor.All(ctx, &sessions)
	return sessions, nil
}

// Get returns a full capture session with lines.
func (s *Store) Get(ctx context.Context, captureID string) (*Session, error) {
	if s.coll == nil {
		return nil, nil
	}
	oid, err := bson.ObjectIDFromHex(captureID)
	if err != nil {
		return nil, err
	}
	var sess Session
	err = s.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&sess)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// ExportText returns a capture as plain text.
func (s *Session) ExportText() string {
	var b strings.Builder
	b.WriteString("=== Session Capture: " + s.Player + " ===\n")
	b.WriteString("Started: " + s.StartedAt.Format(time.RFC1123) + "\n")
	if s.EndedAt != nil {
		b.WriteString("Ended: " + s.EndedAt.Format(time.RFC1123) + "\n")
	}
	b.WriteString("\n")
	for _, line := range s.Lines {
		ts := line.Time.Format("15:04:05")
		if line.Type == "input" {
			b.WriteString(ts + " > " + line.Text + "\n")
		} else {
			b.WriteString(ts + "   " + line.Text + "\n")
		}
	}
	return b.String()
}
