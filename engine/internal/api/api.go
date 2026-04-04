package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jonradoff/lofp/internal/auth"
	"github.com/jonradoff/lofp/internal/capture"
	"github.com/jonradoff/lofp/internal/engine"
	"github.com/jonradoff/lofp/internal/gamelog"
	"github.com/jonradoff/lofp/internal/gameworld"
	"github.com/jonradoff/lofp/internal/hub"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Server struct {
	engine      *engine.GameEngine
	parsed      *gameworld.ParsedData
	auth        *auth.Service
	gamelog     *gamelog.Logger
	hub         *hub.Hub
	captures    *capture.Store
	router      *mux.Router
	upgrader    websocket.Upgrader
	sessions    map[string]*Session
	mu          sync.RWMutex
	frontendURL string
}

type Session struct {
	Player    *engine.Player
	Conn      *websocket.Conn
	mu        sync.Mutex
	CaptureID string // active capture session ID, empty if not recording
}


func NewServer(ge *engine.GameEngine, parsed *gameworld.ParsedData, authSvc *auth.Service, gl *gamelog.Logger, h *hub.Hub, cs *capture.Store, frontendURL string) *Server {
	s := &Server{
		engine:      ge,
		parsed:      parsed,
		auth:        authSvc,
		gamelog:     gl,
		hub:         h,
		captures:    cs,
		sessions:    make(map[string]*Session),
		frontendURL: frontendURL,
	}
	s.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // Allow non-browser clients
			}
			return strings.HasPrefix(origin, s.frontendURL)
		},
	}
	s.router = mux.NewRouter()
	s.setupRoutes()
	ge.SetSessionProvider(s)

	// Set up cross-machine coordination
	h.SetDeliveryFunc(s.deliverRemoteEvent)
	ge.SetRoomChangeCallback(func(change engine.RoomChange) {
		s.hub.Publish(&hub.Event{
			Type: "room_state_change",
			Data: bson.M{
				"roomNumber": change.RoomNumber,
				"type":       change.Type,
				"itemRef":    change.ItemRef,
				"newState":   change.NewState,
				"item":       change.Item,
			},
		})
	})

	// Set up room broadcast for background tasks (monsters, CEVENTs)
	ge.SetRoomBroadcast(func(roomNumber int, messages []string) {
		s.broadcastToRoom(roomNumber, "", messages)
	})

	// Local-only broadcast for monster activity (no hub, this machine only)
	ge.SetLocalRoomBroadcast(func(roomNumber int, messages []string) {
		s.mu.RLock()
		for _, sess := range s.sessions {
			if sess.Player == nil || sess.Player.RoomNumber != roomNumber {
				continue
			}
			s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": messages})
		}
		s.mu.RUnlock()
	})

	// Set up player-targeted messages for background tasks (combat, etc.)
	ge.SetSendToPlayer(func(playerName string, messages []string) {
		s.sendToPlayer(playerName, messages)
	})

	return s
}

