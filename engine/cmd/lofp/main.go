package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"crypto/rand"
	"encoding/hex"

	"github.com/jonradoff/lofp/internal/api"
	"github.com/jonradoff/lofp/internal/auth"
	"github.com/jonradoff/lofp/internal/capture"
	"github.com/jonradoff/lofp/internal/config"
	"github.com/jonradoff/lofp/internal/engine"
	"github.com/jonradoff/lofp/internal/gamelog"
	"github.com/jonradoff/lofp/internal/gameworld"
	"github.com/jonradoff/lofp/internal/hub"
	"github.com/jonradoff/lofp/internal/scriptparser"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	configPath := "config/dev.yaml"
	if p := os.Getenv("LOFP_CONFIG"); p != "" {
		configPath = p
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Parse game scripts
	log.Println("Parsing game scripts...")
	start := time.Now()
	result, err := scriptparser.ParseConfig(cfg.Game.ConfigFile)
	if err != nil {
		log.Fatalf("Failed to parse scripts: %v", err)
	}
	log.Printf("Parsed %d rooms, %d items, %d monsters, %d nouns, %d adjectives in %v",
		len(result.Rooms), len(result.Items), len(result.Monsters),
		len(result.Nouns), len(result.Adjectives), time.Since(start))

	// Convert to ParsedData
	parsed := &gameworld.ParsedData{
		Rooms:        result.Rooms,
		Items:        result.Items,
		Monsters:     result.Monsters,
		Nouns:        result.Nouns,
		Adjectives:   result.Adjectives,
		MonsterAdjs:  result.MonsterAdjs,
		Variables:    result.Variables,
		Regions:      result.Regions,
		MonsterLists: result.MonsterLists,
		StartRoom:    result.StartRoom,
		BumpRoom:     result.BumpRoom,
	}

	// Connect to MongoDB (optional — game works without it)
	var db *mongo.Database
	if cfg.MongoDB.URI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoDB.URI))
		if err != nil {
			log.Printf("Warning: MongoDB connection failed: %v (continuing without persistence)", err)
		} else {
			if err := client.Ping(ctx, nil); err != nil {
				log.Printf("Warning: MongoDB ping failed: %v (continuing without persistence)", err)
			} else {
				db = client.Database(cfg.MongoDB.Database)
				log.Printf("Connected to MongoDB: %s", cfg.MongoDB.Database)
			}
		}
	}

	// Generate unique machine ID
	machineBytes := make([]byte, 8)
	rand.Read(machineBytes)
	machineID := hex.EncodeToString(machineBytes)

	// Create game engine
	ge := engine.NewGameEngine(db, parsed)

	// Create game logger
	gl := gamelog.New(db)

	// Create cross-machine hub
	h := hub.New(db, machineID)

	// Create auth service
	var authSvc *auth.Service
	if cfg.Auth.GoogleClientID != "" {
		authSvc = auth.NewService(db, cfg.Auth.GoogleClientID, cfg.Auth.JWTSecret)
		log.Printf("Google OAuth enabled (client ID: %s...)", cfg.Auth.GoogleClientID[:min(20, len(cfg.Auth.GoogleClientID))])
	} else {
		log.Println("Google OAuth not configured (set GOOGLE_CLIENT_ID to enable)")
	}

	// Create capture store
	cs := capture.NewStore(db)

	// Create API server
	srv := api.NewServer(ge, parsed, authSvc, gl, h, cs, cfg.Server.FrontendURL)
	h.Start()
	ge.StartTimeCycle()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Legends of Future Past server starting on %s", addr)
	log.Printf("Frontend URL: %s", cfg.Server.FrontendURL)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