// OnlinePlayers returns all currently connected players (implements engine.SessionProvider).
func (s *Server) OnlinePlayers() []*engine.Player {
	// Start with local players
	s.mu.RLock()
	localNames := make(map[string]bool, len(s.sessions))
	players := make([]*engine.Player, 0, len(s.sessions))
	for _, sess := range s.sessions {
		players = append(players, sess.Player)
		localNames[sess.Player.FirstName] = true
	}
	s.mu.RUnlock()

	// Add remote players from hub
	for _, rp := range s.hub.AllOnlinePlayers() {
		if localNames[rp.FirstName] {
			continue // already have this player locally
		}
		// Create a lightweight Player for remote presence
		lastName := ""
		if parts := strings.SplitN(rp.FullName, " ", 2); len(parts) > 1 {
			lastName = parts[1]
		}
		players = append(players, &engine.Player{
			FirstName:  rp.FirstName,
			LastName:   lastName,
			RoomNumber: rp.RoomNumber,
			Race:       rp.Race,
			Position:   rp.Position,
			IsGM:       rp.IsGM,
			GMHat:      rp.GMHat,
			GMHidden:   rp.GMHidden,
			GMInvis:    rp.GMInvis,
			Hidden:     rp.Hidden,
		})
	}
	return players
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.Use(s.corsMiddleware)

	// Health check (vibectl compatible)
	s.router.HandleFunc("/healthz", s.handleHealthz).Methods("GET", "HEAD")

	// Game WebSocket
	s.router.HandleFunc("/ws/game", s.handleGameWS)
	s.router.HandleFunc("/ws/events", s.handleEventsWS)

	// REST endpoints
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Auth (public)
	api.HandleFunc("/auth/google", s.handleGoogleAuth).Methods("POST")
	api.HandleFunc("/auth/me", s.handleAuthMe).Methods("GET")

	// Characters (authenticated)
	api.HandleFunc("/characters", s.handleListCharacters).Methods("GET")
	api.HandleFunc("/characters", s.handleCreateCharacter).Methods("POST")
	api.HandleFunc("/characters/{firstName}/gm", s.handleToggleGM).Methods("PUT")

	// Session captures (authenticated)
	api.HandleFunc("/captures", s.handleListCaptures).Methods("GET")
	api.HandleFunc("/captures/{id}", s.handleGetCapture).Methods("GET")
	api.HandleFunc("/captures/{id}/text", s.handleGetCaptureText).Methods("GET")
	api.HandleFunc("/captures/{id}", s.handleDeleteCapture).Methods("DELETE")

	// Game world data (admin only)
	api.HandleFunc("/stats", s.handleStats).Methods("GET")
	api.HandleFunc("/rooms", s.handleListRooms).Methods("GET")
	api.HandleFunc("/rooms/{number}", s.handleGetRoom).Methods("GET")
	api.HandleFunc("/items", s.handleListItems).Methods("GET")
	api.HandleFunc("/items/{number}", s.handleGetItem).Methods("GET")
	api.HandleFunc("/monsters", s.handleListMonsters).Methods("GET")
	api.HandleFunc("/monsters/{number}", s.handleGetMonster).Methods("GET")
	api.HandleFunc("/nouns", s.handleListNouns).Methods("GET")
	api.HandleFunc("/adjectives", s.handleListAdjectives).Methods("GET")

	// Admin bootstrap (only works if no admins exist yet)
	api.HandleFunc("/admin/bootstrap", s.handleBootstrapAdmin).Methods("POST")

	// Admin endpoints (require admin)
	api.HandleFunc("/admin/accounts", s.handleListAccounts).Methods("GET")
	api.HandleFunc("/admin/accounts/{id}", s.handleGetAccount).Methods("GET")
	api.HandleFunc("/admin/accounts/{id}/admin", s.handleToggleAdmin).Methods("PUT")
	api.HandleFunc("/admin/characters", s.handleAdminListCharacters).Methods("GET")
	api.HandleFunc("/admin/characters/{firstName}", s.handleAdminGetCharacter).Methods("GET")
	api.HandleFunc("/admin/characters/{firstName}/reassign", s.handleAdminReassignCharacter).Methods("PUT")
	api.HandleFunc("/admin/logs", s.handleAdminLogs).Methods("GET")

	// Serve static frontend files in production (if /app/static exists)
	staticDir := os.Getenv("LOFP_STATIC_DIR")
	if staticDir == "" {
		staticDir = "/app/static"
	}
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		spa := spaHandler{staticDir: staticDir}
		s.router.PathPrefix("/").Handler(spa)
		log.Printf("Serving static frontend from %s", staticDir)
	}
}

// spaHandler serves static files, falling back to index.html for client-side routing.
type spaHandler struct {
	staticDir string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := h.staticDir + r.URL.Path
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		http.ServeFile(w, r, h.staticDir+"/index.html")
		return
	}
	http.ServeFile(w, r, path)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WebSocket message types
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type CommandMsg struct {
	Input string `json:"input"`
}

type CreateCharMsg struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Race      int    `json:"race"`
	Gender    int    `json:"gender"`
}

type AuthMsg struct {
	Token string `json:"token"`
}

func (s *Server) handleEventsWS(w http.ResponseWriter, r *http.Request) {
	// Validate admin auth via query param token
	token := r.URL.Query().Get("token")
	if token == "" || s.auth == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	accountID, err := s.auth.ValidateJWT(token)
	if err != nil || accountID == "" {
		http.Error(w, "unauthorized", 401)
		return
	}
	account, err := s.auth.GetAccount(r.Context(), accountID)
	if err != nil || account == nil || !account.IsAdmin {
		http.Error(w, "forbidden", 403)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Subscribe to engine events
	ch := s.engine.Events.Subscribe()
	defer s.engine.Events.Unsubscribe(ch)

	// Send initial status
	conn.WriteJSON(map[string]interface{}{
		"type": "event",
		"data": engine.EngineEvent{
			Time:     time.Now(),
			Category: "system",
			Message:  fmt.Sprintf("Event monitor connected. Game time: Hour %d, Day %d of %s, Year %d", engine.GameHour(), engine.GameDay(), engine.GameMonthName(), engine.GameYear()),
		},
	})

	// Read pump (just to detect disconnect)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Write pump
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.WriteJSON(map[string]interface{}{
				"type": "event",
				"data": event,
			}); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func (s *Server) handleGameWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}
	defer conn.Close()
	conn.SetReadLimit(65536) // 64KB max message size

	session := &Session{Conn: conn}
	var accountID, authName, authEmail string

	// Send welcome
	s.sendWSResult(session, &engine.CommandResult{
		Messages: []string{
			"",
			"====================================",
			" LEGENDS OF FUTURE PAST",
			" The Shattered Realms Await...",
			"====================================",
			"",
		},
	})

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WS read error: %v", err)
			}
			break
		}

		switch msg.Type {
		case "auth":
			var authMsg AuthMsg
			json.Unmarshal(msg.Data, &authMsg)
			if s.auth != nil && authMsg.Token != "" {
				claims, err := s.auth.ValidateJWTFull(authMsg.Token)
				if err != nil {
					s.sendWSMessage(session, "auth_result", map[string]interface{}{
						"success": false,
						"error":   "invalid token",
					})
					continue
				}
				accountID = claims.AccountID
				authName = claims.Name
				authEmail = claims.Email
				s.gamelog.Log(gamelog.EventLogin, claims.Name, claims.AccountID, claims.Email, 0, "")
				s.sendWSMessage(session, "auth_result", map[string]interface{}{
					"success": true,
				})
			} else {
				s.sendWSMessage(session, "auth_result", map[string]interface{}{
					"success": false,
					"error":   "auth not configured",
				})
			}

		case "create_character":
			var create CreateCharMsg
			json.Unmarshal(msg.Data, &create)

			if err := engine.ValidateCharacterInput(create.FirstName, create.LastName, create.Race, create.Gender); err != nil {
				s.sendWSMessage(session, "error", map[string]interface{}{"message": err.Error()})
				continue
			}

			ctx := context.Background()
			player, err := s.engine.LoadPlayer(ctx, create.FirstName, create.LastName)
			if err != nil || player == nil {
				player = s.engine.CreateNewPlayer(ctx, create.FirstName, create.LastName, create.Race, create.Gender, accountID)
				s.gamelog.Log(gamelog.EventCharacterCreate, player.FullName(), accountID,
					fmt.Sprintf("Race: %s, Gender: %d", player.RaceName(), player.Gender),
					player.RoomNumber, "")
				s.sendWSResult(session, &engine.CommandResult{
					Messages: []string{
						fmt.Sprintf("Welcome to the Shattered Realms, %s the %s!", player.FullName(), player.RaceName()),
						"",
					},
				})
			} else if player.AccountID != accountID {
				s.sendWSMessage(session, "error", map[string]interface{}{
					"message": fmt.Sprintf("The name '%s %s' is already taken. Please choose a different name.", create.FirstName, create.LastName),
				})
				continue
			} else {
				s.sendWSResult(session, &engine.CommandResult{
					Messages: []string{
						fmt.Sprintf("Welcome back, %s the %s!", player.FullName(), player.RaceName()),
						"",
					},
				})
			}
			session.Player = player

			s.mu.Lock()
			s.sessions[player.FirstName] = session
			s.mu.Unlock()

			// Register in cross-machine presence
			s.hub.RegisterPlayer(player.FirstName, player.FullName(), player.RoomNumber,
				player.Race, player.RaceName(), player.Position,
				player.IsGM, player.GMHat, player.GMHidden, player.GMInvis, player.Hidden)

			// Log character entering the game world
			s.gamelog.Log(gamelog.EventGameEnter, player.FullName(), accountID,
				fmt.Sprintf("%s (%s)", authName, authEmail), player.RoomNumber, "")
			s.broadcastGlobal(player.FirstName,
				[]string{fmt.Sprintf("** %s has just entered the Realms.", player.FirstName)})

			result := s.engine.EnterRoom(ctx, player)
			s.sendWSResult(session, result)
			if len(result.GMBroadcast) > 0 {
				s.broadcastToGMs(result.GMBroadcast)
			}

		case "start_capture":
			if session.Player == nil || accountID == "" {
				log.Printf("capture: cannot start — player=%v accountID=%q", session.Player != nil, accountID)
				continue
			}
			if session.CaptureID != "" {
				log.Printf("capture: already recording %s", session.CaptureID)
				continue
			}
			ctx := context.Background()
			id, err := s.captures.Start(ctx, accountID, session.Player.FullName())
			if err != nil {
				log.Printf("capture: start failed: %v", err)
			} else if id != "" {
				session.CaptureID = id
				log.Printf("capture: started %s for %s", id, session.Player.FullName())
				s.sendWSMessage(session, "capture_status", map[string]interface{}{"recording": true, "id": id})
			}
			continue

		case "stop_capture":
			if session.CaptureID != "" {
				ctx := context.Background()
				s.captures.Stop(ctx, session.CaptureID)
				s.sendWSMessage(session, "capture_status", map[string]interface{}{"recording": false})
				session.CaptureID = ""
			}
			continue

		case "command":
			if session.Player == nil {
				s.sendWSResult(session, &engine.CommandResult{
					Messages: []string{"You must create a character first."},
				})
				continue
			}
			var cmd CommandMsg
			json.Unmarshal(msg.Data, &cmd)
			ctx := context.Background()
			playerRoom := session.Player.RoomNumber
			result := s.engine.ProcessCommand(ctx, session.Player, cmd.Input)
			result.PlayerState = session.Player
			result.PromptIndicators = session.Player.PromptIndicators()
			s.sendWSResult(session, result)

			// Capture input and output
			if session.CaptureID != "" {
				var lines []capture.Line
				lines = append(lines, capture.Line{Time: time.Now(), Type: "input", Text: cmd.Input})
				for _, m := range result.Messages {
					lines = append(lines, capture.Line{Time: time.Now(), Type: "output", Text: m})
				}
				log.Printf("capture: appending %d lines to %s", len(lines), session.CaptureID)
				go func() {
					if err := s.captures.AppendLines(context.Background(), session.CaptureID, lines); err != nil {
						log.Printf("capture: append failed: %v", err)
					}
				}()
			}

			// Broadcast to others in the room (excluding emote target if they get a special message)
			if len(result.RoomBroadcast) > 0 {
				if result.TargetName != "" {
					s.broadcastToRoom(session.Player.RoomNumber, session.Player.FirstName, result.RoomBroadcast, result.TargetName)
				} else {
					s.broadcastToRoom(session.Player.RoomNumber, session.Player.FirstName, result.RoomBroadcast)
				}
			}
			// Send second-person message to emote target
			if result.TargetName != "" && len(result.TargetMsg) > 0 {
				s.sendToPlayer(result.TargetName, result.TargetMsg)
			}
			// Broadcast departure to old room (movement)
			if result.OldRoom > 0 && len(result.OldRoomMsg) > 0 {
				s.broadcastToRoom(result.OldRoom, session.Player.FirstName, result.OldRoomMsg)
				// Update cross-machine presence with new room
				s.hub.UpdatePlayerRoom(session.Player.FirstName, session.Player.RoomNumber)
			}
			// Whisper to specific target
			if result.WhisperTarget != "" && result.WhisperMsg != "" {
				s.sendToPlayer(result.WhisperTarget, []string{result.WhisperMsg})
			}
			// Global broadcast (e.g. quit)
			if len(result.GlobalBroadcast) > 0 {
				s.broadcastGlobal(session.Player.FirstName, result.GlobalBroadcast)
			}
			// GM broadcast
			if len(result.GMBroadcast) > 0 {
				s.broadcastToGMs(result.GMBroadcast)
			}
			// Telepathy broadcast — send to all players with TelepathyActive
			if result.TelepathyMsg != "" {
				telepathyLines := []string{
					fmt.Sprintf("You feel the touch of %s's mind:", result.TelepathySender),
					fmt.Sprintf("\"%s\"", result.TelepathyMsg),
				}
				s.mu.RLock()
				for _, sess := range s.sessions {
					if sess.Player == nil || sess.Player.FirstName == result.TelepathySender {
						continue
					}
					if sess.Player.TelepathyActive {
						s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": telepathyLines})
					}
				}
				s.mu.RUnlock()
			}
			_ = playerRoom
		}
	}

	// Cleanup — global departure broadcast
	if session.Player != nil {
		s.gamelog.Log(gamelog.EventGameExit, session.Player.FullName(), session.Player.AccountID,
			fmt.Sprintf("%s (%s)", authName, authEmail), session.Player.RoomNumber, "")
		s.broadcastGlobal(session.Player.FirstName,
			[]string{fmt.Sprintf("** %s has just left the Realms.", session.Player.FirstName)})
		if session.CaptureID != "" {
			s.captures.Stop(context.Background(), session.CaptureID)
			session.CaptureID = ""
		}
		s.hub.UnregisterPlayer(session.Player.FirstName)
		s.mu.Lock()
		delete(s.sessions, session.Player.FirstName)
		s.mu.Unlock()
	}
	if accountID != "" {
		s.gamelog.Log(gamelog.EventLogout, authName, accountID, authEmail, 0, "")
	}
}

func (s *Server) sendWSResult(session *Session, result *engine.CommandResult) {
	session.mu.Lock()
	defer session.mu.Unlock()
	data, _ := json.Marshal(result)
	msg := WSMessage{Type: "result", Data: data}
	session.Conn.WriteJSON(msg)
}

func (s *Server) sendWSMessage(session *Session, msgType string, payload interface{}) {
	session.mu.Lock()
	defer session.mu.Unlock()
	data, _ := json.Marshal(payload)
	msg := WSMessage{Type: msgType, Data: data}
	session.Conn.WriteJSON(msg)

	// Capture broadcast messages
	if session.CaptureID != "" && msgType == "broadcast" {
		if m, ok := payload.(map[string]interface{}); ok {
			if msgs, ok := m["messages"].([]string); ok {
				var lines []capture.Line
				for _, text := range msgs {
					lines = append(lines, capture.Line{Time: time.Now(), Type: "broadcast", Text: text})
				}
				go s.captures.AppendLines(context.Background(), session.CaptureID, lines)
			}
		}
	}
}

// broadcastToRoom sends messages to all players in a room except the excluded player.
func (s *Server) broadcastToRoom(roomNumber int, excludeName string, messages []string, excludeNames ...string) {
	allExcludes := append([]string{excludeName}, excludeNames...)

	// Deliver to local connections
	s.mu.RLock()
	for _, sess := range s.sessions {
		if sess.Player == nil || sess.Player.RoomNumber != roomNumber {
			continue
		}
		if isExcluded(sess.Player.FirstName, allExcludes) {
			continue
		}
		s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": messages})
	}
	s.mu.RUnlock()

	// Publish to hub for other machines
	s.hub.Publish(&hub.Event{
		Type:           "room_broadcast",
		RoomNumber:     roomNumber,
		ExcludePlayers: allExcludes,
		Messages:       messages,
	})
}

func (s *Server) sendToPlayer(firstName string, messages []string) {
	// Try local first
	s.mu.RLock()
	if sess, ok := s.sessions[firstName]; ok {
		s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": messages})
	}
	s.mu.RUnlock()

	// Publish to hub for other machines
	s.hub.Publish(&hub.Event{
		Type:         "send_to_player",
		TargetPlayer: firstName,
		Messages:     messages,
	})
}

func (s *Server) broadcastGlobal(excludeName string, messages []string) {
	s.mu.RLock()
	for _, sess := range s.sessions {
		if sess.Player == nil || sess.Player.FirstName == excludeName {
			continue
		}
		s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": messages})
	}
	s.mu.RUnlock()

	s.hub.Publish(&hub.Event{
		Type:           "global_broadcast",
		ExcludePlayers: []string{excludeName},
		Messages:       messages,
	})
}

func (s *Server) broadcastToGMs(messages []string) {
	s.mu.RLock()
	for _, sess := range s.sessions {
		if sess.Player == nil || !sess.Player.IsGM {
			continue
		}
		s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": messages})
	}
	s.mu.RUnlock()

	s.hub.Publish(&hub.Event{
		Type:     "gm_broadcast",
		Messages: messages,
	})
}

// deliverRemoteEvent handles events from other machines, delivering to local connections.
func (s *Server) deliverRemoteEvent(event *hub.Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch event.Type {
	case "room_broadcast":
		for _, sess := range s.sessions {
			if sess.Player == nil || sess.Player.RoomNumber != event.RoomNumber {
				continue
			}
			if isExcluded(sess.Player.FirstName, event.ExcludePlayers) {
				continue
			}
			s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": event.Messages})
		}
	case "global_broadcast":
		for _, sess := range s.sessions {
			if sess.Player == nil || isExcluded(sess.Player.FirstName, event.ExcludePlayers) {
				continue
			}
			s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": event.Messages})
		}
	case "send_to_player":
		if sess, ok := s.sessions[event.TargetPlayer]; ok {
			s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": event.Messages})
		}
	case "gm_broadcast":
		for _, sess := range s.sessions {
			if sess.Player == nil || !sess.Player.IsGM {
				continue
			}
			s.sendWSMessage(sess, "broadcast", map[string]interface{}{"messages": event.Messages})
		}
	case "room_state_change":
		// Apply remote room state change to local game engine
		if event.Data != nil {
			change := engine.RoomChange{}
			if v, ok := event.Data["roomNumber"].(int32); ok {
				change.RoomNumber = int(v)
			} else if v, ok := event.Data["roomNumber"].(int64); ok {
				change.RoomNumber = int(v)
			}
			if v, ok := event.Data["type"].(string); ok {
				change.Type = v
			}
			if v, ok := event.Data["itemRef"].(int32); ok {
				change.ItemRef = int(v)
			} else if v, ok := event.Data["itemRef"].(int64); ok {
				change.ItemRef = int(v)
			}
			if v, ok := event.Data["newState"].(string); ok {
				change.NewState = v
			}
			if v, ok := event.Data["item"]; ok && v != nil {
				// Decode RoomItem from BSON map
				if itemMap, ok := v.(bson.M); ok {
					data, _ := bson.Marshal(itemMap)
					var ri gameworld.RoomItem
					bson.Unmarshal(data, &ri)
					change.Item = &ri
				}
			}
			s.engine.ApplyRoomChange(change)
		}
	}
}

func isExcluded(name string, excludes []string) bool {
	for _, e := range excludes {
		if name == e {
			return true
		}
	}
	return false
}

// REST handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.RLock()
	sessions := len(s.sessions)
	s.mu.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"sessions": sessions,
		"rooms":    len(s.parsed.Rooms),
		"items":    len(s.parsed.Items),
		"monsters": len(s.parsed.Monsters),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms":    len(s.parsed.Rooms),
		"items":    len(s.parsed.Items),
		"monsters": len(s.parsed.Monsters),
		"nouns":    len(s.parsed.Nouns),
		"adjs":     len(s.parsed.Adjectives),
		"sessions": len(s.sessions),
	})
}

func (s *Server) handleListRooms(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	// Return summary list
	type roomSummary struct {
		Number  int    `json:"number"`
		Name    string `json:"name"`
		Terrain string `json:"terrain"`
		Exits   int    `json:"exits"`
		File    string `json:"file"`
	}
	var rooms []roomSummary
	q := strings.ToLower(r.URL.Query().Get("q"))
	for _, room := range s.parsed.Rooms {
		if q != "" && !strings.Contains(strings.ToLower(room.Name), q) {
			continue
		}
		rooms = append(rooms, roomSummary{
			Number: room.Number, Name: room.Name,
			Terrain: room.Terrain, Exits: len(room.Exits), File: room.SourceFile,
		})
	}
	json.NewEncoder(w).Encode(rooms)
}

func (s *Server) resolveItemDef(archetype int) map[string]interface{} {
	for _, item := range s.parsed.Items {
		if item.Number == archetype {
			return map[string]interface{}{
				"name":       s.resolveNoun(item.NameID),
				"type":       item.Type,
				"weight":     item.Weight,
				"volume":     item.Volume,
				"substance":  item.Substance,
				"article":    item.Article,
				"wornSlot":   item.WornSlot,
				"container":  item.Container,
				"flags":      item.Flags,
				"sourceFile": item.SourceFile,
			}
		}
	}
	return nil
}

func (s *Server) resolveRoomName(num int) string {
	for _, room := range s.parsed.Rooms {
		if room.Number == num {
			return room.Name
		}
	}
	return ""
}

func (s *Server) handleGetRoom(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	var num int
	fmt.Sscanf(vars["number"], "%d", &num)
	for _, room := range s.parsed.Rooms {
		if room.Number == num {
			// Enrich items with resolved names and full archetype data
			var items []map[string]interface{}
			for _, ri := range room.Items {
				ei := map[string]interface{}{
					"ref":       ri.Ref,
					"archetype": ri.Archetype,
					"itemDef":   s.resolveItemDef(ri.Archetype),
				}
				if ri.Adj1 != 0 {
					ei["adj1"] = ri.Adj1
					ei["adj1Name"] = s.resolveAdj(ri.Adj1)
				}
				if ri.Adj2 != 0 {
					ei["adj2"] = ri.Adj2
					ei["adj2Name"] = s.resolveAdj(ri.Adj2)
				}
				if ri.Adj3 != 0 {
					ei["adj3"] = ri.Adj3
					ei["adj3Name"] = s.resolveAdj(ri.Adj3)
				}
				if ri.Val1 != 0 {
					ei["val1"] = ri.Val1
				}
				if ri.Val2 != 0 {
					ei["val2"] = ri.Val2
				}
				if ri.Val3 != 0 {
					ei["val3"] = ri.Val3
				}
				if ri.Val4 != 0 {
					ei["val4"] = ri.Val4
				}
				if ri.Val5 != 0 {
					ei["val5"] = ri.Val5
				}
				if ri.State != "" {
					ei["state"] = ri.State
				}
				if ri.Extend != "" {
					ei["extend"] = ri.Extend
				}
				if ri.IsPut {
					ei["isPut"] = true
					ei["putIn"] = ri.PutIn
				}
				items = append(items, ei)
			}

			// Enrich exits with room names
			type enrichedExit struct {
				Room     int    `json:"room"`
				RoomName string `json:"roomName"`
			}
			exits := make(map[string]enrichedExit)
			for dir, destNum := range room.Exits {
				exits[dir] = enrichedExit{
					Room:     destNum,
					RoomName: s.resolveRoomName(destNum),
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"number":       room.Number,
				"name":         room.Name,
				"description":  room.Description,
				"exits":        exits,
				"items":        items,
				"terrain":      room.Terrain,
				"lighting":     room.Lighting,
				"monsterGroup": room.MonsterGroup,
				"modifiers":    room.Modifiers,
				"scripts":      room.Scripts,
				"sourceFile":   room.SourceFile,
			})
			return
		}
	}
	http.Error(w, "room not found", 404)
}

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	type itemSummary struct {
		Number     int    `json:"number"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		Weight     int    `json:"weight"`
		Substance  string `json:"substance"`
		SourceFile string `json:"sourceFile"`
	}
	var items []itemSummary
	for _, item := range s.parsed.Items {
		name := fmt.Sprintf("noun#%d", item.NameID)
		for _, n := range s.parsed.Nouns {
			if n.ID == item.NameID {
				name = n.Name
				break
			}
		}
		items = append(items, itemSummary{
			Number: item.Number, Name: name, Type: item.Type,
			Weight: item.Weight, Substance: item.Substance, SourceFile: item.SourceFile,
		})
	}
	json.NewEncoder(w).Encode(items)
}

func (s *Server) resolveNoun(id int) string {
	for _, n := range s.parsed.Nouns {
		if n.ID == id {
			return n.Name
		}
	}
	return fmt.Sprintf("noun#%d", id)
}

func (s *Server) resolveAdj(id int) string {
	for _, a := range s.parsed.Adjectives {
		if a.ID == id {
			return a.Name
		}
	}
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("adj#%d", id)
}

func (s *Server) handleGetItem(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	var num int
	fmt.Sscanf(vars["number"], "%d", &num)
	for _, item := range s.parsed.Items {
		if item.Number == num {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"number":       item.Number,
				"nameId":       item.NameID,
				"resolvedName": s.resolveNoun(item.NameID),
				"type":         item.Type,
				"weight":       item.Weight,
				"volume":       item.Volume,
				"substance":    item.Substance,
				"article":      item.Article,
				"parameter1":   item.Parameter1,
				"parameter2":   item.Parameter2,
				"parameter3":   item.Parameter3,
				"container":    item.Container,
				"interior":     item.Interior,
				"wornSlot":     item.WornSlot,
				"flags":        item.Flags,
				"scripts":      item.Scripts,
				"sourceFile":   item.SourceFile,
			})
			return
		}
	}
	http.Error(w, "item not found", 404)
}

func (s *Server) handleListMonsters(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	json.NewEncoder(w).Encode(s.parsed.Monsters)
}

func (s *Server) handleGetMonster(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	var num int
	fmt.Sscanf(vars["number"], "%d", &num)
	for _, mon := range s.parsed.Monsters {
		if mon.Number == num {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"number":      mon.Number,
				"name":        mon.Name,
				"adjective":   mon.Adjective,
				"adjName":     s.resolveAdj(mon.Adjective),
				"description": mon.Description,
				"bodyType":    mon.BodyType,
				"body":        mon.Body,
				"attack1":     mon.Attack1,
				"attack2":     mon.Attack2,
				"defense":     mon.Defense,
				"strategy":    mon.Strategy,
				"treasure":    mon.Treasure,
				"speed":       mon.Speed,
				"armor":       mon.Armor,
				"race":        mon.Race,
				"gender":      mon.Gender,
				"unique":      mon.Unique,
				"scripts":     mon.Scripts,
				"sourceFile":  mon.SourceFile,
			})
			return
		}
	}
	http.Error(w, "monster not found", 404)
}

func (s *Server) handleListNouns(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	json.NewEncoder(w).Encode(s.parsed.Nouns)
}

func (s *Server) handleListAdjectives(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	json.NewEncoder(w).Encode(s.parsed.Adjectives)
}

func (s *Server) handleListCharacters(w http.ResponseWriter, r *http.Request) {
	accountID := s.getAccountID(r)
	var players interface{}
	var err error
	if accountID != "" {
		players, err = s.engine.ListPlayersByAccount(r.Context(), accountID)
	} else {
		players, err = s.engine.ListPlayers(r.Context())
	}
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func (s *Server) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	var req CreateCharMsg
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	// Trim whitespace from names
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	// Validate
	if err := engine.ValidateCharacterInput(req.FirstName, req.LastName, req.Race, req.Gender); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	accountID := s.getAccountID(r)
	player := s.engine.CreateNewPlayer(r.Context(), req.FirstName, req.LastName, req.Race, req.Gender, accountID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *Server) handleGetCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	player, err := s.engine.GetPlayer(r.Context(), vars["firstName"])
	if err != nil {
		http.Error(w, "player not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *Server) handleToggleGM(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	var req struct {
		IsGM bool `json:"isGM"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	player, err := s.engine.SetPlayerGM(r.Context(), vars["firstName"], req.IsGM)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	if req.IsGM {
		s.gamelog.Log(gamelog.EventGMGrant, player.FullName(), player.AccountID, "", 0, "")
	} else {
		s.gamelog.Log(gamelog.EventGMRevoke, player.FullName(), player.AccountID, "", 0, "")
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

// getAccountID extracts the account ID from the Authorization header JWT.
func (s *Server) getAccountID(r *http.Request) string {
	if s.auth == nil {
		return ""
	}
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	accountID, err := s.auth.ValidateJWT(token)
	if err != nil {
		return ""
	}
	return accountID
}

func (s *Server) handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		http.Error(w, "auth not configured", 500)
		return
	}
	var req struct {
		Credential string `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}

	claims, err := s.auth.VerifyGoogleToken(req.Credential)
	if err != nil {
		log.Printf("Google token verification failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	account, err := s.auth.FindOrCreateAccount(r.Context(), claims)
	if err != nil {
		log.Printf("Account creation failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	jwt, err := s.auth.IssueJWT(account)
	if err != nil {
		http.Error(w, "token error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":   jwt,
		"account": account,
	})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	accountID := s.getAccountID(r)
	if accountID == "" {
		http.Error(w, "unauthorized", 401)
		return
	}
	account, err := s.auth.GetAccount(r.Context(), accountID)
	if err != nil {
		http.Error(w, "account not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// requireAdmin checks if the request comes from an admin account. Returns the account or writes an error.
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) *auth.Account {
	accountID := s.getAccountID(r)
	if accountID == "" {
		http.Error(w, "unauthorized", 401)
		return nil
	}
	account, err := s.auth.GetAccount(r.Context(), accountID)
	if err != nil || !account.IsAdmin {
		http.Error(w, "forbidden", 403)
		return nil
	}
	return account
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	accounts, err := s.auth.ListAccounts(r.Context())
	if err != nil {
		http.Error(w, "failed to list accounts", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	account, err := s.auth.GetAccount(r.Context(), vars["id"])
	if err != nil {
		http.Error(w, "account not found", 404)
		return
	}
	// Also fetch this account's characters
	players, _ := s.engine.ListPlayersByAccount(r.Context(), vars["id"])
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account":    account,
		"characters": players,
	})
}

func (s *Server) handleToggleAdmin(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	var req struct {
		IsAdmin bool `json:"isAdmin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	account, err := s.auth.SetAdmin(r.Context(), vars["id"], req.IsAdmin)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (s *Server) handleAdminListCharacters(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	players, err := s.engine.ListPlayers(r.Context())
	if err != nil {
		http.Error(w, "failed to list players", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func (s *Server) handleAdminGetCharacter(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	player, err := s.engine.GetPlayer(r.Context(), vars["firstName"])
	if err != nil {
		http.Error(w, "player not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *Server) handleAdminReassignCharacter(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	vars := mux.Vars(r)
	var req struct {
		AccountID string `json:"accountId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	if req.AccountID == "" {
		http.Error(w, "accountId is required", 400)
		return
	}
	player, err := s.engine.ReassignCharacter(r.Context(), vars["firstName"], req.AccountID)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *Server) handleListCaptures(w http.ResponseWriter, r *http.Request) {
	accountID := s.getAccountID(r)
	if accountID == "" {
		http.Error(w, "unauthorized", 401)
		return
	}
	sessions, err := s.captures.List(r.Context(), accountID)
	if err != nil {
		http.Error(w, "failed to list captures", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (s *Server) handleGetCapture(w http.ResponseWriter, r *http.Request) {
	accountID := s.getAccountID(r)
	if accountID == "" {
		http.Error(w, "unauthorized", 401)
		return
	}
	vars := mux.Vars(r)
	sess, err := s.captures.Get(r.Context(), vars["id"])
	if err != nil || sess == nil {
		http.Error(w, "capture not found", 404)
		return
	}
	if sess.AccountID != accountID {
		http.Error(w, "forbidden", 403)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

func (s *Server) handleGetCaptureText(w http.ResponseWriter, r *http.Request) {
	accountID := s.getAccountID(r)
	// Also accept token as query param for direct download links
	if accountID == "" {
		if token := r.URL.Query().Get("token"); token != "" && s.auth != nil {
			accountID, _ = s.auth.ValidateJWT(token)
		}
	}
	if accountID == "" {
		http.Error(w, "unauthorized", 401)
		return
	}
	vars := mux.Vars(r)
	sess, err := s.captures.Get(r.Context(), vars["id"])
	if err != nil || sess == nil {
		http.Error(w, "capture not found", 404)
		return
	}
	if sess.AccountID != accountID {
		http.Error(w, "forbidden", 403)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"capture-%s.txt\"", vars["id"]))
	w.Write([]byte(sess.ExportText()))
}

func (s *Server) handleDeleteCapture(w http.ResponseWriter, r *http.Request) {
	accountID := s.getAccountID(r)
	if accountID == "" {
		http.Error(w, "unauthorized", 401)
		return
	}
	vars := mux.Vars(r)
	sess, err := s.captures.Get(r.Context(), vars["id"])
	if err != nil || sess == nil {
		http.Error(w, "capture not found", 404)
		return
	}
	if sess.AccountID != accountID {
		http.Error(w, "forbidden", 403)
		return
	}
	if err := s.captures.Delete(r.Context(), vars["id"]); err != nil {
		http.Error(w, "failed to delete capture", 500)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) handleAdminLogs(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	event := r.URL.Query().Get("event")
	player := r.URL.Query().Get("player")
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	entries, err := s.gamelog.Query(r.Context(), event, player, limit)
	if err != nil {
		http.Error(w, "failed to query logs", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (s *Server) handleBootstrapAdmin(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		http.Error(w, "auth not configured", 500)
		return
	}
	// Only works if no admins exist
	hasAdmin, err := s.auth.HasAnyAdmin(r.Context())
	if err != nil {
		http.Error(w, "check failed", 500)
		return
	}
	if hasAdmin {
		http.Error(w, "admin already exists", 403)
		return
	}
	// Promote the requesting user
	accountID := s.getAccountID(r)
	if accountID == "" {
		http.Error(w, "unauthorized - log in first", 401)
		return
	}
	account, err := s.auth.SetAdmin(r.Context(), accountID, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("Bootstrap: %s (%s) promoted to admin", account.Name, account.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}
